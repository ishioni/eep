package errorpages

import (
	"bytes"
	"fmt"
	"text/template"
)

// Template is a parsed error page template that is safe for concurrent execution.
type Template struct {
	template *template.Template
}

// NewTemplate parses source with the current error-pages data and function surface.
func NewTemplate(source []byte, version string) (*Template, error) {
	return newTemplate(source, version, defaultTemplateDependencies())
}

func newTemplate(source []byte, version string, dependencies templateDependencies) (*Template, error) {
	parsed, err := template.New("errorpage").
		Funcs(templateFunctions(version, dependencies)).
		Parse(string(source))
	if err != nil {
		return nil, fmt.Errorf("parse error page template: %w", err)
	}

	return &Template{template: parsed}, nil
}

// Render executes the template with an upstream-compatible Data value.
func (t *Template) Render(data Data) ([]byte, error) {
	data = data.withDefaults()

	var output bytes.Buffer
	if err := t.template.Execute(&output, data); err != nil {
		return nil, fmt.Errorf("execute error page template: %w", err)
	}

	return output.Bytes(), nil
}
