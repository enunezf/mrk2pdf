package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	Input        string
	Output       string
	Template     string
	ForceDefault bool
}

func parseFlags() (*Config, error) {
	cfg := &Config{}

	flag.StringVar(&cfg.Input, "i", "", "ruta del archivo Markdown de entrada (obligatorio)")
	flag.StringVar(&cfg.Output, "o", "", "ruta del archivo PDF de salida (por defecto: <input>.pdf)")
	flag.StringVar(&cfg.Template, "t", defaultTemplate, "nombre del template (carpeta dentro de template/)")
	flag.BoolVar(&cfg.ForceDefault, "d", false, "forzar sobrescritura de template/default")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Uso: %s -i <archivo.md> [-o <salida.pdf>] [-t <template>] [-d]\n\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	if cfg.Input == "" {
		flag.Usage()
		return nil, errors.New("el flag -i es obligatorio")
	}

	if cfg.Output == "" {
		ext := filepath.Ext(cfg.Input)
		cfg.Output = strings.TrimSuffix(cfg.Input, ext) + ".pdf"
	}

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

func validateTemplate(cfg *Config) error {
	tplDir := filepath.Join(templateRoot, cfg.Template)
	info, err := os.Stat(tplDir)
	if err != nil || !info.IsDir() {
		return fmt.Errorf("el template '%s' no existe en %s/", cfg.Template, templateRoot)
	}
	return nil
}
