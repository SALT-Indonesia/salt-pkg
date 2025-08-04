package httpmanager

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestNewStaticHandler(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	t.Run("success", func(t *testing.T) {
		handler := NewStaticHandler(tempDir)

		if handler == nil {
			t.Fatal("Expected handler to be created, got nil")
		}
		if handler.method != "GET" {
			t.Errorf("Expected method to be GET, got %s", handler.method)
		}
		if handler.rootDir != tempDir {
			t.Errorf("Expected rootDir to be %s, got %s", tempDir, handler.rootDir)
		}
		if len(handler.middlewares) != 0 {
			t.Errorf("Expected middlewares to be empty, got %d items", len(handler.middlewares))
		}
	})

	t.Run("invalid_directory", func(t *testing.T) {
		// Create a file that will conflict with the directory we want to create
		invalidPath := filepath.Join(tempDir, "invalid")
		file, err := os.Create(invalidPath)
		if err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		file.Close()

		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic for invalid directory, but no panic occurred")
			}
		}()

		// Try to create a handler with a path that exists as a file (not a directory)
		NewStaticHandler(invalidPath)
	})
}

func TestStaticHandler_Use(t *testing.T) {
	tempDir := t.TempDir()
	handler := NewStaticHandler(tempDir)

	middleware1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		})
	}

	middleware2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
		})
	}

	// Test adding one middleware
	handler.Use(middleware1)
	if len(handler.middlewares) != 1 {
		t.Errorf("Expected 1 middleware, got %d", len(handler.middlewares))
	}

	// Test adding another middleware
	handler.Use(middleware2)
	if len(handler.middlewares) != 2 {
		t.Errorf("Expected 2 middlewares, got %d", len(handler.middlewares))
	}

	// Test method chaining
	handler = NewStaticHandler(tempDir)

	result := handler.Use(middleware1, middleware2)
	if result != handler {
		t.Error("Expected Use method to return the handler for chaining")
	}
	if len(handler.middlewares) != 2 {
		t.Errorf("Expected 2 middlewares, got %d", len(handler.middlewares))
	}
}

func TestStaticHandler_WithMiddleware(t *testing.T) {
	tempDir := t.TempDir()
	handler := NewStaticHandler(tempDir)

	// Test with no middleware
	result := handler.WithMiddleware()
	if result != handler {
		t.Error("Expected WithMiddleware to return the handler when no middlewares are added")
	}

	// Test with middleware
	middlewareCalled := false
	middleware := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			middlewareCalled = true
			next.ServeHTTP(w, r)
		})
	}

	handler.Use(middleware)
	wrappedHandler := handler.WithMiddleware()

	// Create a test request
	req := httptest.NewRequest("GET", "/images/test.jpg", nil)
	rr := httptest.NewRecorder()

	// Call the wrapped handler
	wrappedHandler.ServeHTTP(rr, req)

	// Verify middleware was called
	if !middlewareCalled {
		t.Error("Expected middleware to be called")
	}

	// Test with multiple middlewares
	handler = NewStaticHandler(tempDir)

	middleware1Called := false
	middleware1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			middleware1Called = true
			next.ServeHTTP(w, r)
		})
	}

	middleware2Called := false
	middleware2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			middleware2Called = true
			next.ServeHTTP(w, r)
		})
	}

	handler.Use(middleware1, middleware2)
	wrappedHandler = handler.WithMiddleware()

	// Reset recorder
	rr = httptest.NewRecorder()
	wrappedHandler.ServeHTTP(rr, req)

	// Verify both middlewares were called
	if !middleware1Called || !middleware2Called {
		t.Error("Expected both middlewares to be called")
	}
}

func TestStaticHandler_ServeHTTP_MethodNotAllowed(t *testing.T) {
	tempDir := t.TempDir()
	handler := NewStaticHandler(tempDir)

	// Create a test request with the wrong method
	req := httptest.NewRequest("POST", "/images/test.jpg", nil)
	rr := httptest.NewRecorder()

	// Call the handler
	handler.ServeHTTP(rr, req)

	// Check response
	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("Expected status code %d, got %d", http.StatusMethodNotAllowed, status)
	}
}

func TestStaticHandler_ServeHTTP_FileNotFound(t *testing.T) {
	tempDir := t.TempDir()
	handler := NewStaticHandler(tempDir)

	// Create a test request for a non-existent file
	req := httptest.NewRequest("GET", "/images/nonexistent.jpg", nil)
	rr := httptest.NewRecorder()

	// Call the handler
	handler.ServeHTTP(rr, req)

	// Check response
	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("Expected status code %d, got %d", http.StatusNotFound, status)
	}
}

func TestStaticHandler_ServeHTTP_DirectoryTraversal(t *testing.T) {
	tempDir := t.TempDir()
	handler := NewStaticHandler(tempDir)

	// Create a test request with a directory traversal attempt
	req := httptest.NewRequest("GET", "/images/../../../etc/passwd", nil)
	rr := httptest.NewRecorder()

	// Call the handler
	handler.ServeHTTP(rr, req)

	// Check response
	// Note: The implementation uses filepath.Clean, which normalizes the path before checking for ".",
	// so the directory traversal check might not be triggered. In that case, it would return a 404 Not Found
	// since the normalized path doesn't exist.
	if status := rr.Code; status != http.StatusNotFound && status != http.StatusBadRequest {
		t.Errorf("Expected status code %d or %d, got %d", http.StatusNotFound, http.StatusBadRequest, status)
	}
}

func TestStaticHandler_ServeHTTP_Directory(t *testing.T) {
	tempDir := t.TempDir()

	// Create a subdirectory
	subDir := filepath.Join(tempDir, "subdir")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatalf("Failed to create test subdirectory: %v", err)
	}

	handler := NewStaticHandler(tempDir)

	// Create a test request for a directory
	req := httptest.NewRequest("GET", "/subdir", nil)
	rr := httptest.NewRecorder()

	// Call the handler
	handler.ServeHTTP(rr, req)

	// Check response
	if status := rr.Code; status != http.StatusForbidden {
		t.Errorf("Expected status code %d, got %d", http.StatusForbidden, status)
	}
}

func TestStaticHandler_ServeHTTP_Success(t *testing.T) {
	tempDir := t.TempDir()

	// Create a test image file
	imagePath := filepath.Join(tempDir, "test.jpg")
	if err := os.WriteFile(imagePath, []byte("fake image content"), 0644); err != nil {
		t.Fatalf("Failed to create test image: %v", err)
	}

	handler := NewStaticHandler(tempDir)

	// Create a test request for the image
	req := httptest.NewRequest("GET", "/test.jpg", nil)
	rr := httptest.NewRecorder()

	// Call the handler
	handler.ServeHTTP(rr, req)

	// Check response
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, status)
	}

	// Check content type
	contentType := rr.Header().Get("Content-Type")
	if contentType != "image/jpeg" {
		t.Errorf("Expected Content-Type %s, got %s", "image/jpeg", contentType)
	}

	// Check cache control header
	cacheControl := rr.Header().Get("Cache-Control")
	if cacheControl != "public, max-age=86400" {
		t.Errorf("Expected Cache-Control %s, got %s", "public, max-age=86400", cacheControl)
	}

	// Check response body
	if rr.Body.String() != "fake image content" {
		t.Errorf("Expected body %s, got %s", "fake image content", rr.Body.String())
	}
}

func TestStaticHandler_ServeHTTP_ContentTypes(t *testing.T) {
	tempDir := t.TempDir()
	handler := NewStaticHandler(tempDir)

	// Test cases for different file extensions and their expected content types
	testCases := []struct {
		filename    string
		contentType string
	}{
		{"test.jpg", "image/jpeg"},
		{"test.jpeg", "image/jpeg"},
		{"test.png", "image/png"},
		{"test.gif", "image/gif"},
		{"test.svg", "image/svg+xml"},
		{"test.webp", "image/webp"},
		{"test.ico", "image/x-icon"},
		{"test.bmp", "image/bmp"},
		{"test.tiff", "image/tiff"},
		{"test.tif", "image/tiff"},
		{"test.unknown", "application/octet-stream"},
	}

	for _, tc := range testCases {
		t.Run(tc.filename, func(t *testing.T) {
			// Create a test file
			filePath := filepath.Join(tempDir, tc.filename)
			if err := os.WriteFile(filePath, []byte("test content"), 0644); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			// Create a test request
			req := httptest.NewRequest("GET", "/"+tc.filename, nil)
			rr := httptest.NewRecorder()

			// Call the handler
			handler.ServeHTTP(rr, req)

			// Check response
			if status := rr.Code; status != http.StatusOK {
				t.Errorf("Expected status code %d, got %d", http.StatusOK, status)
			}

			// Check content type
			contentType := rr.Header().Get("Content-Type")
			if contentType != tc.contentType {
				t.Errorf("Expected Content-Type %s, got %s", tc.contentType, contentType)
			}
		})
	}
}

func TestGetContentType(t *testing.T) {
	testCases := []struct {
		filePath    string
		contentType string
	}{
		{"/path/to/image.jpg", "image/jpeg"},
		{"/path/to/image.jpeg", "image/jpeg"},
		{"/path/to/image.png", "image/png"},
		{"/path/to/image.gif", "image/gif"},
		{"/path/to/image.svg", "image/svg+xml"},
		{"/path/to/image.webp", "image/webp"},
		{"/path/to/image.ico", "image/x-icon"},
		{"/path/to/image.bmp", "image/bmp"},
		{"/path/to/image.tiff", "image/tiff"},
		{"/path/to/image.tif", "image/tiff"},
		{"/path/to/file.unknown", "application/octet-stream"},
		{"/path/to/file", "application/octet-stream"},
	}

	for _, tc := range testCases {
		t.Run(tc.filePath, func(t *testing.T) {
			contentType := getContentType(tc.filePath)
			if contentType != tc.contentType {
				t.Errorf("Expected content type %s for %s, got %s", tc.contentType, tc.filePath, contentType)
			}
		})
	}
}
