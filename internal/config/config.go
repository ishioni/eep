// Copyright 2020-2024 Tetrate
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package config parses the JSON configuration passed to eep by the Proxy-Wasm host.
package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/ishioni/eep/internal/filtering"
)

const (
	defaultTheme  = "connection"
	defaultLocale = "auto"
)

// Config represents eep's plugin configuration.
type Config struct {
	Theme          string            `json:"theme"`
	ShowDetails    bool              `json:"showDetails"`
	Locale         string            `json:"locale"`
	FilterCodes    filtering.Codes   `json:"filterCodes"`
	ExcludeDomains filtering.Domains `json:"excludeDomains"`
}

// Default returns the configuration used when the host does not provide plugin configuration.
func Default() Config {
	return Config{
		Theme:       defaultTheme,
		ShowDetails: false,
		Locale:      defaultLocale,
	}
}

// Parse parses a JSON plugin configuration. Empty configuration uses Default. Unknown fields are
// rejected so policy mistakes do not silently change the resulting error pages.
func Parse(content []byte) (*Config, error) {
	cfg := Default()
	content = bytes.TrimSpace(content)
	if len(content) == 0 {
		return &cfg, nil
	}
	if content[0] != '{' {
		return nil, fmt.Errorf("plugin configuration must be a JSON object")
	}

	var raw struct {
		Theme          *string           `json:"theme"`
		ShowDetails    *bool             `json:"showDetails"`
		Locale         *string           `json:"locale"`
		FilterCodes    filtering.Codes   `json:"filterCodes"`
		ExcludeDomains filtering.Domains `json:"excludeDomains"`
	}

	decoder := json.NewDecoder(bytes.NewReader(content))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&raw); err != nil {
		return nil, fmt.Errorf("decode plugin configuration: %w", err)
	}
	if err := ensureEOF(decoder); err != nil {
		return nil, err
	}

	if raw.Theme != nil {
		if *raw.Theme == "" {
			return nil, fmt.Errorf("plugin configuration field %q must not be empty", "theme")
		}
		cfg.Theme = *raw.Theme
	}
	if raw.ShowDetails != nil {
		cfg.ShowDetails = *raw.ShowDetails
	}
	if raw.Locale != nil {
		if *raw.Locale == "" {
			return nil, fmt.Errorf("plugin configuration field %q must not be empty", "locale")
		}
		cfg.Locale = *raw.Locale
	}
	cfg.FilterCodes = raw.FilterCodes
	cfg.ExcludeDomains = raw.ExcludeDomains

	return &cfg, nil
}

func ensureEOF(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return fmt.Errorf("decode plugin configuration: multiple JSON values")
		}

		return fmt.Errorf("decode plugin configuration: %w", err)
	}

	return nil
}
