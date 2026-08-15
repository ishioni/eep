# eep

**eep** (Envoy Error Pages) is an Envoy Proxy-Wasm extension written in Go that intercepts backend error responses (4xx and 5xx status codes) and replaces them with custom HTML, JSON, XML, or plain-text error pages.

## Features

- **Automatic Error Interception**: Detects and handles all 4xx and 5xx HTTP status codes
- **Content Negotiation**: Responds with HTML, JSON, XML, or plain text according to the request's `Accept` header
- **Custom Error Pages**: Provides error pages for all client errors (4xx) and server errors (5xx)
- **Template-Based Design**: Built-in response templates are stored in separate files for easy customization without Go knowledge
- **Runtime Configuration**: Select the HTML theme and request-detail visibility through the host's plugin configuration
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

## Local Development

Build the module and start the local backend and Envoy stack:

```bash
mise run build
docker compose up
```

The http-debug service provides endpoints that return different status codes:

- `http://localhost:10000/200` - Returns 200 OK (passes through)
- `http://localhost:10000/400` - Returns 400 (shows 4xx error page)
- `http://localhost:10000/404` - Returns 404 (shows 4xx error page)
- `http://localhost:10000/500` - Returns 500 (shows 5xx error page)
- `http://localhost:10000/503` - Returns 503 (shows 5xx error page)

### Quick Testing

```bash
# Build the module, start Envoy, verify passthrough plus 404/500 rendering, and clean up
mise run smoke

# Or inspect the running stack in a browser
open http://localhost:10000/500
open http://localhost:10000/404
```

### Development Workflow

1. Edit the built-in HTML templates in `templates/html/*.tpl.html`
2. Rebuild with `mise run build`
3. Test with `mise run smoke` or visit the local endpoints in a browser
4. Check Envoy logs with `docker compose logs -f envoy`

### Stopping the Environment

```bash
docker compose down
```

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
  "showDetails": false
}
```

| Field         | Type    | Default      | Description                                                                                     |
| ------------- | ------- | ------------ | ----------------------------------------------------------------------------------------------- |
| `theme`       | string  | `connection` | Built-in HTML theme name from `templates/html/` (for example `connection`, `cats`, or `ghost`). |
| `showDetails` | boolean | `false`      | Whether rendered responses include request metadata.                                            |

Unknown fields, malformed JSON, an empty theme, or a theme that is not embedded in the module cause
plugin startup to fail rather than silently serving an unexpected page. When enabled, `showDetails`
exposes request metadata such as the host, URI, forwarded-for value, Kubernetes service identifiers, and
request ID; keep it `false` for public-facing error responses unless that information is intended to be visible.

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
          "showDetails": false
        }
    vm_config:
      runtime: envoy.wasm.runtime.v8
      code:
        local:
          filename: /etc/envoy/plugin.wasm
```

The repository's [`envoy.yaml`](envoy.yaml) is a complete local development example. It deliberately
selects `ghost` and disables details so the smoke test verifies that the configuration is consumed.
For a release-image Docker Compose deployment, see [`examples/docker`](examples/docker). Run the local
configuration with:

```bash
docker run --rm -it \
  -v $(pwd)/envoy.yaml:/etc/envoy/envoy.yaml \
  -v $(pwd)/plugin.wasm:/etc/envoy/plugin.wasm \
  -p 10000:10000 \
  envoyproxy/envoy:v1.39.0 \
  -c /etc/envoy/envoy.yaml
```

### Using Envoy Gateway

Use `EnvoyExtensionPolicy.spec.wasm[].config` for ordinary eep configuration. Envoy Gateway serializes
this object to JSON and supplies it as the plugin configuration. `env.hostKeys` only forwards existing
Envoy-process environment variables and is not needed for theme or detail settings.

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
```

`rootID` identifies the extension's root context and must be unique among Wasm extensions attached to
the same Envoy instance. A ready-to-adapt policy is available in
[`examples/kubernetes/envoy-gateway`](examples/kubernetes/envoy-gateway).

## Testing

The Compose setup includes a test backend (`http-debug`) that makes testing easy. Run the automated smoke test with `mise run smoke`, or start the stack manually:

```bash
mise run build
docker compose up -d

# Test manually
curl http://localhost:10000/200   # Normal response (passes through)
curl http://localhost:10000/400   # Client error (custom 4xx page)
curl http://localhost:10000/404   # Not found (custom 4xx page)
curl http://localhost:10000/500   # Server error (custom 5xx page)
curl http://localhost:10000/503   # Service unavailable (custom 5xx page)

# View in browser for full styling
open http://localhost:10000/500
```

You can also access the Envoy admin interface at http://localhost:9901

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

### Adding Status-Specific Pages

To handle specific status codes differently, modify the `GetErrorPage()` function in `internal/errorpages/errorpages.go`:

```go
func (h *Handler) GetErrorPage(status string) []byte {
    switch status {
    case "404":
        return error404HTML
    case "500":
        return error500HTML
    case "503":
        return error503HTML
    default:
        if status[0] == '4' {
            return h.error4xxHTML
        }
        return h.error5xxHTML
    }
}
```

### Excluding Certain Error Codes

Modify the `IsErrorStatus()` function in `internal/errorpages/errorpages.go` to exclude specific status codes from being intercepted.

## Development

### Project Structure

```
.
├── main.go                    # Entry point and WASM contexts
├── internal/                  # Internal packages
│   └── errorpages/           # Error page handling
│       └── errorpages.go
├── templates/                 # HTML error page templates
│   ├── default.tpl.json      # Default JSON response template
│   ├── default.tpl.txt       # Default plain text response template
│   ├── default.tpl.xml       # Default XML response template
│   └── html/                 # Built-in HTML themes copied from error-pages

├── Dockerfile                 # Multi-stage Docker build
├── Dockerfile.debug           # Debug build configuration
├── docker-compose.yaml        # Local testing setup
├── envoy.yaml                 # Envoy configuration
├── go.mod                     # Go module dependencies
└── README.md                  # This file
```

### Code Structure

**Main Package (`main.go`):**

- `vmContext`: VM-level context for the plugin
- `pluginContext`: Plugin-level context, handles initialization
- `httpContext`: HTTP request/response context, handles error interception
- runtime configuration is read from the Proxy-Wasm host during plugin startup

**Internal Packages:**

- `internal/errorpages`: Error detection and page template management

### Logging

The plugin uses different log levels:

- `LogInfo`: Plugin initialization and error interception events
- `LogDebug`: Detailed status codes and operation confirmations
- `LogWarn`: Non-critical issues
- `LogError`: Critical failures

View logs in real-time:

```bash
docker compose logs -f envoy
```

## License

Apache License 2.0 - See the license header in source files for details.

## References

- [Proxy-WASM Go SDK](https://github.com/proxy-wasm/proxy-wasm-go-sdk)
- [Envoy WASM Documentation](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/wasm_filter)
