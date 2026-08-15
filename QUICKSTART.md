# Quick Start Guide

## Prerequisites

- [mise](https://mise.jdx.dev/)
- Docker with Compose v2
- Git

## Install the toolchain

```bash
mise install
```

This installs the exact Go, formatting, linting, and workflow-audit versions locked in
`.mise/mise.lock`.

## Build the plugin

```bash
mise run build
```

The build produces `main.wasm`, the release artifact for Envoy Proxy `v1.39.0+` and Envoy Gateway
`v1.9.0+` when using its default Envoy image.

## Test through Envoy

```bash
mise run smoke
```

The smoke task builds the module, starts the local backend and Envoy, verifies `200` passthrough and
rendered `404`/`500` responses, checks Envoy logs for WASM failures, and cleans up.

To keep the stack running for browser testing:

```bash
mise run build
docker compose up -d

open http://localhost:10000/404
open http://localhost:10000/500
```

The Envoy admin interface is available at <http://localhost:9901>.

Stop the stack with:

```bash
docker compose down -v
```

## Common tasks

| Command                 | Description                           |
| ----------------------- | ------------------------------------- |
| `mise tasks`            | List project tasks                    |
| `mise run build`        | Build the WASM release artifact       |
| `mise run test`         | Run Go tests with race detection      |
| `mise run smoke`        | Run the Envoy integration smoke test  |
| `mise run lint`         | Run golangci-lint                     |
| `mise run fmt`          | Format Go code                        |
| `mise run oxfmt`        | Format supported non-Go files         |
| `mise run docker-build` | Build the scratch OCI artifact        |
| `mise run clean`        | Remove generated build/test artifacts |

## Customize a template

Built-in HTML themes live under `templates/html/*.tpl.html`. The selected theme is configured in
`config.yaml`.

After editing a template, run:

```bash
mise run smoke
```

## Build the OCI artifact

```bash
mise run docker-build
```

The scratch image contains the validated module at `/plugin.wasm`. Extract it with:

```bash
docker create --name eep-extract --entrypoint /plugin.wasm eep
docker cp eep-extract:/plugin.wasm ./plugin.wasm
docker rm eep-extract
```

## Troubleshooting

Check the Envoy logs:

```bash
docker compose logs envoy
```

Verify local tooling and tasks:

```bash
mise doctor
mise tasks
```

If ports `10000`, `9901`, or `8080` are already in use, stop the conflicting service or adjust
`docker-compose.yaml`.
