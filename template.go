package main

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
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
