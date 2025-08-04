package httpmanager

import (
	"context"
	"github.com/gorilla/mux"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// StaticHandler handles serving static files, particularly images
type StaticHandler struct {
	rootDir     string
	method      string
	middlewares []mux.MiddlewareFunc
}

// NewStaticHandler creates a new handler for serving static files
func NewStaticHandler(rootDir string) *StaticHandler {
	// Create the root directory if it doesn't exist
	if err := os.MkdirAll(rootDir, 0755); err != nil {
		panic("failed to create static files directory: " + err.Error())
	}

	return &StaticHandler{
		rootDir:     rootDir,
		method:      http.MethodGet,
		middlewares: []mux.MiddlewareFunc{},
	}
}

// Use adds middleware to the handler
func (h *StaticHandler) Use(middleware ...mux.MiddlewareFunc) *StaticHandler {
	h.middlewares = append(h.middlewares, middleware...)
	return h
}

// WithMiddleware returns an http.Handler with the middleware applied
func (h *StaticHandler) WithMiddleware() http.Handler {
	var handler http.Handler = h

	// Apply all middlewares in reverse order
	for i := len(h.middlewares) - 1; i >= 0; i-- {
		handler = h.middlewares[i](handler)
	}

	return handler
}

// ServeHTTP processes incoming HTTP requests for static files
func (h *StaticHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != h.method {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract the file path from the URL
	urlPath := r.URL.Path

	// Extract query parameters from the request URL
	queryParams := QueryParams(r.URL.Query())

	// Add query parameters and the HTTP request to the context
	ctx := context.WithValue(r.Context(), queryParamsKey, queryParams)
	ctx = context.WithValue(ctx, RequestKey, r)

	// Sanitize the path to prevent directory traversal attacks
	cleanPath := filepath.Clean(urlPath)
	if strings.Contains(cleanPath, "..") {
		http.Error(w, "Invalid file path", http.StatusBadRequest)
		return
	}

	// Construct the full file path
	filePath := filepath.Join(h.rootDir, cleanPath)

	// Check if the file exists
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			http.Error(w, "File not found", http.StatusNotFound)
		} else {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Don't serve directories
	if fileInfo.IsDir() {
		http.Error(w, "Cannot serve directories", http.StatusForbidden)
		return
	}

	// Open the file
	file, err := os.Open(filePath)
	if err != nil {
		http.Error(w, "Failed to open file", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	// Set the content type based on file extension
	contentType := getContentType(filePath)
	w.Header().Set("Content-Type", contentType)

	// Set cache control headers for better performance
	w.Header().Set("Cache-Control", "public, max-age=86400") // Cache for 24 hours

	// Serve the file
	http.ServeContent(w, r, fileInfo.Name(), fileInfo.ModTime(), file)
}

// getContentType determines the content type based on file extension
func getContentType(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))

	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".svg":
		return "image/svg+xml"
	case ".webp":
		return "image/webp"
	case ".ico":
		return "image/x-icon"
	case ".bmp":
		return "image/bmp"
	case ".tiff", ".tif":
		return "image/tiff"
	default:
		// Default to binary data for unknown types
		return "application/octet-stream"
	}
}
