package config

import (
	"fmt"
	"strings"
)

// LogLevel controls which eep log messages are sent to Envoy's logger.
type LogLevel uint8

const (
	// LogLevelOff disables eep's normal log messages.
	LogLevelOff LogLevel = iota
	// LogLevelDebug emits debug, info, warning, error, and critical messages.
	LogLevelDebug
	// LogLevelInfo emits info, warning, error, and critical messages.
	LogLevelInfo
	// LogLevelWarn emits warning, error, and critical messages.
	LogLevelWarn
	// LogLevelError emits error and critical messages.
	LogLevelError
	// LogLevelCritical emits only critical messages.
	LogLevelCritical
)

// ParseLogLevel parses a configured log level.
func ParseLogLevel(value string) (LogLevel, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "off":
		return LogLevelOff, nil
	case "debug":
		return LogLevelDebug, nil
	case "info":
		return LogLevelInfo, nil
	case "warn", "warning":
		return LogLevelWarn, nil
	case "error":
		return LogLevelError, nil
	case "critical":
		return LogLevelCritical, nil
	default:
		return 0, fmt.Errorf("unsupported log level %q", value)
	}
}

// String returns the canonical configuration value for the log level.
func (level LogLevel) String() string {
	switch level {
	case LogLevelOff:
		return "off"
	case LogLevelDebug:
		return "debug"
	case LogLevelInfo:
		return "info"
	case LogLevelWarn:
		return "warn"
	case LogLevelError:
		return "error"
	case LogLevelCritical:
		return "critical"
	default:
		return "unknown"
	}
}

// Allows reports whether a message at messageLevel should be emitted.
func (level LogLevel) Allows(messageLevel LogLevel) bool {
	return level != LogLevelOff && messageLevel >= level
}
