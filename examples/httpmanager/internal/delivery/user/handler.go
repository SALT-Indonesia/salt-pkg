package user

import (
	"context"
	"fmt"
	"github.com/SALT-Indonesia/salt-pkg/httpmanager"
)

type GetUserRequest struct{}

type GetUserResponse struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Message string `json:"message"`
}

type UpdateUserRequest struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type UpdateUserResponse struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Email   string `json:"email"`
	Message string `json:"message"`
}

// NewGetUserHandler creates a handler for GET /user/{id}
func NewGetUserHandler() *httpmanager.Handler[GetUserRequest, GetUserResponse] {
	return httpmanager.NewHandler("GET", func(ctx context.Context, req *GetUserRequest) (*GetUserResponse, error) {
		// Extract path parameters
		pathParams := httpmanager.GetPathParams(ctx)
		userID := pathParams.Get("id")

		// Extract query parameters
		queryParams := httpmanager.GetQueryParams(ctx)
		includeEmail := queryParams.Get("include_email")

		response := &GetUserResponse{
			ID:      userID,
			Name:    fmt.Sprintf("User %s", userID),
			Message: fmt.Sprintf("Retrieved user with ID: %s", userID),
		}

		if includeEmail == "true" {
			response.Email = fmt.Sprintf("user%s@example.com", userID)
		}

		return response, nil
	})
}

// NewUpdateUserHandler creates a handler for PUT /user/{id}
func NewUpdateUserHandler() *httpmanager.Handler[UpdateUserRequest, UpdateUserResponse] {
	return httpmanager.NewHandler("PUT", func(ctx context.Context, req *UpdateUserRequest) (*UpdateUserResponse, error) {
		// Extract path parameters
		pathParams := httpmanager.GetPathParams(ctx)
		userID := pathParams.Get("id")

		return &UpdateUserResponse{
			ID:      userID,
			Name:    req.Name,
			Email:   req.Email,
			Message: fmt.Sprintf("Updated user with ID: %s", userID),
		}, nil
	})
}

// NewGetUserProfileHandler creates a handler for GET /user/{id}/profile/{section}
func NewGetUserProfileHandler() *httpmanager.Handler[GetUserRequest, map[string]interface{}] {
	return httpmanager.NewHandler("GET", func(ctx context.Context, req *GetUserRequest) (*map[string]interface{}, error) {
		// Extract path parameters
		pathParams := httpmanager.GetPathParams(ctx)
		userID := pathParams.Get("id")
		section := pathParams.Get("section")

		result := map[string]interface{}{
			"user_id": userID,
			"section": section,
			"message": fmt.Sprintf("Retrieved %s section for user %s", section, userID),
		}

		// Mock different profile sections
		switch section {
		case "settings":
			result["data"] = map[string]interface{}{
				"theme":         "dark",
				"notifications": true,
				"language":      "en",
			}
		case "activity":
			result["data"] = map[string]interface{}{
				"last_login":     "2024-01-15T10:30:00Z",
				"posts_count":    42,
				"comments_count": 128,
			}
		case "preferences":
			result["data"] = map[string]interface{}{
				"email_notifications": true,
				"privacy_level":       "public",
				"show_online_status":  false,
			}
		default:
			result["data"] = map[string]interface{}{
				"message": "Section not found",
			}
		}

		return &result, nil
	})
}
