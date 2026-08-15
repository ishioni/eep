# Direct Envoy Docker example

This example runs eep in Envoy Proxy `v1.39.0` with a local status-code backend. It passes the
configuration through the Wasm filter's `google.protobuf.StringValue` field, which eep reads with
`proxywasm.GetPluginConfiguration()`.

## Prerequisites

- Docker with Compose v2
- An eep OCI image, such as `ghcr.io/ishioni/eep:v0.1.0`

## Run it

Extract the module from the eep image into this directory. The generated `plugin.wasm` is ignored by
Git:

```sh
export EEP_IMAGE=ghcr.io/ishioni/eep:v0.1.0
docker create --name eep-extract --entrypoint /plugin.wasm "${EEP_IMAGE}"
docker cp eep-extract:/plugin.wasm ./plugin.wasm
docker rm eep-extract
```

Start the stack:

```sh
docker compose up --detach
```

Request an error response:

```sh
curl --include http://localhost:10000/404
curl --include --header 'Accept: application/json' http://localhost:10000/404
```

The configured HTML theme is `connection`; request details are disabled. Change the JSON under
`config.configuration.value` in [`envoy.yaml`](envoy.yaml), restart Envoy, and request another error
response to try other settings.

Stop and remove the stack:

```sh
docker compose down --volumes
```
