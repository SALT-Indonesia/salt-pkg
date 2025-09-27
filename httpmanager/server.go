package httpmanager

import (
	"context"
	"fmt"
	"github.com/SALT-Indonesia/salt-pkg/logmanager"
	"github.com/SALT-Indonesia/salt-pkg/logmanager/integrations/lmgorilla"
	"github.com/gorilla/mux"
	"net/http"
)

type Server struct {
	server       *http.Server
	router       *mux.Router
	app          *logmanager.Application
	hasSetUpCORS bool
	*Option
}

// NewServer creates a new Server instance with optional custom configurations provided via OptionFunc parameters.
func NewServer(app *logmanager.Application, opts ...OptionFunc) *Server {
	router := mux.NewRouter()

	s := &Server{
		router: router,
		app:    app,
	}

	s.Option = newDefaultOption()
	for _, o := range opts {
		o(s.Option)
	}

	// Add default middlewares
	s.middlewares = append(s.middlewares, lmgorilla.Middleware(s.app))

	// Register health check endpoint if enabled
	if s.healthCheckEnabled {
		s.registerHealthCheck()
	}

	s.server = &http.Server{
		Handler:      router,
		Addr:         s.addr,
		ReadTimeout:  s.readTimeout,
		WriteTimeout: s.writeTimeout,
	}

	return s
}

// Use adds middleware to the server that will be applied to all handlers
func (s *Server) Use(middleware ...mux.MiddlewareFunc) {
	s.middlewares = append(s.middlewares, middleware...)
}

// Handle registers the handler with the given pattern in the Server's router.
// It applies all server middlewares to the handler.
func (s *Server) Handle(pattern string, handler http.Handler) {
	// Apply all middlewares in reverse order
	finalHandler := handler
	for i := len(s.middlewares) - 1; i >= 0; i-- {
		finalHandler = s.middlewares[i](finalHandler)
	}
	s.router.Handle(pattern, finalHandler)
}

// HandleWithMiddleware registers the handler with the given pattern and applies
// the specified middlewares to the handler, in addition to the server middlewares.
func (s *Server) HandleWithMiddleware(pattern string, handler http.Handler, middleware ...mux.MiddlewareFunc) {
	// Apply handler-specific middlewares first (in reverse order)
	finalHandler := handler
	for i := len(middleware) - 1; i >= 0; i-- {
		finalHandler = middleware[i](finalHandler)
	}

	// Then apply server middlewares
	for i := len(s.middlewares) - 1; i >= 0; i-- {
		finalHandler = s.middlewares[i](finalHandler)
	}

	s.router.Handle(pattern, finalHandler)
}

// HandleFunc registers a handler function with the given pattern in the Server's router.
// It applies all server middlewares to the handler.
func (s *Server) HandleFunc(pattern string, handlerFunc func(http.ResponseWriter, *http.Request)) {
	s.Handle(pattern, http.HandlerFunc(handlerFunc))
}

// GET registers a GET handler with path parameter support
// Pattern can include path parameters like "/user/{id}" or "/user/{id:[0-9]+}"
func (s *Server) GET(pattern string, handler http.Handler) {
	// Apply all middlewares in reverse order
	finalHandler := handler
	for i := len(s.middlewares) - 1; i >= 0; i-- {
		finalHandler = s.middlewares[i](finalHandler)
	}
	s.router.Handle(pattern, finalHandler).Methods("GET")
}

// POST registers a POST handler with path parameter support
// Pattern can include path parameters like "/user/{id}" or "/user/{id:[0-9]+}"
func (s *Server) POST(pattern string, handler http.Handler) {
	// Apply all middlewares in reverse order
	finalHandler := handler
	for i := len(s.middlewares) - 1; i >= 0; i-- {
		finalHandler = s.middlewares[i](finalHandler)
	}
	s.router.Handle(pattern, finalHandler).Methods("POST")
}

// PUT registers a PUT handler with path parameter support
// Pattern can include path parameters like "/user/{id}" or "/user/{id:[0-9]+}"
func (s *Server) PUT(pattern string, handler http.Handler) {
	// Apply all middlewares in reverse order
	finalHandler := handler
	for i := len(s.middlewares) - 1; i >= 0; i-- {
		finalHandler = s.middlewares[i](finalHandler)
	}
	s.router.Handle(pattern, finalHandler).Methods("PUT")
}

// DELETE registers a DELETE handler with path parameter support
// Pattern can include path parameters like "/user/{id}" or "/user/{id:[0-9]+}"
func (s *Server) DELETE(pattern string, handler http.Handler) {
	// Apply all middlewares in reverse order
	finalHandler := handler
	for i := len(s.middlewares) - 1; i >= 0; i-- {
		finalHandler = s.middlewares[i](finalHandler)
	}
	s.router.Handle(pattern, finalHandler).Methods("DELETE")
}

// PATCH registers a PATCH handler with path parameter support
// Pattern can include path parameters like "/user/{id}" or "/user/{id:[0-9]+}"
func (s *Server) PATCH(pattern string, handler http.Handler) {
	// Apply all middlewares in reverse order
	finalHandler := handler
	for i := len(s.middlewares) - 1; i >= 0; i-- {
		finalHandler = s.middlewares[i](finalHandler)
	}
	s.router.Handle(pattern, finalHandler).Methods("PATCH")
}

// Start initializes the server and begins listening for incoming HTTP requests on the configured address.
func (s *Server) Start() error {
	if s.app == nil {
		return fmt.Errorf("logmanager application is not set")
	}
	if !s.hasSetUpCORS {
		return fmt.Errorf("CORS middleware is not set")
	}
	fmt.Println("starting server on: ", s.server.Addr)
	return s.server.ListenAndServe()
}

// Stop gracefully shuts down the server without interrupting any active connections.
func (s *Server) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

// registerHealthCheck registers the health check endpoint on the server
func (s *Server) registerHealthCheck() {
	healthHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	s.router.Handle(s.healthCheckPath, healthHandler).Methods("GET")
}
