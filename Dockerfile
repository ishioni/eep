# syntax=docker/dockerfile:1
# check=skip=InvalidDefaultArgInFrom

# Mise and CI pass the exact Go version from .mise/config.toml.
ARG GO_VERSION

FROM golang:${GO_VERSION}-bookworm AS builder
ARG VERSION=dev
ARG REVISION=dev

WORKDIR /workspace

COPY go.mod go.sum ./
RUN go mod download

COPY main.go ./
COPY internal/ internal/
COPY l10n/ l10n/
COPY templates/ templates/

RUN go generate ./l10n/...
RUN GOOS=wasip1 GOARCH=wasm go build -trimpath -buildmode=c-shared \
    -ldflags "-s -w -X main.version=${VERSION}" \
    -o main.wasm main.go

FROM scratch
ARG VERSION=dev
ARG REVISION=dev

LABEL org.opencontainers.image.title="eep" \
    org.opencontainers.image.description="Envoy error pages Proxy-Wasm extension" \
    org.opencontainers.image.version="${VERSION}" \
    org.opencontainers.image.revision="${REVISION}" \
    org.opencontainers.image.source="https://github.com/ishioni/eep"

COPY --from=builder /workspace/main.wasm /plugin.wasm
