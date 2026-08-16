#!/usr/bin/env sh
set -eu

repo_root=$(CDPATH='' cd -- "$(dirname -- "$0")/.." && pwd)
workdir=$(mktemp -d)
compose_file=${workdir}/compose.yaml
envoy_file=${workdir}/envoy.yaml
response_body=
response_headers=

compose() {
    docker compose --project-name eep-smoke --file "${compose_file}" "$@"
}

cleanup() {
    compose down --volumes >/dev/null 2>&1 || true
    rm -rf "${workdir}"
    rm -f "${response_body:-}" "${response_headers:-}"
}

assert_error_response() {
    accept=$1
    expected_content_type=$2
    expected_body=$3
    status_code=$4

    status=$(curl -sS -D "${response_headers}" -o "${response_body}" -w '%{http_code}' \
        -H "Accept: ${accept}" "http://localhost:10000/${status_code}")
    [ "${status}" = "${status_code}" ]
    grep -Fqi "content-type: ${expected_content_type}" "${response_headers}"
    grep -Fq "${expected_body}" "${response_body}"
}

assert_passthrough_error_response() {
    host=$1
    status_code=$2

    status=$(curl -sS -D "${response_headers}" -o "${response_body}" -w '%{http_code}' \
        -H "Accept: text/html" -H "Host: ${host}" "http://localhost:10000/${status_code}")
    [ "${status}" = "${status_code}" ]
    test -s "${response_body}"

    if grep -Fq '<svg class="ghost"' "${response_body}"; then
        echo "eep replaced a response that should have passed through" >&2
        exit 1
    fi
}

cat >"${compose_file}" <<EOF
---
name: eep-smoke

services:
  http-debug:
    image: ghcr.io/ishioni/http-debug:0.0.3

  envoy:
    image: envoyproxy/envoy:v1.39.0
    command: ["-c", "/etc/envoy/envoy.yaml", "--log-level", "info"]
    ports:
      - "10000:10000"
      - "9901:9901"
    volumes:
      - "${envoy_file}:/etc/envoy/envoy.yaml:ro"
      - "${repo_root}/main.wasm:/etc/envoy/plugin.wasm:ro"
    depends_on:
      http-debug:
        condition: service_started
EOF

cat >"${envoy_file}" <<'EOF'
---
admin:
  address:
    socket_address:
      address: 0.0.0.0
      port_value: 9901

static_resources:
  listeners:
    - name: http
      address:
        socket_address:
          address: 0.0.0.0
          port_value: 10000
      filter_chains:
        - filters:
            - name: envoy.filters.network.http_connection_manager
              typed_config:
                "@type": type.googleapis.com/envoy.extensions.filters.network.http_connection_manager.v3.HttpConnectionManager
                stat_prefix: ingress_http
                route_config:
                  name: local_route
                  virtual_hosts:
                    - name: backend
                      domains: ["*"]
                      routes:
                        - match:
                            prefix: "/"
                          route:
                            cluster: http_debug
                            timeout: 30s
                http_filters:
                  - name: envoy.filters.http.wasm
                    typed_config:
                      "@type": type.googleapis.com/envoy.extensions.filters.http.wasm.v3.Wasm
                      config:
                        name: eep
                        configuration:
                          "@type": type.googleapis.com/google.protobuf.StringValue
                          value: |
                            {
                              "theme": "ghost",
                              "showDetails": false,
                              "locale": "pl",
                              "filterCodes": [404, "500-510"],
                              "excludeDomains": ["^skip\\.example\\.test$"]
                            }
                        vm_config:
                          runtime: envoy.wasm.runtime.v8
                          code:
                            local:
                              filename: /etc/envoy/plugin.wasm
                  - name: envoy.filters.http.router
                    typed_config:
                      "@type": type.googleapis.com/envoy.extensions.filters.http.router.v3.Router

  clusters:
    - name: http_debug
      type: STRICT_DNS
      lb_policy: ROUND_ROBIN
      connect_timeout: 5s
      load_assignment:
        cluster_name: http_debug
        endpoints:
          - lb_endpoints:
              - endpoint:
                  address:
                    socket_address:
                      address: http-debug
                      port_value: 8080
EOF

trap cleanup EXIT INT TERM

cd "${repo_root}"
mise run build
compose up -d --force-recreate

attempt=0
until curl -fsS http://localhost:10000/200 >/dev/null 2>&1; do
    attempt=$((attempt + 1))
    if [ "${attempt}" -ge 30 ]; then
        compose logs envoy
        echo "Envoy did not become ready" >&2
        exit 1
    fi

    sleep 1
done

response_body=$(mktemp)
response_headers=$(mktemp)

assert_error_response '*/*' 'text/html; charset=utf-8' '<title>404: Not Found</title>' 404
assert_error_response 'application/json' 'application/json; charset=utf-8' '"error": true' 404
assert_error_response 'application/xml' 'application/xml; charset=utf-8' '<error>' 404
assert_error_response 'text/plain' 'text/plain; charset=utf-8' 'Error 404: Not Found' 404
assert_error_response 'text/html' 'text/html; charset=utf-8' '<title>500: Internal Server Error</title>' 500
assert_passthrough_error_response 'allowed.example.test' 403
assert_passthrough_error_response 'skip.example.test' 404

assert_error_response 'text/html' 'text/html; charset=utf-8' '<title>500: Internal Server Error</title>' 500
grep -Fq '<svg class="ghost"' "${response_body}"
grep -Fq 'window.l10n.setLocale("pl")' "${response_body}"

if grep -Fq '<table class="details">' "${response_body}"; then
    echo "request details were rendered despite showDetails=false" >&2
    exit 1
fi

if compose logs envoy | grep -Eiq 'missing import|failed to load wasm|failed to render'; then
    compose logs envoy
    echo "Envoy reported a WASM loading or rendering failure" >&2
    exit 1
fi

echo "Envoy smoke test passed"
