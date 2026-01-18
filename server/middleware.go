package server

import (
	"log"
	"net/http"
)

// AuthStrategy is an interface for different authentication methods.
type AuthStrategy interface {
	// Authenticate checks the request and returns true if it's authenticated.
	// It can also return an error if the authentication mechanism itself fails
	// (e.g. malformed header).
	Authenticate(r *http.Request) (bool, error)
}

// APIKeyAuth is a strategy for authenticating with an API key in a header.
type APIKeyAuth struct {
	ExpectedKey string // The API key to check against.
}

// Authenticate implements the AuthStrategy interface for APIKeyAuth.
func (aka *APIKeyAuth) Authenticate(r *http.Request) (bool, error) {
	key := r.Header.Get("X-API-Key")
	if key == "" {
		// No API key provided.
		return false, nil
	}

	if key == aka.ExpectedKey {
		return true, nil
	}

	// Invalid API Key.
	return false, nil
}

type NoAuth struct{}

// Authenticate implements the AuthStrategy interface for NoAuth.
func (na *NoAuth) Authenticate(r *http.Request) (bool, error) {
	return true, nil
}

// --- Middleware to use the strategies ---

// AuthMiddleware is a middleware constructor that uses an AuthStrategy.
// It takes a strategy and returns a function that wraps an http.Handler.
func AuthMiddleware(strategy AuthStrategy) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authenticated, err := strategy.Authenticate(r)

			if err != nil {
				// This case is for when the authentication check itself fails.
				log.Printf("Authentication error: %v", err)
				http.Error(w, "400 Bad Request", http.StatusBadRequest)
				return
			}

			if !authenticated {
				// For Basic Auth, you'd typically send this header to prompt the browser.
				w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
				http.Error(w, "401 Unauthorized", http.StatusUnauthorized)
				return
			}

			// If authenticated, call the next handler in the chain.
			next.ServeHTTP(w, r)
		})
	}
}
