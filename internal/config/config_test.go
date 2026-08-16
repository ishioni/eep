package config

import (
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name               string
		content            string
		wantTheme          string
		wantShowDetails    bool
		wantLocale         string
		wantFilterCodes    int
		wantExcludeDomains int
		wantErr            string
	}{
		{
			name:       "empty configuration uses defaults",
			wantTheme:  defaultTheme,
			wantLocale: defaultLocale,
		},
		{
			name:       "whitespace configuration uses defaults",
			content:    " \n\t ",
			wantTheme:  defaultTheme,
			wantLocale: defaultLocale,
		},
		{
			name:       "theme override",
			content:    `{"theme":"ghost"}`,
			wantTheme:  "ghost",
			wantLocale: defaultLocale,
		},
		{
			name:            "show details can be enabled",
			content:         `{"showDetails":true}`,
			wantTheme:       defaultTheme,
			wantShowDetails: true,
			wantLocale:      defaultLocale,
		},
		{
			name:               "full configuration",
			content:            `{"theme":"l7","showDetails":true,"locale":"pl","filterCodes":[404,"500-510"],"excludeDomains":["^internal\\.example\\.com$"]}`,
			wantTheme:          "l7",
			wantShowDetails:    true,
			wantLocale:         "pl",
			wantFilterCodes:    2,
			wantExcludeDomains: 1,
		},
		{
			name:       "empty filters preserve defaults",
			content:    `{"filterCodes":[],"excludeDomains":[]}`,
			wantTheme:  defaultTheme,
			wantLocale: defaultLocale,
		},
		{name: "configuration must be an object", content: `null`, wantErr: "must be a JSON object"},
		{name: "invalid JSON", content: `{"theme":`, wantErr: "decode plugin configuration"},
		{name: "unknown field", content: `{"show_details":false}`, wantErr: "unknown field"},
		{name: "empty theme", content: `{"theme":""}`, wantErr: "must not be empty"},
		{name: "empty locale", content: `{"locale":""}`, wantErr: "must not be empty"},
		{name: "invalid filter codes", content: `{"filterCodes":[200]}`, wantErr: "filterCodes[0]"},
		{name: "null filter codes", content: `{"filterCodes":null}`, wantErr: "filterCodes must be an array"},
		{name: "invalid excluded domain", content: `{"excludeDomains":["["]}`, wantErr: "invalid regular expression"},
		{name: "null excluded domains", content: `{"excludeDomains":null}`, wantErr: "excludeDomains must be an array"},
		{name: "multiple JSON values", content: `{} {}`, wantErr: "multiple JSON values"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse([]byte(tt.content))
			if tt.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tt.wantErr) {
					t.Fatalf("Parse() error = %v, want error containing %q", err, tt.wantErr)
				}

				return
			}
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
			if got.Theme != tt.wantTheme || got.ShowDetails != tt.wantShowDetails || got.Locale != tt.wantLocale {
				t.Fatalf("Parse() base configuration = %#v, want theme=%q showDetails=%v locale=%q", got, tt.wantTheme, tt.wantShowDetails, tt.wantLocale)
			}
			if got.FilterCodes.Len() != tt.wantFilterCodes {
				t.Fatalf("FilterCodes.Len() = %d, want %d", got.FilterCodes.Len(), tt.wantFilterCodes)
			}
			if got.ExcludeDomains.Len() != tt.wantExcludeDomains {
				t.Fatalf("ExcludeDomains.Len() = %d, want %d", got.ExcludeDomains.Len(), tt.wantExcludeDomains)
			}
		})
	}
}
