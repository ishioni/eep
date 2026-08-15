# Envoy Gateway example

This manifest attaches eep to every `Gateway` labelled `role: production` in the `network`
namespace. It requires Envoy Gateway `v1.9.0+`, whose default Envoy image is compatible with eep.

## Configure and apply

1. Update `metadata.namespace`, the `targetSelectors` labels, and the OCI image tag in
   [`eep.yaml`](eep.yaml) for your deployment.
2. Review `config.showDetails`. When enabled, eep includes request metadata such as host, URI,
   forwarded-for values, service identifiers, and request IDs in error responses. Keep it `false`
   for public-facing gateways unless that disclosure is intentional.
3. Set `config.locale` to `auto`, English, or a supported translated locale such as `pl`.
4. Apply the policy:

   ```sh
   kubectl apply --filename eep.yaml
   ```

`spec.wasm[].config` is the configuration channel for eep. Envoy Gateway serializes the YAML object
as JSON and eep receives it during plugin startup. Do not use `env.hostKeys` for `theme`,
`showDetails`, or `locale`; it is only for forwarding existing environment variables from the Envoy
process.

`rootID` identifies the Wasm extension's root context. It must be unique among Wasm extensions loaded
into the same Envoy instance.
