# Changelog

## [0.10.0] - 2026-08-11

### Added
- `WithDialerControl(control func(network, address string, c syscall.RawConn) error) Option` —
  sets the dialer's Control callback for SSRF protection (e.g. rejecting connections to private
  or reserved IP ranges before connecting). Unwraps `*digest.Transport`, `ntlmssp.Negotiator`,
  `*oauth1.Transport`, and `*oauth2.Transport` to reach the underlying `*http.Transport`.
  Composes with `WithDialContext`.
- `WithMaxResponseBytes(max int64) Option` — limits the response body read by `Call` and
  `CallBytes` via `io.LimitReader`. Bodies exceeding the limit are truncated at the limit.
- `BaseResponse.Header http.Header` — exposes response headers so callers can inspect
  `Content-Type` etc. without switching to `CallStream`.

### Changed
- `getMultipartFormBody` no longer returns an error (writes to `bytes.Buffer` never fail);
  signature simplified to `(*bytes.Buffer, string)`.
- `callBytes` now skips `io.ReadAll` errors for consistency with `call`.
- 100% statement coverage achieved.

## [0.9.0] - 2026-07-28

### Added
- `CallStream(ctx, endpoint, options...) (*StreamResponse, error)` — sends a request and returns
  the raw response body as an `io.ReadCloser` without buffering or decoding. Use this for SSE
  streaming, binary downloads, chunked responses, and proxy-passthrough patterns. The caller MUST
  call `defer streamResp.Close()` to end the logmanager segment and release the body.
- `CallBytes(ctx, endpoint, options...) (*BaseResponse[[]byte], error)` — sends a request and
  returns the full response body as raw `[]byte` without JSON/XML decode. Use this for binary
  file downloads and provider-specific response transformations where the shape is not JSON.
- `WithBodyReader(body io.Reader, contentType string) Option` — sets a raw `io.Reader` as the
  request body, bypassing `WithRequestBody` JSON marshalling. Takes precedence over all other
  body-setting options. Required for proxy-style patterns where the body is already serialised.
- `StreamResponse` type with `StatusCode`, `Header`, `Body io.ReadCloser`, `IsSuccess() bool`,
  and `Close() error` (idempotent; ends the logmanager ApiSegment and closes the body).
- `ClientManager.CallStream` and `ClientManager.CallBytes` methods for structured usage with
  pre-configured clients.

## [0.8.0] - 2026-06-18

### Added
- `WithResponseHeaderTimeout(timeout time.Duration) Option` to set the HTTP transport `ResponseHeaderTimeout`.

### Fixed
- `WithTimeout` now also raises `Transport.ResponseHeaderTimeout` when the requested timeout exceeds the current value.
  Previously `ResponseHeaderTimeout` was hardcoded at 5s in the default transport, causing requests to slow upstreams
  (LLM, report generation, long DB queries) to fail with `timeout awaiting response headers` regardless of `WithTimeout` value.

### Changed
- Updated logmanager dependency from v1.41.0 to v1.43.1
  - Fixes a data race / `concurrent map writes` panic on `Transaction.txnRecords` under concurrent fanout
  - Rolls up `WithSkipHeaders()`, split-level log output, and the async nil-deref fix

## [0.5.2] - 2026-02-12

### Changed
- Updated logmanager dependency from v1.38.1 to v1.41.0

## [0.5.1]

### Previous Changes
- See git history for previous changes