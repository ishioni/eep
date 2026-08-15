# Agent Notes

## Runtime compatibility

This project builds a Go `GOOS=wasip1 GOARCH=wasm` Proxy-Wasm HTTP filter. It requires:

- Envoy Proxy `v1.39.0` or later; or
- Envoy Gateway `v1.9.0` or later when using its default Envoy image.

Envoy `v1.39.0` vendors the WASI hostcalls required by Go's standard library. Build the release
artifact directly as `main.wasm`; do not reintroduce the former post-link WASI import stubbing
workaround without verifying a regression in a supported Envoy image.

## Template compatibility

The project stays compatible with the sibling `error-pages` project's `text/template` behavior.
Do not replace `text/template` with a small custom renderer without discussing it first.

The relevant sibling project is available at `../error-pages`. When changing template data or
functions, add or update compatibility tests based on:

- `error-pages/internal/template/template.go`
- `error-pages/internal/template/data.go`
- `error-pages/internal/template/template_test.go`
- `error-pages/templates/html/*.tpl.html`

## Tooling and validation

Mise is the source of truth for tool versions and development tasks. Run `mise install` after
cloning and `mise tasks` to discover the tasks used by CI.

Before finishing a change, run the relevant focused tasks and then the full checks:

- Format: `mise run fmt && mise run oxfmt`
- Formatting checks: `mise run fmt-check && mise run oxfmt-check`
- Tests: `mise run test`
- Lint and vet: `mise run lint && mise run vet`
- Dependency scan: `mise run vulncheck`
- WASM build: `mise run build`
- Envoy smoke test: `mise run smoke`
- OCI build: `mise run docker-build`
- Workflow audits: `mise run actionlint && mise run zizmor`

The release artifact is `main.wasm`.
