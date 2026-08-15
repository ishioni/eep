package errorpages

import (
	"mime"
	"strconv"
	"strings"
)

// Format identifies the representation used for a rendered error response.
type Format uint8

const (
	// PlainTextFormat renders a text/plain response.
	PlainTextFormat Format = iota
	// JSONFormat renders an application/json response.
	JSONFormat
	// XMLFormat renders an application/xml response.
	XMLFormat
	// HTMLFormat renders a text/html response.
	HTMLFormat
)

// ContentType returns the HTTP Content-Type for f, including UTF-8 encoding.
func (f Format) ContentType() string {
	switch f {
	case PlainTextFormat:
		return "text/plain; charset=utf-8"
	case JSONFormat:
		return "application/json; charset=utf-8"
	case XMLFormat:
		return "application/xml; charset=utf-8"
	case HTMLFormat:
		return "text/html; charset=utf-8"
	default:
		return ""
	}
}

// FormatFromAccept chooses a supported response representation from an Accept
// header. HTML is the fallback for an empty, wildcard-only, malformed, or
// unsupported header. Explicit supported media types are ordered by q-value;
// ties preserve their order in the header.
func FormatFromAccept(accept string) Format {
	bestFormat := HTMLFormat
	bestWeight := -1.0

	for _, value := range strings.Split(accept, ",") {
		mediaType, params, err := mime.ParseMediaType(strings.TrimSpace(value))
		if err != nil {
			continue
		}

		format, ok := formatFromMediaType(mediaType)
		if !ok {
			continue
		}

		weight, ok := quality(params)
		if !ok || weight <= 0 || weight <= bestWeight {
			continue
		}

		bestFormat = format
		bestWeight = weight
	}

	return bestFormat
}

func quality(params map[string]string) (float64, bool) {
	q, ok := params["q"]
	if !ok {
		return 1, true
	}

	weight, err := strconv.ParseFloat(q, 64)
	if err != nil || weight < 0 || weight > 1 {
		return 0, false
	}

	return weight, true
}

func formatFromMediaType(mediaType string) (Format, bool) {
	mediaType = strings.ToLower(mediaType)

	switch {
	case mediaType == "text/plain":
		return PlainTextFormat, true
	case mediaType == "text/html":
		return HTMLFormat, true
	case mediaType == "application/json", mediaType == "text/json", strings.HasSuffix(mediaType, "+json"):
		return JSONFormat, true
	case mediaType == "application/xml", mediaType == "text/xml", strings.HasSuffix(mediaType, "+xml"):
		return XMLFormat, true
	default:
		return HTMLFormat, false
	}
}
