package httpmanager

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetQueryParams(t *testing.T) {
	t.Run("with valid query params in context", func(t *testing.T) {
		// Create a context with query parameters
		params := QueryParams{
			"name":  []string{"John"},
			"age":   []string{"30"},
			"tags":  []string{"tag1", "tag2", "tag3"},
			"empty": []string{},
		}
		ctx := context.WithValue(context.Background(), queryParamsKey, params)

		// Get query parameters from context
		result := GetQueryParams(ctx)

		// Assert the result matches the original params
		assert.Equal(t, params, result)
		assert.Equal(t, "John", result.Get("name"))
		assert.Equal(t, "30", result.Get("age"))
		assert.Equal(t, "tag1", result.Get("tags"))
		assert.Equal(t, []string{"tag1", "tag2", "tag3"}, result.GetAll("tags"))
	})

	t.Run("with no query params in context", func(t *testing.T) {
		// Create a context without query parameters
		ctx := context.Background()

		// Get query parameters from context
		result := GetQueryParams(ctx)

		// Assert the result is an empty QueryParams
		assert.NotNil(t, result)
		assert.Equal(t, QueryParams{}, result)
		assert.Empty(t, result.Get("name"))
		assert.Empty(t, result.GetAll("name"))
	})

	t.Run("with wrong type in context", func(t *testing.T) {
		// Create a context with a value of wrong type
		ctx := context.WithValue(context.Background(), queryParamsKey, "not a QueryParams")

		// Get query parameters from context
		result := GetQueryParams(ctx)

		// Assert the result is an empty QueryParams
		assert.NotNil(t, result)
		assert.Equal(t, QueryParams{}, result)
	})
}

func TestQueryParams_Get(t *testing.T) {
	t.Run("with existing key", func(t *testing.T) {
		params := QueryParams{
			"name": []string{"John"},
			"age":  []string{"30"},
		}

		// Test getting existing keys
		assert.Equal(t, "John", params.Get("name"))
		assert.Equal(t, "30", params.Get("age"))
	})

	t.Run("with non-existing key", func(t *testing.T) {
		params := QueryParams{
			"name": []string{"John"},
		}

		// Test getting a non-existing key
		assert.Equal(t, "", params.Get("unknown"))
	})

	t.Run("with empty values", func(t *testing.T) {
		params := QueryParams{
			"empty": []string{},
		}

		// Test getting a key with empty values
		assert.Equal(t, "", params.Get("empty"))
	})

	t.Run("with multiple values", func(t *testing.T) {
		params := QueryParams{
			"tags": []string{"tag1", "tag2", "tag3"},
		}

		// Test getting the first value of a key with multiple values
		assert.Equal(t, "tag1", params.Get("tags"))
	})
}

func TestQueryParams_GetAll(t *testing.T) {
	t.Run("with existing key", func(t *testing.T) {
		params := QueryParams{
			"name": []string{"John"},
			"tags": []string{"tag1", "tag2", "tag3"},
		}

		// Test getting all values for existing keys
		assert.Equal(t, []string{"John"}, params.GetAll("name"))
		assert.Equal(t, []string{"tag1", "tag2", "tag3"}, params.GetAll("tags"))
	})

	t.Run("with non-existing key", func(t *testing.T) {
		params := QueryParams{
			"name": []string{"John"},
		}

		// Test getting all values for a non-existing key
		assert.Equal(t, []string{}, params.GetAll("unknown"))
	})

	t.Run("with empty values", func(t *testing.T) {
		params := QueryParams{
			"empty": []string{},
		}

		// Test getting all values for a key with empty values
		assert.Equal(t, []string{}, params.GetAll("empty"))
	})
}

func TestQueryParamsType(t *testing.T) {
	t.Run("create and use query params", func(t *testing.T) {
		// Create a new QueryParams instance
		params := QueryParams{
			"name": []string{"John"},
			"age":  []string{"30"},
		}

		// Assert it's properly initialized
		assert.Len(t, params, 2)
		assert.Contains(t, params, "name")
		assert.Contains(t, params, "age")
	})

	t.Run("empty query params", func(t *testing.T) {
		// Create an empty QueryParams instance
		params := QueryParams{}

		// Assert it's properly initialized as empty
		assert.Len(t, params, 0)
		assert.Empty(t, params.Get("anything"))
		assert.Empty(t, params.GetAll("anything"))
	})
}

func TestContextKey(t *testing.T) {
	t.Run("context key string representation", func(t *testing.T) {
		// Test that the contextKey type works as expected
		key := contextKey("testKey")
		assert.Equal(t, "testKey", string(key))
	})

	t.Run("queryParamsKey constant", func(t *testing.T) {
		// Test that the queryParamsKey constant has the expected value
		assert.Equal(t, contextKey("queryParams"), queryParamsKey)
	})

	t.Run("RequestKey constant", func(t *testing.T) {
		// Test that the RequestKey constant has the expected value
		assert.Equal(t, contextKey("httpRequest"), RequestKey)
	})
}

func TestGetHeader(t *testing.T) {
	t.Run("with valid request in context", func(t *testing.T) {
		// Create a request with headers
		req, _ := http.NewRequest("GET", "http://example.com", nil)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Request-ID", "123456")
		req.Header.Set("Authorization", "Bearer token123")

		// Create a context with the request
		ctx := context.WithValue(context.Background(), RequestKey, req)

		// Test getting existing headers
		assert.Equal(t, "application/json", GetHeader(ctx, "Content-Type"))
		assert.Equal(t, "123456", GetHeader(ctx, "X-Request-ID"))
		assert.Equal(t, "Bearer token123", GetHeader(ctx, "Authorization"))
	})

	t.Run("with non-existing header key", func(t *testing.T) {
		// Create a request with headers
		req, _ := http.NewRequest("GET", "http://example.com", nil)
		req.Header.Set("Content-Type", "application/json")

		// Create a context with the request
		ctx := context.WithValue(context.Background(), RequestKey, req)

		// Test getting a non-existing header
		assert.Equal(t, "", GetHeader(ctx, "X-Non-Existent"))
	})

	t.Run("with no request in context", func(t *testing.T) {
		// Create a context without a request
		ctx := context.Background()

		// Test getting a header from a context without a request
		assert.Equal(t, "", GetHeader(ctx, "Content-Type"))
	})

	t.Run("with wrong type in context", func(t *testing.T) {
		// Create a context with a value of wrong type
		ctx := context.WithValue(context.Background(), RequestKey, "not a request")

		// Test getting a header from a context with wrong type
		assert.Equal(t, "", GetHeader(ctx, "Content-Type"))
	})
}

func TestGetHeaders(t *testing.T) {
	t.Run("with valid request in context", func(t *testing.T) {
		// Create a request with headers
		req, _ := http.NewRequest("GET", "http://example.com", nil)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Request-ID", "123456")
		req.Header.Set("Authorization", "Bearer token123")

		// Create a context with the request
		ctx := context.WithValue(context.Background(), RequestKey, req)

		// Get all headers
		headers := GetHeaders(ctx)

		// Assert the headers match the original request headers
		assert.Equal(t, req.Header, headers)
		assert.Equal(t, "application/json", headers.Get("Content-Type"))
		assert.Equal(t, "123456", headers.Get("X-Request-ID"))
		assert.Equal(t, "Bearer token123", headers.Get("Authorization"))
	})

	t.Run("with no request in context", func(t *testing.T) {
		// Create a context without a request
		ctx := context.Background()

		// Get all headers
		headers := GetHeaders(ctx)

		// Assert the result is an empty http.Header
		assert.NotNil(t, headers)
		assert.Equal(t, http.Header{}, headers)
	})

	t.Run("with wrong type in context", func(t *testing.T) {
		// Create a context with a value of wrong type
		ctx := context.WithValue(context.Background(), RequestKey, "not a request")

		// Get all headers
		headers := GetHeaders(ctx)

		// Assert the result is an empty http.Header
		assert.NotNil(t, headers)
		assert.Equal(t, http.Header{}, headers)
	})
}
