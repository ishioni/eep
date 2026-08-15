#!/usr/bin/env sh
set -eu

cleanup() {
    rm -f "${response_body:-}" "${response_headers:-}"
    docker compose down -v >/dev/null 2>&1 || true
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

trap cleanup EXIT INT TERM

mise run build
docker compose up -d --build --force-recreate

attempt=0
until curl -fsS http://localhost:10000/200 >/dev/null 2>&1; do
    attempt=$((attempt + 1))
    if [ "${attempt}" -ge 30 ]; then
        docker compose logs envoy
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

if grep -Fq 'Original URI' "${response_body}"; then
    echo "request details were rendered despite showDetails=false" >&2
    exit 1
fi

if docker compose logs envoy | grep -Eiq 'missing import|failed to load wasm|failed to render'; then
    docker compose logs envoy
    echo "Envoy reported a WASM loading or rendering failure" >&2
    exit 1
fi

echo "Envoy smoke test passed"
