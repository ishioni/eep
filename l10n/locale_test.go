package l10n

import (
	"encoding/json"
	"os"
	"slices"
	"strings"
	"testing"
)

func TestResolve(t *testing.T) {
	for _, test := range []struct {
		name       string
		configured string
		want       string
		disabled   bool
		scriptHas  string
		wantErr    string
	}{
		{name: "automatic", configured: "auto", want: "auto", scriptHas: "Object.defineProperty(window,'l10n'"},
		{name: "normalizes exact locale", configured: " PL ", want: "pl", scriptHas: `window.l10n.setLocale("pl")`},
		{name: "regional fallback", configured: "fr-CA", want: "fr-ca", scriptHas: `window.l10n.setLocale("fr")`},
		{name: "English", configured: "en", want: "en", disabled: true},
		{name: "regional English", configured: "en-US", want: "en-us", disabled: true},
		{name: "numeric pseudo-region", configured: "fr-12", wantErr: "region-qualified"},
		{name: "extended tag", configured: "fr-FR-u-ca-gregory", wantErr: "region-qualified"},
		{name: "empty", wantErr: "region-qualified"},
		{name: "malformed", configured: "not_a_locale", wantErr: "region-qualified"},
		{name: "unsupported", configured: "ja", wantErr: "not supported"},
	} {
		t.Run(test.name, func(t *testing.T) {
			locale, err := Resolve(test.configured)
			if test.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), test.wantErr) {
					t.Fatalf("Resolve() error = %v, want error containing %q", err, test.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("Resolve() error = %v", err)
			}
			if locale.String() != test.want {
				t.Fatalf("Locale.String() = %q, want %q", locale.String(), test.want)
			}
			if locale.Disabled() != test.disabled {
				t.Fatalf("Locale.Disabled() = %v, want %v", locale.Disabled(), test.disabled)
			}
			script := locale.Script()
			if test.disabled && script != "" {
				t.Fatalf("Locale.Script() returned %d bytes for disabled locale", len(script))
			}
			if test.scriptHas != "" && !strings.Contains(script, test.scriptHas) {
				t.Fatalf("Locale.Script() does not contain %q", test.scriptHas)
			}
		})
	}
}

func TestGeneratedSupportedLocalesMatchSource(t *testing.T) {
	content, err := os.ReadFile("locales.json")
	if err != nil {
		t.Fatalf("read locales.json: %v", err)
	}

	var source map[string]json.RawMessage
	if err := json.Unmarshal(content, &source); err != nil {
		t.Fatalf("decode locales.json: %v", err)
	}

	seen := make(map[string]struct{})
	for token, raw := range source {
		if strings.HasPrefix(token, "$") {
			continue
		}
		var translations map[string]string
		if err := json.Unmarshal(raw, &translations); err != nil {
			t.Fatalf("decode translations for %q: %v", token, err)
		}
		for code := range translations {
			seen[strings.ToLower(strings.TrimSpace(code))] = struct{}{}
		}
	}

	generated := Supported()
	for code := range seen {
		if !slices.Contains(generated, code) {
			t.Errorf("generated supported locales do not include %q", code)
		}
	}
	if len(generated) != len(seen) {
		t.Errorf("generated %d supported locales, source contains %d", len(generated), len(seen))
	}
}
