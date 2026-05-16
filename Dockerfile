# Use standard Go 1.26 for building WASM
FROM golang:1.26-bookworm AS builder

WORKDIR /src

# Version can be overridden at build time with --build-arg VERSION=x.y.z
# If not provided, defaults to 'dev'
ARG VERSION=dev
ARG TARGETARCH
ARG WASM_TOOLS_VERSION=1.248.0

# Install wasm-tools for the post-link WASI import stubbing step.
RUN apt-get update \
    && apt-get install -y --no-install-recommends ca-certificates curl tar \
    && rm -rf /var/lib/apt/lists/* \
    && case "${TARGETARCH}" in \
        "amd64"|"") WASM_TOOLS_ARCH="x86_64-linux" ;; \
        "arm64") WASM_TOOLS_ARCH="aarch64-linux" ;; \
        *) echo "unsupported TARGETARCH: ${TARGETARCH}"; exit 1 ;; \
    esac \
    && curl -fsSL "https://github.com/bytecodealliance/wasm-tools/releases/download/v${WASM_TOOLS_VERSION}/wasm-tools-${WASM_TOOLS_VERSION}-${WASM_TOOLS_ARCH}.tar.gz" -o /tmp/wasm-tools.tar.gz \
    && tar -xzf /tmp/wasm-tools.tar.gz -C /tmp \
    && install /tmp/wasm-tools-${WASM_TOOLS_VERSION}-${WASM_TOOLS_ARCH}/wasm-tools /usr/local/bin/wasm-tools \
    && rm -rf /tmp/wasm-tools* \
    && wasm-tools --version

COPY . .

# Build the WASM binary using the new Go WASIP1 target, then post-process it
# for Envoy's limited WASI import surface. The final main.wasm is patched.
RUN make build VERSION=${VERSION}

# Use a minimal base image for the OCI artifact
FROM scratch

# Envoy looks for the wasm file
COPY --from=builder /src/main.wasm ./plugin.wasm
