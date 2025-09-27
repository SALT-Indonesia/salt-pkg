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

func TestBindQueryParams(t *testing.T) {
	t.Run("bind string fields", func(t *testing.T) {
		type QueryRequest struct {
			Name     string `query:"name"`
			Category string `query:"category"`
		}

		params := QueryParams{
			"name":     []string{"John"},
			"category": []string{"tech"},
		}
		ctx := context.WithValue(context.Background(), queryParamsKey, params)

		var req QueryRequest
		err := BindQueryParams(ctx, &req)

		assert.NoError(t, err)
		assert.Equal(t, "John", req.Name)
		assert.Equal(t, "tech", req.Category)
	})

	t.Run("bind integer fields", func(t *testing.T) {
		type QueryRequest struct {
			Age   int   `query:"age"`
			Count int64 `query:"count"`
		}

		params := QueryParams{
			"age":   []string{"30"},
			"count": []string{"12345"},
		}
		ctx := context.WithValue(context.Background(), queryParamsKey, params)

		var req QueryRequest
		err := BindQueryParams(ctx, &req)

		assert.NoError(t, err)
		assert.Equal(t, 30, req.Age)
		assert.Equal(t, int64(12345), req.Count)
	})

	t.Run("bind boolean fields", func(t *testing.T) {
		type QueryRequest struct {
			Active  bool `query:"active"`
			Enabled bool `query:"enabled"`
		}

		params := QueryParams{
			"active":  []string{"true"},
			"enabled": []string{"false"},
		}
		ctx := context.WithValue(context.Background(), queryParamsKey, params)

		var req QueryRequest
		err := BindQueryParams(ctx, &req)

		assert.NoError(t, err)
		assert.Equal(t, true, req.Active)
		assert.Equal(t, false, req.Enabled)
	})

	t.Run("bind slice fields", func(t *testing.T) {
		type QueryRequest struct {
			Tags  []string `query:"tags"`
			IDs   []int    `query:"ids"`
			Flags []bool   `query:"flags"`
		}

		params := QueryParams{
			"tags":  []string{"tag1", "tag2", "tag3"},
			"ids":   []string{"1", "2", "3"},
			"flags": []string{"true", "false", "true"},
		}
		ctx := context.WithValue(context.Background(), queryParamsKey, params)

		var req QueryRequest
		err := BindQueryParams(ctx, &req)

		assert.NoError(t, err)
		assert.Equal(t, []string{"tag1", "tag2", "tag3"}, req.Tags)
		assert.Equal(t, []int{1, 2, 3}, req.IDs)
		assert.Equal(t, []bool{true, false, true}, req.Flags)
	})

	t.Run("bind mixed types", func(t *testing.T) {
		type QueryRequest struct {
			Name     string   `query:"name"`
			Age      int      `query:"age"`
			Active   bool     `query:"active"`
			Tags     []string `query:"tags"`
			NoTag    string   // Field without query tag
			Private  string   `query:"private"` // Field with tag but not in params
		}

		params := QueryParams{
			"name":   []string{"John"},
			"age":    []string{"30"},
			"active": []string{"true"},
			"tags":   []string{"tag1", "tag2"},
		}
		ctx := context.WithValue(context.Background(), queryParamsKey, params)

		var req QueryRequest
		err := BindQueryParams(ctx, &req)

		assert.NoError(t, err)
		assert.Equal(t, "John", req.Name)
		assert.Equal(t, 30, req.Age)
		assert.Equal(t, true, req.Active)
		assert.Equal(t, []string{"tag1", "tag2"}, req.Tags)
		assert.Equal(t, "", req.NoTag)    // Field without tag should remain empty
		assert.Equal(t, "", req.Private)  // Field with tag but no param should remain empty
	})

	t.Run("bind with invalid values", func(t *testing.T) {
		type QueryRequest struct {
			Age    int  `query:"age"`
			Active bool `query:"active"`
		}

		params := QueryParams{
			"age":    []string{"invalid"},
			"active": []string{"not_bool"},
		}
		ctx := context.WithValue(context.Background(), queryParamsKey, params)

		var req QueryRequest
		err := BindQueryParams(ctx, &req)

		assert.NoError(t, err) // Should not error, just skip invalid values
		assert.Equal(t, 0, req.Age)      // Should remain zero value
		assert.Equal(t, false, req.Active) // Should remain zero value
	})

	t.Run("bind int64 slice", func(t *testing.T) {
		type QueryRequest struct {
			Values []int64 `query:"values"`
		}

		params := QueryParams{
			"values": []string{"1", "2", "3", "invalid", "4"},
		}
		ctx := context.WithValue(context.Background(), queryParamsKey, params)

		var req QueryRequest
		err := BindQueryParams(ctx, &req)

		assert.NoError(t, err)
		assert.Equal(t, []int64{1, 2, 3, 4}, req.Values) // Should skip invalid value
	})

	t.Run("with nil destination", func(t *testing.T) {
		params := QueryParams{
			"name": []string{"John"},
		}
		ctx := context.WithValue(context.Background(), queryParamsKey, params)

		err := BindQueryParams(ctx, nil)
		assert.NoError(t, err) // Should handle nil gracefully
	})

	t.Run("with non-pointer destination", func(t *testing.T) {
		type QueryRequest struct {
			Name string `query:"name"`
		}

		params := QueryParams{
			"name": []string{"John"},
		}
		ctx := context.WithValue(context.Background(), queryParamsKey, params)

		var req QueryRequest
		err := BindQueryParams(ctx, req) // Not a pointer

		assert.NoError(t, err) // Should handle gracefully
		assert.Equal(t, "", req.Name) // Should remain empty since not a pointer
	})

	t.Run("with non-struct destination", func(t *testing.T) {
		params := QueryParams{
			"name": []string{"John"},
		}
		ctx := context.WithValue(context.Background(), queryParamsKey, params)

		var name string
		err := BindQueryParams(ctx, &name) // Pointer to string, not struct

		assert.NoError(t, err) // Should handle gracefully
		assert.Equal(t, "", name) // Should remain empty since not a struct
	})

	t.Run("with empty query params", func(t *testing.T) {
		type QueryRequest struct {
			Name string `query:"name"`
		}

		ctx := context.Background() // No query params in context

		var req QueryRequest
		err := BindQueryParams(ctx, &req)

		assert.NoError(t, err) // Should handle gracefully
		assert.Equal(t, "", req.Name) // Should remain empty
	})

	t.Run("with empty param values", func(t *testing.T) {
		type QueryRequest struct {
			Name string `query:"name"`
		}

		params := QueryParams{
			"other": []string{"value"}, // Different param name
		}
		ctx := context.WithValue(context.Background(), queryParamsKey, params)

		var req QueryRequest
		err := BindQueryParams(ctx, &req)

		assert.NoError(t, err) // Should handle gracefully
		assert.Equal(t, "", req.Name) // Should remain empty
	})
}
