package httpmanager

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestPathParams_Get(t *testing.T) {
	params := PathParams{
		"id":   "123",
		"name": "john",
	}

	assert.Equal(t, "123", params.Get("id"))
	assert.Equal(t, "john", params.Get("name"))
	assert.Equal(t, "", params.Get("nonexistent"))
}

func TestPathParams_Has(t *testing.T) {
	params := PathParams{
		"id": "123",
	}

	assert.True(t, params.Has("id"))
	assert.False(t, params.Has("nonexistent"))
}

func TestPathParams_Keys(t *testing.T) {
	params := PathParams{
		"id":   "123",
		"name": "john",
	}

	keys := params.Keys()
	assert.Len(t, keys, 2)
	assert.Contains(t, keys, "id")
	assert.Contains(t, keys, "name")
}

func TestGetPathParams(t *testing.T) {
	expectedParams := PathParams{
		"id": "123",
	}

	ctx := context.WithValue(context.Background(), pathParamsKey, expectedParams)
	params := GetPathParams(ctx)

	assert.Equal(t, expectedParams, params)
}

func TestGetPathParams_EmptyContext(t *testing.T) {
	params := GetPathParams(context.Background())
	assert.Equal(t, PathParams{}, params)
}

func TestExtractPathParams(t *testing.T) {
	// Create a mux router and register a route with path parameters
	router := mux.NewRouter()
	router.HandleFunc("/user/{id}/profile/{section}", func(w http.ResponseWriter, r *http.Request) {
		params := extractPathParams(r)
		assert.Equal(t, "123", params.Get("id"))
		assert.Equal(t, "settings", params.Get("section"))
		w.WriteHeader(http.StatusOK)
	})

	// Create a test request
	req := httptest.NewRequest("GET", "/user/123/profile/settings", nil)
	w := httptest.NewRecorder()

	// Serve the request
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestPathParamsIntegration(t *testing.T) {
	type UserRequest struct {
		Action string `json:"action"`
	}

	type UserResponse struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		Action string `json:"action"`
	}

	handler := NewHandler("GET", func(ctx context.Context, req *UserRequest) (*UserResponse, error) {
		pathParams := GetPathParams(ctx)
		return &UserResponse{
			ID:     pathParams.Get("id"),
			Name:   pathParams.Get("name"),
			Action: req.Action,
		}, nil
	})

	// Create a mux router and register the handler
	router := mux.NewRouter()
	router.Handle("/user/{id}/name/{name}", handler).Methods("GET")

	// Create a test request
	req := httptest.NewRequest("GET", "/user/123/name/john", nil)
	w := httptest.NewRecorder()

	// Serve the request
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"id":"123"`)
	assert.Contains(t, w.Body.String(), `"name":"john"`)
}
