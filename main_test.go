package main

import (
	"testing"

	"github.com/ishioni/eep/internal/config"
	"github.com/ishioni/eep/internal/errorpages"
)

func TestPluginContextOwnsRendererAndConfiguration(t *testing.T) {
	firstRenderer, err := errorpages.NewRenderer(map[errorpages.Format][]byte{
		errorpages.HTMLFormat: []byte("first"),
	}, "test-version")
	if err != nil {
		t.Fatalf("create first renderer: %v", err)
	}
	secondRenderer, err := errorpages.NewRenderer(map[errorpages.Format][]byte{
		errorpages.HTMLFormat: []byte("second"),
	}, "test-version")
	if err != nil {
		t.Fatalf("create second renderer: %v", err)
	}

	firstPlugin := &pluginContext{
		config:   config.Config{ShowDetails: true},
		renderer: firstRenderer,
	}
	secondPlugin := &pluginContext{
		config:   config.Config{ShowDetails: false},
		renderer: secondRenderer,
	}

	firstHTTP := firstPlugin.NewHttpContext(1).(*httpContext)
	secondHTTP := secondPlugin.NewHttpContext(2).(*httpContext)

	if firstHTTP.renderer != firstRenderer || !firstHTTP.showDetails {
		t.Fatal("first HTTP context did not inherit its plugin context state")
	}
	if secondHTTP.renderer != secondRenderer || secondHTTP.showDetails {
		t.Fatal("second HTTP context did not inherit its plugin context state")
	}
}
