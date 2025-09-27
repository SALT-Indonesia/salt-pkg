package httpmanager

import (
	"github.com/SALT-Indonesia/salt-pkg/httpmanager/internal/testdata"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewServer(t *testing.T) {
	const (
		defaultTimeOut = 10 * time.Second
		defaultAddr    = ":8080"
	)

	tests := []struct {
		name         string
		options      []OptionFunc
		expectAddr   string
		readTimeout  time.Duration
		writeTimeout time.Duration
	}{
		{
			name:         "default options",
			options:      nil,
			expectAddr:   defaultAddr,
			readTimeout:  defaultTimeOut,
			writeTimeout: defaultTimeOut,
		},
		{
			name:         "custom address",
			options:      []OptionFunc{WithAddr(":8081")},
			expectAddr:   ":8081",
			readTimeout:  defaultTimeOut,
			writeTimeout: defaultTimeOut,
		},
		{
			name:         "custom read timeout",
			options:      []OptionFunc{WithReadTimeout(5 * time.Second)},
			expectAddr:   defaultAddr,
			readTimeout:  5 * time.Second,
			writeTimeout: defaultTimeOut,
		},
		{
			name:         "custom write timeout",
			options:      []OptionFunc{WithWriteTimeout(10 * time.Second)},
			expectAddr:   defaultAddr,
			readTimeout:  defaultTimeOut,
			writeTimeout: 10 * time.Second,
		},
		{
			name:         "multiple options",
			options:      []OptionFunc{WithAddr(":9090"), WithReadTimeout(3 * time.Second), WithWriteTimeout(8 * time.Second)},
			expectAddr:   ":9090",
			readTimeout:  3 * time.Second,
			writeTimeout: 8 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := NewServer(testdata.NewApplication(), tt.options...)

			assert.Equal(t, server.server.Addr, tt.expectAddr)
			assert.Equal(t, server.server.WriteTimeout, tt.writeTimeout)
			assert.Equal(t, server.server.ReadTimeout, tt.readTimeout)
			assert.NotNil(t, server.server.Handler)
		})
	}
}

func TestServer_Middleware(t *testing.T) {
	// Create a test middleware that adds a header to the response
	testMiddleware := func(key, value string) mux.MiddlewareFunc {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set(key, value)
				next.ServeHTTP(w, r)
			})
		}
	}

	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	t.Run("server with middleware option", func(t *testing.T) {
		// Create a server with middleware
		middleware1 := testMiddleware("X-Test-1", "value1")
		middleware2 := testMiddleware("X-Test-2", "value2")
		server := NewServer(testdata.NewApplication(), WithMiddleware(middleware1, middleware2))

		// Register a handler
		server.Handle("/test", testHandler)

		// Create a test request
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rr := httptest.NewRecorder()

		// Serve the request
		server.router.ServeHTTP(rr, req)

		// Check the response
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "value1", rr.Header().Get("X-Test-1"))
		assert.Equal(t, "value2", rr.Header().Get("X-Test-2"))
	})

	t.Run("server with Use method", func(t *testing.T) {
		// Create a server
		server := NewServer(testdata.NewApplication())

		// Add middleware
		middleware1 := testMiddleware("X-Test-1", "value1")
		middleware2 := testMiddleware("X-Test-2", "value2")
		server.Use(middleware1, middleware2)

		// Register a handler
		server.Handle("/test", testHandler)

		// Create a test request
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rr := httptest.NewRecorder()

		// Serve the request
		server.router.ServeHTTP(rr, req)

		// Check the response
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "value1", rr.Header().Get("X-Test-1"))
		assert.Equal(t, "value2", rr.Header().Get("X-Test-2"))
	})

	t.Run("handler with specific middleware", func(t *testing.T) {
		// Create a server
		server := NewServer(testdata.NewApplication())

		// Add server middleware
		serverMiddleware := testMiddleware("X-Server", "server")
		server.Use(serverMiddleware)

		// Add handler-specific middleware
		handlerMiddleware1 := testMiddleware("X-Handler-1", "handler1")
		handlerMiddleware2 := testMiddleware("X-Handler-2", "handler2")

		// Register a handler with specific middleware
		server.HandleWithMiddleware("/test", testHandler, handlerMiddleware1, handlerMiddleware2)

		// Create a test request
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rr := httptest.NewRecorder()

		// Serve the request
		server.router.ServeHTTP(rr, req)

		// Check the response
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "server", rr.Header().Get("X-Server"))
		assert.Equal(t, "handler1", rr.Header().Get("X-Handler-1"))
		assert.Equal(t, "handler2", rr.Header().Get("X-Handler-2"))
	})
}

func TestServer_HealthCheck_DefaultPath(t *testing.T) {
	server := NewServer(testdata.NewApplication())

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	server.router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Empty(t, rr.Body.String())
}

func TestServer_HealthCheck_CustomPath(t *testing.T) {
	customPath := "/custom-health"
	server := NewServer(testdata.NewApplication(), WithHealthCheckPath(customPath))

	req := httptest.NewRequest(http.MethodGet, customPath, nil)
	rr := httptest.NewRecorder()

	server.router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Empty(t, rr.Body.String())
}

func TestServer_HealthCheck_Disabled(t *testing.T) {
	server := NewServer(testdata.NewApplication(), WithoutHealthCheck())

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()

	server.router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestServer_HealthCheck_OnlyGET(t *testing.T) {
	server := NewServer(testdata.NewApplication())

	methods := []string{http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			req := httptest.NewRequest(method, "/health", nil)
			rr := httptest.NewRecorder()

			server.router.ServeHTTP(rr, req)

			assert.Equal(t, http.StatusMethodNotAllowed, rr.Code)
		})
	}
}

func TestServer_HealthCheck_WithOptions(t *testing.T) {
	server := NewServer(testdata.NewApplication(),
		WithHealthCheckPath("/api/health"),
		WithAddr(":8081"),
		WithReadTimeout(5*time.Second),
	)

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	rr := httptest.NewRecorder()

	server.router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Empty(t, rr.Body.String())
	assert.Equal(t, ":8081", server.server.Addr)
	assert.Equal(t, 5*time.Second, server.server.ReadTimeout)
}
