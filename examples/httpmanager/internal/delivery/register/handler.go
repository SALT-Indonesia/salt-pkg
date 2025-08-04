package register

import (
	"context"
	"examples/httpmanager/internal/application"
	"github.com/SALT-Indonesia/salt-pkg/httpmanager"
	"net/http"
)

type Handler struct {
	usecase application.RegisterUseCase
}

func NewHandler(usecase application.RegisterUseCase) *httpmanager.Handler[Request, Response] {
	return httpmanager.NewHandler(
		http.MethodPost,
		func(ctx context.Context, req *Request) (*Response, error) {
			output, err := usecase.Execute(ctx, application.RegisterInput{Name: req.Name})
			return &Response{Message: output.Message}, err
		},
	)
}
