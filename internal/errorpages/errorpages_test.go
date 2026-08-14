package errorpages

import (
	"encoding/json"
	"envoy-wasm-error-pages/templates"
	"os"
	"strings"
	"testing"
)

func TestRenderCompatibilitySurface(t *testing.T) {
	if err := os.Setenv("ERRORPAGES_TEST_ENV", "env-value"); err != nil {
		t.Fatalf("failed to set test env: %v", err)
	}
	t.Cleanup(func() { _ = os.Unsetenv("ERRORPAGES_TEST_ENV") })

	const tpl = `
code={{ code }}/{{ .Code }}
message={{ message }}/{{ .Message }}
description={{ description }}/{{ .Description }}
original_uri={{ original_uri }}/{{ .OriginalURI }}
namespace={{ namespace }}/{{ .Namespace }}
ingress_name={{ ingress_name }}/{{ .IngressName }}
service_name={{ service_name }}/{{ .ServiceName }}
service_port={{ service_port }}/{{ .ServicePort }}
request_id={{ request_id }}/{{ .RequestID }}
forwarded_for={{ forwarded_for }}/{{ .ForwardedFor }}
host={{ host }}/{{ .Host }}
show_details={{ show_details }}/{{ .ShowRequestDetails }}
l10n_disabled={{ l10n_disabled }}/{{ .L10nDisabled }}
hide_details={{ hide_details }}
l10n_enabled={{ l10n_enabled }}
nowUnix={{ nowUnix }}
json={{ json message }}
int={{ int "42" }}/{{ int 3.14 }}/{{ int "nope" }}
version={{ version }}
strings={{ strCount "test" "t" }}/{{ strContains "test" "es" }}/{{ strTrimSpace "  ok  " }}/{{ strTrimPrefix "test" "te" }}/{{ strTrimSuffix "test" "st" }}/{{ strReplace "test" "t" "z" }}/{{ strIndex "barfoobaz" "foo" }}/{{ strFields "foo bar" }}
env={{ env "ERRORPAGES_TEST_ENV" }}
escape={{ escape "<tag>" }}
{{ if show_details }}details-on{{ else }}details-off{{ end }}
{{ if l10n_enabled }}l10n-on{{ else }}l10n-off{{ end }}
`

	handler, err := NewWithTemplate([]byte(tpl), "test-version")
	if err != nil {
		t.Fatalf("failed to create handler: %v", err)
	}

	rendered, err := handler.RenderErrorPage(&TemplateData{
		Code:         404,
		Message:      "Not Found",
		Description:  "Missing",
		OriginalURI:  "/missing",
		Namespace:    "default",
		IngressName:  "ingress",
		ServiceName:  "service",
		ServicePort:  "8080",
		RequestID:    "request-id",
		ForwardedFor: "192.0.2.1",
		Host:         "example.test",
		ShowDetails:  true,
		NowUnix:      1710000000,
		L10nDisabled: false,
	})
	if err != nil {
		t.Fatalf("failed to render compatibility template: %v", err)
	}

	body := string(rendered)
	for _, want := range []string{
		"code=404/404",
		"message=Not Found/Not Found",
		"description=Missing/Missing",
		"original_uri=/missing//missing",
		"namespace=default/default",
		"ingress_name=ingress/ingress",
		"service_name=service/service",
		"service_port=8080/8080",
		"request_id=request-id/request-id",
		"forwarded_for=192.0.2.1/192.0.2.1",
		"host=example.test/example.test",
		"show_details=true/true",
		"l10n_disabled=false/false",
		"hide_details=false",
		"l10n_enabled=true",
		"nowUnix=1710000000",
		`json="Not Found"`,
		"int=42/3/0",
		"version=test-version",
		"strings=2/true/ok/st/te/zesz/3/[foo bar]",
		"env=env-value",
		"escape=&lt;tag&gt;",
		"details-on",
		"l10n-on",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("rendered compatibility template does not contain %q\nbody:\n%s", want, body)
		}
	}
}

func TestRenderBundledFormatTemplates(t *testing.T) {
	data := &TemplateData{
		Code:         404,
		Message:      "Not Found",
		Description:  "The requested resource was not found.",
		ShowDetails:  true,
		Host:         "example.test",
		OriginalURI:  "/missing",
		ForwardedFor: "192.0.2.10",
		RequestID:    "test-request-id",
		NowUnix:      1710000000,
	}

	for name, templateText := range map[string]string{
		"json": templates.JSON,
		"xml":  templates.XML,
		"text": templates.PlainText,
	} {
		t.Run(name, func(t *testing.T) {
			handler, err := NewWithTemplate([]byte(templateText), "test-version")
			if err != nil {
				t.Fatalf("failed to create handler for template %q: %v", name, err)
			}

			rendered, err := handler.RenderErrorPage(data)
			if err != nil {
				t.Fatalf("failed to render template %q: %v", name, err)
			}

			body := string(rendered)
			if !strings.Contains(body, "404") {
				t.Fatalf("rendered template %q does not contain status code 404", name)
			}
			if !strings.Contains(body, "Not Found") {
				t.Fatalf("rendered template %q does not contain status message", name)
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

	data := &TemplateData{
		Code:         503,
		Message:      "Service Unavailable",
		Description:  "The server is currently unable to handle the request.",
		ShowDetails:  true,
		Host:         "example.test",
		OriginalURI:  "/example",
		ForwardedFor: "192.0.2.10",
		RequestID:    "test-request-id",
		NowUnix:      1710000000,
	}

	for _, name := range templateNames {
		t.Run(name, func(t *testing.T) {
			templateBytes, err := templates.GetTemplate(name)
			if err != nil {
				t.Fatalf("failed to load template %q: %v", name, err)
			}

			handler, err := NewWithTemplate(templateBytes, "test-version")
			if err != nil {
				t.Fatalf("failed to create handler for template %q: %v", name, err)
			}

			rendered, err := handler.RenderErrorPage(data)
			if err != nil {
				t.Fatalf("failed to render template %q: %v", name, err)
			}

			body := string(rendered)
			if !strings.Contains(body, "503") {
				t.Fatalf("rendered template %q does not contain status code 503", name)
			}
			if !strings.Contains(body, "Service Unavailable") {
				t.Fatalf("rendered template %q does not contain status message", name)
			}
		})
	}
}
