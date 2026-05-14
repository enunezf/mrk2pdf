package converter

import (
	"bytes"
	"html"
	"strings"

	"github.com/yuin/goldmark/ast"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/text"
)

type tocEntry struct {
	Level int
	ID    string
	Text  string
}

// headingCollector is a goldmark AST transformer that records every heading
// in source order, along with its level, auto-generated ID and rendered text.
// Runs after parser.WithAutoHeadingID so IDs are already attached.
type headingCollector struct {
	entries []tocEntry
}

func (c *headingCollector) Transform(doc *ast.Document, reader text.Reader, _ parser.Context) {
	src := reader.Source()
	_ = ast.Walk(doc, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		h, ok := n.(*ast.Heading)
		if !ok {
			return ast.WalkContinue, nil
		}
		idVal, ok := h.AttributeString("id")
		if !ok {
			return ast.WalkContinue, nil
		}
		id, ok := idVal.([]byte)
		if !ok {
			return ast.WalkContinue, nil
		}
		c.entries = append(c.entries, tocEntry{
			Level: h.Level,
			ID:    string(id),
			Text:  collectText(h, src),
		})
		return ast.WalkContinue, nil
	})
}

// collectText concatenates every text leaf within node (handles emphasis,
// strong, code, links inside a heading by walking all descendants).
func collectText(node ast.Node, src []byte) string {
	var sb strings.Builder
	_ = ast.Walk(node, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}
		if t, ok := n.(*ast.Text); ok {
			sb.Write(t.Segment.Value(src))
		}
		return ast.WalkContinue, nil
	})
	return sb.String()
}

// injectTOC handles the two TOC placement strategies:
//   - if the rendered HTML contains <p>[TOC]</p>, replace the marker
//   - otherwise, if autoTOC is set, prepend the TOC
//
// If no headings exist or no marker is present and autoTOC is off, the HTML
// is returned unchanged.
func injectTOC(htmlOut []byte, entries []tocEntry, autoTOC bool) []byte {
	tocHTML := renderTOC(entries)
	marker := []byte("<p>[TOC]</p>")
	if bytes.Contains(htmlOut, marker) {
		return bytes.Replace(htmlOut, marker, []byte(tocHTML), 1)
	}
	if autoTOC && tocHTML != "" {
		return append([]byte(tocHTML), htmlOut...)
	}
	return htmlOut
}

// tocNode is a single TOC entry plus its descendants.
type tocNode struct {
	entry    tocEntry
	children []*tocNode
}

// buildTree groups a flat list of entries by their relative depth.
func buildTree(entries []tocEntry) []*tocNode {
	var roots []*tocNode
	var stack []*tocNode
	for _, e := range entries {
		node := &tocNode{entry: e}
		for len(stack) > 0 && stack[len(stack)-1].entry.Level >= e.Level {
			stack = stack[:len(stack)-1]
		}
		if len(stack) == 0 {
			roots = append(roots, node)
		} else {
			parent := stack[len(stack)-1]
			parent.children = append(parent.children, node)
		}
		stack = append(stack, node)
	}
	return roots
}

func renderTOC(entries []tocEntry) string {
	if len(entries) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString(`<nav class="toc"><h2 class="toc-title">Tabla de contenidos</h2>`)
	renderNodes(&sb, buildTree(entries))
	sb.WriteString(`</nav>`)
	return sb.String()
}

func renderNodes(sb *strings.Builder, nodes []*tocNode) {
	if len(nodes) == 0 {
		return
	}
	sb.WriteString(`<ul>`)
	for _, n := range nodes {
		sb.WriteString(`<li><a href="#`)
		sb.WriteString(n.entry.ID)
		sb.WriteString(`">`)
		sb.WriteString(html.EscapeString(n.entry.Text))
		sb.WriteString(`</a>`)
		renderNodes(sb, n.children)
		sb.WriteString(`</li>`)
	}
	sb.WriteString(`</ul>`)
}
