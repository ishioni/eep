# WASI Import Stubbing for Go Proxy-Wasm Plugins

## Context

This project builds an Envoy Proxy-Wasm HTTP filter with the upstream Go compiler:

```text
GOOS=wasip1 GOARCH=wasm go build -buildmode=c-shared
```

The plugin intentionally wants to keep compatibility with the upstream `error-pages` project, whose templates are rendered with Go's `text/template`. That matters because custom templates can use real Go template syntax and helper functions, not just a tiny subset.

Unfortunately, importing `text/template` pulls in `os`, and `os` pulls in Go's WASI filesystem support. The resulting `main.wasm` imports several `wasi_snapshot_preview1` filesystem/socket functions that Envoy's V8 Proxy-Wasm runtime does not provide.

The initial observed Envoy failure was:

```text
Failed to load Wasm module due to a missing import: wasi_snapshot_preview1.fd_filestat_set_size
```

Testing with Go 1.26 and Envoy 1.38 showed that this is still true:

- Go 1.26 still emits the unsupported WASI imports.
- Envoy 1.38 still does not provide them.

## Why this is not fixed in the Go SDK

The missing symbol is not a Proxy-Wasm SDK hostcall. It comes from the Go standard library itself.

In Go 1.26, it is declared in the Go source tree at `src/syscall/fs_wasip1.go`:

```text
//go:wasmimport wasi_snapshot_preview1 fd_filestat_set_size
func fd_filestat_set_size(...)
```

That means the compiled guest module asks the host runtime, Envoy, to provide `wasi_snapshot_preview1.fd_filestat_set_size` before the module can instantiate.

The `proxy-wasm-go-sdk` is linked into the guest module. It can import hostcalls from Envoy, but it cannot satisfy imports required by the same guest module. Adding a Go function to the SDK does not remove or satisfy a `wasi_snapshot_preview1` import.

The SDK test harness can provide WASI imports because it runs the module under `wazero`, where the test harness is the host. Envoy is the host in production, so Envoy controls which WASI imports are available.

## Observed import surface

Using `wasm-tools print main.wasm`, the unpatched Go-built module imported these WASI functions before Envoy hostcalls:

```text
(import "wasi_snapshot_preview1" "sched_yield" ...)
(import "wasi_snapshot_preview1" "proc_exit" ...)
(import "wasi_snapshot_preview1" "args_get" ...)
(import "wasi_snapshot_preview1" "args_sizes_get" ...)
(import "wasi_snapshot_preview1" "clock_time_get" ...)
(import "wasi_snapshot_preview1" "environ_get" ...)
(import "wasi_snapshot_preview1" "environ_sizes_get" ...)
(import "wasi_snapshot_preview1" "fd_write" ...)
(import "wasi_snapshot_preview1" "random_get" ...)
(import "wasi_snapshot_preview1" "poll_oneoff" ...)
(import "wasi_snapshot_preview1" "fd_close" ...)
(import "wasi_snapshot_preview1" "fd_filestat_set_size" ...)
(import "wasi_snapshot_preview1" "fd_pread" ...)
(import "wasi_snapshot_preview1" "fd_pwrite" ...)
(import "wasi_snapshot_preview1" "fd_read" ...)
(import "wasi_snapshot_preview1" "fd_readdir" ...)
(import "wasi_snapshot_preview1" "fd_seek" ...)
(import "wasi_snapshot_preview1" "fd_filestat_get" ...)
(import "wasi_snapshot_preview1" "fd_write" ...)
(import "wasi_snapshot_preview1" "fd_sync" ...)
(import "wasi_snapshot_preview1" "path_filestat_get" ...)
(import "wasi_snapshot_preview1" "fd_fdstat_get" ...)
(import "wasi_snapshot_preview1" "fd_fdstat_set_flags" ...)
(import "wasi_snapshot_preview1" "fd_prestat_get" ...)
(import "wasi_snapshot_preview1" "fd_prestat_dir_name" ...)
(import "wasi_snapshot_preview1" "sock_accept" ...)
(import "wasi_snapshot_preview1" "sock_shutdown" ...)
(import "wasi_snapshot_preview1" "path_filestat_get" ...)
```

Envoy can resolve the runtime-ish imports at the top, but not the filesystem/socket tail. After stubbing only `fd_filestat_set_size`, Envoy failed on the next missing import, `fd_pread`, confirming that more than one import needs attention.

## Experimental post-link patch

The successful experiment was to post-process the generated WASM module:

1. Convert `main.wasm` to WAT using `wasm-tools print`.
2. Remove unsupported WASI filesystem/socket imports.
3. Insert internal stub functions with the same signatures.
4. Rewrite direct call indexes so existing calls target the right imports/functions after the import section changes.
5. Parse and validate the patched WAT back to WASM with `wasm-tools parse` and `wasm-tools validate`.

After this patch, Envoy loaded the module and the plugin initialized with `text/template` still linked in.

## Stubbed imports

The experimental patch removed and stubbed this import range:

| Import | Default stub return |
| --- | --- |
| `fd_filestat_set_size` | `ENOSYS` (`52`) |
| `fd_pread` | `ENOSYS` (`52`) |
| `fd_pwrite` | `ENOSYS` (`52`) |
| `fd_read` | `ENOSYS` (`52`) |
| `fd_readdir` | `ENOSYS` (`52`) |
| `fd_seek` | `ENOSYS` (`52`) |
| `fd_filestat_get` | `ENOSYS` (`52`) |
| duplicate later `fd_write` | `ENOSYS` (`52`) |
| `fd_sync` | `ENOSYS` (`52`) |
| `path_filestat_get` | `ENOSYS` (`52`) |
| `fd_fdstat_get` | `0` |
| `fd_fdstat_set_flags` | `0` |
| `fd_prestat_get` | `EBADF` (`8`) |
| `fd_prestat_dir_name` | `ENOSYS` (`52`) |
| `sock_accept` | `ENOSYS` (`52`) |
| `sock_shutdown` | `ENOSYS` (`52`) |
| duplicate later `path_filestat_get` | `ENOSYS` (`52`) |

### Why some stubs return special values

Most unsupported operations should fail with `ENOSYS` because the plugin should not be using real files or sockets inside Envoy.

However, Go's WASI `syscall` package runs initialization before the Proxy-Wasm plugin starts:

- It calls `SetNonblock(0, true)`, `SetNonblock(1, true)`, and `SetNonblock(2, true)`.
  - This uses `fd_fdstat_get` and `fd_fdstat_set_flags`.
  - Returning `ENOSYS` here causes a Go runtime panic during `_initialize`.
  - Returning `0` lets initialization continue.
- It scans preopened directories starting at fd `3` with `fd_prestat_get`.
  - Returning `EBADF` means there are no more preopened directories.
  - Returning `ENOSYS` causes a panic like `fd_prestat: Not implemented on wasip1`.

A failed experiment where all stubs returned `ENOSYS` produced:

```text
Function: _initialize failed: Uncaught Error: restricted_callback
Proxy-Wasm plugin in-VM backtrace:
  runtime.poll_oneoff
  runtime.usleep
  runtime.freezetheworld
  runtime.startpanic_m
  runtime.fatalpanic.func1
  runtime.gopanic
  syscall.init.1
```

Changing `fd_fdstat_get`, `fd_fdstat_set_flags`, and `fd_prestat_get` as described above allowed initialization to complete.

## Validation result

After stubbing the unsupported imports, Envoy started and logged plugin initialization:

```text
wasm log: WASM Error Pages Plugin initialized
wasm log: Error page template loaded: theme=connection, show_details=true
```

A request to `/500` reached the plugin and was intercepted. Rendering then failed because the current Go function map is incomplete relative to the bundled `error-pages` templates:

```text
failed to render error page: failed to parse template: template: errorpage:322: function "ingress_name" not defined
```

This is a separate template-compatibility issue, not a WASI import issue.

## Important caveats

- This is a post-link workaround, not a proper Envoy or Go SDK fix.
- The exact import indexes can change when Go, dependencies, or code change.
- A production patcher should parse the WASM structure, not rely blindly on fixed line numbers in WAT.
- Stubs should be safe only if the plugin never performs actual file/socket operations at runtime.
- If a future template helper uses real filesystem or environment behavior, it may fail or behave unexpectedly.
- `text/template` compatibility still requires matching the `error-pages` helper functions and token surface.

## Recommended next implementation direction

If continuing this approach, implement a deterministic build-time patch step:

1. Build `main.wasm` with Go.
2. Run a small patcher that removes unsupported WASI imports and injects stubs.
3. Validate the output with `wasm-tools validate` or an equivalent library.
4. Fail the build if an unexpected unsupported WASI import remains.
5. Keep tests that assert `wasm-tools print main.wasm` no longer contains imports like `fd_filestat_set_size`.

Then fix template compatibility separately by aligning this project's template data and functions with `error-pages/internal/template`.
