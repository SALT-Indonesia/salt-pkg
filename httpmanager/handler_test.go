package httpmanager

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/gorilla/mux"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Request and Response types for testing
type Request struct {
	Name string `json:"name"`
}
type Response struct {
	Message string `json:"message"`
}

func TestNewHandler(t *testing.T) {

	tests := []struct {
		name         string
		method       string
		handlerFunc  HandlerFunc[Request, Response]
		expectedErr  bool
		expectedResp *Handler[Request, Response]
	}{
		{
			name:   "valid handler with GET",
			method: http.MethodGet,
			handlerFunc: func(ctx context.Context, req *Request) (*Response, error) {
				return &Response{Message: "Hello, " + req.Name}, nil
			},
			expectedErr: false,
		},
		{
			name:   "valid handler with POST",
			method: http.MethodPost,
			handlerFunc: func(ctx context.Context, req *Request) (*Response, error) {
				return &Response{Message: "Created: " + req.Name}, nil
			},
			expectedErr: false,
		},
		{
			name:   "handler with error",
			method: http.MethodPut,
			handlerFunc: func(ctx context.Context, req *Request) (*Response, error) {
				return nil, errors.New("unexpected error")
			},
			expectedErr: false,
		},
		{
			name:        "nil handlerFunc",
			method:      http.MethodGet,
			handlerFunc: nil,
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.expectedErr {
				assert.Panics(t, func() {
					NewHandler(tt.method, tt.handlerFunc)
				}, "Expected panic for nil handlerFunc")
				return
			}

			handler := NewHandler(tt.method, tt.handlerFunc)
			assert.NotNil(t, handler, "Expected valid handler, got nil")
			assert.Equal(t, tt.method, handler.method, "Method mismatch")

			handlerFuncPtr := reflect.ValueOf(handler.handlerFunc).Pointer()
			ttHandlerFuncPtr := reflect.ValueOf(tt.handlerFunc).Pointer()
			assert.Equal(t, ttHandlerFuncPtr, handlerFuncPtr, "Handler function mismatch")
		})
	}
}

func TestHandler_ServeHTTP(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		handlerMethod  string
		requestBody    string
		handlerFunc    HandlerFunc[Request, Response]
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "method not allowed",
			method:         http.MethodGet,
			handlerMethod:  http.MethodPost,
			requestBody:    `{"name":"Test"}`,
			handlerFunc:    func(ctx context.Context, req *Request) (*Response, error) { return &Response{Message: "Hello"}, nil },
			expectedStatus: http.StatusMethodNotAllowed,
			expectedBody:   "Method not allowed\n",
		},
		{
			name:           "invalid request body",
			method:         http.MethodPost,
			handlerMethod:  http.MethodPost,
			requestBody:    `{"name":Test}`, // Invalid JSON
			handlerFunc:    func(ctx context.Context, req *Request) (*Response, error) { return &Response{Message: "Hello"}, nil },
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Invalid request body\n",
		},
		{
			name:           "handler returns error",
			method:         http.MethodPost,
			handlerMethod:  http.MethodPost,
			requestBody:    `{"name":"Test"}`,
			handlerFunc:    func(ctx context.Context, req *Request) (*Response, error) { return nil, errors.New("handler error") },
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `{"status":false,"code":"500","message":{"title":"unknow error","desc":"unknow error"},"data":null}`,
		},
		{
			name:           "successful request with nil response",
			method:         http.MethodPost,
			handlerMethod:  http.MethodPost,
			requestBody:    `{"name":"Test"}`,
			handlerFunc:    func(ctx context.Context, req *Request) (*Response, error) { return nil, nil },
			expectedStatus: http.StatusOK,
			expectedBody:   "",
		},
		{
			name:          "successful request with valid response",
			method:        http.MethodPost,
			handlerMethod: http.MethodPost,
			requestBody:   `{"name":"Test"}`,
			handlerFunc: func(ctx context.Context, req *Request) (*Response, error) {
				return &Response{Message: "Hello, " + req.Name}, nil
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"message":"Hello, Test"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a request with the specified method and body
			var body io.Reader
			if tt.requestBody != "" {
				body = strings.NewReader(tt.requestBody)
			}
			req := httptest.NewRequest(tt.method, "/", body)
			req.Header.Set("Content-Type", "application/json")

			// Create a response recorder
			rr := httptest.NewRecorder()

			// Create the handler
			handler := NewHandler(tt.handlerMethod, tt.handlerFunc)

			// Call the ServeHTTP method
			handler.ServeHTTP(rr, req)

			// Check the status code
			assert.Equal(t, tt.expectedStatus, rr.Code, "Status code mismatch")

			// Check the response body
			if tt.expectedBody != "" {
				if strings.HasPrefix(tt.expectedBody, "{") {
					// For JSON responses, compare after normalizing
					var expected, actual map[string]interface{}
					err := json.Unmarshal([]byte(tt.expectedBody), &expected)
					require.NoError(t, err, "Failed to parse expected JSON")

					err = json.Unmarshal(rr.Body.Bytes(), &actual)
					require.NoError(t, err, "Failed to parse actual JSON")

					assert.Equal(t, expected, actual, "Response body mismatch")
				} else {
					// For plain text responses, compare directly
					assert.Equal(t, tt.expectedBody, rr.Body.String(), "Response body mismatch")
				}
			} else {
				assert.Empty(t, rr.Body.String(), "Expected empty response body")
			}

			// For successful responses, check the Content-Type header
			if tt.expectedStatus == http.StatusOK {
				assert.Equal(t, "application/json", rr.Header().Get("Content-Type"), "Content-Type header mismatch")
			}
		})
	}
}

func TestHandler_Middleware(t *testing.T) {
	// Create a test middleware that adds a header to the response
	testMiddleware := func(key, value string) mux.MiddlewareFunc {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set(key, value)
				next.ServeHTTP(w, r)
			})
		}
	}

	// Create a test handler function
	handlerFunc := func(ctx context.Context, req *Request) (*Response, error) {
		return &Response{Message: "Hello, " + req.Name}, nil
	}

	t.Run("handler with middleware", func(t *testing.T) {
		// Create a handler with middleware
		handler := NewHandler(http.MethodPost, handlerFunc)
		middleware1 := testMiddleware("X-Test-1", "value1")
		middleware2 := testMiddleware("X-Test-2", "value2")
		handler.Use(middleware1, middleware2)

		// Create a test request
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"Test"}`))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		// Get the handler with middleware applied
		handlerWithMiddleware := handler.WithMiddleware()

		// Serve the request
		handlerWithMiddleware.ServeHTTP(rr, req)

		// Check the response
		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Equal(t, "value1", rr.Header().Get("X-Test-1"))
		assert.Equal(t, "value2", rr.Header().Get("X-Test-2"))

		// Check the response body
		var response map[string]interface{}
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "Hello, Test", response["message"])
	})

	t.Run("handler with middleware order", func(t *testing.T) {
		// Create a handler with middleware
		handler := NewHandler(http.MethodPost, handlerFunc)

		// Create middleware that adds a value to a header
		appendMiddleware := func(key, value string) mux.MiddlewareFunc {
			return func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					next.ServeHTTP(w, r)
					current := w.Header().Get(key)
					if current != "" {
						current += ","
					}
					w.Header().Set(key, current+value)
				})
			}
		}

		// Add middleware in a specific order
		handler.Use(appendMiddleware("X-Order", "first"))
		handler.Use(appendMiddleware("X-Order", "second"))
		handler.Use(appendMiddleware("X-Order", "third"))

		// Create a test request
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"Test"}`))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		// Get the handler with middleware applied
		handlerWithMiddleware := handler.WithMiddleware()

		// Serve the request
		handlerWithMiddleware.ServeHTTP(rr, req)

		// Check the response
		assert.Equal(t, http.StatusOK, rr.Code)

		// The middleware should be applied in reverse order (third, second, first)
		// because each middleware wraps the next one
		assert.Equal(t, "third,second,first", rr.Header().Get("X-Order"))
	})
}

// Test types that implement error interface
type CustomStringError string
func (e CustomStringError) Error() string { return string(e) }

type IncompleteError struct {
	StatusCode int
	SomeOtherField string
}
func (e *IncompleteError) Error() string { return "incomplete error" }

type WrongFieldTypesError struct {
	Err        string // Should be error type
	StatusCode string // Should be int type
	Body       interface{}
}
func (e *WrongFieldTypesError) Error() string { return e.Err }

func TestCheckCustomErrorV2(t *testing.T) {
	t.Run("valid ResponseError pointer", func(t *testing.T) {
		type TestBody struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		}

		err := &ResponseError[TestBody]{
			Err:        errors.New("original error"),
			StatusCode: 422,
			Body:       TestBody{Code: "TEST001", Message: "test error"},
		}

		isCustom, statusCode, body := checkCustomErrorV2(err)

		assert.True(t, isCustom)
		assert.Equal(t, 422, statusCode)
		assert.Equal(t, TestBody{Code: "TEST001", Message: "test error"}, body)
	})

	t.Run("valid ResponseError different type", func(t *testing.T) {
		type TestBody struct {
			Error string `json:"error"`
		}

		err := &ResponseError[TestBody]{
			Err:        nil,
			StatusCode: 400,
			Body:       TestBody{Error: "validation failed"},
		}

		// Test with different body type
		isCustom, statusCode, body := checkCustomErrorV2(err)

		assert.True(t, isCustom)
		assert.Equal(t, 400, statusCode)
		assert.Equal(t, TestBody{Error: "validation failed"}, body)
	})

	t.Run("regular error", func(t *testing.T) {
		err := errors.New("regular error")

		isCustom, statusCode, body := checkCustomErrorV2(err)

		assert.False(t, isCustom)
		assert.Equal(t, 0, statusCode)
		assert.Nil(t, body)
	})

	t.Run("non-struct error", func(t *testing.T) {
		// Create a custom type that implements error but isn't a struct
		err := CustomStringError("string error")

		isCustom, statusCode, body := checkCustomErrorV2(err)

		assert.False(t, isCustom)
		assert.Equal(t, 0, statusCode)
		assert.Nil(t, body)
	})

	t.Run("struct missing required fields", func(t *testing.T) {
		err := &IncompleteError{StatusCode: 500, SomeOtherField: "test"}

		isCustom, statusCode, body := checkCustomErrorV2(err)

		assert.False(t, isCustom)
		assert.Equal(t, 0, statusCode)
		assert.Nil(t, body)
	})

	t.Run("struct with wrong field types", func(t *testing.T) {
		err := &WrongFieldTypesError{
			Err:        "error string",
			StatusCode: "500",
			Body:       "body",
		}

		isCustom, statusCode, body := checkCustomErrorV2(err)

		assert.False(t, isCustom)
		assert.Equal(t, 0, statusCode)
		assert.Nil(t, body)
	})
}
