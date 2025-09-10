# HTTP Client Manager

An HTTP client manager to request HTTP endpoints.

## Options

| Option                    | Example                                                      | Description                                              |
|---------------------------|--------------------------------------------------------------|----------------------------------------------------------|
| WithCertificates          | `WithCertificates(cert)`                                     | Add certificates.                                        |
| WithRootCertificate       | `WithRootCertificate(rootCA)`                                | Add a root certificate.                                  |
| WithFiles                 | `WithFiles(map[string]string{"image":"image.jpg"})`          | Include files, field name with its path, in the request. |
| WithFormURLEncoded        | `WithFormURLEncoded()`                                       | Send the request in URL-encoded form.                    |
| WithHeaders               | `WithHeaders(map[string][]string{"X-Trace-ID": {"abc123"}})` | Include headers in the request.                          |
| WithHost                  | `WithHost("https://httpbin.org")`                            | Set the host for the request. Optional.                  |
| WithInsecure              | `WithInsecure()`                                             | Allow the request to an insecure host.                   |
| WithMethod                | `WithMethod(http.MethodPost)`                                | Set the HTTP Method for the request. Default is GET.     |
| WithRequestBody           | `WithRequestBody(req)`                                       | Set the request body.                                    |
| WithURLValues             | `WithURLValues(urlValues)`                                   | Set the request URL values.                              |
| WithTimeout               | `WithTimeout(time.Second)`                                   | Set the default timeout.                                 |
| WithProxy                 | `proxy, err := WithProxy("http://localhost:8080")`           | Set the proxy for the request.                           |
| WithConnectionLimit       | `WithConnectionLimit(1000, 1000, 100)`                       | Set the connection limit.                                |
| WithIdleConnTimeout       | `WithIdleConnTimeout(time.Minute)`                           | Set the idle connection timeout.                         |
| WithTLSHandshakeTimeout   | `WithTLSHandshakeTimeout(5 * time.Second)`                   | Set the TLS handshake timeout.                           |
| WithExpectContinueTimeout | `WithExpectContinueTimeout(5 * time.Second)`                 | Set the expect continue timeout.                         |
| WithDialContext           | `WithDialContext(10 * time.Second, time.Minute)`             | Set the dial context.                                    |
| WithAuth                  | `WithAuth(AuthBasic("user123", "pass123"))`                  | Set the authorization for the request.                   |
| WithAuthDigest            | `WithAuthDigest("user123", "pass123")`                       | Set the digest auth for the request.                     |
| WithOAuth1                | `WithOAuth1(OAuth1Parameters{"a", "b", "c", "d"})`           | Set the OAuth1 request.                                  |
| WithOAuth2                | `WithOAuth2(OAuth2Parameters[string]{"a", ""})`              | Set the OAuth2 request.                                  |
| WithAuthNTLM              | `WithAuthNTLM(AuthBasic("user123", "pass123"))`              | Set the NTLM request.                                    |

## Authorizations

You can find authorization samples on `salt-pkg/clientmanager/examples/auth`.

| Auth       | Example                                                      | Description   |
|------------|--------------------------------------------------------------|---------------|
| AuthBasic  | `AuthBasic("user123", "pass123")`                            | Basic Auth    |
| AuthBearer | `AuthBearer("pass123")`                                      | Bearer Token  |
| AuthAPIKey | `AuthAPIKey("foo", "bar", false)`                            | API Key       |
| AuthJWT    | `AuthJWT("secret", jwt.SigningMethodHS256, AuthJWTClaims{})` | JWT Bearer    |
| AuthHawk   | `AuthHawk("id", "key", nil)`                                 | Hawk          |
| AuthAWS    | `AuthAWS(AWSParameters{})`                                   | AWS Signature |
| AuthESB    | `AuthESB("user123", "pass123")`                              | TSEL ESB      |

### NTLM

To work with NTLM authentication, you need to pass `AuthBasic` after `WithAuthNTLM`.

## Usage

### Simple

You can find the simplest sample on `salt-pkg/clientmanager/examples/main.go`.

```go
type Request struct {
	Title string  `json:"title"`
	Price float64 `json:"price"`
}

type Response struct {
	ID uint64 `json:"id"`
}

func main() {
	req := &Request{
		Title: "My Product",
		Price: 123.45,
	}
	res, err := clientmanager.Call[Response](
		context.Background(),
		"https://httpbin.org/post",
        clientmanager.WithRequestBody(req),
		clientmanager.WithMethod(http.MethodPost),
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("HTTP Status:", res.StatusCode)
	fmt.Println("Response Struct:", res.Body)
	fmt.Println("Raw Body:", string(res.Raw))
	fmt.Println("Success:", res.IsSuccess())
}
```

### Structured

You can find the structured sample on `salt-pkg/clientmanager/examples/dummyjson`, where it uses `WithHost` to avoid host repetition.
Refer to `salt-pkg/clientmanager/examples/dummyjson/product/http_repository.go` as an example if you want to store options to an object as default options or options with higher scope, but can be overridden on the `Call` method.

```
.
├── product/
│   ├── request.go
│   ├── response.go
│   ├── repository.go
│   ├── http_repository.go
│   └── product.go
└── main.go
```

#### Create a request if needed

Example of `request.go`.

```go
type Request struct {
	Title string  `json:"title"`
	Price float64 `json:"price"`
	Stock uint64  `json:"stock"`
}
```

#### Create a response if needed

Example of `response.go`.

```go
type Response struct {
	Products []Product `json:"products"`
}
```

#### Call with URL parameters 

Example.

```go
res, err := clientmanager.Call[Response](
    context.Background(),
    "https://dummyjson.com/products",
    clientmanager.WithURLValues(url.Values{
        "limit":  {"10"},
        "skip":   {"10"},
        "select": {"title,price"},
    }),
)
if err != nil {
    log.Panic(err)
}
```

#### Call with a proxy.

We recommend that you create a new instance using `clientmanager.New()` instead of providing a proxy to the global `Call` method.

Example.

```go
proxy, err := clientmanager.WithProxy("http://localhost:8080") // create the proxy first
if err != nil {
    log.Panic(err)
}

clientManager := clientmanager.New[string]() // we create a new client manager to create a different HTTP client
res, err := clientManager.Call[string]( // call the `Call` method from the `clientManager` instead of the global
    context.Background(),
    "https://dummyjson.com/",
    proxy, // then put your proxy
)
if err != nil {
    log.Panic(err)
}
```

#### Call without request 

Example.

```go
res, err := clientmanager.Call[Response](
    context.Background(),
    "https://dummyjson.com/products",
)
if err != nil {
    log.Panic(err)
}
```

#### Call without response

Example.

```go
req := &Request{
    Title: "My First Product",
    Price: 5000,
    Stock: 9,
}
if _, err := clientmanager.Call[any](
    context.Background(), 
    "https://dummyjson.com/products/add", 
    clientmanager.WithRequestBody(req),
    clientmanager.WithMethod(http.MethodPost),
); err != nil {
    log.Panic(err)
}
```

### String

Here is a sample of parsing a string response, `salt-pkg/clientmanager/examples/strings/main.go`.

### XML

Here is a sample of parsing an XML response, `salt-pkg/clientmanager/examples/xml/main.go`.

### Classic Form (x-www-form-urlencoded)

Here is a sample of a classic form, `salt-pkg/clientmanager/examples/classicform/main.go`.

### File Upload

Here is a sample of file uploading, `salt-pkg/clientmanager/examples/fileupload/main.go`.

## Validation

The `clientmanager` is using [https://github.com/go-playground/validator](https://github.com/go-playground/validator) to validate the request. You can put the validator tags on your request `struct` if you want to validate your request.

Example.

```go
type Request struct {
	Title string  `json:"title" validate:"required"`
	Price float64 `json:"price" validate:"required"`
}
```

## Test

You don't need to test your infrastructure layers. You make sure that your domain layers are tested. Here is an example.
```go 
type OrderUseCase struct {
  courierService CourierService
}

func (u OrderUseCase) Pay(ctx context.Context, order Order) error {
  ...
  if err := u.courierService.Send(ctx, order); err != nil {
    return err
  }
  ...
  return nil
}

type CourierService interface {
  Send(context.Context, Order) error
}

type gojekCourierService struct {
    host string
}

func (s gojekCourierService) Send(ctx context.Context, order Order) error {
  ...
  if _, err := clientmanager.Call[any](
    context.Background(), 
    fmt.Sprintf("%s/send", s.host), 
    clientmanager.WithRequestBody(toRequest(Order)),
    clientmanager.WithMethod(http.MethodPost),
  ); err != nil {
    return err
  }
  ...
  return nil
}

type CourierServiceMock struct {
    mock.Mock
}

func (m CourierServiceMock) Send(ctx context.Context, order Order) error {
  args := m.Called(ctx, order)
  return args.Error(0)
}
```
You only need to pass `CourierServiceMock` to the `OrderUseCase` in tests.

In case you need to test the `clientmanager.Call`, which is unnecessary, here is an example.
```go
func TestSend(t *testing.T) {
    app := logmanager.NewApplication()
    txn := app.Start("test", "cli", logmanager.TxnTypeOther)
    ctx := txn.ToContext(context.Background())
    defer txn.End()

    ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    }))
    defer ts.Close()

    var courierService CourierService = &gojekCourierService{
        host: ts.URL
    }

    assert.NoError(t, courierService.Send(ctx, data.Order()))
}
```

The best way to test infrastructure layers is by doing **integration tests**. We send real requests to the sandbox. We can mock the responses if the sandbox is not available, like the code above.
