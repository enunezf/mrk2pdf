package converter

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
)

// Options bundles inputs for Render.
type Options struct {
	InputPath   string
	TemplateDir string
	AutoTOC     bool
}

// Meta carries front-matter metadata exposed to the template as .Meta.
type Meta struct {
	Title  string   `yaml:"title"`
	Author string   `yaml:"author"`
	Date   string   `yaml:"date"`
	Tags   []string `yaml:"tags"`
}

// pageData is the value Go templates receive when rendering index.html.
type pageData struct {
	Meta    Meta
	Styles  template.CSS
	Content template.HTML
}

// Render reads the Markdown input, parses any YAML front-matter, converts
// the body via goldmark (+ GFM, image embedding, TOC), and renders it into
// the given template directory's index.html.
func Render(opt Options) ([]byte, error) {
	raw, err := os.ReadFile(opt.InputPath)
	if err != nil {
		return nil, fmt.Errorf("leyendo markdown %s: %w", opt.InputPath, err)
	}

	meta, mdSrc, err := extractFrontmatter(raw)
	if err != nil {
		return nil, fmt.Errorf("parseando frontmatter: %w", err)
	}
	meta = fillMetaDefaults(meta, opt.InputPath)

	htmlContent, err := markdownToHTML(mdSrc, filepath.Dir(opt.InputPath), opt.AutoTOC)
	if err != nil {
		return nil, fmt.Errorf("convirtiendo markdown: %w", err)
	}

	tmplBytes, err := os.ReadFile(filepath.Join(opt.TemplateDir, "index.html"))
	if err != nil {
		return nil, fmt.Errorf("leyendo plantilla index.html: %w", err)
	}
	cssBytes, err := os.ReadFile(filepath.Join(opt.TemplateDir, "style.css"))
	if err != nil {
		return nil, fmt.Errorf("leyendo plantilla style.css: %w", err)
	}

	tmpl, err := template.New("page").Parse(string(tmplBytes))
	if err != nil {
		return nil, fmt.Errorf("parseando plantilla: %w", err)
	}

	var out bytes.Buffer
	err = tmpl.Execute(&out, pageData{
		Meta:    meta,
		Styles:  template.CSS(cssBytes),
		Content: template.HTML(htmlContent),
	})
	if err != nil {
		return nil, fmt.Errorf("renderizando plantilla: %w", err)
	}
	return out.Bytes(), nil
}
