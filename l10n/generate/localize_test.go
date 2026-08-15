package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWriteJSFilesEscapesInlineScriptTerminators(t *testing.T) {
	directory := t.TempDir()
	jsPath := filepath.Join(directory, "localize.js")
	minPath := filepath.Join(directory, "localize.min.js")

	_, err := writeJSFiles(localesData{
		"unsafe": {"fr": `</script><script>alert("x")</script>`},
	}, jsPath, minPath)
	if err != nil {
		t.Fatalf("writeJSFiles() error = %v", err)
	}

	for _, path := range []string{jsPath, minPath} {
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read generated file: %v", err)
		}
		if strings.Contains(strings.ToLower(string(content)), "</script") {
			t.Fatalf("generated JavaScript %q contains an inline script terminator", path)
		}
		if !strings.Contains(string(content), `\u003c/script\u003e`) {
			t.Fatalf("generated JavaScript %q does not contain escaped markup", path)
		}
	}
}
