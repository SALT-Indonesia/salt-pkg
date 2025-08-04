package httpmanager

import (
	"context"
	"github.com/gorilla/mux"
	"net/http"
)

// PathParams represents a map of path parameters from the URL
type PathParams map[string]string

// pathParamsKey is the context key for path parameters
const pathParamsKey contextKey = "pathParams"

// GetPathParams extracts path parameters from the context
func GetPathParams(ctx context.Context) PathParams {
	if params, ok := ctx.Value(pathParamsKey).(PathParams); ok {
		return params
	}
	return PathParams{}
}

// Get returns the value for the given path parameter key
func (p PathParams) Get(key string) string {
	if value, ok := p[key]; ok {
		return value
	}
	return ""
}

// Has checks if a path parameter exists
func (p PathParams) Has(key string) bool {
	_, exists := p[key]
	return exists
}

// Keys returns all parameter keys
func (p PathParams) Keys() []string {
	keys := make([]string, 0, len(p))
	for k := range p {
		keys = append(keys, k)
	}
	return keys
}

// extractPathParams extracts path parameters from the request using gorilla/mux
func extractPathParams(r *http.Request) PathParams {
	vars := mux.Vars(r)
	params := make(PathParams)
	for k, v := range vars {
		params[k] = v
	}
	return params
}
