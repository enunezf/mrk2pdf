package main

import (
	"fmt"
	"log"
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
	fmt.Println("Configuración inicial completa. (La conversión a PDF se implementará en el siguiente paso.)")
}
