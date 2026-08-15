# syntax=docker/dockerfile:1
# check=skip=InvalidDefaultArgInFrom

ARG GO_VERSION

FROM golang:${GO_VERSION}-bookworm AS builder
ARG TARGETARCH
ARG VERSION=dev
ARG REVISION=dev
# renovate: datasource=github-releases depName=bytecodealliance/wasm-tools extractVersion=^v(?<version>.*)$
ARG WASM_TOOLS_VERSION=1.256.0

RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates curl tar \
    && rm -rf /var/lib/apt/lists/* \
    && case "${TARGETARCH}" in \
        "amd64"|"") WASM_TOOLS_ARCH="x86_64-linux" ;; \
        "arm64") WASM_TOOLS_ARCH="aarch64-linux" ;; \
        *) echo "unsupported TARGETARCH: ${TARGETARCH}"; exit 1 ;; \
    esac \
    && curl -fsSL \
        "https://github.com/bytecodealliance/wasm-tools/releases/download/v${WASM_TOOLS_VERSION}/wasm-tools-${WASM_TOOLS_VERSION}-${WASM_TOOLS_ARCH}.tar.gz" \
        -o /tmp/wasm-tools.tar.gz \
    && tar -xzf /tmp/wasm-tools.tar.gz -C /tmp \
    && install \
        "/tmp/wasm-tools-${WASM_TOOLS_VERSION}-${WASM_TOOLS_ARCH}/wasm-tools" \
        /usr/local/bin/wasm-tools \
    && rm -rf /tmp/wasm-tools* \
    && wasm-tools --version

WORKDIR /workspace

COPY go.mod go.sum ./
RUN go mod download

COPY main.go config.yaml ./
COPY internal/ internal/
COPY templates/ templates/
COPY tools/ tools/

RUN CGO_ENABLED=0 go build -trimpath -o /usr/local/bin/patch-wasi-imports ./tools/patch-wasi-imports
RUN GOOS=wasip1 GOARCH=wasm go build -trimpath -buildmode=c-shared \
    -ldflags "-s -w -X main.version=${VERSION}" \
    -o main.raw.wasm main.go
RUN patch-wasi-imports -in main.raw.wasm -out main.wasm \
    && wasm-tools validate main.wasm

FROM scratch
ARG VERSION=dev
ARG REVISION=dev

LABEL org.opencontainers.image.title="eep" \
    org.opencontainers.image.description="Envoy error pages Proxy-Wasm extension" \
    org.opencontainers.image.version="${VERSION}" \
    org.opencontainers.image.revision="${REVISION}" \
    org.opencontainers.image.source="https://github.com/ishioni/eep"

COPY --from=builder /workspace/main.wasm /plugin.wasm
