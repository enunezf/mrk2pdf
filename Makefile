# mrk2pdf — Cross-compilation Makefile
#
# Genera binarios estáticos para Linux, macOS y Windows en la carpeta dist/.
# Los binarios se commitean al repositorio para que los usuarios puedan
# descargarlos directamente desde la página del proyecto.

BINARY  := mrk2pdf
DIST    := dist
LDFLAGS := -s -w

# Default target: mostrar ayuda cuando se ejecuta `make` sin argumentos.
.DEFAULT_GOAL := help

.PHONY: help all clean local linux mac windows \
        linux-amd64 linux-arm64 mac-amd64 mac-arm64 windows-amd64 \
        checksums list

help: ## Lista los targets disponibles
	@echo "mrk2pdf — Makefile"
	@echo ""
	@echo "Uso: make <target>"
	@echo ""
	@echo "Targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

all: linux mac windows ## Compila los binarios para TODAS las plataformas (5 binarios)

linux: linux-amd64 linux-arm64 ## Linux amd64 + arm64
mac:   mac-amd64 mac-arm64     ## macOS Intel + Apple Silicon
windows: windows-amd64         ## Windows amd64

linux-amd64: ## Solo Linux amd64
	@mkdir -p $(DIST)
	@echo "→ Linux amd64"
	@GOOS=linux GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(DIST)/$(BINARY)-linux-amd64 .

linux-arm64: ## Solo Linux arm64 (RPi, AWS Graviton)
	@mkdir -p $(DIST)
	@echo "→ Linux arm64"
	@GOOS=linux GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o $(DIST)/$(BINARY)-linux-arm64 .

mac-amd64: ## Solo macOS Intel (amd64)
	@mkdir -p $(DIST)
	@echo "→ macOS Intel"
	@GOOS=darwin GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(DIST)/$(BINARY)-mac-amd64 .

mac-arm64: ## Solo macOS Apple Silicon (arm64)
	@mkdir -p $(DIST)
	@echo "→ macOS Apple Silicon"
	@GOOS=darwin GOARCH=arm64 go build -ldflags="$(LDFLAGS)" -o $(DIST)/$(BINARY)-mac-arm64 .

windows-amd64: ## Solo Windows amd64
	@mkdir -p $(DIST)
	@echo "→ Windows amd64"
	@GOOS=windows GOARCH=amd64 go build -ldflags="$(LDFLAGS)" -o $(DIST)/$(BINARY)-windows-amd64.exe .

local: ## Compila para la plataforma actual en la raíz del repo (./mrk2pdf)
	@go build -ldflags="$(LDFLAGS)" -o $(BINARY) .
	@echo "→ $(BINARY) compilado para $$(go env GOOS)/$$(go env GOARCH)"

checksums: ## Genera SHA256SUMS de los binarios en dist/
	@if [ ! -d $(DIST) ]; then echo "Error: $(DIST)/ no existe. Ejecutá 'make all' primero."; exit 1; fi
	@cd $(DIST) && \
	  if command -v sha256sum >/dev/null 2>&1; then \
	    sha256sum $(BINARY)-* > SHA256SUMS; \
	  else \
	    shasum -a 256 $(BINARY)-* > SHA256SUMS; \
	  fi
	@echo "→ SHA256SUMS generado en $(DIST)/"

list: ## Lista los binarios y tamaños en dist/
	@if [ ! -d $(DIST) ]; then echo "Error: $(DIST)/ no existe."; exit 1; fi
	@ls -lh $(DIST)/ | awk 'NR>1 {printf "  %-35s %s\n", $$NF, $$5}'

clean: ## Elimina la carpeta dist/
	@rm -rf $(DIST)
	@echo "→ $(DIST)/ eliminado"
