package httpmanager

import (
	"context"
	"net/http"
)

// Additional context keys for redirect functionality
const (
	responseWriterKey contextKey = "responseWriter"
	requestKey        contextKey = "request"
)

// Context wraps the standard context and provides Gin-like functionality
type Context struct {
	context.Context
	Writer  http.ResponseWriter
	Request *http.Request
}

// NewContext creates a new Context
func NewContext(w http.ResponseWriter, r *http.Request) *Context {
	ctx := context.WithValue(r.Context(), responseWriterKey, w)
	ctx = context.WithValue(ctx, requestKey, r)

	return &Context{
		Context: ctx,
		Writer:  w,
		Request: r,
	}
}

// Redirect issues an HTTP redirect to the given URL with the given status code
func (c *Context) Redirect(code int, location string) {
	if code < 300 || code > 399 {
		panic("redirect status code must be 3xx")
	}
	http.Redirect(c.Writer, c.Request, location, code)
}

// RedirectToURL redirects with 302 status code (Found)
func (c *Context) RedirectToURL(location string) {
	c.Redirect(http.StatusFound, location)
}

// RedirectPermanent redirects with 301 status code (Moved Permanently)
func (c *Context) RedirectPermanent(location string) {
	c.Redirect(http.StatusMovedPermanently, location)
}

// GetQueryParams returns query parameters from the context
func (c *Context) GetQueryParams() QueryParams {
	return GetQueryParams(c.Context)
}

// GetPathParams returns path parameters from the context
func (c *Context) GetPathParams() PathParams {
	return GetPathParams(c.Context)
}

// GetHeader returns a specific header value from the request
func (c *Context) GetHeader(key string) string {
	return GetHeader(c.Context, key)
}

// Utility functions for use outside of Context

// Redirect issues an HTTP redirect to the given URL with the given status code
func Redirect(w http.ResponseWriter, r *http.Request, code int, location string) {
	if code < 300 || code > 399 {
		panic("redirect status code must be 3xx")
	}
	http.Redirect(w, r, location, code)
}

// RedirectToURL redirects with 302 status code (Found)
func RedirectToURL(w http.ResponseWriter, r *http.Request, location string) {
	Redirect(w, r, http.StatusFound, location)
}

// RedirectPermanent redirects with 301 status code (Moved Permanently)
func RedirectPermanent(w http.ResponseWriter, r *http.Request, location string) {
	Redirect(w, r, http.StatusMovedPermanently, location)
}

// GetContextFromStdContext extracts httpmanager Context from standard context
func GetContextFromStdContext(ctx context.Context) *Context {
	w, ok := ctx.Value(responseWriterKey).(http.ResponseWriter)
	if !ok {
		return nil
	}

	r, ok := ctx.Value(requestKey).(*http.Request)
	if !ok {
		return nil
	}

	return &Context{
		Context: ctx,
		Writer:  w,
		Request: r,
	}
}

// RedirectHandler provides a handler specifically for redirects
type RedirectHandler struct {
	method      string
	middlewares []func(http.Handler) http.Handler
	handlerFunc func(*Context)
}

// RedirectHandlerFunc is the function signature for redirect handlers
type RedirectHandlerFunc func(*Context)

// NewRedirectHandler creates a new redirect handler
func NewRedirectHandler(method string, handlerFunc RedirectHandlerFunc) *RedirectHandler {
	if handlerFunc == nil {
		panic("handlerFunc cannot be nil")
	}
	return &RedirectHandler{
		method:      method,
		middlewares: []func(http.Handler) http.Handler{},
		handlerFunc: handlerFunc,
	}
}

// Use adds middleware to the redirect handler
func (h *RedirectHandler) Use(middleware ...func(http.Handler) http.Handler) *RedirectHandler {
	h.middlewares = append(h.middlewares, middleware...)
	return h
}

// WithMiddleware returns an http.Handler with the middleware applied
func (h *RedirectHandler) WithMiddleware() http.Handler {
	var handler http.Handler = h

	for i := len(h.middlewares) - 1; i >= 0; i-- {
		handler = h.middlewares[i](handler)
	}

	return handler
}

// ServeHTTP processes incoming HTTP requests for redirects
func (h *RedirectHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != h.method {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Create context with query and path parameters
	queryParams := QueryParams(r.URL.Query())
	pathParams := extractPathParams(r)

	ctx := context.WithValue(r.Context(), queryParamsKey, queryParams)
	ctx = context.WithValue(ctx, pathParamsKey, pathParams)
	ctx = context.WithValue(ctx, RequestKey, r)
	ctx = context.WithValue(ctx, responseWriterKey, w)
	ctx = context.WithValue(ctx, requestKey, r)

	// Create httpmanager Context
	c := &Context{
		Context: ctx,
		Writer:  w,
		Request: r,
	}

	h.handlerFunc(c)
}
