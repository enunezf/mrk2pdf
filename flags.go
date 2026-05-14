package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"mrk2pdf/internal/pdf"
)

type Config struct {
	Input         string
	Output        string
	Template      string
	ForceDefault  bool
	ListTemplates bool
	AutoTOC       bool
	PageSize      pdf.PageSize
	Landscape     bool
}

func parseFlags() (*Config, error) {
	cfg := &Config{}
	var rawSize string

	flag.StringVar(&cfg.Input, "i", "", "ruta del archivo Markdown de entrada (obligatorio salvo con -l)")
	flag.StringVar(&cfg.Output, "o", "", "ruta del archivo PDF de salida (por defecto: <input>.pdf)")
	flag.StringVar(&cfg.Template, "t", defaultTemplate, "nombre del template (carpeta dentro de template/)")
	flag.BoolVar(&cfg.ForceDefault, "d", false, "forzar sobrescritura de template/default")
	flag.BoolVar(&cfg.ListTemplates, "l", false, "listar las plantillas disponibles y salir")
	flag.BoolVar(&cfg.AutoTOC, "toc", false, "prepend tabla de contenidos al inicio del documento")
	flag.StringVar(&rawSize, "size", "A4", "tamaño de página (A4 o Letter)")
	flag.BoolVar(&cfg.Landscape, "landscape", false, "orientación apaisada")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Uso:\n")
		fmt.Fprintf(os.Stderr, "  %s -i <archivo.md> [-o <salida.pdf>] [-t <template>] [--toc] [--size A4|Letter] [--landscape]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -l\n\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	if cfg.ListTemplates {
		// -l is a standalone informational mode; skip the input/page validation.
		return cfg, nil
	}

	if cfg.Input == "" {
		flag.Usage()
		return nil, errors.New("el flag -i es obligatorio")
	}

	if cfg.Output == "" {
		ext := filepath.Ext(cfg.Input)
		cfg.Output = strings.TrimSuffix(cfg.Input, ext) + ".pdf"
	}

	parsed, err := pdf.ParsePageSize(rawSize)
	if err != nil {
		return nil, err
	}
	cfg.PageSize = parsed

	return cfg, nil
}

func validateInput(cfg *Config) error {
	info, err := os.Stat(cfg.Input)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("el archivo de entrada no existe: %s", cfg.Input)
		}
		return fmt.Errorf("no se pudo leer el archivo de entrada: %w", err)
	}
	if info.IsDir() {
		return fmt.Errorf("la ruta de entrada es un directorio, no un archivo: %s", cfg.Input)
	}
	if strings.ToLower(filepath.Ext(cfg.Input)) != ".md" {
		return fmt.Errorf("el archivo de entrada debe tener extensión .md: %s", cfg.Input)
	}
	return nil
}

// validateTemplate verifies the requested template directory exists. When it
// doesn't, the error includes a "did you mean ..." suggestion plus the list
// of templates currently available under template/.
func validateTemplate(cfg *Config) error {
	tplDir := filepath.Join(templateRoot, cfg.Template)
	info, err := os.Stat(tplDir)
	if err == nil && info.IsDir() {
		return nil
	}

	available, _ := listTemplateNames()

	var msg strings.Builder
	fmt.Fprintf(&msg, "el template '%s' no existe en %s/", cfg.Template, templateRoot)
	if suggestion := suggestClosest(cfg.Template, available); suggestion != "" {
		fmt.Fprintf(&msg, "\n¿Quizás quisiste decir '%s'?", suggestion)
	}
	if len(available) > 0 {
		fmt.Fprintf(&msg, "\nDisponibles: %s", strings.Join(available, ", "))
	}
	return errors.New(msg.String())
}
