package product

import (
	"context"
	"github.com/SALT-Indonesia/salt-pkg/httpmanager"
	"go/types"
	"net/http"
)

type Handler struct {
}

func NewHandler() *httpmanager.Handler[types.Nil, Response] {
	return httpmanager.NewHandler(
		http.MethodGet,
		func(ctx context.Context, _ *types.Nil) (*Response, error) {
			queryParams := httpmanager.GetQueryParams(ctx)
			return &Response{ID: queryParams.Get("id")}, nil
		},
	)
}
