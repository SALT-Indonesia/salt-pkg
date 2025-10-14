package create_user

import (
	"context"
	"github.com/SALT-Indonesia/salt-pkg/httpmanager"
	"net/http"
)

// Handler demonstrates using custom HTTP status codes with ResponseSuccess
func NewHandler() *httpmanager.Handler[Request, httpmanager.ResponseSuccess[Response]] {
	return httpmanager.NewHandler(
		http.MethodPost,
		func(ctx context.Context, req *Request) (*httpmanager.ResponseSuccess[Response], error) {
			// Return 201 Created instead of default 200 OK
			return &httpmanager.ResponseSuccess[Response]{
				StatusCode: 201,
				Body: Response{
					ID:      "user-12345",
					Name:    req.Name,
					Email:   req.Email,
					Message: "User successfully created",
				},
			}, nil
		},
	)
}
