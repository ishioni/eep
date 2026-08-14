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

package errorpages

import (
	"encoding/json"
	"fmt"
	"html"
	"os"
	"reflect"
	"strconv"
	"strings"
	"text/template"
	"time"
)

// TemplateData holds all the data that can be used in error page templates
type TemplateData struct {
	// StatusCode is the upstream error-pages field used by the current bundled templates.
	StatusCode int
	// Code is the legacy local field kept for compatibility with older local/custom templates.
	Code int `token:"code"`

	Message     string `token:"message"`
	Description string `token:"description"`

	OriginalURI  string `token:"original_uri"`
	Namespace    string `token:"namespace"`
	IngressName  string `token:"ingress_name"`
	ServiceName  string `token:"service_name"`
	ServicePort  string `token:"service_port"`
	RequestID    string `token:"request_id"`
	ForwardedFor string `token:"forwarded_for"`
	Host         string `token:"host"`

	HomepageURL string
	Links       []TemplateLink
	Config      TemplateConfig

	ShowDetails        bool `token:"show_details"`
	ShowRequestDetails bool
	L10nDisabled       bool  `token:"l10n_disabled"`
	NowUnix            int64 // registered as builtin function
	L10nEnabled        bool  // deprecated compatibility field; l10n_enabled is derived from L10nDisabled
	L10nScript         string
}

// TemplateLink represents an additional link shown by upstream error-pages templates.
type TemplateLink struct {
	Label string
	URL   string
}

// TemplateConfig holds template-facing configuration used by upstream error-pages templates.
type TemplateConfig struct {
	ShowRequestDetails bool
	L10nDisabled       bool
}

// Values converts TemplateData fields into a map keyed by their token tags,
// suitable for registering as template functions.
func (d *TemplateData) Values() map[string]any {
	result := make(map[string]any)
	v := reflect.ValueOf(*d)
	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		if token, ok := t.Field(i).Tag.Lookup("token"); ok {
			result[token] = v.Field(i).Interface()
		}
	}
	return result
}

// Handler manages error page templates and detection
type Handler struct {
	templateText string // preprocessed template content
	version      string
}

// NewWithTemplate creates a handler that uses a Go template for error pages.
func NewWithTemplate(templateBytes []byte, version string) (*Handler, error) {
	return &Handler{
		templateText: string(templateBytes),
		version:      version,
	}, nil
}

// IsErrorStatus checks if a status code is in the 4xx or 5xx range
func IsErrorStatus(status string) bool {
	if len(status) != 3 {
		return false
	}
	return status[0] == '4' || status[0] == '5'
}

// RenderErrorPage renders the template with the provided data
func (h *Handler) RenderErrorPage(data *TemplateData) ([]byte, error) {
	if data.StatusCode == 0 {
		data.StatusCode = data.Code
	}
	if data.Code == 0 {
		data.Code = data.StatusCode
	}
	if data.NowUnix == 0 {
		data.NowUnix = time.Now().Unix()
	}
	if data.Message == "" {
		data.Message = getStatusMessage(data.StatusCode)
	}
	if data.Description == "" {
		data.Description = getStatusDescription(data.StatusCode)
	}

	data.ShowRequestDetails = data.ShowDetails
	data.Config.ShowRequestDetails = data.ShowRequestDetails
	data.Config.L10nDisabled = data.L10nDisabled

	toJSON := func(v any) string { b, _ := json.Marshal(v); return string(b) }
	fns := template.FuncMap{
		"escape":        html.EscapeString,
		"now":           func() time.Time { return time.Unix(data.NowUnix, 0) },
		"nowUnix":       func() int64 { return data.NowUnix },
		"hostname":      func() string { h, _ := os.Hostname(); return h },
		"json":          toJSON,
		"toJson":        toJSON,
		"toJSON":        toJSON,
		"int":           templateInt,
		"toInt":         templateInt,
		"version":       func() string { return h.version },
		"strCount":      strings.Count,
		"strContains":   strings.Contains,
		"strTrimSpace":  strings.TrimSpace,
		"strTrimPrefix": strings.TrimPrefix,
		"strTrimSuffix": strings.TrimSuffix,
		"strReplace":    strings.ReplaceAll,
		"strIndex":      strings.Index,
		"strFields":     strings.Fields,
		"env":           os.Getenv,
		"l10nScript":    func() string { return data.L10nScript },
		"hide_details":  func() bool { return !data.ShowDetails },
		"l10n_enabled":  func() bool { return !data.L10nDisabled },
	}

	for k, v := range data.Values() {
		val := v
		fns[k] = func() any { return val }
	}

	tmpl, err := template.New("errorpage").Funcs(fns).Parse(h.templateText)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template: %w", err)
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("failed to execute template: %w", err)
	}

	return []byte(buf.String()), nil
}

func templateInt(v any) int {
	switch v := v.(type) {
	case string:
		if i, err := strconv.Atoi(strings.TrimSpace(v)); err == nil {
			return i
		}
	case int:
		return v
	case int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		if i, err := strconv.Atoi(fmt.Sprintf("%d", v)); err == nil {
			return i
		}
	case float32, float64:
		if f, err := strconv.ParseFloat(fmt.Sprintf("%f", v), 64); err == nil {
			return int(f)
		}
	case fmt.Stringer:
		if i, err := strconv.Atoi(v.String()); err == nil {
			return i
		}
	}
	return 0
}

// getStatusMessage returns the standard HTTP status message for a code
func getStatusMessage(code int) string {
	messages := map[int]string{
		// 4xx Client Errors
		400: "Bad Request",
		401: "Unauthorized",
		402: "Payment Required",
		403: "Forbidden",
		404: "Not Found",
		405: "Method Not Allowed",
		406: "Not Acceptable",
		407: "Proxy Authentication Required",
		408: "Request Timeout",
		409: "Conflict",
		410: "Gone",
		411: "Length Required",
		412: "Precondition Failed",
		413: "Payload Too Large",
		414: "URI Too Long",
		415: "Unsupported Media Type",
		416: "Range Not Satisfiable",
		417: "Expectation Failed",
		418: "I'm a teapot",
		421: "Misdirected Request",
		422: "Unprocessable Entity",
		423: "Locked",
		424: "Failed Dependency",
		425: "Too Early",
		426: "Upgrade Required",
		428: "Precondition Required",
		429: "Too Many Requests",
		431: "Request Header Fields Too Large",
		451: "Unavailable For Legal Reasons",

		// 5xx Server Errors
		500: "Internal Server Error",
		501: "Not Implemented",
		502: "Bad Gateway",
		503: "Service Unavailable",
		504: "Gateway Timeout",
		505: "HTTP Version Not Supported",
		506: "Variant Also Negotiates",
		507: "Insufficient Storage",
		508: "Loop Detected",
		510: "Not Extended",
		511: "Network Authentication Required",
	}

	if msg, ok := messages[code]; ok {
		return msg
	}

	if code >= 400 && code < 500 {
		return "Client Error"
	}
	return "Server Error"
}

// getStatusDescription returns a description for common HTTP status codes
func getStatusDescription(code int) string {
	descriptions := map[int]string{
		400: "The request could not be understood by the server due to malformed syntax.",
		401: "The request requires user authentication.",
		403: "The server understood the request, but is refusing to fulfill it.",
		404: "The requested resource could not be found.",
		405: "The method specified in the request is not allowed for the resource.",
		408: "The server timed out waiting for the request.",
		429: "Too many requests have been sent in a given amount of time.",
		500: "The server encountered an unexpected condition that prevented it from fulfilling the request.",
		502: "The server received an invalid response from the upstream server.",
		503: "The server is currently unable to handle the request due to temporary overloading or maintenance.",
		504: "The server did not receive a timely response from the upstream server.",
	}

	if desc, ok := descriptions[code]; ok {
		return desc
	}

	if code >= 400 && code < 500 {
		return "An error occurred while processing your request."
	}
	return "The server encountered an error while processing your request."
}
