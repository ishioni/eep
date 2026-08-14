#!/usr/bin/env sh
set -eu

cleanup() {
    rm -f "${response_body:-}"
    docker compose down -v >/dev/null 2>&1 || true
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

status=$(curl -sS -o "${response_body}" -w '%{http_code}' http://localhost:10000/404)
[ "${status}" = "404" ]
grep -q '<title>404 | Not Found</title>' "${response_body}"

status=$(curl -sS -o "${response_body}" -w '%{http_code}' http://localhost:10000/500)
[ "${status}" = "500" ]
grep -q '<title>500 | Internal Server Error</title>' "${response_body}"

if docker compose logs envoy | grep -Eiq 'missing import|failed to load wasm|failed to render'; then
    docker compose logs envoy
    echo "Envoy reported a WASM loading or rendering failure" >&2
    exit 1
fi

echo "Envoy smoke test passed"
