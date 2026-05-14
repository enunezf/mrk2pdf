# mrk2pdf

Convertidor CLI de Markdown a PDF con plantillas embebidas, syntax highlighting, diagramas mermaid, modo batch y modo watch. Un binario único sin dependencias en tiempo de ejecución más allá de un navegador Chromium en el sistema.

```bash
./mrk2pdf -i documento.md
# → genera documento.pdf en el mismo directorio
```

## Tabla de contenidos

- [Características](#características)
- [Instalación](#instalación)
- [Uso rápido](#uso-rápido)
- [Flags](#flags)
- [Frontmatter YAML](#frontmatter-yaml)
- [Tabla de contenidos en el documento](#tabla-de-contenidos-en-el-documento)
- [Plantillas](#plantillas)
- [Modo batch](#modo-batch)
- [Modo watch](#modo-watch)
- [Variables de entorno](#variables-de-entorno)
- [Plataformas soportadas](#plataformas-soportadas)
- [Limitaciones conocidas](#limitaciones-conocidas)
- [Compilar desde fuentes](#compilar-desde-fuentes)

## Características

- **Markdown completo**: CommonMark + GFM (tablas con alineación, strikethrough, autolinks, task lists)
- **Frontmatter YAML**: `title`, `author`, `date`, `tags`, con fallbacks razonables si faltan
- **Tabla de contenidos automática** vía marcador `[TOC]` o flag `--toc`
- **Plantillas personalizables**: 3 embebidas en el binario (`default`, `elegante`, `cian`), más cualquier custom en `template/`
- **Syntax highlighting** vía Prism.js (bash, go, yaml, json, python, typescript, sql, html, css)
- **Diagramas mermaid** renderizados en el navegador antes de imprimir
- **Imágenes self-contained**: se embeben como `data:` URIs base64 en el HTML, los PDFs no dependen de archivos externos
- **Multiplataforma**: descubrimiento automático del navegador en Linux, macOS, Windows y WSL
- **Modo batch**: procesa archivos, directorios o globs con un único arranque de browser
- **Modo watch**: regenera al detectar cambios; los errores no matan el proceso
- **Layout configurable**: A4 / Letter, vertical / apaisado
- **Auto-discovery de navegador**: Chrome, Chromium, Edge, Brave; override con `BROWSER_PATH`

## Instalación

### Pre-requisitos

Un navegador basado en Chromium en el sistema:

| Plataforma | Lo que busca primero | Fallback |
|---|---|---|
| Linux | `google-chrome`, `chromium`, `chromium-browser`, `microsoft-edge`, `brave-browser` en `$PATH` | — |
| macOS | Lo anterior en PATH | `/Applications/Google Chrome.app/...`, `/Applications/Chromium.app/...`, etc. |
| Windows nativo | `chrome`, `msedge`, `brave` en PATH | `C:\Program Files\...\chrome.exe`, `%LOCALAPPDATA%\...` |
| WSL | Chromium en WSL primero | `/mnt/c/Program Files/.../chrome.exe` |

Edge viene preinstalado en Windows 10/11. macOS suele tener Chrome o se instala fácil. En Linux, `chromium-browser` cubre todo. Para WSL puro con solo Windows .exe, [ver limitaciones](#limitaciones-conocidas).

### Compilar

```bash
git clone <url-del-repo>
cd mrk2pdf
go build -o mrk2pdf .
```

Requiere Go 1.22+. El resultado es un binario único de ~16 MB con todos los assets embebidos.

## Uso rápido

```bash
# Convertir un archivo
./mrk2pdf -i documento.md
# → documento.pdf

# Especificar salida explícita
./mrk2pdf -i documento.md -o final.pdf

# Usar una plantilla diferente
./mrk2pdf -i documento.md -t elegante

# Listar plantillas disponibles
./mrk2pdf -l

# Procesar todos los .md de una carpeta
./mrk2pdf -i docs/

# Modo watch: regenera al cambiar el archivo
./mrk2pdf -i documento.md -w
```

## Flags

Resumen rápido:

| Flag | Tipo | Default | Descripción |
|---|---|---|---|
| `-i` | string | _(obligatorio salvo con `-l`)_ | Archivo, directorio o glob de entrada |
| `-o` | string | `<input>.pdf` | Archivo PDF o directorio destino |
| `-t` | string | `default` | Nombre de la plantilla |
| `-d` | bool | `false` | Forzar sobrescritura de `template/default/` |
| `-l` | bool | `false` | Listar plantillas disponibles y salir |
| `-w` | bool | `false` | Modo watch |
| `-R` | bool | `false` | Recursivo cuando `-i` es un directorio |
| `--toc` | bool | `false` | Auto-prepend TOC al inicio del documento |
| `--size` | string | `A4` | Tamaño de página: `A4` o `Letter` |
| `--landscape` | bool | `false` | Orientación apaisada |

A continuación, descripción exhaustiva de cada uno.

### `-i <ruta>` — Entrada

**Tipo:** string · **Default:** _(vacío)_ · **Obligatorio salvo con `-l`**

Especifica los documentos a procesar. La interpretación es smart:

| Forma | Comportamiento |
|---|---|
| `-i documento.md` | Procesa exactamente ese archivo |
| `-i docs/` | Escanea `docs/` por archivos `.md` (solo top-level, agregá `-R` para recursivo) |
| `-i "docs/*.md"` | Glob (las comillas evitan que el shell expanda; pasa el patrón crudo al binario) |
| `-i docs/*.md` (sin comillas) | El shell expande; el primer archivo va a `-i`, el resto como args posicionales |
| Sin `-i` + args posicionales | `./mrk2pdf documento.md` también funciona |

Si la ruta no existe o no es `.md`, el programa falla con un mensaje claro antes de hacer nada.

### `-o <ruta>` — Salida

**Tipo:** string · **Default:** _(vacío → derivado por archivo)_

Define dónde escribir los PDFs. La interpretación depende del valor y de cuántos inputs hay:

| Caso | Resultado |
|---|---|
| `-o` vacío | Cada `<archivo>.md` produce `<archivo>.pdf` al lado |
| `-o salida.pdf` con un único input | Escribe en `salida.pdf` |
| `-o resultados/` (termina en `/` o ya existe como dir) | Crea el dir si no existe; cada `<archivo>.md` produce `resultados/<archivo>.pdf` |
| `-o salida.pdf` con múltiples inputs | **Error**: ambiguo. Hay que usar un directorio destino. |

Ejemplos:

```bash
./mrk2pdf -i doc.md                    # → doc.pdf
./mrk2pdf -i doc.md -o /tmp/out.pdf    # → /tmp/out.pdf
./mrk2pdf -i docs/ -o pdfs/            # → pdfs/doc1.pdf, pdfs/doc2.pdf, ...
```

### `-t <nombre>` — Plantilla

**Tipo:** string · **Default:** `default`

Nombre de la subcarpeta de `template/` a usar. Si la plantilla no existe, `mrk2pdf` sugiere la más cercana por distancia de Levenshtein y lista las disponibles:

```bash
$ ./mrk2pdf -i doc.md -t elegantte
2026/05/13 error: el template 'elegantte' no existe en template/
¿Quizás quisiste decir 'elegante'?
Disponibles: cian, default, elegante
```

Plantillas embebidas en el binario:

- `default` — tipografía sans-serif, encabezado simple con metadata; estilo limpio para reports y notas
- `elegante` — serif (Playfair + Lora), navy + dorado, cover page completa, running headers
- `cian` — sans-serif (Inter), bandas cian en cada página vía `@page` margin boxes

Para crear una plantilla custom, ver [`template/README.md`](template/README.md).

### `-d` — Regenerar `template/default/`

**Tipo:** bool · **Default:** `false`

Sobrescribe `template/default/` con los archivos embebidos en el binario. Útil si:

- Querés deshacer cambios que hiciste en el default
- Actualizaste el binario y querés agarrar la nueva versión del template

**Importante:** `-d` solo afecta a `default`. Las otras plantillas (incluidas las embebidas como `elegante` y `cian`) nunca se sobrescriben automáticamente; tus customizaciones persisten.

### `-l` — Listar plantillas

**Tipo:** bool · **Default:** `false`

Modo informacional: lista los nombres de plantillas presentes en `template/` y sale. No requiere `-i`.

```bash
$ ./mrk2pdf -l
Plantillas disponibles:
  - cian
  - default
  - elegante
```

Si `template/` no existe aún, el binario lo crea y extrae las plantillas embebidas antes de listar.

### `-w` — Modo watch

**Tipo:** bool · **Default:** `false`

Activa la vigilancia de los inputs vía `fsnotify`. Tras un render inicial, el proceso queda esperando eventos y regenera al detectar cambios. Detalles en [Modo watch](#modo-watch).

### `-R` — Recursivo

**Tipo:** bool · **Default:** `false`

Cuando `-i` es un directorio, escanea recursivamente buscando `.md` en todos los subdirectorios. Sin `-R`, solo se procesan los archivos del nivel superior.

```bash
./mrk2pdf -i docs/         # docs/*.md (no entra a docs/subdir/)
./mrk2pdf -i docs/ -R      # docs/**/*.md (toda la jerarquía)
```

> **Nota:** la combinación `-w -R` (watch recursivo) hace el snapshot inicial recursivo pero el watching efectivo es solo flat — `fsnotify` en Linux usa `inotify`, que no es recursivo. Cambios en subdirectorios después del arranque no se detectan en esta versión.

### `--toc` — Tabla de contenidos automática

**Tipo:** bool · **Default:** `false`

Cuando se pasa, prepend una tabla de contenidos al inicio del documento, generada a partir de todas las heading IDs auto-generadas por goldmark. Si el documento ya contiene un marcador `[TOC]`, este flag **no** agrega un segundo TOC — gana el marcador (el usuario eligió la posición).

```bash
./mrk2pdf -i doc.md --toc                       # TOC al principio
./mrk2pdf -i doc.md                             # solo si doc.md tiene [TOC]
```

### `--size A4|Letter` — Tamaño de página

**Tipo:** string · **Default:** `A4`

Configura las dimensiones de la página enviadas a Chrome:

| Valor | Dimensiones |
|---|---|
| `A4` | 210 × 297 mm (8.27" × 11.69") |
| `Letter` | 8.5" × 11" |

Si pasás otro valor, el programa rechaza con error explícito. Es case-insensitive (`a4`, `LETTER`, etc. todos funcionan).

### `--landscape` — Orientación apaisada

**Tipo:** bool · **Default:** `false`

Imprime el PDF apaisado (rotado 90°). Las dimensiones de `--size` siguen aplicando — Chrome se encarga de la rotación. Útil para tablas anchas, presentaciones, o diagramas.

```bash
./mrk2pdf -i wide-table.md --landscape
./mrk2pdf -i slides.md --size Letter --landscape
```

## Frontmatter YAML

`mrk2pdf` parsea un bloque YAML opcional al inicio del archivo `.md` con `github.com/adrg/frontmatter`. Es el patrón habitual de Jekyll, Hugo, etc.

```yaml
---
title: "Mi documento"
author: "Equipo de Plataforma"
date: "2026-05-13"
tags:
  - migration
  - auth
  - golang
---

# Contenido del markdown...
```

Campos reconocidos por la plantilla por defecto:

| Campo | Tipo | Uso |
|---|---|---|
| `title` | string | Título; aparece en `<title>` y como header del documento |
| `author` | string | Autor; mostrado en running footers donde aplique |
| `date` | string | Fecha; se preserva como string (sin parseo) |
| `tags` | lista de strings | Lista de etiquetas; renderizado depende del template |

**Fallbacks** cuando falta un campo:

- `title` vacío → nombre del archivo sin extensión (`docs/manual.md` → `manual`)
- `date` vacío → fecha actual en formato `2026-05-13`
- `author` y `tags` vacíos → no se renderizan

Si no hay bloque `---...---` al inicio del archivo, todo el contenido se trata como markdown puro y los fallbacks se aplican igual.

Si el YAML está malformado (sintaxis rota), el render falla con un error claro indicando línea y motivo. En [modo batch](#modo-batch) o [modo watch](#modo-watch), el error se reporta pero el proceso sigue con los otros archivos.

## Tabla de contenidos en el documento

Hay dos formas mutuamente excluyentes de inyectar el TOC:

### Marcador `[TOC]` en el markdown

```markdown
# Título

[TOC]

## Sección 1
...
```

Donde aparezca `[TOC]` (en su propio párrafo), se reemplaza por el TOC. Útil cuando querés que vaya después de una introducción.

### Flag `--toc`

Si el `.md` **no contiene** `[TOC]`, pasar `--toc` prepend el TOC al inicio:

```bash
./mrk2pdf -i doc.md --toc
```

Si pasás `--toc` **y** el documento tiene `[TOC]`, gana el marcador — no se generan dos TOCs.

## Plantillas

Cada plantilla es una carpeta dentro de `template/` con dos archivos: `index.html` (Go template) y `style.css`. El binario las descubre dinámicamente en cada arranque.

Las plantillas embebidas (`default`, `elegante`, `cian`) se extraen automáticamente al primer arranque. Después podés editarlas libremente; tus cambios persisten en runs subsiguientes (solo `-d` regenera el default).

Para detalles sobre cómo crear plantillas, las variables disponibles y un ejemplo completo, ver [`template/README.md`](template/README.md).

## Modo batch

Múltiples archivos se procesan en una sola invocación, reusando la misma instancia del navegador entre renders:

```bash
# Todos los .md de un dir
./mrk2pdf -i docs/

# Recursivo
./mrk2pdf -i docs/ -R

# Glob con quotes
./mrk2pdf -i "docs/*.md"

# Args posicionales (después de los flags)
./mrk2pdf -t elegante docs/intro.md docs/api.md docs/release.md

# Output a un directorio
./mrk2pdf -i docs/ -o pdfs/
```

Salida típica:

```
Usando template: default
Usando navegador: /snap/bin/chromium
Procesando 5 archivos en A4 vertical...

[1/5] docs/intro.md → docs/intro.pdf (41 KB)
[2/5] docs/api.md → docs/api.pdf (89 KB)
[3/5] docs/release.md → docs/release.pdf (52 KB)
[4/5] docs/migration.md → docs/migration.pdf (118 KB)
[5/5] docs/troubleshooting.md → docs/troubleshooting.pdf (76 KB)

5 documentos procesados correctamente.
```

**Política de errores en batch:**

- **Fail-fast** en validaciones obvias (path inexistente, extensión != `.md`, glob sin matches, template inexistente)
- **Continue-on-error** en runtime (frontmatter roto, render fallido): cada falla se reporta y el resto sigue. Al final, resumen de fallos y exit code `1` si alguno falló.

## Modo watch

`-w` activa la regeneración automática al cambiar los archivos. Tras un render inicial, queda escuchando vía `fsnotify`.

```bash
# Watch un solo archivo
./mrk2pdf -i documento.md -w

# Watch un directorio (cualquier .md que cambie se regenera)
./mrk2pdf -i docs/ -w

# Watch con template específico
./mrk2pdf -i docs/ -t elegante -w
```

Salida en una sesión típica:

```
Usando template: default
Usando navegador: /snap/bin/chromium
Modo watch activo. Render inicial:
[15:20:33] docs/manual.md → docs/manual.pdf (54 KB)

Vigilando 1 directorio(s): /home/user/docs
Ctrl+C para salir.
[15:21:02] docs/manual.md → docs/manual.pdf (54 KB)
[15:21:18] docs/api.md → docs/api.pdf (78 KB)
[15:21:45] docs/manual.md → ERROR (convirtiendo): parseando frontmatter: yaml: line 3: ...
[15:22:01] docs/manual.md → docs/manual.pdf (54 KB)

^C
Deteniendo watch.
```

**Comportamiento:**

- Vigila **directorios** internamente (los editores hacen save-by-rename, lo que rompería el watch por filename). Para entradas tipo archivo, vigila el directorio padre y filtra por path exacto.
- **Debounce de 250 ms** por archivo: los editores emiten Create+Write+Rename para un mismo save; se consolida en un único render.
- **Errores no matan el proceso**: se loguean con timestamp y el watcher sigue activo.
- **Ctrl+C** (o SIGTERM) → shutdown limpio del watcher y del browser.

Reacciona a:

- Edición de un archivo vigilado (Write event)
- Creación de un archivo nuevo en un directorio vigilado (Create event)
- Save-by-rename de editores (Rename llega como Create del target)

No reacciona a: Remove, Chmod, Rename-out.

## Variables de entorno

### `BROWSER_PATH`

Override absoluto del navegador a usar. Toma prioridad sobre todo el auto-discovery. Útil para forzar una versión específica o para entornos donde el navegador no está en `$PATH`:

```bash
export BROWSER_PATH=/opt/chromium/chrome
./mrk2pdf -i doc.md

# Una sola vez
BROWSER_PATH=/usr/bin/google-chrome-beta ./mrk2pdf -i doc.md
```

Si `BROWSER_PATH` apunta a un archivo inexistente, el programa falla con error claro sin caer al auto-discovery.

## Plataformas soportadas

| Plataforma | Estado |
|---|---|
| Linux nativo (Chrome, Chromium, Edge o Brave en PATH) | ✅ Pleno |
| macOS (Chrome/Chromium/Edge/Brave en `/Applications` o PATH) | ✅ Pleno |
| Windows 10/11 (Edge preinstalado siempre funciona; Chrome en Program Files o `%LOCALAPPDATA%`) | ✅ Pleno |
| WSL con Chromium nativo (recomendado: `sudo apt install chromium-browser`) | ✅ Pleno |
| WSL usando chrome.exe de Windows (sin Chromium en WSL) | ⚠️ Funciona si Windows 11 22H2+ con networking mirrored, o si abrís el Windows Firewall para chrome.exe. Si no, el error indica las tres opciones para resolver. |

El binario cross-compila para `linux/amd64`, `darwin/amd64`, `darwin/arm64` y `windows/amd64` sin cambios:

```bash
GOOS=darwin  GOARCH=arm64 go build -o dist/mrk2pdf-mac-arm    .
GOOS=windows GOARCH=amd64 go build -o dist/mrk2pdf-win.exe    .
GOOS=linux   GOARCH=amd64 go build -o dist/mrk2pdf-linux      .
```

## Limitaciones conocidas

- **Watch recursivo limitado** (`-w -R`): el snapshot inicial es recursivo pero el `inotify` subyacente no propaga eventos de subdirectorios creados durante la sesión.
- **CDN requerido para Prism/mermaid**: el template default carga `prism.min.js` y `mermaid.min.js` desde `cdnjs` y `jsdelivr`. Sin internet, los bloques de código no tendrán highlighting y los diagramas mermaid quedarán como texto plano. Workaround: clonar los assets a tu template y servirlos desde disco.
- **WSL2 + chrome.exe**: cuando el único navegador es Windows chrome.exe y el Firewall bloquea inbound desde la red de WSL, la comunicación DevTools falla. La salida del error sugiere instalar Chromium en WSL, habilitar mirrored networking, o agregar una regla al Firewall.
- **Prism: lenguajes incluidos**: bash, go, yaml, json, python, typescript, sql. Para otros (rust, ruby, etc.) hay que editar `assets/<template>/index.html` para agregar el `<script>` correspondiente y recompilar, o agregarlos a un template custom sin recompilar.

## Compilar desde fuentes

```bash
git clone <url-del-repo>
cd mrk2pdf

# Build local
go build -o mrk2pdf .

# Tests (vet)
go vet ./...

# Cross-compile
GOOS=darwin  GOARCH=arm64 go build -o dist/mrk2pdf-mac-arm    .
GOOS=darwin  GOARCH=amd64 go build -o dist/mrk2pdf-mac-intel  .
GOOS=windows GOARCH=amd64 go build -o dist/mrk2pdf-win.exe    .
GOOS=linux   GOARCH=amd64 go build -o dist/mrk2pdf-linux      .
```

**Estructura del proyecto:**

```
mrk2pdf/
├── README.md                    ← este archivo
├── go.mod / go.sum
├── main.go                      ← orquestación
├── flags.go                     ← parseo y validación de flags
├── inputs.go                    ← resolución de inputs (file/dir/glob) y outputs
├── template.go                  ← gestión de plantillas + embed FS
├── watch.go                     ← modo watch (fsnotify + debounce + signal)
├── assets/                      ← templates source, embebidos via go:embed
│   ├── default/
│   ├── elegante/
│   └── cian/
├── internal/
│   ├── converter/               ← Markdown → HTML
│   │   ├── converter.go
│   │   ├── frontmatter.go
│   │   ├── markdown.go
│   │   └── toc.go
│   └── pdf/                     ← HTML → PDF
│       ├── pdf.go               ← Renderer reusable
│       ├── pagesize.go          ← A4 / Letter
│       ├── finder.go            ← Auto-discovery de navegador
│       └── wsl.go               ← Path WSL → chrome.exe con rewrite de URL DevTools
├── template/                    ← extracted runtime (gitignored excepto README)
│   └── README.md                ← guía para crear plantillas
└── ejemplos/                    ← markdown de ejemplo para pruebas
```

**Dependencias principales:**

| Librería | Uso |
|---|---|
| `github.com/yuin/goldmark` | Parser de Markdown (CommonMark + GFM) |
| `github.com/adrg/frontmatter` | Parser de YAML frontmatter |
| `github.com/chromedp/chromedp` | Control de Chrome headless para PrintToPDF |
| `github.com/fsnotify/fsnotify` | Eventos de filesystem para modo watch |

Más Prism.js y mermaid vía CDN en runtime (no son dependencias Go).
