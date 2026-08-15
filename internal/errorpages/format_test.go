package errorpages

import "testing"

func TestFormatFromAccept(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name   string
		accept string
		want   Format
	}{
		{name: "empty defaults to HTML", want: HTMLFormat},
		{name: "wildcard defaults to HTML", accept: "*/*", want: HTMLFormat},
		{name: "HTML", accept: "text/html", want: HTMLFormat},
		{name: "JSON", accept: "application/json", want: JSONFormat},
		{name: "JSON structured suffix", accept: "application/problem+json", want: JSONFormat},
		{name: "XML", accept: "application/xml", want: XMLFormat},
		{name: "XML structured suffix", accept: "application/xhtml+xml", want: XMLFormat},
		{name: "plain text", accept: "text/plain", want: PlainTextFormat},
		{name: "highest quality supported type wins", accept: "text/html;q=0.5, application/json;q=0.9", want: JSONFormat},
		{name: "equal quality preserves header order", accept: "application/xml;q=0.9, application/json;q=0.9", want: XMLFormat},
		{name: "zero quality is unacceptable", accept: "application/json;q=0, text/plain;q=0.5", want: PlainTextFormat},
		{name: "unsupported higher quality type is ignored", accept: "image/png;q=1, application/json;q=0.5", want: JSONFormat},
		{name: "unsupported defaults to HTML", accept: "image/png", want: HTMLFormat},
		{name: "malformed defaults to HTML", accept: "application/json;q=not-a-number", want: HTMLFormat},
		{name: "case insensitive", accept: "Application/JSON", want: JSONFormat},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := FormatFromAccept(tt.accept); got != tt.want {
				t.Errorf("FormatFromAccept(%q) = %v, want %v", tt.accept, got, tt.want)
			}
		})
	}
}

func TestFormatContentType(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name   string
		format Format
		want   string
	}{
		{name: "plain text", format: PlainTextFormat, want: "text/plain; charset=utf-8"},
		{name: "JSON", format: JSONFormat, want: "application/json; charset=utf-8"},
		{name: "XML", format: XMLFormat, want: "application/xml; charset=utf-8"},
		{name: "HTML", format: HTMLFormat, want: "text/html; charset=utf-8"},
		{name: "unknown", format: Format(255), want: ""},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.format.ContentType(); got != tt.want {
				t.Errorf("ContentType() = %q, want %q", got, tt.want)
			}
		})
	}
}
