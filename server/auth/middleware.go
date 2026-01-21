package auth

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

const (
	userContextKey    contextKey = "user"
	sessionContextKey contextKey = "session"
)

// AuthMiddleware is middleware that validates user sessions
type AuthMiddleware struct {
	db *sql.DB
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(db *sql.DB) *AuthMiddleware {
	return &AuthMiddleware{db: db}
}

// RequireAuth is middleware that requires authentication
// It validates the session and attaches user info to the request context
func (m *AuthMiddleware) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get session from cookie
		sessionID, err := GetSessionFromRequest(r)
		if err != nil {
			respondUnauthorized(w, "Not authenticated")
			return
		}

		// Validate session and get user
		session, user, err := ValidateAndRefreshSession(sessionID, m.db)
		if err != nil {
			ClearSessionCookie(w)
			respondUnauthorized(w, "Session expired or invalid")
			return
		}

		// Add user and session to request context
		ctx := context.WithValue(r.Context(), userContextKey, user)
		ctx = context.WithValue(ctx, sessionContextKey, session)

		// Call the next handler with updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// OptionalAuth is middleware that checks for authentication but doesn't require it
// If authenticated, it attaches user info to the request context
func (m *AuthMiddleware) OptionalAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get session from cookie
		sessionID, err := GetSessionFromRequest(r)
		if err != nil {
			// No session, continue without user context
			next.ServeHTTP(w, r)
			return
		}

		// Validate session and get user
		session, user, err := ValidateAndRefreshSession(sessionID, m.db)
		if err != nil {
			// Invalid session, clear cookie and continue
			ClearSessionCookie(w)
			next.ServeHTTP(w, r)
			return
		}

		// Add user and session to request context
		ctx := context.WithValue(r.Context(), userContextKey, user)
		ctx = context.WithValue(ctx, sessionContextKey, session)

		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// GetUserFromContext retrieves the authenticated user from the request context
func GetUserFromContext(r *http.Request) (*User, bool) {
	user, ok := r.Context().Value(userContextKey).(*User)
	return user, ok
}

// GetSessionFromContext retrieves the session from the request context
func GetSessionFromContext(r *http.Request) (*Session, bool) {
	session, ok := r.Context().Value(sessionContextKey).(*Session)
	return session, ok
}

// respondUnauthorized sends a JSON unauthorized response
func respondUnauthorized(w http.ResponseWriter, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	json.NewEncoder(w).Encode(AuthResponse{
		Success: false,
		Message: message,
	})
}
