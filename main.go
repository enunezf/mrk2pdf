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

	if cfg.ListTemplates {
		runListTemplates(cfg.ForceDefault)
		return
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
		AutoTOC:     cfg.AutoTOC,
	})
	if err != nil {
		log.Fatalf("error generando HTML: %v", err)
	}

	orientation := "vertical"
	if cfg.Landscape {
		orientation = "apaisada"
	}
	fmt.Printf("Generando PDF (%s, %s)...\n", cfg.PageSize.Name, orientation)

	if err := pdf.GeneratePDF(pdf.Options{
		HTML:        html,
		OutputPath:  cfg.Output,
		BrowserPath: browserPath,
		PageSize:    cfg.PageSize,
		Landscape:   cfg.Landscape,
	}); err != nil {
		log.Fatalf("error generando PDF: %v", err)
	}
	fmt.Printf("PDF generado: %s\n", cfg.Output)
}

func runListTemplates(force bool) {
	if err := ensureTemplates(force); err != nil {
		log.Fatalf("error preparando plantillas: %v", err)
	}
	names, err := listTemplateNames()
	if err != nil {
		log.Fatalf("error listando plantillas: %v", err)
	}
	if len(names) == 0 {
		fmt.Println("No hay plantillas disponibles.")
		return
	}
	fmt.Println("Plantillas disponibles:")
	for _, n := range names {
		fmt.Printf("  - %s\n", n)
	}
}
