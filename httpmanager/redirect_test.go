package httpmanager

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContext_Redirect(t *testing.T) {
	tests := []struct {
		name           string
		code           int
		location       string
		expectedCode   int
		expectedHeader string
		expectPanic    bool
	}{
		{
			name:           "valid redirect with 302",
			code:           http.StatusFound,
			location:       "http://example.com",
			expectedCode:   http.StatusFound,
			expectedHeader: "http://example.com",
			expectPanic:    false,
		},
		{
			name:           "valid redirect with 301",
			code:           http.StatusMovedPermanently,
			location:       "http://example.com",
			expectedCode:   http.StatusMovedPermanently,
			expectedHeader: "http://example.com",
			expectPanic:    false,
		},
		{
			name:           "valid redirect with 303",
			code:           http.StatusSeeOther,
			location:       "/relative/path",
			expectedCode:   http.StatusSeeOther,
			expectedHeader: "/relative/path",
			expectPanic:    false,
		},
		{
			name:        "invalid redirect code 200",
			code:        http.StatusOK,
			location:    "http://example.com",
			expectPanic: true,
		},
		{
			name:        "invalid redirect code 400",
			code:        http.StatusBadRequest,
			location:    "http://example.com",
			expectPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rr := httptest.NewRecorder()

			ctx := NewContext(rr, req)

			if tt.expectPanic {
				assert.Panics(t, func() {
					ctx.Redirect(tt.code, tt.location)
				})
				return
			}

			ctx.Redirect(tt.code, tt.location)

			assert.Equal(t, tt.expectedCode, rr.Code)
			assert.Equal(t, tt.expectedHeader, rr.Header().Get("Location"))
		})
	}
}

func TestContext_RedirectToURL(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	ctx := NewContext(rr, req)
	ctx.RedirectToURL("http://example.com")

	assert.Equal(t, http.StatusFound, rr.Code)
	assert.Equal(t, "http://example.com", rr.Header().Get("Location"))
}

func TestContext_RedirectPermanent(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	ctx := NewContext(rr, req)
	ctx.RedirectPermanent("http://example.com")

	assert.Equal(t, http.StatusMovedPermanently, rr.Code)
	assert.Equal(t, "http://example.com", rr.Header().Get("Location"))
}

func TestRedirectUtilityFunctions(t *testing.T) {
	t.Run("Redirect function", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rr := httptest.NewRecorder()

		Redirect(rr, req, http.StatusFound, "http://example.com")

		assert.Equal(t, http.StatusFound, rr.Code)
		assert.Equal(t, "http://example.com", rr.Header().Get("Location"))
	})

	t.Run("RedirectToURL function", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rr := httptest.NewRecorder()

		RedirectToURL(rr, req, "http://example.com")

		assert.Equal(t, http.StatusFound, rr.Code)
		assert.Equal(t, "http://example.com", rr.Header().Get("Location"))
	})

	t.Run("RedirectPermanent function", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rr := httptest.NewRecorder()

		RedirectPermanent(rr, req, "http://example.com")

		assert.Equal(t, http.StatusMovedPermanently, rr.Code)
		assert.Equal(t, "http://example.com", rr.Header().Get("Location"))
	})

	t.Run("Redirect function with invalid code", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rr := httptest.NewRecorder()

		assert.Panics(t, func() {
			Redirect(rr, req, http.StatusOK, "http://example.com")
		})
	})
}

func TestNewRedirectHandler(t *testing.T) {
	t.Run("valid redirect handler", func(t *testing.T) {
		handlerFunc := func(c *Context) {
			c.RedirectToURL("http://example.com")
		}

		handler := NewRedirectHandler(http.MethodGet, handlerFunc)
		assert.NotNil(t, handler)
		assert.Equal(t, http.MethodGet, handler.method)
	})

	t.Run("nil handler function should panic", func(t *testing.T) {
		assert.Panics(t, func() {
			NewRedirectHandler(http.MethodGet, nil)
		})
	})
}

func TestRedirectHandler_ServeHTTP(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		handlerMethod  string
		handlerFunc    RedirectHandlerFunc
		expectedStatus int
		expectedHeader string
	}{
		{
			name:          "successful redirect",
			method:        http.MethodGet,
			handlerMethod: http.MethodGet,
			handlerFunc: func(c *Context) {
				c.RedirectToURL("http://example.com")
			},
			expectedStatus: http.StatusFound,
			expectedHeader: "http://example.com",
		},
		{
			name:          "method not allowed",
			method:        http.MethodPost,
			handlerMethod: http.MethodGet,
			handlerFunc: func(c *Context) {
				c.RedirectToURL("http://example.com")
			},
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:          "redirect with path parameters",
			method:        http.MethodGet,
			handlerMethod: http.MethodGet,
			handlerFunc: func(c *Context) {
				pathParams := c.GetPathParams()
				id := pathParams.Get("id")
				c.RedirectToURL("http://example.com/user/" + id)
			},
			expectedStatus: http.StatusFound,
			expectedHeader: "http://example.com/user/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/", nil)
			rr := httptest.NewRecorder()

			handler := NewRedirectHandler(tt.handlerMethod, tt.handlerFunc)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectedHeader != "" {
				assert.Equal(t, tt.expectedHeader, rr.Header().Get("Location"))
			}
		})
	}
}

func TestRedirectHandler_WithMiddleware(t *testing.T) {
	// Create a test middleware that adds a header
	testMiddleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Test-Middleware", "applied")
			next.ServeHTTP(w, r)
		})
	}

	handlerFunc := func(c *Context) {
		c.RedirectToURL("http://example.com")
	}

	handler := NewRedirectHandler(http.MethodGet, handlerFunc)
	handler.Use(testMiddleware)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	handlerWithMiddleware := handler.WithMiddleware()
	handlerWithMiddleware.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusFound, rr.Code)
	assert.Equal(t, "http://example.com", rr.Header().Get("Location"))
	assert.Equal(t, "applied", rr.Header().Get("X-Test-Middleware"))
}

func TestGetContextFromStdContext(t *testing.T) {
	t.Run("valid context extraction", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rr := httptest.NewRecorder()

		originalCtx := NewContext(rr, req)

		// Test the context extraction
		extractedCtx := GetContextFromStdContext(originalCtx.Context)
		require.NotNil(t, extractedCtx)

		assert.Equal(t, rr, extractedCtx.Writer)
		assert.Equal(t, req, extractedCtx.Request)
	})

	t.Run("invalid context - missing response writer", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		ctx := req.Context()

		extractedCtx := GetContextFromStdContext(ctx)
		assert.Nil(t, extractedCtx)
	})
}

func TestContext_ParameterAccess(t *testing.T) {
	t.Run("access query parameters", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/?name=test&page=1", nil)
		rr := httptest.NewRecorder()

		// Simulate what happens in ServeHTTP
		queryParams := QueryParams(req.URL.Query())
		ctx := req.Context()
		ctx = contextWithValue(ctx, queryParamsKey, queryParams)

		httpCtx := &Context{
			Context: ctx,
			Writer:  rr,
			Request: req,
		}

		params := httpCtx.GetQueryParams()
		assert.Equal(t, "test", params.Get("name"))
		assert.Equal(t, "1", params.Get("page"))
	})

	t.Run("access headers", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("Authorization", "Bearer token123")
		rr := httptest.NewRecorder()

		// Simulate what happens in ServeHTTP
		ctx := contextWithValue(req.Context(), RequestKey, req)

		httpCtx := &Context{
			Context: ctx,
			Writer:  rr,
			Request: req,
		}

		auth := httpCtx.GetHeader("Authorization")
		assert.Equal(t, "Bearer token123", auth)
	})
}

// Helper function for context manipulation in tests
func contextWithValue(parent context.Context, key interface{}, val interface{}) context.Context {
	return context.WithValue(parent, key, val)
}
