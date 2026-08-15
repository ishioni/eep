package errorpages

import (
	"encoding/json"
	"fmt"
	"html"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
	"text/template"
	"time"
	"unicode/utf8"
)

type templateDependencies struct {
	now      func() time.Time
	hostname func() string
}

func defaultTemplateDependencies() templateDependencies {
	return templateDependencies{now: time.Now, hostname: systemHostname}
}

func templateFunctions(version string, dependencies templateDependencies) template.FuncMap {
	return template.FuncMap{
		"now":         dependencies.now,
		"hostname":    dependencies.hostname,
		"toJson":      toJSON,
		"toJSON":      toJSON,
		"toInt":       templateInt,
		"int":         templateInt,
		"version":     func() string { return version },
		"env":         getEnv,
		"escape":      html.EscapeString,
		"trim":        strings.TrimSpace,
		"trimPrefix":  trimPrefix,
		"trimSuffix":  trimSuffix,
		"trimPostfix": trimSuffix,
		"replace":     replace,
		"contains":    contains,
		"count":       count,
		"fields":      strings.Fields,
		"lower":       strings.ToLower,
		"upper":       strings.ToUpper,
		"default":     defaultValue,
		"hasPrefix":   hasPrefix,
		"hasSuffix":   hasSuffix,
		"hasPostfix":  hasSuffix,
		"join":        join,
		"split":       split,
		"quote":       strconv.Quote,
		"squote":      singleQuote,
		"repeat":      repeat,
		"substr":      substring,
		"toString":    toString,
		"str":         toString,
		"ternary":     ternary,
		"coalesce":    coalesce,
		"urlEncode":   url.QueryEscape,
		"isEmpty":     isEmpty,
		"isNotEmpty":  func(value any) bool { return !isEmpty(value) },
		"truncate":    truncate,
		"trimAll":     trimAll,
		// Localization remains disabled in eep. The function must exist so upstream templates parse.
		"l10nScript": func() string { return "" },
	}
}

// toJSON intentionally matches error-pages: unsupported values render as an empty string instead
// of failing template execution. Template-visible eep data contains only JSON-compatible values.
func toJSON(value any) string {
	encoded, err := json.Marshal(value)
	if err != nil {
		return ""
	}

	return string(encoded)
}

// systemHostname intentionally matches error-pages by returning an empty string when the host does
// not expose a hostname to the WASM guest.
func systemHostname() string {
	hostname, err := os.Hostname()
	if err != nil {
		return ""
	}

	return hostname
}

func templateInt(value any) int { //nolint:cyclop,funlen
	switch value := value.(type) {
	case int:
		return value
	case int8:
		return int(value)
	case int16:
		return int(value)
	case int32:
		return int(value)
	case int64:
		return int(value)
	case uint:
		return int(value)
	case uint8:
		return int(value)
	case uint16:
		return int(value)
	case uint32:
		return int(value)
	case uint64:
		return int(value) //nolint:gosec
	case float32:
		return int(value)
	case float64:
		return int(value)
	case complex64:
		return int(real(value))
	case complex128:
		return int(real(value))
	case bool:
		if value {
			return 1
		}
		return 0
	case nil:
		return 0
	case string:
		return stringToInt(value)
	case fmt.Stringer:
		return stringToInt(value.String())
	default:
		reflected := reflect.ValueOf(value)
		for reflected.Kind() == reflect.Pointer {
			if reflected.IsNil() {
				return 0
			}
			reflected = reflected.Elem()
		}

		switch reflected.Kind() { //nolint:exhaustive
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return int(reflected.Int())
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return int(reflected.Uint()) //nolint:gosec
		case reflect.Float32, reflect.Float64:
			return int(reflected.Float())
		case reflect.Complex64, reflect.Complex128:
			return int(real(reflected.Complex()))
		case reflect.Bool:
			if reflected.Bool() {
				return 1
			}
			return 0
		case reflect.String:
			return stringToInt(reflected.String())
		default:
			return 0
		}
	}
}

func stringToInt(value string) int {
	value = strings.TrimSpace(value)
	if result, err := strconv.Atoi(value); err == nil {
		return result
	}
	if result, err := strconv.ParseFloat(value, 64); err == nil {
		return int(result)
	}
	return 0
}

func getEnv(key string) string {
	value, ok := os.LookupEnv(key)
	if !ok {
		return ""
	}

	for segment := range strings.SplitSeq(strings.ToUpper(key), "_") {
		switch segment {
		case "PASSWORD", "SECRET", "KEY", "TOKEN", "PASS", "PWD", "CRED":
			return strings.Repeat("*", utf8.RuneCountInString(value))
		}
	}

	return value
}

func trimPrefix(prefix, source string) string { return strings.TrimPrefix(source, prefix) }
func trimSuffix(suffix, source string) string { return strings.TrimSuffix(source, suffix) }
func replace(old, replacement, source string) string {
	return strings.ReplaceAll(source, old, replacement)
}
func contains(substring, source string) bool  { return strings.Contains(source, substring) }
func count(substring, source string) int      { return strings.Count(source, substring) }
func hasPrefix(prefix, source string) bool    { return strings.HasPrefix(source, prefix) }
func hasSuffix(suffix, source string) bool    { return strings.HasSuffix(source, suffix) }
func split(separator, source string) []string { return strings.Split(source, separator) }
func singleQuote(value string) string         { return "'" + value + "'" }
func repeat(count int, value string) string   { return strings.Repeat(value, count) }
func trimAll(cutset, value string) string     { return strings.Trim(value, cutset) }

func defaultValue(fallback any, values ...any) any {
	if isEmpty(values) || isEmpty(values[0]) {
		return fallback
	}
	return values[0]
}

func isEmpty(value any) bool { //nolint:cyclop,funlen
	switch value := value.(type) {
	case nil:
		return true
	case bool:
		return !value
	case string:
		return value == ""
	case int:
		return value == 0
	case int8:
		return value == 0
	case int16:
		return value == 0
	case int32:
		return value == 0
	case int64:
		return value == 0
	case uint:
		return value == 0
	case uint8:
		return value == 0
	case uint16:
		return value == 0
	case uint32:
		return value == 0
	case uint64:
		return value == 0
	case uintptr:
		return value == 0
	case float32:
		return value == 0
	case float64:
		return value == 0
	case complex64:
		return value == 0
	case complex128:
		return value == 0
	default:
		reflected := reflect.ValueOf(value)
		if !reflected.IsValid() {
			return true
		}
		for reflected.Kind() == reflect.Pointer {
			if reflected.IsNil() {
				return true
			}
			reflected = reflected.Elem()
		}

		switch reflected.Kind() { //nolint:exhaustive
		case reflect.Array, reflect.Slice, reflect.Map, reflect.String:
			return reflected.Len() == 0
		case reflect.Bool:
			return !reflected.Bool()
		case reflect.Complex64, reflect.Complex128:
			return reflected.Complex() == 0
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return reflected.Int() == 0
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			return reflected.Uint() == 0
		case reflect.Float32, reflect.Float64:
			return reflected.Float() == 0
		case reflect.Struct:
			return false
		default:
			return reflected.IsNil()
		}
	}
}

func join(separator string, value any) string {
	switch value := value.(type) {
	case []string:
		return strings.Join(value, separator)
	case []any:
		values := make([]string, len(value))
		for index, item := range value {
			values[index] = fmt.Sprint(item)
		}
		return strings.Join(values, separator)
	default:
		reflected := reflect.ValueOf(value)
		if reflected.Kind() != reflect.Slice && reflected.Kind() != reflect.Array {
			return fmt.Sprint(value)
		}
		values := make([]string, reflected.Len())
		for index := range reflected.Len() {
			values[index] = fmt.Sprint(reflected.Index(index).Interface())
		}
		return strings.Join(values, separator)
	}
}

func substring(start, length int, value string) string {
	runes := []rune(value)
	if start < 0 {
		start = 0
	}
	if start >= len(runes) {
		return ""
	}
	if length < 0 {
		return string(runes[start:])
	}
	return string(runes[start:min(start+length, len(runes))])
}

func toString(value any) string { //nolint:cyclop,funlen
	switch value := value.(type) {
	case string:
		return value
	case []byte:
		return string(value)
	case int:
		return strconv.Itoa(value)
	case int8:
		return strconv.FormatInt(int64(value), 10)
	case int16:
		return strconv.FormatInt(int64(value), 10)
	case int32:
		return strconv.FormatInt(int64(value), 10)
	case int64:
		return strconv.FormatInt(value, 10)
	case uint:
		return strconv.FormatUint(uint64(value), 10)
	case uint8:
		return strconv.FormatUint(uint64(value), 10)
	case uint16:
		return strconv.FormatUint(uint64(value), 10)
	case uint32:
		return strconv.FormatUint(uint64(value), 10)
	case uint64:
		return strconv.FormatUint(value, 10)
	case float32:
		return strconv.FormatFloat(float64(value), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(value, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(value)
	case nil:
		return ""
	case fmt.Stringer:
		return value.String()
	default:
		reflected := reflect.ValueOf(value)
		for reflected.Kind() == reflect.Pointer {
			if reflected.IsNil() {
				return ""
			}
			reflected = reflected.Elem()
		}

		switch reflected.Kind() { //nolint:exhaustive
		case reflect.String:
			return reflected.String()
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return strconv.FormatInt(reflected.Int(), 10)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return strconv.FormatUint(reflected.Uint(), 10)
		case reflect.Float32:
			return strconv.FormatFloat(reflected.Float(), 'f', -1, 32)
		case reflect.Float64:
			return strconv.FormatFloat(reflected.Float(), 'f', -1, 64)
		case reflect.Bool:
			return strconv.FormatBool(reflected.Bool())
		case reflect.Slice:
			if reflected.Type().Elem().Kind() == reflect.Uint8 {
				return string(reflected.Bytes())
			}
			return fmt.Sprint(reflected.Interface())
		default:
			return fmt.Sprint(reflected.Interface())
		}
	}
}

func ternary(trueValue, falseValue any, condition bool) any {
	if condition {
		return trueValue
	}
	return falseValue
}

func coalesce(values ...any) any {
	for _, value := range values {
		if !isEmpty(value) {
			return value
		}
	}
	return ""
}

func truncate(length int, value string) string {
	const ellipsis = "..."
	runes := []rune(value)
	if len(runes) <= length {
		return value
	}
	if length < len(ellipsis) {
		return string(runes[:length])
	}
	return string(runes[:length-len(ellipsis)]) + ellipsis
}
