package converter

import (
	"bytes"
	"encoding/base64"
	"fmt"
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

// markdownToHTML converts CommonMark+GFM markdown to HTML, embedding local
// images as base64 data URIs and (depending on flags / inline markers)
// injecting a table of contents.
func markdownToHTML(src []byte, baseDir string, autoTOC bool) ([]byte, error) {
	headings := &headingCollector{}

	md := goldmark.New(
		goldmark.WithExtensions(extension.GFM),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
			parser.WithASTTransformers(
				util.Prioritized(&imageEmbedder{baseDir: baseDir}, 100),
				util.Prioritized(headings, 1000),
			),
		),
		goldmark.WithRendererOptions(html.WithUnsafe()),
	)
	var buf bytes.Buffer
	if err := md.Convert(src, &buf); err != nil {
		return nil, err
	}
	return injectTOC(buf.Bytes(), headings.entries, autoTOC), nil
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
