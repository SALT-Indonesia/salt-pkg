package protected

import (
	"context"
	"errors"
	"examples/httpmanager/internal/middleware"
	"fmt"

	"github.com/SALT-Indonesia/salt-pkg/httpmanager"
)

// MockTokenService is a mock implementation for demonstration
type MockTokenService struct{}

func (m *MockTokenService) ValidateToken(token string) (*middleware.User, error) {
	// In real implementation, this would validate JWT token
	if token == "valid-token" {
		return &middleware.User{
			ID:    "user-123",
			Email: "user@example.com",
			Role:  "admin",
		}, nil
	}
	return nil, errors.New("invalid token")
}

// Request/Response types
type EmptyRequest struct{}

type ProfileResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

// NewGetProfileHandler returns a handler for GET /protected/me
func NewGetProfileHandler() *httpmanager.Handler[EmptyRequest, ProfileResponse] {
	return httpmanager.NewHandler("GET", func(ctx context.Context, _ *EmptyRequest) (*ProfileResponse, error) {
		user := middleware.GetUser(ctx)
		if user == nil {
			return nil, &httpmanager.CustomError{
				Err:        fmt.Errorf("user not found in context"),
				Code:       "401",
				Title:      "Unauthorized",
				Desc:       "User not found",
				StatusCode: 401,
			}
		}

		return &ProfileResponse{
			ID:    user.ID,
			Email: user.Email,
			Role:  user.Role,
		}, nil
	})
}
