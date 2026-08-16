package filtering

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestCodesUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		allows  []uint16
		denies  []uint16
		wantLen int
		wantErr string
	}{
		{
			name:   "omitted equivalent empty array",
			value:  `[]`,
			allows: []uint16{400, 404, 500, 599},
		},
		{
			name:    "numeric and quoted codes",
			value:   `[404,"503"]`,
			allows:  []uint16{404, 503},
			denies:  []uint16{400, 500, 599},
			wantLen: 2,
		},
		{
			name:    "inclusive range",
			value:   `["500-510"]`,
			allows:  []uint16{500, 505, 510},
			denies:  []uint16{499, 511},
			wantLen: 1,
		},
		{
			name:    "mixed selectors",
			value:   `[404,"500-510",511]`,
			allows:  []uint16{404, 500, 510, 511},
			denies:  []uint16{403, 499, 512},
			wantLen: 3,
		},
		{name: "reject null", value: `null`, wantErr: "must be an array"},
		{name: "reject scalar", value: `404`, wantErr: "must be an array"},
		{name: "reject object", value: `{}`, wantErr: "must be an array"},
		{name: "reject boolean entry", value: `[true]`, wantErr: "must be an HTTP error code"},
		{name: "reject null entry", value: `[null]`, wantErr: "must be a 4xx or 5xx"},
		{name: "reject non-error code", value: `[200]`, wantErr: "must be a 4xx or 5xx"},
		{name: "reject malformed code", value: `["five hundred"]`, wantErr: "must be a 4xx or 5xx"},
		{name: "reject malformed range", value: `["500-other"]`, wantErr: "must be a 4xx or 5xx"},
		{name: "reject reversed range", value: `["510-500"]`, wantErr: "must not be greater"},
		{name: "reject range below errors", value: `["300-404"]`, wantErr: "must be a 4xx or 5xx"},
		{name: "reject range above errors", value: `["500-600"]`, wantErr: "must be a 4xx or 5xx"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got Codes
			err := json.Unmarshal([]byte(tt.value), &got)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("UnmarshalJSON() error = %v, want containing %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("UnmarshalJSON() error = %v", err)
			}
			if got.Len() != tt.wantLen {
				t.Fatalf("Len() = %d, want %d", got.Len(), tt.wantLen)
			}
			for _, code := range tt.allows {
				if !got.Allows(code) {
					t.Errorf("Allows(%d) = false, want true", code)
				}
			}
			for _, code := range tt.denies {
				if got.Allows(code) {
					t.Errorf("Allows(%d) = true, want false", code)
				}
			}
		})
	}
}

func TestCodesRejectNonErrorStatusEvenWhenEmpty(t *testing.T) {
	var codes Codes
	for _, code := range []uint16{0, 200, 399, 600} {
		if codes.Allows(code) {
			t.Errorf("Allows(%d) = true, want false", code)
		}
	}
}

func TestDomainsUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		value   string
		matches []string
		misses  []string
		wantLen int
		wantErr string
	}{
		{name: "empty array excludes nothing", value: `[]`, misses: []string{"example.com"}},
		{name: "missing host never matches", value: `[".*"]`, matches: []string{"example.com"}, misses: []string{""}, wantLen: 1},
		{name: "significant whitespace is preserved", value: `["^ example\\.com $"]`, matches: []string{" example.com "}, misses: []string{"example.com"}, wantLen: 1},
		{
			name:    "exact and wildcard-like regexes",
			value:   `["^admin\\.example\\.com$","(^|\\.)internal\\.example\\.com(:[0-9]+)?$"]`,
			matches: []string{"admin.example.com", "api.internal.example.com", "internal.example.com:8443"},
			misses:  []string{"public.example.com", "admin.example.com:443"},
			wantLen: 2,
		},
		{name: "reject null", value: `null`, wantErr: "must be an array"},
		{name: "reject scalar", value: `"example.com"`, wantErr: "must be an array of strings"},
		{name: "reject mixed array", value: `["example.com",404]`, wantErr: "must be an array of strings"},
		{name: "reject empty expression", value: `[""]`, wantErr: "must not be empty"},
		{name: "reject whitespace expression", value: `["  "]`, wantErr: "must not be empty"},
		{name: "reject invalid regex", value: `["["]`, wantErr: "invalid regular expression"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got Domains
			err := json.Unmarshal([]byte(tt.value), &got)
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("UnmarshalJSON() error = %v, want containing %q", err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("UnmarshalJSON() error = %v", err)
			}
			if got.Len() != tt.wantLen {
				t.Fatalf("Len() = %d, want %d", got.Len(), tt.wantLen)
			}
			for _, host := range tt.matches {
				if !got.Matches(host) {
					t.Errorf("Matches(%q) = false, want true", host)
				}
			}
			for _, host := range tt.misses {
				if got.Matches(host) {
					t.Errorf("Matches(%q) = true, want false", host)
				}
			}
		})
	}
}

func TestPolicyShouldHandle(t *testing.T) {
	var codes Codes
	if err := json.Unmarshal([]byte(`[404,"500-510"]`), &codes); err != nil {
		t.Fatalf("unmarshal codes: %v", err)
	}
	var domains Domains
	if err := json.Unmarshal([]byte(`["(^|\\.)example\\.com$"]`), &domains); err != nil {
		t.Fatalf("unmarshal domains: %v", err)
	}
	policy := NewPolicy(codes, domains)

	tests := []struct {
		name string
		code uint16
		host string
		want bool
	}{
		{name: "selected code and allowed host", code: 404, host: "public.test", want: true},
		{name: "selected range and allowed host", code: 505, host: "public.test", want: true},
		{name: "unselected code", code: 403, host: "public.test", want: false},
		{name: "excluded domain", code: 404, host: "example.com", want: false},
		{name: "excluded subdomain", code: 500, host: "api.example.com", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := policy.ShouldHandle(tt.code, tt.host); got != tt.want {
				t.Fatalf("ShouldHandle(%d, %q) = %v, want %v", tt.code, tt.host, got, tt.want)
			}
		})
	}
}
