package httpmanager

import (
	"context"
	"net/http"
)

// QueryParams represents a map of query parameters from the URL
type QueryParams map[string][]string

// contextKey is a type for context keys
type contextKey string

// queryParamsKey is the context key for query parameters
const queryParamsKey contextKey = "queryParams"

// RequestKey is the context key for the HTTP request
const RequestKey contextKey = "httpRequest"

// GetQueryParams extracts query parameters from the context
func GetQueryParams(ctx context.Context) QueryParams {
	if params, ok := ctx.Value(queryParamsKey).(QueryParams); ok {
		return params
	}
	return QueryParams{}
}

// Get returns the first value for the given query parameter key
func (q QueryParams) Get(key string) string {
	if values, ok := q[key]; ok && len(values) > 0 {
		return values[0]
	}
	return ""
}

// GetAll returns all values for the given query parameter key
func (q QueryParams) GetAll(key string) []string {
	if values, ok := q[key]; ok {
		return values
	}
	return []string{}
}

// GetHeader returns a single header value from the context for the given key
func GetHeader(ctx context.Context, key string) string {
	if req, ok := ctx.Value(RequestKey).(*http.Request); ok {
		return req.Header.Get(key)
	}
	return ""
}

// GetHeaders returns all headers from the context
func GetHeaders(ctx context.Context) http.Header {
	if req, ok := ctx.Value(RequestKey).(*http.Request); ok {
		return req.Header
	}
	return http.Header{}
}
