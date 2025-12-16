package middleware

import (
	"context"
	"net/http"

	"github.com/gorilla/mux"
)

// UserContextKey is the context key for user information
type UserContextKey struct{}

// User represents authenticated user information
type User struct {
	ID    string
	Email string
	Role  string
}

// TokenService interface for token validation
type TokenService interface {
	ValidateToken(token string) (*User, error)
}

// KongAuth validates the token using Kong gateway simulation.
// In production, Kong handles this, but in local environment we need to validate manually.
func KongAuth(tokenService TokenService) mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("Authorization")
			if token == "" {
				http.Error(w, "Unauthorized: missing token", http.StatusUnauthorized)
				return
			}

			// Remove "Bearer " prefix if present
			if len(token) > 7 && token[:7] == "Bearer " {
				token = token[7:]
			}

			user, err := tokenService.ValidateToken(token)
			if err != nil {
				http.Error(w, "Unauthorized: invalid token", http.StatusUnauthorized)
				return
			}

			// Add user to context
			ctx := context.WithValue(r.Context(), UserContextKey{}, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// ExtractUser extracts user information from headers (set by Kong in production)
// or from context (set by KongAuth middleware in local).
// This middleware expects X-User-ID, X-User-Email headers to be present in production.
func ExtractUser() mux.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if user already in context (from KongAuth middleware)
			if user := r.Context().Value(UserContextKey{}); user != nil {
				next.ServeHTTP(w, r)
				return
			}

			// Extract from headers (production mode - Kong sets these)
			userID := r.Header.Get("X-User-ID")
			userEmail := r.Header.Get("X-User-Email")
			userRole := r.Header.Get("X-User-Role")

			if userID == "" {
				http.Error(w, "Unauthorized: missing user information", http.StatusUnauthorized)
				return
			}

			user := &User{
				ID:    userID,
				Email: userEmail,
				Role:  userRole,
			}

			ctx := context.WithValue(r.Context(), UserContextKey{}, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUser retrieves user information from context
func GetUser(ctx context.Context) *User {
	if user, ok := ctx.Value(UserContextKey{}).(*User); ok {
		return user
	}
	return nil
}
