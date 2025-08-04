package profile

import (
	"context"
	"examples/httpmanager/internal/application"
	"github.com/SALT-Indonesia/salt-pkg/httpmanager"
	"go/types"
	"net/http"
)

type Handler struct {
	usecase application.ProfileUseCase
}

func NewHandler(usecase application.ProfileUseCase) *httpmanager.Handler[types.Nil, Response] {
	return httpmanager.NewHandler(
		http.MethodGet,
		func(ctx context.Context, _ *types.Nil) (*Response, error) {
			output, err := usecase.Execute(ctx)

			if err != nil {
				return nil, &httpmanager.CustomError{
					Err:        err,
					Code:       "002",
					Title:      "error happened",
					Desc:       "this is because bla bla bla",
					StatusCode: 400,
				}
			}

			return &Response{Name: output.Name}, err
		},
	)
}
