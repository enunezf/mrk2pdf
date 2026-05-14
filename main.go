package main

import (
	"flag"
	"fmt"
	"log"
	"os"
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

	inputs, err := collectInputs(cfg.Input, flag.Args(), cfg.Recursive)
	if err != nil {
		log.Fatalf("error resolviendo entradas: %v", err)
	}
	if len(inputs) == 0 {
		log.Fatalf("error: no se encontraron archivos .md para procesar")
	}

	for _, in := range inputs {
		if err := validateInputFile(in); err != nil {
			log.Fatalf("error: %v", err)
		}
	}

	outputs, err := resolveOutputs(inputs, cfg.Output)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	if err := ensureTemplates(cfg.ForceDefault); err != nil {
		log.Fatalf("error preparando plantillas: %v", err)
	}
	if err := validateTemplate(cfg); err != nil {
		log.Fatalf("error: %v", err)
	}

	browserPath, err := pdf.FindBrowser()
	if err != nil {
		log.Fatalf("%v", err)
	}

	fmt.Printf("Usando template: %s\n", cfg.Template)
	fmt.Printf("Usando navegador: %s\n", browserPath)

	renderer, err := pdf.NewRenderer(browserPath)
	if err != nil {
		log.Fatalf("error iniciando navegador: %v", err)
	}
	defer renderer.Close()

	orientation := "vertical"
	if cfg.Landscape {
		orientation = "apaisada"
	}
	if len(inputs) == 1 {
		fmt.Printf("Generando PDF (%s, %s)...\n", cfg.PageSize.Name, orientation)
	} else {
		fmt.Printf("Procesando %d archivos en %s %s...\n\n", len(inputs), cfg.PageSize.Name, orientation)
	}

	type failure struct {
		input string
		err   error
	}
	var failures []failure

	for i, in := range inputs {
		prefix := ""
		if len(inputs) > 1 {
			prefix = fmt.Sprintf("[%d/%d] ", i+1, len(inputs))
		}
		out := outputs[i]
		if err := processOne(in, out, cfg, renderer); err != nil {
			fmt.Printf("%s%s → ERROR: %v\n", prefix, in, err)
			failures = append(failures, failure{input: in, err: err})
			continue
		}
		fmt.Printf("%s%s → %s (%s)\n", prefix, in, out, humanSize(out))
	}

	if len(failures) > 0 {
		fmt.Fprintf(os.Stderr, "\n%d/%d documentos fallaron:\n", len(failures), len(inputs))
		for _, f := range failures {
			fmt.Fprintf(os.Stderr, "  - %s: %v\n", f.input, f.err)
		}
		os.Exit(1)
	}
	if len(inputs) > 1 {
		fmt.Printf("\n%d documentos procesados correctamente.\n", len(inputs))
	}
}

func processOne(input, output string, cfg *Config, r *pdf.Renderer) error {
	html, err := converter.Render(converter.Options{
		InputPath:   input,
		TemplateDir: filepath.Join(templateRoot, cfg.Template),
		AutoTOC:     cfg.AutoTOC,
	})
	if err != nil {
		return fmt.Errorf("convirtiendo: %w", err)
	}
	return r.Render(pdf.RenderOpts{
		HTML:       html,
		OutputPath: output,
		PageSize:   cfg.PageSize,
		Landscape:  cfg.Landscape,
	})
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

func humanSize(path string) string {
	info, err := os.Stat(path)
	if err != nil {
		return "?"
	}
	n := info.Size()
	switch {
	case n >= 1<<20:
		return fmt.Sprintf("%.1f MB", float64(n)/(1<<20))
	case n >= 1<<10:
		return fmt.Sprintf("%d KB", n/(1<<10))
	default:
		return fmt.Sprintf("%d B", n)
	}
}
