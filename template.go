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

//go:embed assets/index.html assets/style.css
var defaultAssets embed.FS

func ensureTemplates(force bool) error {
	if _, err := os.Stat(templateRoot); os.IsNotExist(err) {
		fmt.Println("Creando carpeta de plantillas...")
		if err := os.MkdirAll(templateRoot, 0o755); err != nil {
			return fmt.Errorf("creando %s: %w", templateRoot, err)
		}
	}

	defaultDir := filepath.Join(templateRoot, defaultTemplate)
	_, statErr := os.Stat(defaultDir)
	missing := os.IsNotExist(statErr)

	if !missing && !force {
		return nil
	}

	if force && !missing {
		fmt.Printf("Sobrescribiendo template/%s/ (flag -d)...\n", defaultTemplate)
	} else {
		fmt.Printf("Creando template/%s/ con archivos por defecto...\n", defaultTemplate)
	}

	if err := os.MkdirAll(defaultDir, 0o755); err != nil {
		return fmt.Errorf("creando %s: %w", defaultDir, err)
	}

	return fs.WalkDir(defaultAssets, "assets", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		data, err := defaultAssets.ReadFile(path)
		if err != nil {
			return fmt.Errorf("leyendo asset embebido %s: %w", path, err)
		}
		dst := filepath.Join(defaultDir, filepath.Base(path))
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
