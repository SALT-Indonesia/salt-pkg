package main

import (
	"context"
	"log"

	"github.com/SALT-Indonesia/salt-pkg/httpmanager"
	"github.com/SALT-Indonesia/salt-pkg/logmanager"
)

type GetUserRequest struct{}

type GetUserResponse struct {
	ID      string `json:"id"`
	Message string `json:"message"`
}

func NewGetUserHandler() *httpmanager.Handler[GetUserRequest, GetUserResponse] {
	return httpmanager.NewHandler("GET", func(ctx context.Context, req *GetUserRequest) (*GetUserResponse, error) {
		// Extract path parameters from context
		pathParams := httpmanager.GetPathParams(ctx)
		userID := pathParams.Get("id")

		// Log the user ID using logmanager
		logmanager.InfoWithContext(ctx, "Received request for user", map[string]string{
			"user_id": userID,
		})

		return &GetUserResponse{
			ID:      userID,
			Message: "User retrieved successfully",
		}, nil
	})
}

func main() {
	// Create application with debug mode enabled
	app := logmanager.NewApplication(
		logmanager.WithDebug(),
		logmanager.WithAppName("httpmanager-path-param-example"),
	)

	// Create HTTP server
	server := httpmanager.NewServer(app)

	// Enable CORS middleware
	server.EnableCORS(
		[]string{"*"},    // allowed origins
		[]string{"GET"},  // allowed methods
		[]string{"*"},    // allowed headers
		false,            // allow credentials
	)

	// Register GET /users/{id} route with dynamic path parameter
	server.GET("/users/{id}", NewGetUserHandler())

	log.Println("Server starting on :8080")
	log.Println("Try: GET http://localhost:8080/users/123")
	log.Println("Try: GET http://localhost:8080/users/abc")

	log.Panic(server.Start())
}
