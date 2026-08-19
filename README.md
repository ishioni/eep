# eep

**eep** (Envoy Error Pages) is a Go Proxy-Wasm HTTP filter for Envoy. It replaces backend 4xx and
5xx response bodies with branded error pages while preserving the original status code.

It renders an HTML theme for browsers and negotiates JSON, XML, or plain-text responses for API
clients from the request's `Accept` header. Built-in HTML pages can be localized automatically or
configured for a specific locale.

## Requirements

To run eep, use one of the following:

- Envoy Proxy **v1.39.0 or later**.
- Envoy Gateway **v1.9.0 or later** when using its default Envoy image, which is Envoy v1.39.0.

If you override Envoy Gateway's Envoy image, it must still be Envoy v1.39.0 or later. Earlier
versions do not provide the WASI hostcalls used by the Go runtime.

## Quickstart

Download `eep.wasm` and `eep.wasm.sha256` from a GitHub release for a standalone Envoy deployment,
or use the published OCI image with Envoy Gateway.

- **Direct Envoy with Docker Compose:** follow [`examples/docker`](examples/docker). It extracts the
  WASM module from an eep image and loads it into Envoy.
- **Envoy Gateway:** adapt and apply
  [`examples/kubernetes/envoy-gateway/eep.yaml`](examples/kubernetes/envoy-gateway/eep.yaml) and
  [`examples/kubernetes/envoy-gateway/client-traffic-policy.yaml`](examples/kubernetes/envoy-gateway/client-traffic-policy.yaml).
  The first attaches eep to `Gateway` resources with an `EnvoyExtensionPolicy`; the second
  configures the response buffer required by larger error-page templates.

For example, after starting the Docker example:

```sh
curl --include http://localhost:10000/404
curl --include --header 'Accept: application/json' http://localhost:10000/404
curl --include --header 'Accept: application/xml' http://localhost:10000/500
curl --include --header 'Accept: text/plain' http://localhost:10000/503
```

Supported response representations are:

| Request `Accept` media type                                     | Response `Content-Type`           |
| --------------------------------------------------------------- | --------------------------------- |
| `text/html`                                                     | `text/html; charset=utf-8`        |
| `application/json`, `text/json`, or a `+json` structured suffix | `application/json; charset=utf-8` |
| `application/xml`, `text/xml`, or a `+xml` structured suffix    | `application/xml; charset=utf-8`  |
| `text/plain`                                                    | `text/plain; charset=utf-8`       |

When multiple supported types are accepted, eep selects the highest `q` value and keeps request
header order for ties. Missing, wildcard-only, malformed, or unsupported `Accept` headers use the
configured HTML theme.

## Configuration

Eep receives a strict JSON configuration object through the Proxy-Wasm host. Omitting it uses these
defaults:

```json
{
  "theme": "connection",
  "showDetails": false,
  "locale": "auto",
  "logLevel": "warn",
  "filterCodes": [],
  "excludeDomains": []
}
```

| Field            | Type    | Default      | Description                                                                                                                                                                                                 |
| ---------------- | ------- | ------------ | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `theme`          | string  | `connection` | Built-in HTML theme. Available themes are the filenames in [`templates/html`](templates/html) without the `.tpl.html` suffix, for example `connection`, `cats`, and `ghost`.                                |
| `showDetails`    | boolean | `false`      | Adds request metadata to rendered responses. Keep this disabled for public-facing errors unless exposing the host, URI, forwarding information, Kubernetes service metadata, and request ID is intentional. |
| `locale`         | string  | `auto`       | HTML locale. `auto` uses browser language preferences; `en` keeps English; a supported base or regional language tag forces a locale.                                                                       |
| `logLevel`       | string  | `warn`       | Minimum eep log level sent to Envoy: `debug`, `info`, `warn`, `error`, `critical`, or `off`. Critical startup failures remain visible.                                                                      |
| `filterCodes`    | array   | `[]`         | Error statuses to replace. Entries may be numeric or quoted individual codes, or quoted inclusive ranges such as `"500-510"`. Omitted or empty means every 4xx/5xx status.                                  |
| `excludeDomains` | array   | `[]`         | Go/RE2 regexes matched against the request authority/Host. A match leaves the original response untouched, even when its status is selected by `filterCodes`.                                               |

Unknown configuration fields, malformed JSON, empty `theme`, `locale`, or domain expressions,
unavailable themes, unsupported locales, invalid log levels, invalid regular expressions, and invalid
filter codes fail plugin startup rather than silently selecting a different response. Filter codes must be 4xx or 5xx
values, and ranges are inclusive and must be ordered from low to high.

Domain expressions are case-sensitive and unanchored unless the expression says otherwise. The matched
value may include a port, so use an expression such as `(?i)^admin\\.example\\.com(:[0-9]+)?$` when an
exact, case-insensitive domain match with an optional port is required.

### Direct Envoy

Pass the JSON object through the Wasm filter's `google.protobuf.StringValue` configuration. The
[Docker example](examples/docker/envoy.yaml) shows the complete filter configuration:

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
          "locale": "auto",
          "logLevel": "warn",
          "filterCodes": [404, "500-510"],
          "excludeDomains": ["(?i)^admin\\.example\\.com(:[0-9]+)?$"]
        }
```

### Envoy Gateway

Configure eep through `EnvoyExtensionPolicy.spec.wasm[].config`. Envoy Gateway serializes that
object as JSON and supplies it to the plugin at startup. The complete policy is in
[`examples/kubernetes/envoy-gateway/eep.yaml`](examples/kubernetes/envoy-gateway/eep.yaml):

```yaml
wasm:
  - name: eep
    rootID: eep
    code:
      type: Image
      image:
        url: oci://ghcr.io/ishioni/eep:<release-tag>
    config:
      theme: connection
      showDetails: false
      locale: auto
      logLevel: warn
      filterCodes: [404, "500-510"]
      excludeDomains: ["(?i)^admin\\.example\\.com(:[0-9]+)?$"]
```

Envoy Gateway's default response buffer is 32 KiB. Several built-in HTML themes exceed that size,
which causes Envoy to report `response_payload_too_large` and return the upstream error instead of
eep's page. Configure a `ClientTrafficPolicy` with `connection.bufferLimit: 1Mi` and `http2.initialStreamWindowSize: 1Mi` for gateways using
eep; the recommended policy is included in
[`examples/kubernetes/envoy-gateway/client-traffic-policy.yaml`](examples/kubernetes/envoy-gateway/client-traffic-policy.yaml).
Increase the limit further for larger custom templates.

`rootID` must be unique among Wasm extensions attached to the same Envoy instance. Do not use
`env.hostKeys` for eep's settings: it only forwards existing Envoy-process environment variables and
is unnecessary for theme, detail, and locale configuration.

### Localization

Localization applies only to HTML responses; JSON, XML, and plain-text responses stay in English.
Supported translated locales are `de`, `es`, `fr`, `hu`, `id`, `it`, `ko`, `nl`, `no`, `pl`, `pt`,
`ro`, `ru`, `uk`, and `zh`.

- `locale: auto` embeds the localization runtime and selects a supported language from
  `navigator.languages`.
- A configured locale, such as `fr` or `fr-CA`, is forced for every rendered HTML page. Regional tags
  fall back to their supported base locale.
- English (`en`, `en-US`, and similar tags) leaves the page in its original English and omits the
  localization script.

Pages remain readable in English when JavaScript is disabled. The localization runtime is inline, so
a strict Content Security Policy must allow it.

## Development

[Mise](https://mise.jdx.dev/) is the development toolchain source of truth. It provides the pinned Go
version and all project tasks.

```sh
mise install
mise tasks
```

Common tasks:

```sh
mise run fmt && mise run oxfmt
mise run test
mise run build       # writes main.wasm for local development
mise run smoke       # builds and tests eep through Envoy in temporary containers
mise run docker-run  # starts the persistent local stack in tools/
```

The smoke test generates its own Compose and Envoy configuration and does not depend on checked-in
runtime configuration. Stop the persistent local stack with:

```sh
docker compose --file tools/docker-compose.yaml down --volumes
```

GitHub releases publish the local build output as `eep.wasm` together with `eep.wasm.sha256`.

### Templates

Templates use Go's standard [`text/template`](https://pkg.go.dev/text/template) behavior and are
embedded in the WASM module at build time:

- HTML themes: `templates/html/*.tpl.html`
- JSON response: `templates/default.tpl.json`
- XML response: `templates/default.tpl.xml`
- Plain-text response: `templates/default.tpl.txt`

To add an HTML theme, add a non-empty `*.tpl.html` file under `templates/html/`; eep discovers
built-in themes at build time. Update or add renderer tests when changing template data or functions,
then run `mise run build` and `mise run smoke`.

### Locales

`l10n/locales.json` is the localization source. Generated localization artifacts are derived from it
at build time:

```sh
mise run l10n-generate
```

Update `l10n/locales.json` to add or change translations, regenerate the artifacts, and verify the
result with the smoke test or a browser request using the relevant configured locale.

## Acknowledgements

- HTML templates and localization data are synchronized from the sibling
  [`error-pages`](https://github.com/tarampampam/error-pages) project, under its MIT License. See
  [`THIRD_PARTY_NOTICES.md`](THIRD_PARTY_NOTICES.md) for the required copyright and license notice.
- Eep is built on the [Proxy-Wasm Go SDK](https://github.com/proxy-wasm/proxy-wasm-go-sdk) and
  Envoy's [Wasm HTTP filter](https://www.envoyproxy.io/docs/envoy/latest/configuration/http/http_filters/wasm_filter).
- Licensed under the [Apache License 2.0](LICENSE).
