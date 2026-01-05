# HTTP Redirects

The httpmanager module provides comprehensive HTTP redirect functionality similar to Gin's implementation. You can redirect users to different URLs with various HTTP status codes.

## Basic Redirect Functions

The module provides utility functions for common redirect scenarios:

```go
// Basic redirect with custom status code
httpmanager.Redirect(w, r, http.StatusFound, "http://example.com")

// Redirect with 302 (Found) status code
httpmanager.RedirectToURL(w, r, "http://example.com")

// Redirect with 301 (Moved Permanently) status code
httpmanager.RedirectPermanent(w, r, "http://example.com")
```

## Context-Based Redirects

For more convenient usage, you can use the `Context` type which provides Gin-like redirect methods:

```go
// Using RedirectHandler for context-based redirects
redirectHandler := httpmanager.NewRedirectHandler(http.MethodGet, func(c *httpmanager.Context) {
    // Redirect with custom status code
    c.Redirect(http.StatusFound, "http://example.com")

    // Or use convenience methods
    c.RedirectToURL("http://example.com")         // 302 Found
    c.RedirectPermanent("http://example.com")     // 301 Moved Permanently
})

server.Handle("/old-path", redirectHandler.WithMiddleware())
```

## Redirect Handler

The `RedirectHandler` provides a specialized handler for redirect operations:

```go
// Create a redirect handler
redirectHandler := httpmanager.NewRedirectHandler(http.MethodGet, func(c *httpmanager.Context) {
    // Access query parameters
    targetURL := c.GetQueryParams().Get("redirect_to")
    if targetURL == "" {
        targetURL = "http://default-example.com"
    }

    // Redirect to the target URL
    c.RedirectToURL(targetURL)
})

// Add middleware if needed
redirectHandler.Use(authMiddleware, loggingMiddleware)

// Register with the server
server.GET("/redirect", redirectHandler.WithMiddleware())
```

## Dynamic Redirects with Path Parameters

You can create dynamic redirects using path parameters:

```go
// Redirect handler with path parameters
redirectHandler := httpmanager.NewRedirectHandler(http.MethodGet, func(c *httpmanager.Context) {
    // Get path parameters
    pathParams := c.GetPathParams()
    userID := pathParams.Get("id")

    // Get query parameters
    queryParams := c.GetQueryParams()
    section := queryParams.Get("section")

    // Build redirect URL
    redirectURL := fmt.Sprintf("https://newdomain.com/users/%s", userID)
    if section != "" {
        redirectURL += "?section=" + section
    }

    c.RedirectPermanent(redirectURL)
})

// Register with path parameter
server.GET("/old-user/{id}", redirectHandler.WithMiddleware())
```

## Redirect Status Codes

The module supports all standard HTTP redirect status codes:

| Status Code | Constant                        | Method              | Description                          |
|-------------|--------------------------------|---------------------|--------------------------------------|
| 301         | `http.StatusMovedPermanently`  | `RedirectPermanent` | Permanent redirect                   |
| 302         | `http.StatusFound`             | `RedirectToURL`     | Temporary redirect (default)         |
| 303         | `http.StatusSeeOther`          | `Redirect`          | See other (POST to GET)              |
| 307         | `http.StatusTemporaryRedirect` | `Redirect`          | Temporary redirect (method preserved) |
| 308         | `http.StatusPermanentRedirect` | `Redirect`          | Permanent redirect (method preserved) |

## Complete Redirect Examples

### Example 1: Simple Domain Migration

```go
func main() {
    server := httpmanager.NewServer(logmanager.NewApplication())
    server.EnableCORS([]string{"*"}, nil, nil, false)

    // Redirect all old domain traffic to new domain
    migrationHandler := httpmanager.NewRedirectHandler(http.MethodGet, func(c *httpmanager.Context) {
        oldPath := c.Request.URL.Path
        c.RedirectPermanent("https://newdomain.com" + oldPath)
    })

    server.GET("/old-api/{path:.*}", migrationHandler.WithMiddleware())

    log.Panic(server.Start())
}
```

### Example 2: Conditional Redirects

```go
func main() {
    server := httpmanager.NewServer(logmanager.NewApplication())
    server.EnableCORS([]string{"*"}, nil, nil, false)

    // Conditional redirect based on user agent
    mobileRedirectHandler := httpmanager.NewRedirectHandler(http.MethodGet, func(c *httpmanager.Context) {
        userAgent := c.GetHeader("User-Agent")

        if strings.Contains(strings.ToLower(userAgent), "mobile") {
            c.RedirectToURL("https://m.example.com" + c.Request.URL.Path)
        } else {
            c.RedirectToURL("https://www.example.com" + c.Request.URL.Path)
        }
    })

    server.GET("/", mobileRedirectHandler.WithMiddleware())

    log.Panic(server.Start())
}
```

### Example 3: POST to GET Redirect

```go
func main() {
    server := httpmanager.NewServer(logmanager.NewApplication())
    server.EnableCORS([]string{"*"}, nil, nil, false)

    // Handle form submission with redirect
    formHandler := httpmanager.NewRedirectHandler(http.MethodPost, func(c *httpmanager.Context) {
        // Process form data here (if needed)
        // Then redirect to success page
        c.Redirect(http.StatusSeeOther, "/success")
    })

    server.POST("/submit-form", formHandler.WithMiddleware())

    log.Panic(server.Start())
}
```

## Context Methods

The `Context` type provides the following redirect methods:

| Method                                  | Description                                    |
|----------------------------------------|------------------------------------------------|
| `Redirect(code int, location string)`  | Redirect with custom HTTP status code         |
| `RedirectToURL(location string)`       | Redirect with 302 status (Found)              |
| `RedirectPermanent(location string)`   | Redirect with 301 status (Moved Permanently)  |
| `GetQueryParams()`                     | Access query parameters for dynamic redirects  |
| `GetPathParams()`                      | Access path parameters for dynamic redirects   |
| `GetHeader(key string)`                | Access request headers                         |

## Error Handling

The redirect functions will panic if an invalid HTTP status code is provided:

```go
// This will panic - status code must be 3xx
c.Redirect(http.StatusOK, "http://example.com") // Panics!

// Valid redirect status codes (3xx)
c.Redirect(http.StatusMovedPermanently, "http://example.com")  // 301 ✓
c.Redirect(http.StatusFound, "http://example.com")             // 302 ✓
c.Redirect(http.StatusSeeOther, "http://example.com")          // 303 ✓
```

HTTP redirects are essential for URL migration, mobile detection, form processing, and API versioning. The httpmanager module provides flexible redirect functionality that integrates seamlessly with the existing middleware and routing system.
