package httpmanager

import (
	"context"
	"encoding/json"
	"github.com/gorilla/mux"
	"net/http"
)

type HandlerFunc[Req any, Resp any] func(ctx context.Context, req *Req) (*Resp, error)

// DetailedErrorResponse represents an alternative JSON error response format
type DetailedErrorResponse struct {
	Status  bool        `json:"status"`
	Code    string      `json:"code"`
	Message MessageInfo `json:"message"`
	Data    interface{} `json:"data"`
}

// MessageInfo contains detailed error message information
type MessageInfo struct {
	Title string `json:"title"`
	Desc  string `json:"desc"`
}

// Handler restricts HTTP requests to a specific method
type Handler[Req any, Resp any] struct {
	handlerFunc HandlerFunc[Req, Resp]
	method      string
	middlewares []mux.MiddlewareFunc
}

// NewHandler creates and returns a new Handler with the specified handler function and HTTP method.
// It will panic if handlerFunc is nil.
func NewHandler[Req any, Resp any](method string, handlerFunc HandlerFunc[Req, Resp]) *Handler[Req, Resp] {
	if handlerFunc == nil {
		panic("handlerFunc cannot be nil")
	}
	return &Handler[Req, Resp]{
		handlerFunc: handlerFunc,
		method:      method,
		middlewares: []mux.MiddlewareFunc{},
	}
}

// Use adds middleware to the handler
func (h *Handler[Req, Resp]) Use(middleware ...mux.MiddlewareFunc) *Handler[Req, Resp] {
	h.middlewares = append(h.middlewares, middleware...)
	return h
}

// WithMiddleware returns an http.Handler with the middleware applied
func (h *Handler[Req, Resp]) WithMiddleware() http.Handler {
	var handler http.Handler = h

	// Apply all middlewares in reverse order
	for i := len(h.middlewares) - 1; i >= 0; i-- {
		handler = h.middlewares[i](handler)
	}

	return handler
}

// ServeHTTP processes incoming HTTP requests, decodes the request body, executes the handler func, and writes a JSON response.
func (h *Handler[Req, Resp]) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != h.method {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract query parameters from the request URL
	queryParams := QueryParams(r.URL.Query())

	// Extract path parameters from the request URL
	pathParams := extractPathParams(r)

	// Add query parameters, path parameters, and the HTTP request to the context
	ctx := context.WithValue(r.Context(), queryParamsKey, queryParams)
	ctx = context.WithValue(ctx, pathParamsKey, pathParams)
	ctx = context.WithValue(ctx, RequestKey, r)

	var req Req
	if r.Body != nil && r.ContentLength > 0 {
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&req); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}
	}

	resp, err := h.handlerFunc(ctx, &req)
	if err != nil {
		// Return error as JSON
		statusCode := http.StatusInternalServerError
		errorResp := DetailedErrorResponse{
			Status: false,
			Code:   "500",
			Message: MessageInfo{
				Title: "unknow error",
				Desc:  "unknow error",
			},
			Data: nil,
		}

		// Check if the error is a CustomError to use client-provided values
		if detailedErr, ok := IsCustomError(err); ok {
			// Use client-provided values
			errorResp = DetailedErrorResponse{
				Status: false,
				Code:   detailedErr.Code,
				Message: MessageInfo{
					Title: detailedErr.Title,
					Desc:  detailedErr.Desc,
				},
				Data: nil,
			}
			statusCode = detailedErr.StatusCode
		} else {
			errorResp = DetailedErrorResponse{
				Status: false,
				Code:   "500",
				Message: MessageInfo{
					Title: "unknow error",
					Desc:  "unknow error",
				},
				Data: nil,
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		encoder := json.NewEncoder(w)
		_ = encoder.Encode(errorResp)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if resp != nil {
		encoder := json.NewEncoder(w)
		_ = encoder.Encode(resp)
	}
}
