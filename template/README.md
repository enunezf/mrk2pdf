# Plantillas de `mrk2pdf`

Cada plantilla es una carpeta dentro de `template/` con **dos archivos obligatorios**:

| Archivo | Rol |
|---|---|
| `index.html` | Esqueleto Go template del documento HTML que se imprime a PDF |
| `style.css` | Hoja de estilos cargada via `{{.Styles}}` en el `<head>` |

El binario descubre las plantillas dinámicamente en tiempo de ejecución: alcanza con crear una nueva carpeta bajo `template/`. **No es necesario recompilar** para usar una plantilla nueva.

## Cómo se descubre una plantilla

1. Al primer arranque, el binario extrae sus plantillas embebidas (`default`, `elegante`, `cian`, `github`) en `template/<nombre>/` si no existen todavía
2. En cada ejecución, `mrk2pdf -l` lista todos los subdirectorios de `template/`
3. `mrk2pdf -i archivo.md -t <nombre>` carga `template/<nombre>/index.html` y `template/<nombre>/style.css` en cada render

## Plantillas embebidas

| Nombre | Estilo | Uso típico |
|---|---|---|
| `default` | Sans-serif simple con metadata header | Reports, notas, documentación interna |
| `elegante` | Serif (Playfair + Lora), navy + dorado, cover page | Informes formales, propuestas |
| `cian` | Sans-serif (Inter), bandas cian arriba y abajo en cada página | Documentación corporativa con identidad visual |
| `github` | Clon visual de cómo GitHub renderiza un README en su web (vía `github-markdown-css`) | READMEs, documentación open-source, archivos pensados para GitHub |

El flag `-d` regenera **solo** `template/default/` desde el embed; el resto de plantillas nunca es sobrescrito automáticamente, así que tus customizaciones se preservan.

## Variables disponibles en `index.html`

`index.html` se procesa con el paquete `html/template` de Go. Recibe esta estructura de datos:

```go
type pageData struct {
    Meta    Meta             // metadata del frontmatter YAML
    Styles  template.CSS     // contenido íntegro de style.css
    Content template.HTML    // HTML generado desde el markdown
}

type Meta struct {
    Title  string    // {{.Meta.Title}}
    Author string    // {{.Meta.Author}}
    Date   string    // {{.Meta.Date}}
    Tags   []string  // {{range .Meta.Tags}}{{.}}{{end}}
}
```

Si el `.md` no tiene frontmatter o le faltan campos, el converter aplica fallbacks:
- `Title` vacío → nombre del archivo sin extensión
- `Date` vacío → fecha de hoy en formato `2026-05-13`

### Placeholders típicos

| Placeholder | Renderiza |
|---|---|
| `{{.Meta.Title}}` | Título del documento (con fallback) |
| `{{.Meta.Author}}` | Autor (vacío si no se definió) |
| `{{.Meta.Date}}` | Fecha (hoy si no se definió) |
| `{{range .Meta.Tags}}{{.}}{{end}}` | Itera sobre los tags |
| `{{.Styles}}` | Contenido completo del `style.css`, listo para `<style>...</style>` |
| `{{.Content}}` | HTML del markdown (ya parseado con goldmark + GFM) |
| `{{if .Meta.Author}}…{{end}}` | Renderiza solo si Author está presente |

### Inyección dinámica en CSS

Los placeholders también se pueden usar dentro de un bloque `<style>` en `index.html`. Esto es útil para `@page` con margin boxes que necesitan el título o autor en runtime:

```html
<style>
    @page {
        @top-center { content: "{{.Meta.Title}}"; }
        @bottom-left { content: "{{.Meta.Author}}"; }
        @bottom-right { content: counter(page) " / " counter(pages); }
    }
</style>
```

## Crear una plantilla nueva paso a paso

1. Crear la carpeta: `mkdir template/mi-plantilla`
2. Crear `template/mi-plantilla/index.html` con al menos `{{.Content}}` dentro
3. Crear `template/mi-plantilla/style.css` (puede estar vacío)
4. Usar: `./mrk2pdf -i documento.md -t mi-plantilla`

Si te equivocás con el nombre (`-t mi-pantilla`), `mrk2pdf` sugiere el más parecido por distancia de Levenshtein y lista todos los disponibles.

## Ejemplo completo: plantilla `cian`

A continuación se muestra una plantilla completa con **encabezado y pie de página en color cian** que aparece en cada página del PDF. Esta plantilla **viene incluida en el binario** — la podés inspeccionar en `template/cian/` después del primer arranque, o copiar el código de abajo para empezar la tuya.

La plantilla usa **CSS @page margin boxes** para las bandas. Chrome (vía chromedp) los respeta cuando genera el PDF.

### `template/cian/index.html`

```html
<!DOCTYPE html>
<html lang="es">
<head>
    <meta charset="UTF-8">
    <title>{{.Meta.Title}}</title>
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
    <link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&display=swap" rel="stylesheet">
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/prism/1.29.0/themes/prism-tomorrow.min.css">
    <style>
        /* Bandas cian: header (arriba) y footer (abajo) en CADA página.   */
        /* Se cubren las 3 zonas del margen (left/center/right) con el     */
        /* mismo background para conseguir el efecto de banda continua.    */
        @page {
            margin: 2.6cm 1.8cm 2.6cm 1.8cm;

            @top-left   { content: ""; background-color: #0891b2; }
            @top-center {
                content: "{{.Meta.Title}}";
                background-color: #0891b2;
                color: white;
                font-family: 'Inter', sans-serif;
                font-size: 10pt;
                font-weight: 600;
            }
            @top-right  { content: ""; background-color: #0891b2; }

            @bottom-left {
                content: "{{if .Meta.Author}}{{.Meta.Author}}{{end}}";
                background-color: #0891b2;
                color: white;
                font-family: 'Inter', sans-serif;
                font-size: 9pt;
                padding: 0 0.6em;
            }
            @bottom-center { content: ""; background-color: #0891b2; }
            @bottom-right {
                content: "Página " counter(page) " / " counter(pages);
                background-color: #0891b2;
                color: white;
                font-family: 'Inter', sans-serif;
                font-size: 9pt;
                padding: 0 0.6em;
            }
        }
    </style>
    <style>{{.Styles}}</style>
</head>
<body>
    {{if .Meta.Title}}
    <header class="meta">
        {{if .Meta.Author}}<div class="meta-author">{{.Meta.Author}}</div>{{end}}
        {{if .Meta.Date}}<div class="meta-date">{{.Meta.Date}}</div>{{end}}
        {{if .Meta.Tags}}
        <div class="meta-tags">{{range .Meta.Tags}}<span class="tag">{{.}}</span>{{end}}</div>
        {{end}}
    </header>
    {{end}}

    <main class="content">
        {{.Content}}
    </main>

    <!-- Prism + mermaid: scripts que toda plantilla pdf-ready debería incluir -->
    <script src="https://cdnjs.cloudflare.com/ajax/libs/prism/1.29.0/prism.min.js"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/prism/1.29.0/components/prism-bash.min.js"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/prism/1.29.0/components/prism-go.min.js"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/prism/1.29.0/components/prism-yaml.min.js"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/prism/1.29.0/components/prism-json.min.js"></script>
    <script src="https://cdn.jsdelivr.net/npm/mermaid@10/dist/mermaid.min.js"></script>
    <script>
        (function () {
            document.querySelectorAll('pre > code.language-mermaid').forEach(function (code) {
                var div = document.createElement('div');
                div.className = 'mermaid';
                div.textContent = code.textContent;
                code.parentElement.replaceWith(div);
            });
            if (window.Prism) { Prism.highlightAll(); }
            if (window.mermaid) {
                window.mermaid.initialize({ startOnLoad: false, securityLevel: 'loose' });
                window.mermaid.run().catch(function (e) { console.error('mermaid:', e); });
            }
        })();
    </script>
</body>
</html>
```

### `template/cian/style.css`

```css
:root {
    --cian: #0891b2;
    --cian-bg: #ecfeff;
    --text: #1a1a1a;
    --muted: #666;
    --rule: #e5e5e5;
}

body {
    font-family: 'Inter', sans-serif;
    font-size: 11pt;
    line-height: 1.6;
    color: var(--text);
}

h1, h2, h3 { color: var(--cian); font-weight: 700; }
h2 { border-bottom: 3px solid var(--cian); padding-bottom: 0.2em; }

.meta {
    margin-bottom: 1.8em;
    padding-bottom: 1em;
    border-bottom: 2px solid var(--cian-bg);
}
.meta-author { font-weight: 600; font-size: 12pt; }
.meta-date { color: var(--muted); font-size: 10pt; }
.tag {
    display: inline-block;
    background: var(--cian);
    color: white;
    padding: 0.2em 0.8em;
    margin: 0.15em;
    border-radius: 12px;
    font-size: 9pt;
}

code { background: var(--cian-bg); color: var(--cian); padding: 0.1em 0.4em; border-radius: 3px; }
pre[class*="language-"] { border-left: 4px solid var(--cian); }
blockquote { border-left: 4px solid var(--cian); background: var(--cian-bg); padding: 0.8em 1.2em; }
th { background: var(--cian); color: white; padding: 0.6em 0.9em; }
tr:nth-child(even) td { background: var(--cian-bg); }
```

> ⚠️ El bloque anterior es una **versión abreviada** para legibilidad en el README. La versión completa (con listas, hr, TOC y resto) está en el archivo real `template/cian/style.css` que se extrae al primer arranque.

### Probarlo

```bash
./mrk2pdf -i ejemplos/07_frontmatter.md -t cian
```

Abrí el PDF y verás:

- **En cada página:** banda cyan arriba con el título centrado, banda cian abajo con autor a la izquierda y `Página N / total` a la derecha
- **En la primera página:** además, los metadatos (autor, fecha, tags) renderizados en el cuerpo arriba del contenido
- **Headings** en cyan con línea inferior 3 px
- **Code blocks** con tema "tomorrow" (oscuro) y borde izquierdo cyan
- **Tablas** con header cyan y filas alternadas en cyan tenue

## Modificar una plantilla embebida

Las plantillas `default`, `elegante` y `cian` se extraen del binario al primer arranque. Después podés editar libremente los archivos en `template/<nombre>/` y tus cambios persisten:

- `default`: el flag `-d` la regenera desde el embed, descartando tus cambios. Usalo solo si querés volver al original
- `elegante`, `cian`, o cualquier plantilla custom: **nunca** son sobrescritas automáticamente

## Distribuir una plantilla

Para compartir tu plantilla con otra persona:

```bash
# Empaquetar
tar czf mi-plantilla.tar.gz -C template mi-plantilla

# La otra persona la descomprime en su template/
tar xzf mi-plantilla.tar.gz -C template/
mrk2pdf -i doc.md -t mi-plantilla
```

Si querés que la plantilla viaje **dentro del binario** (para usuarios que solo bajan el ejecutable), tenés que agregarla a `assets/<nombre>/` en las fuentes del proyecto y recompilar `mrk2pdf`. El embed Go (`//go:embed assets` en `template.go`) lo levanta automáticamente.
