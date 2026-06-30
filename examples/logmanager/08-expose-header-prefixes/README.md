# Expose Header Prefixes Example

Selectively log infrastructure headers by **wildcard prefix** — even in production
where `debug=false`.

Teams behind Cloudflare / CloudFront usually want only a handful of infrastructure
headers (`CF-Connecting-IP`, `CF-Ray`, `X-Amz-Cf-Id`, `X-Amzn-Trace-Id`, …) in their
logs, without enumerating each one by exact name. `WithExposeHeaders` accepts a
trailing-`*` wildcard so a single entry like `"CF-*"` exposes every header sharing
that prefix.

## Features

- Wildcard/prefix header exposure via `WithExposeHeaders("CF-*", ...)`
- Works in **production mode** (`WithEnvironment("production")`, `debug=false`)
- Case-insensitive matching (Go canonicalizes `CF-Connecting-IP` → `Cf-Connecting-Ip`)
- Non-matching headers (e.g. `User-Agent`) are dropped from logs

## Key Concept

```go
app := logmanager.NewApplication(
    logmanager.WithAppName("expose-header-prefixes-example"),
    logmanager.WithEnvironment("production"), // debug=false, no WithDebug()
    logmanager.WithExposeHeaders("CF-*", "X-Amz-Cf-*", "X-Amzn-*", "X-Request-Id"),
)
```

Entries without `*` keep exact-match behavior; entries ending in `*` match by
prefix. A bare `"*"` exposes all headers (equivalent to debug mode for headers).

## Running

```bash
cd 08-expose-header-prefixes && go run main.go   # Port 8008
```

Then send a request with a mix of matching and non-matching headers:

```bash
curl http://localhost:8008/ping \
  -H "CF-Ray: 8abc" \
  -H "CF-Connecting-IP: 1.2.3.4" \
  -H "X-Amzn-Trace-Id: Root=1-xyz" \
  -H "X-Amz-Cf-Id: zzz" \
  -H "User-Agent: noise"
```

## Expected Result

The logged `headers` object contains only the prefix-matched headers — `User-Agent`
is omitted — even though `debug` is `false`:

```json
{
  "headers": {
    "Cf-Ray": "8abc",
    "Cf-Connecting-Ip": "1.2.3.4",
    "X-Amzn-Trace-Id": "Root=1-xyz",
    "X-Amz-Cf-Id": "zzz"
  },
  "type": "http"
}
```
