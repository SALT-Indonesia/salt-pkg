package home

import (
	"context"
	"github.com/SALT-Indonesia/salt-pkg/httpmanager"
	"go/types"
	"log"
	"net/http"
)

type Handler struct {
}

// NewHandler creates a new HTTP handler for the home endpoint
func NewHandler() *httpmanager.Handler[types.Nil, Response] {
	return httpmanager.NewHandler(
		http.MethodGet,
		func(ctx context.Context, _ *types.Nil) (*Response, error) {
			// Extract a specific header using the new GetHeader method
			// Common headers: Authorization, Content-Type, Accept, X-Request-ID
			headerKey := "X-Request-ID" // Change this to the desired header key
			headerValue := httpmanager.GetHeader(ctx, headerKey)
			log.Println("Header value:", headerValue)

			// You can also get all headers using httpmanager.GetHeaders(ctx)
			// headers := httpmanager.GetHeaders(ctx)

			return &Response{
				Message: "ok",
			}, nil
		},
	)
}
