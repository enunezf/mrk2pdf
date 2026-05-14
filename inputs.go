package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// collectInputs expands every source (from the -i flag plus positional args)
// into a flat, sorted, de-duplicated list of file paths. Each source can be a
// file, a directory, or a glob pattern.
func collectInputs(iFlag string, positional []string, recursive bool) ([]string, error) {
	sources := make([]string, 0, len(positional)+1)
	if iFlag != "" {
		sources = append(sources, iFlag)
	}
	sources = append(sources, positional...)

	seen := map[string]struct{}{}
	var out []string
	for _, s := range sources {
		expanded, err := expandInput(s, recursive)
		if err != nil {
			return nil, err
		}
		for _, p := range expanded {
			abs, err := filepath.Abs(p)
			if err != nil {
				abs = p
			}
			if _, dup := seen[abs]; dup {
				continue
			}
			seen[abs] = struct{}{}
			out = append(out, p)
		}
	}
	sort.Strings(out)
	return out, nil
}

func expandInput(s string, recursive bool) ([]string, error) {
	if hasGlobChars(s) {
		matches, err := filepath.Glob(s)
		if err != nil {
			return nil, fmt.Errorf("glob %q: %w", s, err)
		}
		if len(matches) == 0 {
			return nil, fmt.Errorf("el patrón %q no coincide con ningún archivo", s)
		}
		return matches, nil
	}
	info, err := os.Stat(s)
	if err != nil {
		// Pass through so the per-file validator can produce a clear error.
		return []string{s}, nil
	}
	if !info.IsDir() {
		return []string{s}, nil
	}
	return scanDir(s, recursive)
}

func hasGlobChars(s string) bool {
	return strings.ContainsAny(s, "*?[")
}

func scanDir(dir string, recursive bool) ([]string, error) {
	var files []string
	if recursive {
		err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() && strings.EqualFold(filepath.Ext(path), ".md") {
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
		return files, nil
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		if !e.IsDir() && strings.EqualFold(filepath.Ext(e.Name()), ".md") {
			files = append(files, filepath.Join(dir, e.Name()))
		}
	}
	return files, nil
}

// resolveOutputs maps each input to its destination PDF path. The -o flag
// can be empty (PDF next to each MD), a single file path (only valid with
// exactly one input) or a directory (forced via trailing slash or existing
// directory; created if absent).
func resolveOutputs(inputs []string, oFlag string) ([]string, error) {
	if oFlag == "" {
		return derivePerInput(inputs), nil
	}

	if looksLikeDir(oFlag) {
		dir := strings.TrimRight(oFlag, "/"+string(os.PathSeparator))
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("creando directorio de salida %s: %w", dir, err)
		}
		outs := make([]string, len(inputs))
		for i, in := range inputs {
			name := strings.TrimSuffix(filepath.Base(in), filepath.Ext(in)) + ".pdf"
			outs[i] = filepath.Join(dir, name)
		}
		return outs, nil
	}

	if len(inputs) != 1 {
		return nil, fmt.Errorf(
			"-o '%s' apunta a un archivo, pero hay %d entradas; usá un directorio (terminá con /)",
			oFlag, len(inputs),
		)
	}
	return []string{oFlag}, nil
}

func derivePerInput(inputs []string) []string {
	outs := make([]string, len(inputs))
	for i, in := range inputs {
		outs[i] = strings.TrimSuffix(in, filepath.Ext(in)) + ".pdf"
	}
	return outs
}

func looksLikeDir(p string) bool {
	if strings.HasSuffix(p, "/") || strings.HasSuffix(p, string(os.PathSeparator)) {
		return true
	}
	info, err := os.Stat(p)
	return err == nil && info.IsDir()
}
