package main

import (
	"fmt"
	"log"
	"path/filepath"

	"mrk2pdf/internal/converter"
	"mrk2pdf/internal/pdf"
)

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

	browserPath, err := pdf.FindBrowser()
	if err != nil {
		log.Fatalf("%v", err)
	}
	fmt.Printf("Usando navegador: %s\n", browserPath)

	html, err := converter.Render(converter.Options{
		InputPath:   cfg.Input,
		TemplateDir: filepath.Join(templateRoot, cfg.Template),
	})
	if err != nil {
		log.Fatalf("error generando HTML: %v", err)
	}

	fmt.Println("Generando PDF...")
	if err := pdf.GeneratePDF(html, cfg.Output, browserPath); err != nil {
		log.Fatalf("error generando PDF: %v", err)
	}
	fmt.Printf("PDF generado: %s\n", cfg.Output)
}
