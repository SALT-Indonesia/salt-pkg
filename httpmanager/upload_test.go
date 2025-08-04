package httpmanager

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewUploadHandler(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	t.Run("success", func(t *testing.T) {
		handler := NewUploadHandler("POST", tempDir, func(ctx context.Context, files map[string][]*UploadedFile, form map[string][]string) (interface{}, error) {
			return nil, nil
		})

		if handler == nil {
			t.Fatal("Expected handler to be created, got nil")
		}
		if handler.method != "POST" {
			t.Errorf("Expected method to be POST, got %s", handler.method)
		}
		if handler.uploadDir != tempDir {
			t.Errorf("Expected uploadDir to be %s, got %s", tempDir, handler.uploadDir)
		}
		if handler.maxFileSize != 10<<20 {
			t.Errorf("Expected maxFileSize to be %d, got %d", 10<<20, handler.maxFileSize)
		}
		if len(handler.middlewares) != 0 {
			t.Errorf("Expected middlewares to be empty, got %d items", len(handler.middlewares))
		}
	})

	t.Run("nil_handler_func", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Error("Expected panic for nil handlerFunc, but no panic occurred")
			}
		}()

		NewUploadHandler("POST", tempDir, nil)
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
		NewUploadHandler("POST", invalidPath, func(ctx context.Context, files map[string][]*UploadedFile, form map[string][]string) (interface{}, error) {
			return nil, nil
		})
	})
}

func TestUploadHandler_Use(t *testing.T) {
	tempDir := t.TempDir()
	handler := NewUploadHandler("POST", tempDir, func(ctx context.Context, files map[string][]*UploadedFile, form map[string][]string) (interface{}, error) {
		return nil, nil
	})

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
	handler = NewUploadHandler("POST", tempDir, func(ctx context.Context, files map[string][]*UploadedFile, form map[string][]string) (interface{}, error) {
		return nil, nil
	})

	result := handler.Use(middleware1, middleware2)
	if result != handler {
		t.Error("Expected Use method to return the handler for chaining")
	}
	if len(handler.middlewares) != 2 {
		t.Errorf("Expected 2 middlewares, got %d", len(handler.middlewares))
	}
}

func TestUploadHandler_WithMiddleware(t *testing.T) {
	tempDir := t.TempDir()
	handler := NewUploadHandler("POST", tempDir, func(ctx context.Context, files map[string][]*UploadedFile, form map[string][]string) (interface{}, error) {
		return nil, nil
	})

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
	req := httptest.NewRequest("POST", "/upload", nil)
	rr := httptest.NewRecorder()

	// Call the wrapped handler
	wrappedHandler.ServeHTTP(rr, req)

	// Verify middleware was called
	if !middlewareCalled {
		t.Error("Expected middleware to be called")
	}

	// Test with multiple middlewares
	handler = NewUploadHandler("POST", tempDir, func(ctx context.Context, files map[string][]*UploadedFile, form map[string][]string) (interface{}, error) {
		return nil, nil
	})

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

func TestUploadHandler_WithMaxFileSize(t *testing.T) {
	tempDir := t.TempDir()
	handler := NewUploadHandler("POST", tempDir, func(ctx context.Context, files map[string][]*UploadedFile, form map[string][]string) (interface{}, error) {
		return nil, nil
	})

	// Test default max file size
	if handler.maxFileSize != 10<<20 {
		t.Errorf("Expected default maxFileSize to be %d, got %d", 10<<20, handler.maxFileSize)
	}

	// Test setting max file size
	newSize := int64(5 << 20) // 5 MB
	result := handler.WithMaxFileSize(newSize)

	if result != handler {
		t.Error("Expected WithMaxFileSize to return the handler for chaining")
	}

	if handler.maxFileSize != newSize {
		t.Errorf("Expected maxFileSize to be %d, got %d", newSize, handler.maxFileSize)
	}
}

func TestUploadHandler_ServeHTTP_MethodNotAllowed(t *testing.T) {
	tempDir := t.TempDir()
	handler := NewUploadHandler("POST", tempDir, func(ctx context.Context, files map[string][]*UploadedFile, form map[string][]string) (interface{}, error) {
		return nil, nil
	})

	// Create a test request with wrong method
	req := httptest.NewRequest("GET", "/upload", nil)
	rr := httptest.NewRecorder()

	// Call the handler
	handler.ServeHTTP(rr, req)

	// Check response
	if status := rr.Code; status != http.StatusMethodNotAllowed {
		t.Errorf("Expected status code %d, got %d", http.StatusMethodNotAllowed, status)
	}
}

func TestUploadHandler_ServeHTTP_InvalidMultipartForm(t *testing.T) {
	tempDir := t.TempDir()
	handler := NewUploadHandler("POST", tempDir, func(ctx context.Context, files map[string][]*UploadedFile, form map[string][]string) (interface{}, error) {
		return nil, nil
	})

	// Create a test request with invalid content type
	req := httptest.NewRequest("POST", "/upload", strings.NewReader("not a multipart form"))
	req.Header.Set("Content-Type", "text/plain")
	rr := httptest.NewRecorder()

	// Call the handler
	handler.ServeHTTP(rr, req)

	// Check response
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, status)
	}
}

func TestUploadHandler_ServeHTTP_Success(t *testing.T) {
	tempDir := t.TempDir()

	// Create a handler that returns a response
	handler := NewUploadHandler("POST", tempDir, func(ctx context.Context, files map[string][]*UploadedFile, form map[string][]string) (interface{}, error) {
		return map[string]string{"status": "success"}, nil
	})

	// Create a multipart form with a file
	var b bytes.Buffer
	writer := multipart.NewWriter(&b)

	// Add a form field
	if err := writer.WriteField("field1", "value1"); err != nil {
		t.Fatalf("Failed to write form field: %v", err)
	}

	// Add a file
	fileWriter, err := writer.CreateFormFile("file", "test.txt")
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}

	_, err = fileWriter.Write([]byte("test file content"))
	if err != nil {
		t.Fatalf("Failed to write to form file: %v", err)
	}

	writer.Close()

	// Create a test request
	req := httptest.NewRequest("POST", "/upload", &b)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rr := httptest.NewRecorder()

	// Call the handler
	handler.ServeHTTP(rr, req)

	// Check response
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, status)
	}

	// Check content type
	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type %s, got %s", "application/json", contentType)
	}

	// Check response body
	var response map[string]string
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if response["status"] != "success" {
		t.Errorf("Expected response status 'success', got '%s'", response["status"])
	}
}

func TestUploadHandler_ServeHTTP_NoContent(t *testing.T) {
	tempDir := t.TempDir()

	// Create a handler that returns nil
	handler := NewUploadHandler("POST", tempDir, func(ctx context.Context, files map[string][]*UploadedFile, form map[string][]string) (interface{}, error) {
		return nil, nil
	})

	// Create a multipart form
	var b bytes.Buffer
	writer := multipart.NewWriter(&b)
	writer.Close()

	// Create a test request
	req := httptest.NewRequest("POST", "/upload", &b)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rr := httptest.NewRecorder()

	// Call the handler
	handler.ServeHTTP(rr, req)

	// Check response
	if status := rr.Code; status != http.StatusNoContent {
		t.Errorf("Expected status code %d, got %d", http.StatusNoContent, status)
	}
}

func TestUploadHandler_ServeHTTP_HandlerError(t *testing.T) {
	tempDir := t.TempDir()

	// Create a handler that returns an error
	handler := NewUploadHandler("POST", tempDir, func(ctx context.Context, files map[string][]*UploadedFile, form map[string][]string) (interface{}, error) {
		return nil, errors.New("handler error")
	})

	// Create a multipart form
	var b bytes.Buffer
	writer := multipart.NewWriter(&b)
	writer.Close()

	// Create a test request
	req := httptest.NewRequest("POST", "/upload", &b)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rr := httptest.NewRecorder()

	// Call the handler
	handler.ServeHTTP(rr, req)

	// Check response
	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, status)
	}
}

func TestUploadHandler_processUploadedFiles(t *testing.T) {
	tempDir := t.TempDir()

	handler := NewUploadHandler("POST", tempDir, func(ctx context.Context, files map[string][]*UploadedFile, form map[string][]string) (interface{}, error) {
		return nil, nil
	})

	t.Run("nil_multipart_form", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/upload", nil)

		files, err := handler.processUploadedFiles(req)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if files != nil {
			t.Errorf("Expected nil files, got %v", files)
		}
	})

	t.Run("with_files", func(t *testing.T) {
		// Create a multipart form with a file
		var b bytes.Buffer
		writer := multipart.NewWriter(&b)

		fileWriter, err := writer.CreateFormFile("file", "test.txt")
		if err != nil {
			t.Fatalf("Failed to create form file: %v", err)
		}

		_, err = fileWriter.Write([]byte("test file content"))
		if err != nil {
			t.Fatalf("Failed to write to form file: %v", err)
		}

		writer.Close()

		// Create a test request
		req := httptest.NewRequest("POST", "/upload", &b)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		// Parse the multipart form
		err = req.ParseMultipartForm(10 << 20)
		if err != nil {
			t.Fatalf("Failed to parse multipart form: %v", err)
		}

		files, err := handler.processUploadedFiles(req)

		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		if files == nil {
			t.Fatal("Expected files, got nil")
		}

		if len(files) != 1 {
			t.Errorf("Expected 1 file field, got %d", len(files))
		}

		if fileList, ok := files["file"]; !ok || len(fileList) != 1 {
			t.Errorf("Expected 1 file in 'file' field, got %v", files)
		}
	})
}

func TestUploadHandler_saveFile(t *testing.T) {
	tempDir := t.TempDir()

	handler := NewUploadHandler("POST", tempDir, func(ctx context.Context, files map[string][]*UploadedFile, form map[string][]string) (interface{}, error) {
		return nil, nil
	})

	// Create a test file
	var b bytes.Buffer
	writer := multipart.NewWriter(&b)

	fileWriter, err := writer.CreateFormFile("file", "test.txt")
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}

	_, err = fileWriter.Write([]byte("test file content"))
	if err != nil {
		t.Fatalf("Failed to write to form file: %v", err)
	}

	writer.Close()

	// Create a test request
	req := httptest.NewRequest("POST", "/upload", &b)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// Parse the multipart form
	err = req.ParseMultipartForm(10 << 20)
	if err != nil {
		t.Fatalf("Failed to parse multipart form: %v", err)
	}

	// Get the file header
	fileHeader := req.MultipartForm.File["file"][0]

	// Test saving the file
	uploadedFile, err := handler.saveFile(fileHeader)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if uploadedFile == nil {
		t.Fatal("Expected uploadedFile, got nil")
	}

	// Check file metadata
	if uploadedFile.Filename != "test.txt" {
		t.Errorf("Expected filename 'test.txt', got '%s'", uploadedFile.Filename)
	}

	// Check if file exists
	if _, err := os.Stat(uploadedFile.SavedPath); os.IsNotExist(err) {
		t.Errorf("Expected file to exist at '%s'", uploadedFile.SavedPath)
	}

	// Check file content
	content, err := os.ReadFile(uploadedFile.SavedPath)
	if err != nil {
		t.Fatalf("Failed to read saved file: %v", err)
	}

	if string(content) != "test file content" {
		t.Errorf("Expected file content 'test file content', got '%s'", string(content))
	}
}

func TestUploadHandler_saveFile_Errors(t *testing.T) {
	tempDir := t.TempDir()

	// We'll create handlers in each test case as needed

	t.Run("create_error", func(t *testing.T) {
		// Make the upload directory read-only to cause a creation error
		readOnlyDir := filepath.Join(tempDir, "readonly")
		if err := os.MkdirAll(readOnlyDir, 0o555); err != nil {
			t.Fatalf("Failed to create test directory: %v", err)
		}

		log.Printf("Created read-only directory: %s", readOnlyDir)

		// Create a handler with the read-only directory
		readOnlyHandler := NewUploadHandler("POST", readOnlyDir, func(ctx context.Context, files map[string][]*UploadedFile, form map[string][]string) (interface{}, error) {
			return nil, nil
		})

		// Create a test file
		var b bytes.Buffer
		writer := multipart.NewWriter(&b)

		fileWriter, err := writer.CreateFormFile("file", "test.txt")
		if err != nil {
			t.Fatalf("Failed to create form file: %v", err)
		}

		_, err = fileWriter.Write([]byte("test file content"))
		if err != nil {
			t.Fatalf("Failed to write to form file: %v", err)
		}

		writer.Close()

		// Create a test request
		req := httptest.NewRequest("POST", "/upload", &b)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		// Parse the multipart form
		err = req.ParseMultipartForm(10 << 20)
		if err != nil {
			t.Fatalf("Failed to parse multipart form: %v", err)
		}

		// Get the file header
		fileHeader := req.MultipartForm.File["file"][0]

		// Test saving the file - this should fail because the directory is read-only
		_, err = readOnlyHandler.saveFile(fileHeader)

		if err == nil {
			t.Error("Expected error when destination file cannot be created, got nil")
		}
	})
}

// TestMain is used to set up and tear down the test environment
func TestMain(m *testing.M) {
	// Run tests
	exitCode := m.Run()

	// Clean up any temporary files or directories

	// Exit with the test exit code
	os.Exit(exitCode)
}
