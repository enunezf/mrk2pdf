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
	Recursive     bool
	PageSize      pdf.PageSize
	Landscape     bool
}

func parseFlags() (*Config, error) {
	cfg := &Config{}
	var rawSize string

	flag.StringVar(&cfg.Input, "i", "", "archivo, directorio o glob (obligatorio salvo con -l)")
	flag.StringVar(&cfg.Output, "o", "", "archivo PDF o directorio destino (por defecto: <input>.pdf)")
	flag.StringVar(&cfg.Template, "t", defaultTemplate, "nombre del template (carpeta dentro de template/)")
	flag.BoolVar(&cfg.ForceDefault, "d", false, "forzar sobrescritura de template/default")
	flag.BoolVar(&cfg.ListTemplates, "l", false, "listar las plantillas disponibles y salir")
	flag.BoolVar(&cfg.AutoTOC, "toc", false, "prepend tabla de contenidos al inicio del documento")
	flag.BoolVar(&cfg.Recursive, "R", false, "buscar recursivamente cuando -i es un directorio")
	flag.StringVar(&rawSize, "size", "A4", "tamaño de página (A4 o Letter)")
	flag.BoolVar(&cfg.Landscape, "landscape", false, "orientación apaisada")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Uso:\n")
		fmt.Fprintf(os.Stderr, "  %s -i <archivo.md|dir|glob> [-o <salida|dir>] [-t <template>] [--toc] [--size A4|Letter] [--landscape] [-R]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s [opciones] <archivo.md>...\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -l\n\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	if cfg.ListTemplates {
		return cfg, nil
	}

	if cfg.Input == "" && len(flag.Args()) == 0 {
		flag.Usage()
		return nil, errors.New("se requiere al menos un archivo .md (vía -i o como argumento posicional)")
	}

	parsed, err := pdf.ParsePageSize(rawSize)
	if err != nil {
		return nil, err
	}
	cfg.PageSize = parsed

	return cfg, nil
}

func validateInputFile(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("el archivo no existe: %s", path)
		}
		return fmt.Errorf("no se pudo leer %s: %w", path, err)
	}
	if info.IsDir() {
		return fmt.Errorf("la ruta es un directorio, no un archivo: %s", path)
	}
	if strings.ToLower(filepath.Ext(path)) != ".md" {
		return fmt.Errorf("el archivo debe tener extensión .md: %s", path)
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
