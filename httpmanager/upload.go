package httpmanager

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/SALT-Indonesia/salt-pkg/logmanager"
	"github.com/gorilla/mux"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// UploadedFile represents a file that has been uploaded
type UploadedFile struct {
	Filename    string
	Size        int64
	ContentType string
	SavedPath   string
}

// UploadHandler handles file uploads via multipart form data
type UploadHandler struct {
	handlerFunc func(ctx context.Context, files map[string][]*UploadedFile, form map[string][]string) (interface{}, error)
	method      string
	uploadDir   string
	maxFileSize int64
	middlewares []mux.MiddlewareFunc
}

// NewUploadHandler creates a new handler for file uploads
func NewUploadHandler(method string, uploadDir string, handlerFunc func(ctx context.Context, files map[string][]*UploadedFile, form map[string][]string) (interface{}, error)) *UploadHandler {
	if handlerFunc == nil {
		panic("handlerFunc cannot be nil")
	}

	// Create an upload directory if it doesn't exist
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		panic(fmt.Sprintf("failed to create upload directory: %v", err))
	}

	return &UploadHandler{
		handlerFunc: handlerFunc,
		method:      method,
		uploadDir:   uploadDir,
		maxFileSize: 10 << 20, // 10 MB default
		middlewares: []mux.MiddlewareFunc{},
	}
}

// Use adds middleware to the handler
func (h *UploadHandler) Use(middleware ...mux.MiddlewareFunc) *UploadHandler {
	h.middlewares = append(h.middlewares, middleware...)
	return h
}

// WithMiddleware returns an http.Handler with the middleware applied
func (h *UploadHandler) WithMiddleware() http.Handler {
	var handler http.Handler = h

	// Apply all middlewares in reverse order
	for i := len(h.middlewares) - 1; i >= 0; i-- {
		handler = h.middlewares[i](handler)
	}

	return handler
}

// WithMaxFileSize sets the maximum file size allowed for uploads
func (h *UploadHandler) WithMaxFileSize(maxSize int64) *UploadHandler {
	h.maxFileSize = maxSize
	return h
}

// ServeHTTP processes incoming HTTP requests with multipart form data
func (h *UploadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != h.method {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Limit the request body size
	r.Body = http.MaxBytesReader(w, r.Body, h.maxFileSize)

	// Parse the multipart form
	if err := r.ParseMultipartForm(h.maxFileSize); err != nil {
		http.Error(w, "Request too large or invalid multipart form", http.StatusBadRequest)
		return
	}

	ctx := r.Context()

	// Process uploaded files
	files, err := h.processUploadedFiles(r)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error processing uploads: %v", err), http.StatusInternalServerError)
		return
	}

	// Get form values
	formValues := make(map[string][]string)
	for key, values := range r.MultipartForm.Value {
		formValues[key] = values
	}

	// Update transaction with parsed multipart form data for logging
	if txn := logmanager.FromContext(ctx); txn != nil {
		txn.SetWebRequest(r)
	}

	// Call the handler function
	resp, err := h.handlerFunc(ctx, files, formValues)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// If response is nil, return 204 No Content
	if resp == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Return JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	encoder := json.NewEncoder(w)
	_ = encoder.Encode(resp)
}

// processUploadedFiles processes all uploaded files from the request
func (h *UploadHandler) processUploadedFiles(r *http.Request) (map[string][]*UploadedFile, error) {
	if r.MultipartForm == nil || r.MultipartForm.File == nil {
		return nil, nil
	}

	result := make(map[string][]*UploadedFile)

	for fieldName, fileHeaders := range r.MultipartForm.File {
		var files []*UploadedFile

		for _, fileHeader := range fileHeaders {
			uploadedFile, err := h.saveFile(fileHeader)
			if err != nil {
				// Clean up any files that were already saved
				for _, fileList := range result {
					for _, file := range fileList {
						os.Remove(file.SavedPath)
					}
				}
				return nil, err
			}

			files = append(files, uploadedFile)
		}

		if len(files) > 0 {
			result[fieldName] = files
		}
	}

	return result, nil
}

// saveFile saves an uploaded file to disk and returns metadata about the saved file
func (h *UploadHandler) saveFile(fileHeader *multipart.FileHeader) (*UploadedFile, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return nil, errors.New("failed to open uploaded file")
	}
	defer file.Close()

	// Create a unique filename to prevent overwriting
	filename := filepath.Base(fileHeader.Filename)
	timestamp := time.Now().UnixNano()
	uniqueFilename := fmt.Sprintf("%d_%s", timestamp, filename)
	filePath := filepath.Join(h.uploadDir, uniqueFilename)

	// Create the destination file
	dst, err := os.Create(filePath)
	if err != nil {
		return nil, errors.New("failed to create destination file")
	}
	defer dst.Close()

	// Copy the uploaded file to the destination file
	if _, err := io.Copy(dst, file); err != nil {
		return nil, errors.New("failed to save uploaded file")
	}

	// Get content type
	contentType := fileHeader.Header.Get("Content-Type")

	return &UploadedFile{
		Filename:    filename,
		Size:        fileHeader.Size,
		ContentType: contentType,
		SavedPath:   filePath,
	}, nil
}
