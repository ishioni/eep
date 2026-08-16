# eep

**eep** (Envoy Error Pages) is an Envoy Proxy-Wasm extension written in Go that intercepts backend error responses (4xx and 5xx status codes) and replaces them with custom HTML, JSON, XML, or plain-text error pages.

## Features

- **Automatic Error Interception**: Detects and handles all 4xx and 5xx HTTP status codes
- **Content Negotiation**: Responds with HTML, JSON, XML, or plain text according to the request's `Accept` header
- **Custom Error Pages**: Provides error pages for all client errors (4xx) and server errors (5xx)
- **Template-Based Design**: Built-in response templates are stored in separate files for easy customization without Go knowledge
- **Runtime Configuration**: Select the HTML theme, request-detail visibility, and locale through the host's plugin configuration
- **Localization**: Localize HTML pages automatically from browser preferences or force a configured locale
- **Version Tracking**: Automatically embeds the git commit SHA into the plugin for easy version identification
- **Lightweight**: Compiled to WASM for minimal overhead

## Prerequisites

- [mise](https://mise.jdx.dev/)
- Docker with Compose v2
- Git

Install the pinned Go and project tools with:

```bash
mise install
```

Tool versions and development tasks are defined in `.mise/config.toml` and locked in
`.mise/mise.lock`.

## Building

```bash
# Build main.wasm
mise run build

# Run Go tests with the race detector
mise run test

# Build the scratch OCI artifact
mise run docker-build

# List all available tasks
mise tasks
```

The release artifact is `main.wasm`, produced directly by `mise run build`.

## Runtime compatibility

Eep requires Envoy Proxy `v1.39.0` or later. It also supports Envoy Gateway `v1.9.0` or later
when using Gateway's default Envoy image, which is Envoy `v1.39.0`. If you override the Gateway
Envoy image, use an Envoy `v1.39.0+` image.

## Response formats

Eep chooses an error representation from the request's `Accept` header and preserves the original
4xx or 5xx status code. The supported media types are:

| `Accept` media type                                             | Response `Content-Type`           |
| --------------------------------------------------------------- | --------------------------------- |
| `text/html`                                                     | `text/html; charset=utf-8`        |
| `application/json`, `text/json`, or a `+json` structured suffix | `application/json; charset=utf-8` |
| `application/xml`, `text/xml`, or a `+xml` structured suffix    | `application/xml; charset=utf-8`  |
| `text/plain`                                                    | `text/plain; charset=utf-8`       |

When several supported media types are present, eep selects the highest `q` value and preserves
header order for ties. Missing, wildcard-only, malformed, or unsupported `Accept` headers fall back
to the configured HTML theme.

## Local development

Run the self-contained Envoy integration test:

```bash
mise run smoke
```

The smoke task builds `main.wasm`, generates a temporary Envoy and Compose configuration, starts the
`http-debug` backend with Envoy `v1.39.0`, verifies passthrough plus HTML/JSON/XML/plain-text error
responses, checks localization and configuration, and cleans up the containers. It does not depend on
any checked-in runtime configuration.

For a persistent Docker deployment using a released eep image, follow the
[`examples/docker`](examples/docker) guide. For an Envoy Gateway deployment, see the
[`examples/kubernetes/envoy-gateway`](examples/kubernetes/envoy-gateway) manifests and guide.

### Development workflow

1. Edit the built-in HTML templates in `templates/html/*.tpl.html`.
2. Edit localization source in `l10n/locales.json` when needed.
3. Rebuild with `mise run build`; localization artifacts are generated automatically.
4. Run `mise run smoke`.

## Extracting the WASM File

To extract the WASM file from the Docker image for standalone use:

```bash
docker create --name eep-extract --entrypoint /plugin.wasm eep:latest
docker cp eep-extract:/plugin.wasm ./plugin.wasm
docker rm eep-extract
```

## Running with Envoy

This extension requires Envoy `v1.39.0` or later. Envoy Gateway users require `v1.9.0` or later when using its default Envoy image.

### Configuration

Eep accepts a strict JSON object through the Proxy-Wasm plugin configuration. If no configuration is
provided, it uses the following defaults:

```json
{
  "theme": "connection",
  "showDetails": false,
  "locale": "auto"
}
```

| Field         | Type    | Default      | Description                                                                                     |
| ------------- | ------- | ------------ | ----------------------------------------------------------------------------------------------- |
| `theme`       | string  | `connection` | Built-in HTML theme name from `templates/html/` (for example `connection`, `cats`, or `ghost`). |
| `showDetails` | boolean | `false`      | Whether rendered responses include request metadata.                                            |
| `locale`      | string  | `auto`       | HTML locale: browser-selected `auto`, English, or a supported base/region-qualified language.   |

Unknown fields, malformed JSON, an empty theme or locale, an unavailable theme, or an unsupported locale
cause plugin startup to fail rather than silently serving an unexpected page. When enabled, `showDetails`
exposes request metadata such as the host, URI, forwarded-for value, Kubernetes service identifiers, and
request ID; keep it `false` for public-facing error responses unless that information is intended to be visible.

#### Localization

Localization applies to HTML responses only. JSON, XML, and plain-text responses remain English. Supported
translated locales are `de`, `es`, `fr`, `hu`, `id`, `it`, `ko`, `nl`, `no`, `pl`, `pt`, `ro`, `ru`, `uk`,
and `zh`.

- `locale: auto` embeds the client-side localization runtime and selects from `navigator.languages`.
- An explicit locale such as `pl` forces that language for every HTML page rendered by the plugin instance.
- Regional tags fall back to a supported base language, for example `fr-CA` to `fr`.
- English (`en` or a region-qualified tag such as `en-US`) keeps the original content and omits the script.

The pages remain readable in English if JavaScript is disabled. Localization uses an inline script, so a strict
Content Security Policy must explicitly permit it. The embedded runtime adds roughly 54 KiB to localized HTML
responses before Envoy compression; compression remains the proxy's responsibility.

### Using Envoy Directly

1. Extract the WASM file (see above).
2. Configure the Wasm filter's `configuration` as a `google.protobuf.StringValue`. Its `value` is the
   JSON passed to eep:

```yaml
name: envoy.filters.http.wasm
typed_config:
  "@type": type.googleapis.com/envoy.extensions.filters.http.wasm.v3.Wasm
  config:
    name: eep
    configuration:
      "@type": type.googleapis.com/google.protobuf.StringValue
      value: |
        {
          "theme": "connection",
          "showDetails": false,
          "locale": "auto"
        }
    vm_config:
      runtime: envoy.wasm.runtime.v8
      code:
        local:
          filename: /etc/envoy/plugin.wasm
```

The smoke test generates its own Envoy configuration and deliberately selects `ghost`, disables
request details, and forces Polish localization so configuration consumption is tested together with
runtime behavior. For a persistent Docker deployment using a released image, see
[`examples/docker`](examples/docker).

### Using Envoy Gateway

Use `EnvoyExtensionPolicy.spec.wasm[].config` for ordinary eep configuration. Envoy Gateway serializes
this object to JSON and supplies it as the plugin configuration. `env.hostKeys` only forwards existing
Envoy-process environment variables and is not needed for theme, detail, or locale settings.

```yaml
apiVersion: gateway.envoyproxy.io/v1alpha1
kind: EnvoyExtensionPolicy
metadata:
  name: eep
  namespace: network
spec:
  targetSelectors:
    - group: gateway.networking.k8s.io
      kind: Gateway
      matchLabels:
        role: production
  wasm:
    - name: eep
      rootID: eep
      code:
        type: Image
        image:
          url: oci://ghcr.io/ishioni/eep:v0.1.0
      config:
        theme: connection
        showDetails: false
        locale: auto
```

`rootID` identifies the extension's root context and must be unique among Wasm extensions attached to
the same Envoy instance. A ready-to-adapt policy is available in
[`examples/kubernetes/envoy-gateway`](examples/kubernetes/envoy-gateway).

## Testing

Run the automated integration test through Envoy:

```bash
mise run smoke
```

For manual requests against a persistent deployment, use the commands in
[`examples/docker/README.md`](examples/docker/README.md). The Envoy admin interface is available at
`http://localhost:9901` when that example is running.

## How It Works

### Response Processing

1. **Interception**: The plugin monitors all HTTP response headers
2. **Detection**: When it detects a 4xx or 5xx status code, it flags the response for modification
3. **Replacement**: The original response body is replaced with a custom HTML error page
4. **Headers**: Content-Type, Content-Length, and Content-Encoding headers are updated appropriately

### Supported Error Codes

- **4xx (Client Errors)**: 400, 401, 402, 403, 404, 405, 406, 407, 408, 409, 410, etc.
  - Displays an orange-themed "Client Error" page
- **5xx (Server Errors)**: 500, 501, 502, 503, 504, 505, etc.
  - Displays a red-themed "Server Error" page

## Customization

### Modifying Error Pages

The built-in error page templates are copied from the sibling `error-pages` project and stored under `templates/`:

- `templates/html/*.tpl.html` - Built-in HTML themes selected by `theme`
- `templates/default.tpl.json` - Default JSON response template
- `templates/default.tpl.xml` - Default XML response template
- `templates/default.tpl.txt` - Default plain-text response template

You can edit these template files directly. They are embedded into the WASM binary at compile time using Go's `embed` package, so after editing them, you'll need to rebuild:

```bash
mise run build
# or build the OCI artifact
mise run docker-build
```

The templates include:

- Modern, responsive design
- Gradient backgrounds
- Action buttons (Go Back, Return Home, Retry)
- Mobile-friendly layout
- Customizable colors, text, and styling

### Status-specific content

Templates can branch on `.StatusCode` with standard `text/template` actions. For example:

```gotemplate
{{ if eq .StatusCode 404 }}The requested page does not exist.{{ else }}{{ .Description }}{{ end }}
```

Eep intercepts all valid 4xx and 5xx statuses. To change that policy, update `ParseErrorStatus` in
`internal/errorpages/status.go` and its table-driven tests.

## Development

### Project Structure

```
.
├── cmd/eep/main.go            # Entry point and WASM contexts
├── internal/                  # Internal packages
│   ├── config/               # Runtime plugin configuration
│   └── errorpages/           # Negotiation and rendering
│       ├── data.go           # Upstream-compatible template data
│       ├── format.go         # Accept-header content negotiation
│       ├── functions.go      # error-pages-compatible template functions
│       ├── renderer.go       # Per-format template selection
│       ├── status.go         # Error status parsing and defaults
│       └── template.go       # Parsed text/template execution
├── l10n/                      # Localization source, generator, and embedded runtime
├── templates/                 # Error response templates
│   ├── default.tpl.json      # Default JSON response template
│   ├── default.tpl.txt       # Default plain text response template
│   ├── default.tpl.xml       # Default XML response template
│   └── html/                 # Built-in HTML themes copied from error-pages
├── tools/smoke-envoy.sh       # Self-contained Envoy integration test
├── examples/                  # Docker and Envoy Gateway deployment examples
├── Dockerfile                 # Multi-stage Docker build
├── go.mod                     # Go module dependencies
└── README.md                  # This file
```

### Code Structure

**Main Package (`cmd/eep/main.go`):**

- `vmContext`: VM-level context for the plugin
- `pluginContext`: owns one plugin instance's runtime configuration and immutable renderer
- `httpContext`: captures request metadata and performs response interception
- runtime configuration is read from the Proxy-Wasm host during plugin startup

**Internal packages:**

- `internal/config`: strict JSON runtime configuration and defaults
- `internal/errorpages`: pure status parsing, content negotiation, and rendering
- `l10n`: locale validation plus generated, embedded client-side localization

### Logging

The plugin uses different log levels:

- `LogInfo`: Plugin initialization and error interception events
- `LogDebug`: Detailed status codes and operation confirmations
- `LogWarn`: Non-critical issues
- `LogError`: Critical failures

View logs in real-time:

```bash
docker compose --file examples/docker/compose.yaml logs -f envoy
```

## License

Apache License 2.0 - See the license header in source files for details.

## References

- [Proxy-WASM Go SDK](https://github.com/proxy-wasm/proxy-wasm-go-sdk)
- [Envoy WASM Documentation](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/wasm_filter)
