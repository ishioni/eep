# Envoy Gateway example

This manifest attaches eep to every `Gateway` labelled `role: production` in the `network`
namespace. It requires Envoy Gateway `v1.9.0+`, whose default Envoy image is compatible with eep.

## Configure and apply

1. Update `metadata.namespace`, the `targetSelectors` labels, and the OCI image tag in
   [`eep.yaml`](eep.yaml) and [`client-traffic-policy.yaml`](client-traffic-policy.yaml) for your deployment.
2. Review `config.showDetails`. When enabled, eep includes request metadata such as host, URI,
   forwarded-for values, service identifiers, and request IDs in error responses. Keep it `false`
   for public-facing gateways unless that disclosure is intentional.
3. Set `config.locale` to `auto`, English, or a supported translated locale such as `pl`.
4. Optionally set `config.logLevel` to `debug`, `info`, `warn`, `error`, `critical`, or `off`. The default
   is `warn`.
5. Optionally set `config.filterCodes` to individual error codes and inclusive ranges. Omitted or empty
   means every 4xx/5xx status.
6. Optionally add Go/RE2 regexes to `config.excludeDomains`. A matching request authority/Host bypasses
   eep and preserves the upstream error response.
7. Apply both policies:

   ```sh
   kubectl apply --filename eep.yaml --filename client-traffic-policy.yaml
   ```

Envoy Gateway's default response buffer is only 32 KiB. Several built-in HTML themes are larger than
that, which can result in Envoy returning `response_payload_too_large` instead of the rendered error
page. [`client-traffic-policy.yaml`](client-traffic-policy.yaml) raises the per-request buffer to the
recommended 1 MiB. Increase it further if you use a larger custom template.

`spec.wasm[].config` is the configuration channel for eep. Envoy Gateway serializes the YAML object
as JSON and eep receives it during plugin startup. Do not use `env.hostKeys` for `theme`,
`showDetails`, `locale`, or `logLevel`; it is only for forwarding existing environment variables from
the Envoy process.

`rootID` identifies the Wasm extension's root context. It must be unique among Wasm extensions loaded
into the same Envoy instance.
