package config

import "testing"

func TestParseLogLevel(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  LogLevel
	}{
		{name: "off", value: "off", want: LogLevelOff},
		{name: "debug", value: "debug", want: LogLevelDebug},
		{name: "info", value: "info", want: LogLevelInfo},
		{name: "warn", value: "warn", want: LogLevelWarn},
		{name: "warning alias", value: "WARNING", want: LogLevelWarn},
		{name: "error", value: "error", want: LogLevelError},
		{name: "critical", value: "critical", want: LogLevelCritical},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseLogLevel(tt.value)
			if err != nil {
				t.Fatalf("ParseLogLevel() error = %v", err)
			}
			if got != tt.want {
				t.Fatalf("ParseLogLevel() = %q, want %q", got.String(), tt.want.String())
			}
		})
	}
}

func TestLogLevelAllows(t *testing.T) {
	tests := []struct {
		threshold LogLevel
		message   LogLevel
		want      bool
	}{
		{threshold: LogLevelOff, message: LogLevelCritical, want: false},
		{threshold: LogLevelDebug, message: LogLevelDebug, want: true},
		{threshold: LogLevelInfo, message: LogLevelDebug, want: false},
		{threshold: LogLevelWarn, message: LogLevelInfo, want: false},
		{threshold: LogLevelWarn, message: LogLevelWarn, want: true},
		{threshold: LogLevelError, message: LogLevelCritical, want: true},
		{threshold: LogLevelCritical, message: LogLevelError, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.threshold.String()+"/"+tt.message.String(), func(t *testing.T) {
			if got := tt.threshold.Allows(tt.message); got != tt.want {
				t.Fatalf("Allows() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLogLevelString(t *testing.T) {
	if got := LogLevel(255).String(); got != "unknown" {
		t.Fatalf("String() = %q, want unknown", got)
	}
}
