---
name: salt-clientmanager
description: Type-safe Go HTTP client with generic response decoding, 7 auth schemes, multipart uploads, proxy/TLS configuration, request validation, and automatic APM tracing via the SALT-Indonesia/salt-pkg clientmanager library. Use when making outbound HTTP calls with this library.
---

# clientmanager — type-safe Go HTTP client

`clientmanager` is a generic HTTP client library. A single `Call[T]` function
dispatches a request and decodes the response body into the type you specify --
JSON, XML, or raw string -- without any manual (de)serialization. Every call is
traced through `logmanager` automatically. Request bodies are validated with
`go-playground/validator` struct tags before they are sent.

## Install

```bash
go get github.com/SALT-Indonesia/salt-pkg/clientmanager
# pulls in logmanager — every Call requires a logmanager Transaction in the context
```

```go
import (
    "github.com/SALT-Indonesia/salt-pkg/clientmanager"
    "github.com/SALT-Indonesia/salt-pkg/logmanager"
)
```

## Minimal GET request

```go
type ProductResponse struct {
    Products []Product `json:"products"`
}

app := logmanager.NewApplication()
txn := app.Start("fetch-products", "cli", logmanager.TxnTypeOther)
ctx := txn.ToContext(context.Background())
defer txn.End()

res, err := clientmanager.Call[ProductResponse](
    ctx,
    "https://api.example.com/products",
)
if err != nil {
    log.Fatal(err)
}
fmt.Println(res.StatusCode)       // 200
fmt.Println(len(res.Body.Products))
fmt.Println(res.IsSuccess())      // true
```

`Call[T]` is a package-level function that uses a shared `*http.Client`. It
defaults to `GET`. The context must carry a `logmanager` transaction; otherwise
the call returns an error.

The return type is `*BaseResponse[T]`, which exposes `StatusCode`, `Body` (the
decoded `T`), `Raw` (original bytes), and `IsSuccess()` (true for 2xx).

## Scoped client

For repeated calls to the same host, or when you need an isolated HTTP client
(e.g. for proxy, custom TLS), use `New[T]` to create a `ClientManager[T]`.

```go
cm := clientmanager.New[ProductResponse](
    clientmanager.WithHost("https://api.example.com"),
    clientmanager.WithTimeout(30 * time.Second),
)

res, err := cm.Call(ctx, "/products")
res, err = cm.Call(ctx, "/products/featured")
```

Options passed to `New` become defaults. Options passed to the instance `Call`
are applied on top, so you can override per-request.

## POST with request body

```go
type CreateRequest struct {
    Title string  `json:"title" validate:"required"`
    Price float64 `json:"price" validate:"required"`
}
type CreateResponse struct {
    ID uint64 `json:"id"`
}

res, err := clientmanager.Call[CreateResponse](
    ctx,
    "https://api.example.com/products",
    clientmanager.WithMethod(http.MethodPost),
    clientmanager.WithRequestBody(&CreateRequest{Title: "Widget", Price: 9.99}),
)
```

The request body is JSON-encoded by default. Struct fields tagged with
`validate:"..."` are checked before the request is sent; a validation error
stops the call.

## Query parameters

```go
res, err := clientmanager.Call[ProductResponse](
    ctx,
    "https://api.example.com/products",
    clientmanager.WithURLValues(url.Values{
        "limit":  {"10"},
        "skip":   {"20"},
        "select": {"title,price"},
    }),
)
// GET https://api.example.com/products?limit=10&skip=20&select=title%2Cprice
```

Query parameters can also be combined with `WithHost` and passed to a scoped
client's `Call`.

## Authentication

Pass an auth function via `WithAuth`. Seven schemes are included.

```go
// Bearer token
clientmanager.WithAuth(clientmanager.AuthBearer("eyJhbGci..."))

// Basic auth
clientmanager.WithAuth(clientmanager.AuthBasic("user", "pass"))

// API key (header or query param)
clientmanager.WithAuth(clientmanager.AuthAPIKey("X-API-Key", "abc123", false))

// JWT with claims
clientmanager.WithAuth(clientmanager.AuthJWT(
    "secret",
    jwt.SigningMethodHS256,
    clientmanager.AuthJWTClaims{
        Sub: "user-42",
        Iss: "my-service",
        Exp: time.Now().Add(time.Hour),
        Extra: map[string]any{"role": "admin"},
    },
))

// AWS Signature v4
clientmanager.WithAuth(clientmanager.AuthAWS(clientmanager.AWSParameters{
    Key:     "AKID...",
    Secret:  "wJalrX...",
    Service: "execute-api",
    Region:  "ap-southeast-3",
}))
```

Additional auth options: `AuthHawk`, `AuthESB`. For digest auth and NTLM, use
`WithAuthDigest(username, password)` and `WithAuthNTLM(auth)` instead of
`WithAuth`. OAuth1 and OAuth2 are wired through `WithOAuth1` and `WithOAuth2`,
which replace the underlying HTTP client.

## File uploads (multipart)

```go
imageBytes, _ := os.ReadFile("logo.png")

res, err := clientmanager.Call[UploadResponse](
    ctx,
    "https://api.example.com/upload",
    clientmanager.WithMethod(http.MethodPost),
    clientmanager.WithMultipartForm(clientmanager.MultipartForm{
        Files: map[string]clientmanager.FilePart{
            "file": {
                Filename:    "logo.png",
                Content:     imageBytes,
                ContentType: "image/png",
            },
        },
        Values: map[string]string{
            "category": "images",
        },
    }),
)
```

`WithMultipartForm` accepts in-memory file content with custom MIME types and
string form fields. The older `WithFiles` (disk paths only) is deprecated.

## Form URL-encoded

```go
type LoginForm struct {
    Username string `json:"username"`
    Password string `json:"password"`
}

res, err := clientmanager.Call[any](
    ctx,
    "https://api.example.com/login",
    clientmanager.WithRequestBody(&LoginForm{Username: "alice", Password: "s3cr3t"}),
    clientmanager.WithMethod(http.MethodPost),
    clientmanager.WithFormURLEncoded(),
)
```

The struct is JSON-marshalled then converted to `application/x-www-form-urlencoded`.

## XML responses

Response decoding is automatic based on the `Content-Type` header. If the
response is `application/xml` or `text/xml`, the body is XML-unmarshalled into
`T`.

```go
type RSS struct {
    Channel struct {
        Title string `xml:"title"`
    } `xml:"channel"`
}

res, err := clientmanager.Call[RSS](ctx, "https://example.com/feed.xml")
fmt.Println(res.Body.Channel.Title)
```

When `T` is `string`, the raw body is returned without any decoding.

## Proxy and TLS

```go
// Proxy (returns (Option, error); use a scoped client to avoid shared state)
proxy, err := clientmanager.WithProxy("http://proxy.internal:8080")
if err != nil {
    log.Fatal(err)
}
cm := clientmanager.New[string](proxy)
res, err := cm.Call(ctx, "https://example.com")

// Skip TLS verification (development only)
clientmanager.WithInsecure()

// Client certificates
clientmanager.WithCertificates(tlsCert)

// Custom root CA
clientmanager.WithRootCertificate(rootCAPool)

// Disable HTTP/2
clientmanager.WithDisabledHTTP2()
```

Connection tuning options: `WithConnectionLimit(maxIdle, maxIdlePerHost,
maxPerHost)`, `WithIdleConnTimeout`, `WithTLSHandshakeTimeout`,
`WithExpectContinueTimeout`, `WithDialContext(timeout, keepAlive)`.

## Key options

- `WithMethod(m)` -- HTTP method (default `GET`)
- `WithHost(url)` -- base URL, prepended to endpoint
- `WithRequestBody(v)` -- request body (JSON by default)
- `WithURLValues(v)` -- query parameters
- `WithHeaders(h)` -- custom headers (`http.Header`)
- `WithTimeout(d)` -- per-request timeout
- `WithAuth(auth)` -- authentication function
- `WithMultipartForm(form)` -- file upload with form fields
- `WithFormURLEncoded()` -- URL-encoded form body
- `WithProxy(url)` -- HTTP proxy
- `WithInsecure()` -- skip TLS verification
- `WithCertificates(certs...)` -- client TLS certificates
- `WithRootCertificate(pool)` -- custom root CA
- `WithDisabledHTTP2()` -- force HTTP/1.1

## More

See [`clientmanager/README.md`](../../clientmanager/README.md) for the full
reference including OAuth1/OAuth2 setup, NTLM, digest auth, and testing
patterns.
