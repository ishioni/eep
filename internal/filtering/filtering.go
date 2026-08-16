package filtering

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Codes is a validated set of individual HTTP error codes and inclusive ranges. Its zero value
// accepts every 4xx and 5xx status.
type Codes struct {
	ranges []codeRange
}

// UnmarshalJSON accepts an array containing numeric codes, quoted codes, and quoted inclusive
// ranges such as [404, "500-510"].
func (codes *Codes) UnmarshalJSON(data []byte) error {
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		return fmt.Errorf("filterCodes must be an array")
	}

	var values []json.RawMessage
	if err := json.Unmarshal(data, &values); err != nil {
		return fmt.Errorf("filterCodes must be an array: %w", err)
	}

	ranges := make([]codeRange, 0, len(values))
	for index, value := range values {
		parsed, err := parseCodeRange(value)
		if err != nil {
			return fmt.Errorf("filterCodes[%d]: %w", index, err)
		}

		ranges = append(ranges, parsed)
	}

	codes.ranges = ranges
	return nil
}

// Len returns the number of configured code selectors. A range counts as one selector.
func (codes Codes) Len() int {
	return len(codes.ranges)
}

// Allows reports whether code is selected. An empty set allows every 4xx and 5xx status.
func (codes Codes) Allows(code uint16) bool {
	if code < 400 || code > 599 {
		return false
	}
	if len(codes.ranges) == 0 {
		return true
	}

	for _, candidate := range codes.ranges {
		if code >= candidate.start && code <= candidate.end {
			return true
		}
	}

	return false
}

// Domains is a validated set of exclusionary Go/RE2 regular expressions. Its zero value excludes
// no domains.
type Domains struct {
	patterns []*regexp.Regexp
}

// UnmarshalJSON accepts an array of non-empty regular-expression strings.
func (domains *Domains) UnmarshalJSON(data []byte) error {
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		return fmt.Errorf("excludeDomains must be an array")
	}

	var values []string
	if err := json.Unmarshal(data, &values); err != nil {
		return fmt.Errorf("excludeDomains must be an array of strings: %w", err)
	}

	patterns := make([]*regexp.Regexp, 0, len(values))
	for index, value := range values {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("excludeDomains[%d] must not be empty", index)
		}

		pattern, err := regexp.Compile(value)
		if err != nil {
			return fmt.Errorf("excludeDomains[%d]: invalid regular expression: %w", index, err)
		}

		patterns = append(patterns, pattern)
	}

	domains.patterns = patterns
	return nil
}

// Len returns the number of configured domain-exclusion patterns.
func (domains Domains) Len() int {
	return len(domains.patterns)
}

// Matches reports whether host matches any configured exclusion pattern.
func (domains Domains) Matches(host string) bool {
	if host == "" {
		return false
	}

	for _, pattern := range domains.patterns {
		if pattern.MatchString(host) {
			return true
		}
	}

	return false
}

// Policy combines inclusive status-code filtering with exclusionary domain filtering.
type Policy struct {
	codes          Codes
	excludeDomains Domains
}

// NewPolicy returns an immutable response-filtering policy.
func NewPolicy(codes Codes, excludeDomains Domains) Policy {
	return Policy{codes: codes, excludeDomains: excludeDomains}
}

// ShouldHandle reports whether eep should replace the error response for code and host.
func (policy Policy) ShouldHandle(code uint16, host string) bool {
	return policy.codes.Allows(code) && !policy.excludeDomains.Matches(host)
}

type codeRange struct {
	start uint16
	end   uint16
}

func parseCodeRange(raw json.RawMessage) (codeRange, error) {
	raw = bytes.TrimSpace(raw)

	var value string
	if len(raw) > 0 && raw[0] == '"' {
		if err := json.Unmarshal(raw, &value); err != nil {
			return codeRange{}, fmt.Errorf("must be an HTTP error code or inclusive range: %w", err)
		}
	} else {
		var code int
		if err := json.Unmarshal(raw, &code); err != nil {
			return codeRange{}, fmt.Errorf("must be an HTTP error code or inclusive range")
		}
		value = strconv.Itoa(code)
	}

	startValue, endValue, isRange := strings.Cut(strings.TrimSpace(value), "-")
	start, err := parseErrorCode(startValue)
	if err != nil {
		return codeRange{}, err
	}
	if !isRange {
		return codeRange{start: start, end: start}, nil
	}

	end, err := parseErrorCode(endValue)
	if err != nil {
		return codeRange{}, err
	}
	if start > end {
		return codeRange{}, fmt.Errorf("range start %d must not be greater than range end %d", start, end)
	}

	return codeRange{start: start, end: end}, nil
}

func parseErrorCode(value string) (uint16, error) {
	code, err := strconv.ParseUint(strings.TrimSpace(value), 10, 16)
	if err != nil || code < 400 || code > 599 {
		return 0, fmt.Errorf("must be a 4xx or 5xx HTTP status code")
	}

	return uint16(code), nil
}
