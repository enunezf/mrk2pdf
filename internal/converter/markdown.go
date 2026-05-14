package converter

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
	"github.com/yuin/goldmark/text"
	"github.com/yuin/goldmark/util"
)

// pageBreakRegex matches a line that contains only the literal '\newpage'
// (optionally surrounded by whitespace). The multiline flag (?m) anchors
// ^ and $ to line boundaries.
var pageBreakRegex = regexp.MustCompile(`(?m)^[ \t]*\\newpage[ \t]*$`)

// pageBreakHTML is the HTML emitted in place of a \newpage marker. The
// .pagebreak class is defined in each shipped template's style.css; the
// inline style provides a safety net for custom templates that forget to
// add the rule.
var pageBreakHTML = []byte(`<div class="pagebreak" style="break-after: page; page-break-after: always;"></div>`)

// expandPageBreaks converts standalone '\newpage' lines into the HTML
// page-break div before the markdown reaches goldmark. To keep a literal
// '\newpage' that should NOT trigger a break, escape it as '\\newpage'.
func expandPageBreaks(src []byte) []byte {
	return pageBreakRegex.ReplaceAll(src, pageBreakHTML)
}

// markdownToHTML converts CommonMark+GFM markdown to HTML, embedding local
// images as base64 data URIs and (depending on flags / inline markers)
// injecting a table of contents.
func markdownToHTML(src []byte, baseDir string, autoTOC bool) ([]byte, error) {
	src = expandPageBreaks(src)

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
