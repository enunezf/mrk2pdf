package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/fsnotify/fsnotify"

	"mrk2pdf/internal/converter"
	"mrk2pdf/internal/pdf"
)

const watchDebounce = 250 * time.Millisecond

// watchScope distinguishes between files the user listed explicitly (-i foo.md,
// glob expansions, positional args) and directories opened wholesale (-i docs/).
// Events on explicit files only trigger when their path matches exactly;
// events on directories trigger for any .md inside them.
type watchScope struct {
	explicitFiles map[string]bool
	openDirs      map[string]bool
}

func buildWatchScope(iFlag string, positional []string) (*watchScope, error) {
	scope := &watchScope{
		explicitFiles: map[string]bool{},
		openDirs:      map[string]bool{},
	}
	sources := make([]string, 0, 1+len(positional))
	if iFlag != "" {
		sources = append(sources, iFlag)
	}
	sources = append(sources, positional...)

	for _, s := range sources {
		if hasGlobChars(s) {
			matches, err := filepath.Glob(s)
			if err != nil {
				return nil, fmt.Errorf("glob %q: %w", s, err)
			}
			for _, m := range matches {
				abs, _ := filepath.Abs(m)
				scope.explicitFiles[abs] = true
			}
			continue
		}
		abs, _ := filepath.Abs(s)
		info, err := os.Stat(s)
		if err != nil {
			scope.explicitFiles[abs] = true
			continue
		}
		if info.IsDir() {
			scope.openDirs[abs] = true
		} else {
			scope.explicitFiles[abs] = true
		}
	}
	return scope, nil
}

// dirsToWatch returns the fsnotify watch targets: parent dirs of every
// explicit file plus every opened directory. Deduplicated.
func (s *watchScope) dirsToWatch() []string {
	set := map[string]bool{}
	for f := range s.explicitFiles {
		set[filepath.Dir(f)] = true
	}
	for d := range s.openDirs {
		set[d] = true
	}
	out := make([]string, 0, len(set))
	for d := range set {
		out = append(out, d)
	}
	return out
}

// matches reports whether the given absolute path should trigger a render.
func (s *watchScope) matches(abs string) bool {
	if s.explicitFiles[abs] {
		return true
	}
	return s.openDirs[filepath.Dir(abs)]
}

// debouncer coalesces repeated events on the same key, calling fn only after
// `delay` has elapsed without further activity on that key.
type debouncer struct {
	mu     sync.Mutex
	timers map[string]*time.Timer
}

func newDebouncer() *debouncer {
	return &debouncer{timers: map[string]*time.Timer{}}
}

func (d *debouncer) Do(key string, delay time.Duration, fn func()) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if t, ok := d.timers[key]; ok {
		t.Stop()
	}
	d.timers[key] = time.AfterFunc(delay, fn)
}

// runWatch executes the initial render of every input, then blocks watching
// the relevant directories until the user interrupts with Ctrl+C. Render
// failures are reported on stdout but do not terminate the process.
func runWatch(cfg *Config, inputs []string, renderer *pdf.Renderer) error {
	scope, err := buildWatchScope(cfg.Input, flag.Args())
	if err != nil {
		return err
	}
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return fmt.Errorf("creando watcher: %w", err)
	}
	defer watcher.Close()

	dirs := scope.dirsToWatch()
	for _, d := range dirs {
		if err := watcher.Add(d); err != nil {
			return fmt.Errorf("watching %s: %w", d, err)
		}
	}

	fmt.Println("Modo watch activo. Render inicial:")
	for _, in := range inputs {
		renderOne(in, cfg, renderer)
	}
	fmt.Printf("\nVigilando %d directorio(s): %s\n", len(dirs), strings.Join(dirs, ", "))
	fmt.Println("Ctrl+C para salir.")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigs
		fmt.Println("\nDeteniendo watch.")
		cancel()
	}()

	deb := newDebouncer()

	for {
		select {
		case <-ctx.Done():
			return nil

		case err, ok := <-watcher.Errors:
			if !ok {
				return nil
			}
			fmt.Fprintf(os.Stderr, "[watcher] %v\n", err)

		case ev, ok := <-watcher.Events:
			if !ok {
				return nil
			}
			// Only react to writes and creates. Rename-target arrives as Create.
			if ev.Op&(fsnotify.Write|fsnotify.Create) == 0 {
				continue
			}
			if !strings.EqualFold(filepath.Ext(ev.Name), ".md") {
				continue
			}
			abs, err := filepath.Abs(ev.Name)
			if err != nil {
				continue
			}
			if !scope.matches(abs) {
				continue
			}
			path := ev.Name
			deb.Do(abs, watchDebounce, func() {
				renderOne(path, cfg, renderer)
			})
		}
	}
}

// renderOne converts and prints one document. Errors are surfaced on stdout
// but never propagated — the watch loop must survive bad input.
func renderOne(input string, cfg *Config, renderer *pdf.Renderer) {
	ts := time.Now().Format("15:04:05")
	output := outputFor(input, cfg.Output)

	html, err := converter.Render(converter.Options{
		InputPath:   input,
		TemplateDir: filepath.Join(templateRoot, cfg.Template),
		AutoTOC:     cfg.AutoTOC,
	})
	if err != nil {
		fmt.Printf("[%s] %s → ERROR (convirtiendo): %v\n", ts, input, err)
		return
	}
	if err := renderer.Render(pdf.RenderOpts{
		HTML:       html,
		OutputPath: output,
		PageSize:   cfg.PageSize,
		Landscape:  cfg.Landscape,
	}); err != nil {
		fmt.Printf("[%s] %s → ERROR (PDF): %v\n", ts, input, err)
		return
	}
	fmt.Printf("[%s] %s → %s (%s)\n", ts, input, output, humanSize(output))
}

// outputFor derives a single output path from an input and the -o flag.
// Mirrors resolveOutputs but for one file at a time (used per-event in watch
// mode, where new files may appear inside a watched directory).
func outputFor(input, oFlag string) string {
	if oFlag == "" {
		return strings.TrimSuffix(input, filepath.Ext(input)) + ".pdf"
	}
	if looksLikeDir(oFlag) {
		dir := strings.TrimRight(oFlag, "/"+string(os.PathSeparator))
		return filepath.Join(dir, strings.TrimSuffix(filepath.Base(input), filepath.Ext(input))+".pdf")
	}
	return oFlag
}
