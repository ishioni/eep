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

package main

import (
	"errors"

	"github.com/ishioni/eep/internal/config"
	"github.com/ishioni/eep/internal/errorpages"
	"github.com/ishioni/eep/l10n"
	"github.com/ishioni/eep/templates"

	"github.com/proxy-wasm/proxy-wasm-go-sdk/proxywasm"
	"github.com/proxy-wasm/proxy-wasm-go-sdk/proxywasm/types"
)

// version is set at compile time via ldflags
var version = "dev"

func main() {}

func init() {
	proxywasm.SetVMContext(&vmContext{})
}

// vmContext implements types.VMContext.
type vmContext struct {
	types.DefaultVMContext
}

// NewPluginContext implements types.VMContext.
func (*vmContext) NewPluginContext(contextID uint32) types.PluginContext {
	return &pluginContext{}
}

// pluginContext implements types.PluginContext.
type pluginContext struct {
	types.DefaultPluginContext

	config               config.Config
	renderer             *errorpages.Renderer
	localizationDisabled bool
}

// NewHttpContext implements types.PluginContext.
func (ctx *pluginContext) NewHttpContext(contextID uint32) types.HttpContext {
	return &httpContext{
		renderer:             ctx.renderer,
		showDetails:          ctx.config.ShowDetails,
		localizationDisabled: ctx.localizationDisabled,
	}
}

// OnPluginStart implements types.PluginContext.
func (ctx *pluginContext) OnPluginStart(pluginConfigurationSize int) types.OnPluginStartStatus {
	proxywasm.LogInfo("eep initialized (version: " + version + ")")

	pluginConfiguration, err := proxywasm.GetPluginConfiguration()
	if err != nil && !errors.Is(err, types.ErrorStatusNotFound) {
		proxywasm.LogCriticalf("failed to read plugin configuration: %v", err)
		return types.OnPluginStartStatusFailed
	}

	pluginConfig, err := config.Parse(pluginConfiguration)
	if err != nil {
		proxywasm.LogCriticalf("failed to parse JSON plugin configuration: %v", err)
		return types.OnPluginStartStatusFailed
	}

	locale, err := l10n.Resolve(pluginConfig.Locale)
	if err != nil {
		proxywasm.LogCriticalf("invalid locale configuration: %v", err)
		return types.OnPluginStartStatusFailed
	}

	// Validate the configured theme at startup rather than silently serving a different theme.
	templateBytes, err := templates.GetTemplate(pluginConfig.Theme)
	if err != nil {
		proxywasm.LogCriticalf("failed to load configured theme %q: %v", pluginConfig.Theme, err)
		return types.OnPluginStartStatusFailed
	}

	renderer, err := errorpages.NewRenderer(map[errorpages.Format][]byte{
		errorpages.HTMLFormat:      templateBytes,
		errorpages.JSONFormat:      []byte(templates.JSON),
		errorpages.XMLFormat:       []byte(templates.XML),
		errorpages.PlainTextFormat: []byte(templates.PlainText),
	}, errorpages.RendererOptions{
		Version:            version,
		LocalizationScript: locale.Script(),
	})
	if err != nil {
		proxywasm.LogCriticalf("failed to initialize error page renderer: %v", err)
		return types.OnPluginStartStatusFailed
	}

	ctx.config = *pluginConfig
	ctx.renderer = renderer
	ctx.localizationDisabled = locale.Disabled()

	proxywasm.LogInfof(
		"error page templates loaded: theme=%s, show_details=%v, locale=%s",
		ctx.config.Theme,
		ctx.config.ShowDetails,
		locale.String(),
	)
	return types.OnPluginStartStatusOK
}

// httpContext implements types.HttpContext.
type httpContext struct {
	types.DefaultHttpContext

	renderer             *errorpages.Renderer
	showDetails          bool
	localizationDisabled bool
	shouldReplaceBody    bool
	responseFormat       errorpages.Format
	statusCode           uint16
	// Request data for template rendering. Ingress metadata follows the
	// ingress-nginx custom errors header convention used by the upstream
	// error-pages project.
	host         string
	originalURI  string
	namespace    string
	ingressName  string
	serviceName  string
	servicePort  string
	forwardedFor string
	requestID    string
}

// OnHttpRequestHeaders implements types.HttpContext.
func (ctx *httpContext) OnHttpRequestHeaders(numHeaders int, endOfStream bool) types.Action {
	ctx.responseFormat = errorpages.HTMLFormat
	if accept, err := proxywasm.GetHttpRequestHeader("accept"); err == nil {
		ctx.responseFormat = errorpages.FormatFromAccept(accept)
	}

	// Capture request data for error page rendering
	if host, err := proxywasm.GetHttpRequestHeader(":authority"); err == nil {
		ctx.host = host
	} else if host, err := proxywasm.GetHttpRequestHeader("host"); err == nil {
		ctx.host = host
	}

	if originalURI, err := proxywasm.GetHttpRequestHeader("x-original-uri"); err == nil {
		ctx.originalURI = originalURI
	} else if path, err := proxywasm.GetHttpRequestHeader(":path"); err == nil {
		ctx.originalURI = path
	}

	if namespace, err := proxywasm.GetHttpRequestHeader("x-namespace"); err == nil {
		ctx.namespace = namespace
	}

	if ingressName, err := proxywasm.GetHttpRequestHeader("x-ingress-name"); err == nil {
		ctx.ingressName = ingressName
	}

	if serviceName, err := proxywasm.GetHttpRequestHeader("x-service-name"); err == nil {
		ctx.serviceName = serviceName
	}

	if servicePort, err := proxywasm.GetHttpRequestHeader("x-service-port"); err == nil {
		ctx.servicePort = servicePort
	}

	if xff, err := proxywasm.GetHttpRequestHeader("x-forwarded-for"); err == nil {
		ctx.forwardedFor = xff
	}

	if reqID, err := proxywasm.GetHttpRequestHeader("x-request-id"); err == nil {
		ctx.requestID = reqID
	}

	return types.ActionContinue
}

// OnHttpResponseHeaders implements types.HttpContext.
func (ctx *httpContext) OnHttpResponseHeaders(numHeaders int, endOfStream bool) types.Action {
	status, err := proxywasm.GetHttpResponseHeader(":status")
	if err != nil {
		proxywasm.LogWarnf("failed to get status code: %v", err)
		return types.ActionContinue
	}

	proxywasm.LogDebugf("response status code: %s", status)

	// Check if this is a 4xx or 5xx error.
	if statusCode, ok := errorpages.ParseErrorStatus(status); ok {
		if ctx.renderer == nil || !ctx.renderer.HasFormat(ctx.responseFormat) {
			proxywasm.LogErrorf("no error page template for response format %q", ctx.responseFormat.ContentType())
			return types.ActionContinue
		}

		ctx.shouldReplaceBody = true
		ctx.statusCode = statusCode
		proxywasm.LogInfof("intercepting error response: %s as %s", status, ctx.responseFormat.ContentType())

		// Remove headers that could conflict with our custom error page.
		for _, header := range []string{"content-length", "content-encoding", "content-type"} {
			if err := proxywasm.RemoveHttpResponseHeader(header); err != nil {
				proxywasm.LogWarnf("failed to remove response header %q: %v", header, err)
			}
		}

		if err := proxywasm.AddHttpResponseHeader("content-type", ctx.responseFormat.ContentType()); err != nil {
			proxywasm.LogWarnf("failed to set error page content type: %v", err)
		}
	}

	return types.ActionContinue
}

// OnHttpResponseBody implements types.HttpContext.
func (ctx *httpContext) OnHttpResponseBody(bodySize int, endOfStream bool) types.Action {
	if !ctx.shouldReplaceBody {
		return types.ActionContinue
	}

	if !endOfStream {
		// Wait until we see the entire body to replace.
		return types.ActionPause
	}

	templateData := errorpages.Data{
		StatusCode:   ctx.statusCode,
		Host:         ctx.host,
		OriginalURI:  ctx.originalURI,
		Namespace:    ctx.namespace,
		IngressName:  ctx.ingressName,
		ServiceName:  ctx.serviceName,
		ServicePort:  ctx.servicePort,
		ForwardedFor: ctx.forwardedFor,
		RequestID:    ctx.requestID,
		Config: errorpages.TemplateConfig{
			ShowRequestDetails: ctx.showDetails,
			L10nDisabled:       ctx.localizationDisabled,
		},
	}

	if ctx.renderer == nil {
		proxywasm.LogError("error page renderer is not initialized")
		return types.ActionContinue
	}

	errorPage, err := ctx.renderer.Render(ctx.responseFormat, templateData)
	if err != nil {
		proxywasm.LogErrorf("failed to render error page: %v", err)
		return types.ActionContinue
	}

	// Replace the response body with our custom error page
	err = proxywasm.ReplaceHttpResponseBody(errorPage)
	if err != nil {
		proxywasm.LogErrorf("failed to replace response body: %v", err)
		return types.ActionContinue
	}

	proxywasm.LogDebugf("replaced error page for status: %d", ctx.statusCode)
	return types.ActionContinue
}
