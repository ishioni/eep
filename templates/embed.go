package templates

import (
	"embed"
	"fmt"
	"io/fs"
	"strings"
)

// TemplatesFS contains the copied upstream error-pages template assets.
//
//go:embed default.tpl.json default.tpl.txt default.tpl.xml html/*.tpl.html
var TemplatesFS embed.FS

// JSON holds the embedded JSON template for error responses.
//
//go:embed default.tpl.json
var JSON string

// XML holds the embedded XML template for error responses.
//
//go:embed default.tpl.xml
var XML string

// PlainText holds the embedded plain text template for error responses.
//
//go:embed default.tpl.txt
var PlainText string

// PlaintText keeps compatibility with upstream error-pages' exported typo.
var PlaintText = PlainText

// GetTemplate returns a built-in HTML template by canonical theme name.
func GetTemplate(theme string) ([]byte, error) {
	filename := htmlTemplatePath(theme)

	data, err := TemplatesFS.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("template %q not found: %w", theme, err)
	}

	return data, nil
}

// GetTemplateNames returns all built-in HTML template names.
func GetTemplateNames() ([]string, error) {
	entries, err := fs.ReadDir(TemplatesFS, "html")
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}

		name := e.Name()
		if before, ok := strings.CutSuffix(name, ".tpl.html"); ok {
			names = append(names, before)
		}
	}

	return names, nil
}

func htmlTemplatePath(theme string) string {
	filename := strings.TrimPrefix(theme, "html/")

	switch {
	case strings.HasSuffix(filename, ".tpl.html"):
		return "html/" + filename
	case strings.HasSuffix(filename, ".html"):
		return "html/" + strings.TrimSuffix(filename, ".html") + ".tpl.html"
	default:
		return "html/" + filename + ".tpl.html"
	}
}
