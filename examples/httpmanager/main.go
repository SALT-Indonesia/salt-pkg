package main

import (
	"examples/httpmanager/internal/application"
	"examples/httpmanager/internal/delivery/home"
	"examples/httpmanager/internal/delivery/product"
	"examples/httpmanager/internal/delivery/profile"
	"examples/httpmanager/internal/delivery/register"
	"examples/httpmanager/internal/delivery/upload"
	"examples/httpmanager/internal/delivery/user"
	"github.com/SALT-Indonesia/salt-pkg/httpmanager"
	"github.com/SALT-Indonesia/salt-pkg/logmanager"
	"log"
)

func main() {
	// Create a server with CORS middleware enabled
	server := httpmanager.NewServer(logmanager.NewApplication())
	server.EnableCORS(
		[]string{"http://localhost:3000", "https://example.com"}, // allowed origins
		[]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},      // allowed methods
		[]string{"*"}, // allowed all headers
		true,          // allow credentials
	)

	// Alternatively, you can enable CORS after server creation:
	// server := httpmanager.NewServer()
	// server.EnableCORS([]string{"*"}, nil, nil, false)

	server.Handle("/", home.NewHandler())
	server.Handle("/me", profile.NewHandler(application.NewProfileUseCaseImpl()))
	server.Handle("/register", register.NewHandler(application.NewUseCaseImpl()))
	server.Handle("/upload", upload.NewHandler())
	server.Handle("/product", product.NewHandler())

	// Path parameter routes - demonstrating Gin-like functionality
	server.GET("/user/{id}", user.NewGetUserHandler())
	server.PUT("/user/{id}", user.NewUpdateUserHandler())
	server.GET("/user/{id}/profile/{section}", user.NewGetUserProfileHandler())

	// static directory for serving images
	staticDir := "/home/static"
	server.Handle("/images/", httpmanager.NewStaticHandler(staticDir))

	log.Println("Try accessing: http://localhost:8080/images/avatar.jpg")
	log.Println("Or a file in a subdirectory: http://localhost:8080/images/others/orange.jpg")
	log.Println("")
	log.Println("Path parameter examples:")
	log.Println("GET http://localhost:8080/user/123")
	log.Println("GET http://localhost:8080/user/123?include_email=true")
	log.Println("PUT http://localhost:8080/user/123 (with JSON body)")
	log.Println("GET http://localhost:8080/user/123/profile/settings")
	log.Println("GET http://localhost:8080/user/456/profile/activity")
	log.Println("GET http://localhost:8080/user/789/profile/preferences")

	log.Panic(server.Start())
}
