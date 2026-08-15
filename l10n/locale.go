package l10n

import (
	"fmt"
	"slices"
	"strings"
)

const Automatic = "auto"

// Locale is a validated localization mode for one plugin instance.
type Locale struct {
	configured string
	resolved   string
	automatic  bool
	english    bool
}

// Resolve validates and resolves a configured locale. Regional tags fall back to a supported base
// language, while English disables the client-side localization runtime.
func Resolve(configured string) (Locale, error) {
	configured = strings.ToLower(strings.TrimSpace(configured))
	if configured == Automatic {
		return Locale{configured: Automatic, automatic: true}, nil
	}
	if !validConfiguredLocale(configured) {
		return Locale{}, fmt.Errorf("locale %q must be %q, a base language, or a region-qualified language such as fr-CA", configured, Automatic)
	}
	if configured == "en" || strings.HasPrefix(configured, "en-") {
		return Locale{configured: configured, resolved: "en", english: true}, nil
	}
	if supports(configured) {
		return Locale{configured: configured, resolved: configured}, nil
	}
	if base, _, ok := strings.Cut(configured, "-"); ok && supports(base) {
		return Locale{configured: configured, resolved: base}, nil
	}

	return Locale{}, fmt.Errorf("locale %q is not supported (supported: auto, en, %s)", configured, strings.Join(supportedLocaleCodes[:], ", "))
}

// Script returns the embedded localization runtime and, for an explicit locale, a safe override.
func (l Locale) Script() string {
	if l.english {
		return ""
	}
	if l.automatic {
		return ScriptContent()
	}

	return ScriptContent() + "\nwindow.l10n.setLocale(" + fmt.Sprintf("%q", l.resolved) + ");"
}

// Disabled reports whether templates should omit localization entirely.
func (l Locale) Disabled() bool { return l.english }

// String returns the normalized configured locale.
func (l Locale) String() string { return l.configured }

// Supported returns the generated list of translated locale codes.
func Supported() []string { return slices.Clone(supportedLocaleCodes[:]) }

func supports(locale string) bool { return slices.Contains(supportedLocaleCodes[:], locale) }

func validConfiguredLocale(locale string) bool {
	parts := strings.Split(locale, "-")
	if len(parts) < 1 || len(parts) > 2 || len(parts[0]) < 2 || len(parts[0]) > 3 || !asciiLetters(parts[0]) {
		return false
	}
	if len(parts) == 1 {
		return true
	}

	region := parts[1]
	return (len(region) == 2 && asciiLetters(region)) || (len(region) == 3 && asciiDigits(region))
}

func asciiLetters(value string) bool {
	for _, char := range value {
		if char < 'a' || char > 'z' {
			return false
		}
	}
	return true
}

func asciiDigits(value string) bool {
	for _, char := range value {
		if char < '0' || char > '9' {
			return false
		}
	}
	return true
}
