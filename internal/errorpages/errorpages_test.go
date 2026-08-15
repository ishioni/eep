package errorpages

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/ishioni/eep/templates"
)

func testTemplate(t *testing.T, source []byte) *Template {
	t.Helper()

	parsed, err := newTemplate(source, "test-version", templateDependencies{
		now:      func() time.Time { return time.Unix(1710000000, 0) },
		hostname: func() string { return "test-host" },
	})
	if err != nil {
		t.Fatalf("failed to parse template: %v", err)
	}

	return parsed
}

func TestRenderBundledFormatTemplates(t *testing.T) {
	data := Data{
		StatusCode:   404,
		Message:      "Not Found",
		Description:  "The requested resource was not found.",
		Host:         "example.test",
		OriginalURI:  "/missing",
		ForwardedFor: "192.0.2.10",
		RequestID:    "test-request-id",
		Config: TemplateConfig{
			ShowRequestDetails: true,
			L10nDisabled:       true,
		},
	}

	for name, templateText := range map[string]string{
		"json": templates.JSON,
		"xml":  templates.XML,
		"text": templates.PlainText,
	} {
		t.Run(name, func(t *testing.T) {
			rendered, err := testTemplate(t, []byte(templateText)).Render(data)
			if err != nil {
				t.Fatalf("failed to render template %q: %v", name, err)
			}

			body := string(rendered)
			for _, want := range []string{"404", "Not Found", "1710000000"} {
				if !strings.Contains(body, want) {
					t.Fatalf("rendered template %q does not contain %q", name, want)
				}
			}
			if name == "json" && !json.Valid(rendered) {
				t.Fatalf("rendered JSON template is not valid JSON:\n%s", body)
			}
		})
	}
}

func TestRenderBundledHTMLTemplates(t *testing.T) {
	templateNames, err := templates.GetTemplateNames()
	if err != nil {
		t.Fatalf("failed to list bundled templates: %v", err)
	}

	data := Data{
		StatusCode:   503,
		Message:      "Service Unavailable",
		Description:  "The server is currently unable to handle the request.",
		Host:         "example.test",
		OriginalURI:  "/example",
		ForwardedFor: "192.0.2.10",
		RequestID:    "test-request-id",
		Config: TemplateConfig{
			ShowRequestDetails: true,
			L10nDisabled:       true,
		},
	}

	for _, name := range templateNames {
		t.Run(name, func(t *testing.T) {
			templateBytes, err := templates.GetTemplate(name)
			if err != nil {
				t.Fatalf("failed to load template %q: %v", name, err)
			}

			parsed, err := NewTemplate(templateBytes, "test-version")
			if err != nil {
				t.Fatalf("failed to parse template %q: %v", name, err)
			}
			rendered, err := parsed.Render(data)
			if err != nil {
				t.Fatalf("failed to render template %q: %v", name, err)
			}

			body := string(rendered)
			for _, want := range []string{"503", "Service Unavailable"} {
				if !strings.Contains(body, want) {
					t.Fatalf("rendered template %q does not contain %q", name, want)
				}
			}
		})
	}
}

func TestNewTemplateRejectsInvalidAndLegacySyntax(t *testing.T) {
	if _, err := NewTemplate([]byte("{{"), "test-version"); err == nil {
		t.Fatal("NewTemplate() error = nil, want parse error")
	}
	if _, err := NewTemplate([]byte("{{ code }}"), "test-version"); err == nil {
		t.Fatal("NewTemplate() accepted legacy token syntax")
	}

	parsed, err := NewTemplate([]byte("{{ .Code }}"), "test-version")
	if err != nil {
		t.Fatalf("NewTemplate() error = %v", err)
	}
	if _, err := parsed.Render(Data{}); err == nil {
		t.Fatal("Render() accepted legacy field syntax")
	}
}

func TestToJSONMatchesUpstreamErrorBehavior(t *testing.T) {
	if got := toJSON(make(chan int)); got != "" {
		t.Fatalf("toJSON(unsupported) = %q, want empty string", got)
	}
}

func TestSystemHostname(t *testing.T) {
	if _, err := os.Hostname(); err != nil {
		t.Skipf("hostname unavailable on test host: %v", err)
	}
	if got := systemHostname(); got == "" {
		t.Fatal("systemHostname() returned an empty hostname")
	}
}
