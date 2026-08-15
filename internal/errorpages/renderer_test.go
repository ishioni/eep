package errorpages

import (
	"strings"
	"testing"
)

func TestRenderer(t *testing.T) {
	renderer, err := NewRenderer(map[Format][]byte{
		HTMLFormat: []byte(`{{ .StatusCode }} {{ .Message }}`),
	}, "test-version")
	if err != nil {
		t.Fatalf("NewRenderer() error = %v", err)
	}

	if !renderer.HasFormat(HTMLFormat) {
		t.Fatal("HasFormat(HTMLFormat) = false, want true")
	}
	if renderer.HasFormat(JSONFormat) {
		t.Fatal("HasFormat(JSONFormat) = true, want false")
	}

	rendered, err := renderer.Render(HTMLFormat, Data{StatusCode: 404})
	if err != nil {
		t.Fatalf("Render() error = %v", err)
	}
	if got := string(rendered); got != "404 Not Found" {
		t.Fatalf("Render() = %q, want %q", got, "404 Not Found")
	}

	if _, err := renderer.Render(JSONFormat, Data{StatusCode: 404}); err == nil ||
		!strings.Contains(err.Error(), JSONFormat.ContentType()) {
		t.Fatalf("Render(missing format) error = %v", err)
	}
}

func TestNewRendererRejectsInvalidTemplate(t *testing.T) {
	if _, err := NewRenderer(map[Format][]byte{HTMLFormat: []byte("{{")}, "test-version"); err == nil {
		t.Fatal("NewRenderer() error = nil, want parse error")
	}
}
