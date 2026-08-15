package config

import (
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    Config
		wantErr string
	}{
		{
			name: "empty configuration uses defaults",
			want: Default(),
		},
		{
			name:    "whitespace configuration uses defaults",
			content: " \n\t ",
			want:    Default(),
		},
		{
			name:    "theme override",
			content: `{"theme":"ghost"}`,
			want: Config{
				Theme:       "ghost",
				ShowDetails: false,
				Locale:      "auto",
			},
		},
		{
			name:    "show details can be enabled",
			content: `{"showDetails":true}`,
			want: Config{
				Theme:       "connection",
				ShowDetails: true,
				Locale:      "auto",
			},
		},
		{
			name:    "show details can be disabled explicitly",
			content: `{"showDetails":false}`,
			want: Config{
				Theme:       "connection",
				ShowDetails: false,
				Locale:      "auto",
			},
		},
		{
			name:    "full configuration",
			content: `{"theme":"l7","showDetails":true,"locale":"pl"}`,
			want: Config{
				Theme:       "l7",
				ShowDetails: true,
				Locale:      "pl",
			},
		},
		{
			name:    "configuration must be an object",
			content: `null`,
			wantErr: "must be a JSON object",
		},
		{
			name:    "invalid JSON",
			content: `{"theme":`,
			wantErr: "decode plugin configuration",
		},
		{
			name:    "unknown field",
			content: `{"show_details":false}`,
			wantErr: "unknown field",
		},
		{
			name:    "empty theme",
			content: `{"theme":""}`,
			wantErr: "must not be empty",
		},
		{
			name:    "empty locale",
			content: `{"locale":""}`,
			wantErr: "must not be empty",
		},
		{
			name:    "multiple JSON values",
			content: `{} {}`,
			wantErr: "multiple JSON values",
		},
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
			if *got != tt.want {
				t.Fatalf("Parse() = %#v, want %#v", *got, tt.want)
			}
		})
	}
}
