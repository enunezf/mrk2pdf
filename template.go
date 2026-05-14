package main

import (
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	templateRoot    = "template"
	defaultTemplate = "default"
)

//go:embed assets
var embeddedTemplates embed.FS

// ensureTemplates walks every template embedded under assets/ and extracts
// each one to template/<name>/ if it doesn't already exist on disk. The
// -d flag additionally re-extracts the default template, overwriting any
// local edits — other templates are never overwritten so user
// customisations are preserved across runs.
func ensureTemplates(force bool) error {
	if _, err := os.Stat(templateRoot); os.IsNotExist(err) {
		fmt.Println("Creando carpeta de plantillas...")
		if err := os.MkdirAll(templateRoot, 0o755); err != nil {
			return fmt.Errorf("creando %s: %w", templateRoot, err)
		}
	}

	entries, err := fs.ReadDir(embeddedTemplates, "assets")
	if err != nil {
		return fmt.Errorf("listando templates embebidos: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		if err := extractTemplate(entry.Name(), force); err != nil {
			return err
		}
	}
	return nil
}

func extractTemplate(name string, force bool) error {
	dstDir := filepath.Join(templateRoot, name)
	_, statErr := os.Stat(dstDir)
	missing := os.IsNotExist(statErr)

	// -d only overrides the default template; user customisations on other
	// templates are always preserved (extract only if missing).
	mustExtract := missing || (force && name == defaultTemplate)
	if !mustExtract {
		return nil
	}

	if !missing {
		fmt.Printf("Sobrescribiendo template/%s/ (flag -d)...\n", name)
	} else {
		fmt.Printf("Creando template/%s/ con archivos por defecto...\n", name)
	}

	if err := os.MkdirAll(dstDir, 0o755); err != nil {
		return fmt.Errorf("creando %s: %w", dstDir, err)
	}

	srcDir := "assets/" + name
	return fs.WalkDir(embeddedTemplates, srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		data, err := embeddedTemplates.ReadFile(path)
		if err != nil {
			return fmt.Errorf("leyendo asset embebido %s: %w", path, err)
		}
		dst := filepath.Join(dstDir, filepath.Base(path))
		if err := os.WriteFile(dst, data, 0o644); err != nil {
			return fmt.Errorf("escribiendo %s: %w", dst, err)
		}
		return nil
	})
}

// listTemplateNames returns the names of every subdirectory under
// template/, sorted alphabetically. Returns (nil, nil) if template/
// doesn't exist yet.
func listTemplateNames() ([]string, error) {
	entries, err := os.ReadDir(templateRoot)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	return names, nil
}

// suggestClosest returns the candidate most similar to name by Levenshtein
// distance, or empty string if no candidate is reasonably close.
func suggestClosest(name string, candidates []string) string {
	if len(candidates) == 0 || name == "" {
		return ""
	}
	var best string
	bestDist := len(name)*2 + 1
	for _, c := range candidates {
		d := levenshtein(strings.ToLower(name), strings.ToLower(c))
		if d < bestDist {
			best = c
			bestDist = d
		}
	}
	// Only suggest if distance is small (typo-grade), not unrelated names.
	if bestDist <= 3 || bestDist*2 <= len(name) {
		return best
	}
	return ""
}

func levenshtein(a, b string) int {
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}
	prev := make([]int, len(b)+1)
	curr := make([]int, len(b)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(a); i++ {
		curr[0] = i
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			curr[j] = min3(prev[j]+1, curr[j-1]+1, prev[j-1]+cost)
		}
		prev, curr = curr, prev
	}
	return prev[len(b)]
}

func min3(a, b, c int) int {
	m := a
	if b < m {
		m = b
	}
	if c < m {
		m = c
	}
	return m
}
