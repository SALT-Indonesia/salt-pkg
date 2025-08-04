package httpmanager

import (
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

// CORSMiddleware creates a middleware that adds CORS headers to the response
func CORSMiddleware(allowedOrigins, allowedMethods, allowedHeaders []string, allowCredentials bool) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Set default values if not provided
			origins := "*"
			if len(allowedOrigins) > 0 {
				origins = allowedOrigins[0]
				for _, origin := range allowedOrigins[1:] {
					origins += ", " + origin
				}
			}

			methods := "GET, POST, PUT, DELETE, OPTIONS"
			if len(allowedMethods) > 0 {
				methods = allowedMethods[0]
				for _, method := range allowedMethods[1:] {
					methods += ", " + method
				}
			}

			headers := "Accept, Content-Type, Content-Length, Accept-Encoding, Authorization, X-CSRF-Token"
			if len(allowedHeaders) > 0 {
				headers = allowedHeaders[0]
				for _, header := range allowedHeaders[1:] {
					headers += ", " + header
				}
			}

			// Set CORS headers
			w.Header().Set("Access-Control-Allow-Origin", origins)
			w.Header().Set("Access-Control-Allow-Methods", methods)
			w.Header().Set("Access-Control-Allow-Headers", headers)

			if allowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			// Handle preflight requests
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}

			log.Println("CORS")

			// Call the next handler
			next.ServeHTTP(w, r)
		})
	}
}

// EnableCORS adds CORS middleware to an existing server
func (s *Server) EnableCORS(allowedOrigins, allowedMethods, allowedHeaders []string, allowCredentials bool) {
	s.Use(CORSMiddleware(allowedOrigins, allowedMethods, allowedHeaders, allowCredentials))
	s.hasSetUpCORS = true
}
