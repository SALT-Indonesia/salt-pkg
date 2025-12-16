package main

import (
	"examples/httpmanager/internal/delivery/protected"
	"examples/httpmanager/internal/middleware"
	"log"
	"os"

	"github.com/SALT-Indonesia/salt-pkg/httpmanager"
	"github.com/SALT-Indonesia/salt-pkg/logmanager"
	"github.com/gorilla/mux"
)

func main() {
	app := logmanager.NewApplication(
		logmanager.WithDebug(),
		logmanager.WithAppName("conditional-middleware-example"),
	)

	server := httpmanager.NewServer(app)
	server.EnableCORS(
		[]string{"*"},
		[]string{"GET", "POST", "PUT", "DELETE"},
		[]string{"*"},
		false,
	)

	// ============================================================
	// CONDITIONAL MIDDLEWARE EXAMPLE
	// ============================================================
	//
	// Problem: In production, Kong gateway handles token validation
	// and sets X-User-* headers. In local environment, we need to
	// validate tokens ourselves. This leads to duplicated route definitions:
	//
	//   if isLocal {
	//       server.Handle("/users/me", handler.Use(KongAuth(ts), ExtractUser()).WithMiddleware())
	//   } else {
	//       server.Handle("/users/me", handler.Use(ExtractUser()).WithMiddleware())
	//   }
	//
	// Solution: Build middleware array dynamically and pass to Use()

	isLocal := os.Getenv("APP_ENV") == "local"
	tokenService := &protected.MockTokenService{}

	if isLocal {
		log.Println("Mode: LOCAL - KongAuth middleware enabled")
	} else {
		log.Println("Mode: PRODUCTION - KongAuth middleware disabled (handled by gateway)")
	}

	// Build middleware array dynamically
	middlewares := []mux.MiddlewareFunc{}
	if isLocal {
		middlewares = append(middlewares, middleware.KongAuth(tokenService))
	}
	middlewares = append(middlewares, middleware.ExtractUser())

	// Clean approach: pass middleware array to Use()
	server.Handle("/protected/me",
		protected.NewGetProfileHandler().
			Use(middlewares...).
			WithMiddleware(),
	)

	log.Println("")
	log.Println("Server starting on :8080")
	log.Println("")
	log.Println("Test commands:")
	if isLocal {
		log.Println("  curl -H 'Authorization: Bearer valid-token' http://localhost:8080/protected/me")
		log.Println("  curl -H 'Authorization: Bearer invalid' http://localhost:8080/protected/me  # 401")
	} else {
		log.Println("  curl -H 'X-User-ID: user-123' -H 'X-User-Email: user@example.com' http://localhost:8080/protected/me")
	}
	log.Println("")
	log.Println("To switch mode: APP_ENV=local go run main.go")

	log.Panic(server.Start())
}
