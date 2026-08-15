package errorpages

import (
	"strings"
	"testing"
)

func TestRenderer(t *testing.T) {
	renderer, err := NewRenderer(map[Format][]byte{
		HTMLFormat: []byte(`{{ .StatusCode }} {{ .Message }} {{ l10nScript }}`),
	}, RendererOptions{Version: "test-version", LocalizationScript: "localization-script"})
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
	if got := string(rendered); got != "404 Not Found localization-script" {
		t.Fatalf("Render() = %q, want %q", got, "404 Not Found localization-script")
	}

	if _, err := renderer.Render(JSONFormat, Data{StatusCode: 404}); err == nil ||
		!strings.Contains(err.Error(), JSONFormat.ContentType()) {
		t.Fatalf("Render(missing format) error = %v", err)
	}
}

func TestRendererLocalization(t *testing.T) {
	renderer, err := NewRenderer(map[Format][]byte{
		HTMLFormat: []byte(`{{ if not .Config.L10nDisabled }}{{ l10nScript }}{{ end }}`),
	}, RendererOptions{LocalizationScript: "localization-script"})
	if err != nil {
		t.Fatalf("NewRenderer() error = %v", err)
	}

	enabled, err := renderer.Render(HTMLFormat, Data{})
	if err != nil {
		t.Fatalf("Render(enabled) error = %v", err)
	}
	if string(enabled) != "localization-script" {
		t.Fatalf("Render(enabled) = %q", enabled)
	}

	disabled, err := renderer.Render(HTMLFormat, Data{Config: TemplateConfig{L10nDisabled: true}})
	if err != nil {
		t.Fatalf("Render(disabled) error = %v", err)
	}
	if len(disabled) != 0 {
		t.Fatalf("Render(disabled) = %q, want empty", disabled)
	}
}

func TestNewRendererRejectsInvalidTemplate(t *testing.T) {
	if _, err := NewRenderer(
		map[Format][]byte{HTMLFormat: []byte("{{")},
		RendererOptions{Version: "test-version"},
	); err == nil {
		t.Fatal("NewRenderer() error = nil, want parse error")
	}
}
