package validation

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/SALT-Indonesia/salt-pkg/httpmanager"
)

// NewHandler demonstrates simple error handling with ResponseError (400/500 status codes)
func NewHandler() *httpmanager.Handler[CreateUserRequest, CreateUserResponse] {
	return httpmanager.NewHandler(
		http.MethodPost,
		func(ctx context.Context, req *CreateUserRequest) (*CreateUserResponse, error) {
			// Validation error - 400 status using ResponseError
			if strings.TrimSpace(req.Name) == "" {
				return nil, &httpmanager.ResponseError[ErrorResponse]{
					Err:        fmt.Errorf("name is required"),
					StatusCode: http.StatusBadRequest,
					Body: ErrorResponse{
						Code:    "VIRB01001",
						Message: "Name field is required and cannot be empty",
						Data:    nil,
					},
				}
			}

			if strings.TrimSpace(req.Email) == "" {
				return nil, &httpmanager.ResponseError[ErrorResponse]{
					Err:        fmt.Errorf("email is required"),
					StatusCode: http.StatusBadRequest,
					Body: ErrorResponse{
						Code:    "VIRB01002",
						Message: "Email field is required and cannot be empty",
						Data:    nil,
					},
				}
			}

			if req.Age <= 0 {
				return nil, &httpmanager.ResponseError[ErrorResponse]{
					Err:        fmt.Errorf("invalid age"),
					StatusCode: http.StatusBadRequest,
					Body: ErrorResponse{
						Code:    "VIRB01003",
						Message: "Age must be greater than 0",
						Data:    nil,
					},
				}
			}

			// Internal server error - 500 status using ResponseError
			if req.Name == "database_error" {
				return nil, &httpmanager.ResponseError[ErrorResponse]{
					Err:        fmt.Errorf("database connection failed"),
					StatusCode: http.StatusInternalServerError,
					Body: ErrorResponse{
						Code:    "VISE01001",
						Message: "Database is currently unavailable",
						Data:    nil,
					},
				}
			}

			// Success response
			return &CreateUserResponse{
				ID:      12345,
				Message: fmt.Sprintf("User %s created successfully", req.Name),
			}, nil
		},
	)
}