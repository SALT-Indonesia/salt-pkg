package httpmanager

import (
	"github.com/SALT-Indonesia/salt-pkg/httpmanager/internal/testdata"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCORSMiddleware(t *testing.T) {
	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	t.Run("default CORS settings", func(t *testing.T) {
		// Create middleware with default settings
		middleware := CORSMiddleware(nil, nil, nil, false)
		handler := middleware(testHandler)

		// Create a test request
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rr := httptest.NewRecorder()

		// Serve the request
		handler.ServeHTTP(rr, req)

		// Check the response
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "*", rr.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "GET, POST, PUT, DELETE, OPTIONS", rr.Header().Get("Access-Control-Allow-Methods"))
		assert.Equal(t, "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization, X-CSRF-Token", rr.Header().Get("Access-Control-Allow-Headers"))
		assert.Equal(t, "", rr.Header().Get("Access-Control-Allow-Credentials"))
	})

	t.Run("custom CORS settings", func(t *testing.T) {
		// Create middleware with custom settings
		origins := []string{"https://example.com", "https://api.example.com"}
		methods := []string{"GET", "POST"}
		headers := []string{"Content-Type", "Authorization"}
		middleware := CORSMiddleware(origins, methods, headers, true)
		handler := middleware(testHandler)

		// Create a test request
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rr := httptest.NewRecorder()

		// Serve the request
		handler.ServeHTTP(rr, req)

		// Check the response
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "https://example.com, https://api.example.com", rr.Header().Get("Access-Control-Allow-Origin"))
		assert.Equal(t, "GET, POST", rr.Header().Get("Access-Control-Allow-Methods"))
		assert.Equal(t, "Content-Type, Authorization", rr.Header().Get("Access-Control-Allow-Headers"))
		assert.Equal(t, "true", rr.Header().Get("Access-Control-Allow-Credentials"))
	})

	t.Run("preflight request", func(t *testing.T) {
		// Create middleware
		middleware := CORSMiddleware(nil, nil, nil, false)
		handler := middleware(testHandler)

		// Create a preflight OPTIONS request
		req := httptest.NewRequest(http.MethodOptions, "/test", nil)
		rr := httptest.NewRecorder()

		// Serve the request
		handler.ServeHTTP(rr, req)

		// Check the response - should return 200 OK without calling the next handler
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "*", rr.Header().Get("Access-Control-Allow-Origin"))
	})
}

func TestWithCORS(t *testing.T) {
	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Create a server with CORS middleware
	server := NewServer(testdata.NewApplication())
	server.EnableCORS([]string{"https://example.com"}, nil, nil, true)

	// Register a handler
	server.Handle("/test", testHandler)

	// Create a test request
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	// Serve the request
	server.router.ServeHTTP(rr, req)

	// Check the response
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "https://example.com", rr.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "true", rr.Header().Get("Access-Control-Allow-Credentials"))
}

func TestServer_EnableCORS(t *testing.T) {
	// Create a test handler
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Create a server
	server := NewServer(testdata.NewApplication())

	// Enable CORS
	server.EnableCORS([]string{"https://api.example.com"}, []string{"GET", "POST"}, nil, false)

	// Register a handler
	server.Handle("/test", testHandler)

	// Create a test request
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rr := httptest.NewRecorder()

	// Serve the request
	server.router.ServeHTTP(rr, req)

	// Check the response
	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "https://api.example.com", rr.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "GET, POST", rr.Header().Get("Access-Control-Allow-Methods"))
}
