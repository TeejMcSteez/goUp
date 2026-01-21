package auth

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"net/http"
	"time"
)

const (
	// Session token length in bytes (32 bytes = 256 bits)
	sessionTokenBytes = 32

	// Session expiration times
	sessionDuration = 24 * time.Hour  // Hard expiration
	idleTimeout     = 30 * time.Minute // Idle timeout

	// Cookie name for session
	sessionCookieName = "session_token"
)

// GenerateSessionToken creates a cryptographically secure random session token
func GenerateSessionToken() (string, error) {
	bytes := make([]byte, sessionTokenBytes)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}

	// Encode to base64 URL-safe format
	token := base64.URLEncoding.EncodeToString(bytes)
	return token, nil
}

// NewSession creates a new session for a user
func NewSession(userID int64, ipAddress, userAgent string) (*Session, error) {
	token, err := GenerateSessionToken()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	return &Session{
		ID:           token,
		UserID:       userID,
		CreatedAt:    now,
		ExpiresAt:    now.Add(sessionDuration),
		LastActivity: now,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
	}, nil
}

// SetSessionCookie sets the session cookie on the HTTP response
func SetSessionCookie(w http.ResponseWriter, sessionID string, secure bool) {
	cookie := &http.Cookie{
		Name:     sessionCookieName,
		Value:    sessionID,
		Path:     "/",
		HttpOnly: true,                // Prevents JavaScript access (XSS protection)
		Secure:   secure,              // Only send over HTTPS (set to true in production)
		SameSite: http.SameSiteLaxMode, // CSRF protection
		MaxAge:   int(sessionDuration.Seconds()),
	}

	http.SetCookie(w, cookie)
}

// ClearSessionCookie removes the session cookie
func ClearSessionCookie(w http.ResponseWriter) {
	cookie := &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1, // Expire immediately
	}

	http.SetCookie(w, cookie)
}

// GetSessionFromRequest retrieves the session ID from the request cookie
func GetSessionFromRequest(r *http.Request) (string, error) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		return "", err
	}

	return cookie.Value, nil
}

// ValidateAndRefreshSession checks if a session is valid and refreshes activity
// Returns the session and user if valid, error otherwise
func ValidateAndRefreshSession(sessionID string, db *sql.DB) (*Session, *User, error) {
	// Get session from database
	session, err := GetSession(sessionID, db)
	if err != nil {
		return nil, nil, err
	}

	// Check idle timeout
	if time.Since(session.LastActivity) > idleTimeout {
		DeleteSession(sessionID, db)
		return nil, nil, ErrSessionExpired
	}

	// Update last activity
	err = UpdateSessionActivity(sessionID, db)
	if err != nil {
		return nil, nil, err
	}

	// Get user
	user, err := GetUserByID(session.UserID, db)
	if err != nil {
		return nil, nil, err
	}

	// Don't expose password hash
	user.Password = nil

	return session, user, nil
}

// GetClientIP extracts the client's IP address from the request
func GetClientIP(r *http.Request) string {
	// Check for forwarded IP first (if behind a proxy)
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		return forwarded
	}

	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	return r.RemoteAddr
}
