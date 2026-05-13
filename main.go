package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"mrk2pdf/internal/converter"
)

const previewPath = "preview.html"

func main() {
	cfg, err := parseFlags()
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	fmt.Printf("Archivo de salida configurado en: %s\n", cfg.Output)

	if err := validateInput(cfg); err != nil {
		log.Fatalf("error: %v", err)
	}
	if err := ensureTemplates(cfg.ForceDefault); err != nil {
		log.Fatalf("error preparando plantillas: %v", err)
	}
	if err := validateTemplate(cfg); err != nil {
		log.Fatalf("error: %v", err)
	}

	fmt.Printf("Usando template: %s\n", cfg.Template)

	html, err := converter.Render(converter.Options{
		InputPath:   cfg.Input,
		TemplateDir: filepath.Join(templateRoot, cfg.Template),
	})
	if err != nil {
		log.Fatalf("error generando HTML: %v", err)
	}

	if err := os.WriteFile(previewPath, html, 0o644); err != nil {
		log.Fatalf("error escribiendo %s: %v", previewPath, err)
	}
	fmt.Printf("Previa HTML generada: %s\n", previewPath)
	fmt.Println("(La conversión a PDF se implementará en el siguiente paso.)")
}
