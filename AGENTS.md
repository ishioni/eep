# Agent Notes

This repository has a non-obvious Envoy/Go/WASI compatibility issue. Read this before changing the build or template rendering code.

## Required background

Start with:

- `README.md` for the project goal and normal development workflow.
- `docs/wasi-import-stubbing.md` for the current investigation into Go `text/template`, unsupported WASI imports, and the post-link stubbing experiment.

## Current direction

The project is trying to stay compatible with the sibling `error-pages` project's template behavior. Do not replace `text/template` with a tiny custom renderer without discussing it first. `error-pages` supports custom Go templates and helper functions, so renderer compatibility matters.

The relevant sibling project is available at:

- `../error-pages`

Important upstream files to compare against:

- `error-pages/internal/template/template.go`
- `error-pages/internal/template/props.go`
- `error-pages/internal/template/template_test.go`
- `error-pages/templates/*.html`

## Known findings

- Go `GOOS=wasip1 GOARCH=wasm` with `text/template` pulls in `os` and Go's WASI filesystem imports.
- Envoy's V8 Proxy-Wasm runtime does not provide the full filesystem/socket WASI surface.
- Go 1.26 and Envoy 1.38 were tested and still have this problem.
- A post-link WASM stubbing experiment allowed Envoy to instantiate the module while keeping `text/template` linked.

## Before making changes

- Prefer documenting experiments and asking before changing architectural direction.
- Keep `text/template` compatibility as a requirement unless the user explicitly changes that goal.
- If implementing WASI stubbing, avoid brittle fixed-index hacks in production code; parse the WASM structure or use a reliable library/tooling path.
- If changing template data/functions, add or update compatibility tests based on `error-pages/internal/template/template_test.go`.

## Useful commands

- Build WASM: `make build`
- Run tests: `go test ./...`
- Inspect WASM imports: `wasm-tools print main.wasm | grep '^  (import'`
- Validate WASM: `wasm-tools validate main.wasm`
- Start local Envoy stack: `docker-compose up -d --build`
- Check Envoy logs: `docker-compose logs envoy`
