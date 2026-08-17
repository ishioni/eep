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
	defaultTheme    = "connection"
	defaultLocale   = "auto"
	defaultLogLevel = LogLevelWarn
)

// Config represents eep's plugin configuration.
type Config struct {
	Theme          string            `json:"theme"`
	ShowDetails    bool              `json:"showDetails"`
	Locale         string            `json:"locale"`
	LogLevel       LogLevel          `json:"logLevel"`
	FilterCodes    filtering.Codes   `json:"filterCodes"`
	ExcludeDomains filtering.Domains `json:"excludeDomains"`
}

// Default returns the configuration used when the host does not provide plugin configuration.
func Default() Config {
	return Config{
		Theme:       defaultTheme,
		ShowDetails: false,
		Locale:      defaultLocale,
		LogLevel:    defaultLogLevel,
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
		Theme          json.RawMessage `json:"theme"`
		ShowDetails    json.RawMessage `json:"showDetails"`
		Locale         json.RawMessage `json:"locale"`
		LogLevel       json.RawMessage `json:"logLevel"`
		FilterCodes    json.RawMessage `json:"filterCodes"`
		ExcludeDomains json.RawMessage `json:"excludeDomains"`
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
		theme, err := parseStringField("theme", raw.Theme)
		if err != nil {
			return nil, err
		}
		if theme == "" {
			return nil, fmt.Errorf("plugin configuration field %q must not be empty", "theme")
		}
		cfg.Theme = theme
	}
	if raw.ShowDetails != nil {
		showDetails, err := parseBoolField("showDetails", raw.ShowDetails)
		if err != nil {
			return nil, err
		}
		cfg.ShowDetails = showDetails
	}
	if raw.Locale != nil {
		locale, err := parseStringField("locale", raw.Locale)
		if err != nil {
			return nil, err
		}
		if locale == "" {
			return nil, fmt.Errorf("plugin configuration field %q must not be empty", "locale")
		}
		cfg.Locale = locale
	}
	if raw.LogLevel != nil {
		logLevel, err := parseStringField("logLevel", raw.LogLevel)
		if err != nil {
			return nil, err
		}
		if logLevel == "" {
			return nil, fmt.Errorf("plugin configuration field %q must not be empty", "logLevel")
		}
		level, err := ParseLogLevel(logLevel)
		if err != nil {
			return nil, fmt.Errorf("plugin configuration field %q: %w", "logLevel", err)
		}
		cfg.LogLevel = level
	}
	if raw.FilterCodes != nil {
		if err := json.Unmarshal(raw.FilterCodes, &cfg.FilterCodes); err != nil {
			return nil, fmt.Errorf("plugin configuration field %q: %w", "filterCodes", err)
		}
	}
	if raw.ExcludeDomains != nil {
		if err := json.Unmarshal(raw.ExcludeDomains, &cfg.ExcludeDomains); err != nil {
			return nil, fmt.Errorf("plugin configuration field %q: %w", "excludeDomains", err)
		}
	}

	return &cfg, nil
}

func parseStringField(name string, raw json.RawMessage) (string, error) {
	if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return "", fmt.Errorf("plugin configuration field %q must be a string", name)
	}

	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		return "", fmt.Errorf("plugin configuration field %q must be a string", name)
	}

	return value, nil
}

func parseBoolField(name string, raw json.RawMessage) (bool, error) {
	if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return false, fmt.Errorf("plugin configuration field %q must be a boolean", name)
	}

	var value bool
	if err := json.Unmarshal(raw, &value); err != nil {
		return false, fmt.Errorf("plugin configuration field %q must be a boolean", name)
	}

	return value, nil
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
