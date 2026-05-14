package converter

import (
	"bytes"
	"path/filepath"
	"strings"
	"time"

	"github.com/adrg/frontmatter"
)

// extractFrontmatter parses an optional YAML/TOML/JSON front-matter block
// at the top of src into Meta and returns the remaining body bytes.
// If no front-matter block is present, meta is the zero value and body == src.
func extractFrontmatter(src []byte) (Meta, []byte, error) {
	var meta Meta
	rest, err := frontmatter.Parse(bytes.NewReader(src), &meta)
	if err != nil {
		return Meta{}, nil, err
	}
	return meta, rest, nil
}

// fillMetaDefaults applies sensible fallbacks for fields the author may have
// omitted: title from the source filename, date from today.
func fillMetaDefaults(meta Meta, inputPath string) Meta {
	if strings.TrimSpace(meta.Title) == "" {
		meta.Title = titleFromPath(inputPath)
	}
	if strings.TrimSpace(meta.Date) == "" {
		meta.Date = time.Now().Format("2006-01-02")
	}
	return meta
}

func titleFromPath(p string) string {
	base := filepath.Base(p)
	return strings.TrimSuffix(base, filepath.Ext(base))
}
