# Agent Notes

This repository has a non-obvious Envoy/Go/WASI compatibility issue. Read this before changing the build or template rendering code.

## Required background

Start with:

- `README.md` for the project goal and normal development workflow.
- `docs/wasi-import-stubbing.md` for the current investigation into Go `text/template`, unsupported WASI imports, upstream fixes, and the post-link stubbing workaround.

## Current direction

The project is trying to stay compatible with the sibling `error-pages` project's template behavior. Do not replace `text/template` with a tiny custom renderer without discussing it first. `error-pages` supports custom Go templates and helper functions, so renderer compatibility matters.

The relevant sibling project is available at:

- `../error-pages`

Important upstream files to compare against:

- `error-pages/internal/template/template.go`
- `error-pages/internal/template/data.go`
- `error-pages/internal/template/template_test.go`
- `error-pages/templates/html/*.tpl.html`

## Known findings

- Go `GOOS=wasip1 GOARCH=wasm` with `text/template` pulls in `os` and Go's WASI filesystem imports.
- Envoy's V8 Proxy-Wasm runtime does not provide the full filesystem/socket WASI surface.
- Go 1.26 and Envoy 1.38 were tested and still have this problem.
- A post-link WASM stubbing step allows current Envoy images to instantiate the module while keeping `text/template` linked.
- The upstream hostcall fix was merged in `proxy-wasm/proxy-wasm-cpp-host#533`: https://github.com/proxy-wasm/proxy-wasm-cpp-host/pull/533
- Envoy still needs to vendor that host update. Track `envoyproxy/envoy#44534`: https://github.com/envoyproxy/envoy/pull/44534
- Once Envoy ships with that update, test raw `main.raw.wasm` without patching; the patcher should become optional or removable.

## Before making changes

- Prefer documenting experiments and asking before changing architectural direction.
- Keep `text/template` compatibility as a requirement unless the user explicitly changes that goal.
- If changing WASI stubbing, remember it is expected to be temporary until Envoy includes the upstream hostcall support.
- Avoid brittle fixed-index hacks in production code; parse the WASM structure or use a reliable library/tooling path.
- If changing template data/functions, add or update compatibility tests based on `error-pages/internal/template/template_test.go`.

## Tooling and validation

Mise is the source of truth for tool versions and development tasks. Run `mise install` after cloning and
`mise tasks` to discover the same commands used by CI. The Makefile is only a compatibility wrapper.

Before finishing a change, run the relevant focused tasks and then the full checks:

- Format Go and non-Go files: `mise run fmt && mise run oxfmt`
- Check formatting: `mise run fmt-check && mise run oxfmt-check`
- Run tests: `mise run test`
- Run lint and vet: `mise run lint && mise run vet`
- Scan dependencies: `mise run vulncheck`
- Build and validate WASM: `mise run build`
- Test through Envoy: `mise run smoke`
- Build the OCI artifact: `mise run docker-build`
- Audit workflows: `mise run actionlint && mise run zizmor`

The release artifact is `main.wasm`. `main.raw.wasm` is intentionally unpatched and exists only for compatibility
experiments. Current Envoy images still require the patched module.
