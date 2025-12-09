package main

import (
	"examples/httpmanager/internal/application"
	"examples/httpmanager/internal/delivery/home"
	"examples/httpmanager/internal/delivery/product"
	"examples/httpmanager/internal/delivery/profile"
	"examples/httpmanager/internal/delivery/register"
	"examples/httpmanager/internal/delivery/upload"
	"examples/httpmanager/internal/delivery/user"
	"examples/httpmanager/internal/delivery/validation"
	"log"

	"github.com/SALT-Indonesia/salt-pkg/httpmanager"
	"github.com/SALT-Indonesia/salt-pkg/logmanager"
)

func main() {
	// Create application with debug mode enabled
	// Debug mode enables 404 logging which helps identify missing routes
	app := logmanager.NewApplication(
		logmanager.WithDebug(),
		logmanager.WithAppName("httpmanager-example"),
	)

	// Create a server with CORS middleware enabled
	// Health check is enabled by default at GET /health
	// When debug mode is enabled, 404 responses will be logged
	server := httpmanager.NewServer(app)
	server.EnableCORS(
		[]string{"http://localhost:3000", "https://example.com"}, // allowed origins
		[]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},      // allowed methods
		[]string{"*"}, // allowed all headers
		true,          // allow credentials
	)

	// Health check examples:
	// Default: Health check is enabled at GET /health
	// To customize path: httpmanager.NewServer(app, httpmanager.WithHealthCheckPath("/api/health"))
	// To disable: httpmanager.NewServer(app, httpmanager.WithoutHealthCheck())

	// Alternatively, you can enable CORS after server creation:
	// server := httpmanager.NewServer()
	// server.EnableCORS([]string{"*"}, nil, nil, false)

	server.Handle("/", home.NewHandler())
	server.Handle("/me", profile.NewHandler(application.NewProfileUseCaseImpl()))
	server.Handle("/register", register.NewHandler(application.NewUseCaseImpl()))
	server.Handle("/upload", upload.NewHandler())
	server.Handle("/product", product.NewHandler())
	server.Handle("/validation/create-user", validation.NewHandler())

	// Path parameter routes - demonstrating Gin-like functionality
	server.GET("/user/{id}", user.NewGetUserHandler())
	server.PUT("/user/{id}", user.NewUpdateUserHandler())
	server.GET("/user/{id}/profile/{section}", user.NewGetUserProfileHandler())

	// Query parameter binding route - demonstrating automatic query parameter binding
	server.GET("/users/search", user.NewUserSearchHandler())

	// static directory for serving images
	staticDir := "static"
	server.Handle("/images/", httpmanager.NewStaticHandler(staticDir))

	log.Println("Health check endpoint: GET http://localhost:8080/health")
	log.Println("")
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
	log.Println("")
	log.Println("Automatic query parameter binding examples:")
	log.Println("GET http://localhost:8080/users/search?name=john&min_age=18&max_age=65&active=true&tags=developer&tags=golang&include_email=true")
	log.Println("GET http://localhost:8080/users/search?name=alice&min_age=25&active=false")
	log.Println("GET http://localhost:8080/users/search?tags=frontend&tags=react&tags=typescript")
	log.Println("")
	log.Println("Simple error handling examples (CustomErrorV2):")
	log.Println("POST http://localhost:8080/validation/create-user")
	log.Println("  Valid: {\"name\":\"John\",\"email\":\"john@example.com\",\"age\":25}")
	log.Println("  400 error: {\"name\":\"\",\"email\":\"john@example.com\",\"age\":25}")
	log.Println("  500 error: {\"name\":\"database_error\",\"email\":\"test@example.com\",\"age\":25}")
	log.Println("")
	log.Println("CustomErrorV2 examples:")
	log.Println("POST http://localhost:8080/customv2/process-order")
	log.Println("  Valid request: {\"order_id\":\"ORD123\",\"customer_id\":\"CUST456\",\"amount\":100.50,\"payment_type\":\"credit_card\"}")
	log.Println("  Validation error (400): {\"order_id\":\"\",\"customer_id\":\"CUST456\",\"amount\":100.50,\"payment_type\":\"credit_card\"}")
	log.Println("  Business error (422): {\"order_id\":\"ORD123\",\"customer_id\":\"blocked_customer\",\"amount\":100.50,\"payment_type\":\"credit_card\"}")
	log.Println("  System error (500): {\"order_id\":\"ORD_db_error\",\"customer_id\":\"CUST456\",\"amount\":100.50,\"payment_type\":\"credit_card\"}")
	log.Println("")
	log.Println("404 Debug Logging examples (debug mode enabled):")
	log.Println("GET http://localhost:8080/non-existent-path")
	log.Println("GET http://localhost:8080/api/unknown?foo=bar")
	log.Println("POST http://localhost:8080/missing-endpoint")
	log.Println("  -> These will return 404 and log debug messages with method, path, and query params")

	log.Panic(server.Start())
}
