package errorpages

import "fmt"

// Renderer selects and executes an immutable parsed template for each response format.
type Renderer struct {
	templates map[Format]*Template
}

// RendererOptions contains plugin-instance values exposed through template functions.
type RendererOptions struct {
	Version            string
	LocalizationScript string
}

// NewRenderer parses the supplied templates and returns a renderer ready for concurrent use.
func NewRenderer(sources map[Format][]byte, options RendererOptions) (*Renderer, error) {
	dependencies := defaultTemplateDependencies()
	dependencies.localizationScript = options.LocalizationScript

	templates := make(map[Format]*Template, len(sources))
	for format, source := range sources {
		parsed, err := newTemplate(source, options.Version, dependencies)
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
