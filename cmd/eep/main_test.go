package main

import (
	"testing"

	"github.com/ishioni/eep/internal/config"
	"github.com/ishioni/eep/internal/errorpages"
	"github.com/ishioni/eep/internal/filtering"
)

func TestPluginContextOwnsRendererAndConfiguration(t *testing.T) {
	firstRenderer, err := errorpages.NewRenderer(map[errorpages.Format][]byte{
		errorpages.HTMLFormat: []byte("first"),
	}, errorpages.RendererOptions{Version: "test-version"})
	if err != nil {
		t.Fatalf("create first renderer: %v", err)
	}
	secondRenderer, err := errorpages.NewRenderer(map[errorpages.Format][]byte{
		errorpages.HTMLFormat: []byte("second"),
	}, errorpages.RendererOptions{Version: "test-version"})
	if err != nil {
		t.Fatalf("create second renderer: %v", err)
	}
	secondConfig, err := config.Parse([]byte(`{"filterCodes":[404],"excludeDomains":["^excluded\\.example$"]}`))
	if err != nil {
		t.Fatalf("parse second config: %v", err)
	}

	firstPlugin := &pluginContext{
		config:               config.Config{ShowDetails: true, Locale: "auto"},
		renderer:             firstRenderer,
		localizationDisabled: false,
	}
	secondPlugin := &pluginContext{
		config:               *secondConfig,
		filter:               filtering.NewPolicy(secondConfig.FilterCodes, secondConfig.ExcludeDomains),
		renderer:             secondRenderer,
		localizationDisabled: true,
	}

	firstHTTP := firstPlugin.NewHttpContext(1).(*httpContext)
	secondHTTP := secondPlugin.NewHttpContext(2).(*httpContext)

	if firstHTTP.renderer != firstRenderer || !firstHTTP.showDetails || firstHTTP.localizationDisabled {
		t.Fatal("first HTTP context did not inherit its plugin context state")
	}
	if secondHTTP.renderer != secondRenderer || secondHTTP.showDetails || !secondHTTP.localizationDisabled {
		t.Fatal("second HTTP context did not inherit its plugin context state")
	}
	if !secondHTTP.filter.ShouldHandle(404, "allowed.example") || secondHTTP.filter.ShouldHandle(404, "excluded.example") ||
		secondHTTP.filter.ShouldHandle(500, "allowed.example") {
		t.Fatal("second HTTP context did not inherit its filtering policy")
	}
}
