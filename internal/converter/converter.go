package converter

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"html/template"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

type Options struct {
	InputPath   string
	TemplateDir string
}

type pageData struct {
	Title   string
	Styles  template.CSS
	Content template.HTML
}

func Render(opt Options) ([]byte, error) {
	mdSrc, err := os.ReadFile(opt.InputPath)
	if err != nil {
		return nil, fmt.Errorf("leyendo markdown %s: %w", opt.InputPath, err)
	}

	contentHTML, err := markdownToHTML(mdSrc, filepath.Dir(opt.InputPath))
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
		Title:   titleFromPath(opt.InputPath),
		Styles:  template.CSS(cssBytes),
		Content: template.HTML(contentHTML),
	})
	if err != nil {
		return nil, fmt.Errorf("renderizando plantilla: %w", err)
	}
	return out.Bytes(), nil
}

func markdownToHTML(src []byte, baseDir string) ([]byte, error) {
	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
			parser.WithASTTransformers(
				util.Prioritized(&imageEmbedder{baseDir: baseDir}, 100),
			),
		),
		goldmark.WithRendererOptions(html.WithUnsafe()),
	)
	var buf bytes.Buffer
	if err := md.Convert(src, &buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func titleFromPath(p string) string {
	base := filepath.Base(p)
	return strings.TrimSuffix(base, filepath.Ext(base))
}

// imageEmbedder reads local image files referenced from the markdown and
// inlines them as base64 data URIs. This avoids broken images across browsers,
// filesystems (WSL/Windows boundary) and downstream PDF renderers.
type imageEmbedder struct {
	baseDir string
}

func (r *imageEmbedder) Transform(doc *ast.Document, _ text.Reader, _ parser.Context) {
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		img, ok := n.(*ast.Image)
		if !ok {
			return ast.WalkContinue, nil
		}
		dest := string(img.Destination)
		if dest == "" || isAbsoluteRef(dest) {
			return ast.WalkContinue, nil
		}
		path := filepath.Join(r.baseDir, dest)
		data, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "aviso: no se pudo embeber imagen %s: %v\n", path, err)
			return ast.WalkContinue, nil
		}
		mimeType := http.DetectContentType(data)
		img.Destination = []byte("data:" + mimeType + ";base64," + base64.StdEncoding.EncodeToString(data))
		return ast.WalkContinue, nil
	})
}

func isAbsoluteRef(s string) bool {
	if strings.HasPrefix(s, "/") || strings.HasPrefix(s, "data:") {
		return true
	}
	if u, err := url.Parse(s); err == nil && u.IsAbs() {
		return true
	}
	return false
}
