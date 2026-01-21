package auth

import "time"

// User represents a user in the system
type User struct {
	ID       int64
	Username string
	Password []byte // bcrypt hashed password
}

// Session represents an active user session
type Session struct {
	ID           string    // Cryptographically random session token
	UserID       int64
	CreatedAt    time.Time
	ExpiresAt    time.Time
	LastActivity time.Time
	IPAddress    string // Optional: for security monitoring
	UserAgent    string // Optional: for security monitoring
}

// LoginRequest represents the login payload
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// RegisterRequest represents the registration payload
type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// AuthResponse represents the response after login/register
type AuthResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	User    *User  `json:"user,omitempty"`
}
