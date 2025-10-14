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

func TestHandler_ServeHTTP_CustomStatusCode(t *testing.T) {
	t.Run("handler returns 201 Created with ResponseSuccess", func(t *testing.T) {
		type CreateRequest struct {
			Name string `json:"name"`
		}
		type CreateResponse struct {
			ID      string `json:"id"`
			Message string `json:"message"`
		}

		handlerFunc := func(ctx context.Context, req *CreateRequest) (*ResponseSuccess[CreateResponse], error) {
			return &ResponseSuccess[CreateResponse]{
				StatusCode: 201,
				Body: CreateResponse{
					ID:      "12345",
					Message: "Resource created: " + req.Name,
				},
			}, nil
		}

		handler := NewHandler(http.MethodPost, handlerFunc)
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"Test"}`))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, 201, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		var response CreateResponse
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "12345", response.ID)
		assert.Equal(t, "Resource created: Test", response.Message)
	})

	t.Run("handler returns 202 Accepted with ResponseSuccess", func(t *testing.T) {
		type AsyncRequest struct {
			Data string `json:"data"`
		}
		type AsyncResponse struct {
			Status string `json:"status"`
			TaskID string `json:"task_id"`
		}

		handlerFunc := func(ctx context.Context, req *AsyncRequest) (*ResponseSuccess[AsyncResponse], error) {
			return &ResponseSuccess[AsyncResponse]{
				StatusCode: 202,
				Body: AsyncResponse{
					Status: "Accepted",
					TaskID: "task-001",
				},
			}, nil
		}

		handler := NewHandler(http.MethodPost, handlerFunc)
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"data":"test"}`))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, 202, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		var response AsyncResponse
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "Accepted", response.Status)
		assert.Equal(t, "task-001", response.TaskID)
	})

	t.Run("handler returns 204 No Content with ResponseSuccess", func(t *testing.T) {
		type DeleteRequest struct {
			ID string `json:"id"`
		}
		type EmptyResponse struct{}

		handlerFunc := func(ctx context.Context, req *DeleteRequest) (*ResponseSuccess[EmptyResponse], error) {
			return &ResponseSuccess[EmptyResponse]{
				StatusCode: 204,
				Body:       EmptyResponse{},
			}, nil
		}

		handler := NewHandler(http.MethodDelete, handlerFunc)
		req := httptest.NewRequest(http.MethodDelete, "/", strings.NewReader(`{"id":"123"}`))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, 204, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
	})

	t.Run("handler returns default 200 OK when not using ResponseSuccess", func(t *testing.T) {
		handlerFunc := func(ctx context.Context, req *Request) (*Response, error) {
			return &Response{Message: "Hello, " + req.Name}, nil
		}

		handler := NewHandler(http.MethodPost, handlerFunc)
		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(`{"name":"Test"}`))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.ServeHTTP(rr, req)

		assert.Equal(t, 200, rr.Code)
		assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

		var response Response
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "Hello, Test", response.Message)
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

func TestCheckResponseSuccess(t *testing.T) {
	t.Run("valid ResponseSuccess pointer with 201 status", func(t *testing.T) {
		type TestBody struct {
			ID      string `json:"id"`
			Message string `json:"message"`
		}

		resp := &ResponseSuccess[TestBody]{
			StatusCode: 201,
			Body:       TestBody{ID: "123", Message: "Resource created"},
		}

		isCustom, statusCode, body := checkResponseSuccess(resp)

		assert.True(t, isCustom)
		assert.Equal(t, 201, statusCode)
		assert.Equal(t, TestBody{ID: "123", Message: "Resource created"}, body)
	})

	t.Run("valid ResponseSuccess with 202 status", func(t *testing.T) {
		type TestBody struct {
			Status string `json:"status"`
		}

		resp := &ResponseSuccess[TestBody]{
			StatusCode: 202,
			Body:       TestBody{Status: "Accepted"},
		}

		isCustom, statusCode, body := checkResponseSuccess(resp)

		assert.True(t, isCustom)
		assert.Equal(t, 202, statusCode)
		assert.Equal(t, TestBody{Status: "Accepted"}, body)
	})

	t.Run("valid ResponseSuccess with 204 and nil body", func(t *testing.T) {
		type EmptyBody struct{}

		resp := &ResponseSuccess[EmptyBody]{
			StatusCode: 204,
			Body:       EmptyBody{},
		}

		isCustom, statusCode, body := checkResponseSuccess(resp)

		assert.True(t, isCustom)
		assert.Equal(t, 204, statusCode)
		assert.Equal(t, EmptyBody{}, body)
	})

	t.Run("nil response", func(t *testing.T) {
		isCustom, statusCode, body := checkResponseSuccess(nil)

		assert.False(t, isCustom)
		assert.Equal(t, 0, statusCode)
		assert.Nil(t, body)
	})

	t.Run("regular struct response", func(t *testing.T) {
		resp := Response{Message: "regular response"}

		isCustom, statusCode, body := checkResponseSuccess(&resp)

		assert.False(t, isCustom)
		assert.Equal(t, 0, statusCode)
		assert.Nil(t, body)
	})

	t.Run("struct with wrong field types", func(t *testing.T) {
		type InvalidSuccess struct {
			StatusCode string      // Should be int
			Body       interface{}
		}

		resp := &InvalidSuccess{
			StatusCode: "201",
			Body:       "body",
		}

		isCustom, statusCode, body := checkResponseSuccess(resp)

		assert.False(t, isCustom)
		assert.Equal(t, 0, statusCode)
		assert.Nil(t, body)
	})

	t.Run("struct missing Body field", func(t *testing.T) {
		type IncompleteSuccess struct {
			StatusCode int
		}

		resp := &IncompleteSuccess{StatusCode: 201}

		isCustom, statusCode, body := checkResponseSuccess(resp)

		assert.False(t, isCustom)
		assert.Equal(t, 0, statusCode)
		assert.Nil(t, body)
	})
}

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
