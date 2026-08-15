package errorpages

import "fmt"

// Renderer selects and executes an immutable parsed template for each response format.
type Renderer struct {
	templates map[Format]*Template
}

// NewRenderer parses the supplied templates and returns a renderer ready for concurrent use.
func NewRenderer(sources map[Format][]byte, version string) (*Renderer, error) {
	templates := make(map[Format]*Template, len(sources))
	for format, source := range sources {
		parsed, err := NewTemplate(source, version)
		if err != nil {
			return nil, fmt.Errorf("initialize %s template: %w", format.ContentType(), err)
		}
		templates[format] = parsed
	}

	return &Renderer{templates: templates}, nil
}

// HasFormat reports whether the renderer has a template for format.
func (r *Renderer) HasFormat(format Format) bool {
	_, ok := r.templates[format]
	return ok
}

// Render renders data using the template registered for format.
func (r *Renderer) Render(format Format, data Data) ([]byte, error) {
	template, ok := r.templates[format]
	if !ok {
		return nil, fmt.Errorf("no error page template for content type %q", format.ContentType())
	}

	return template.Render(data)
}
