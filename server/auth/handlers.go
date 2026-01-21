package auth

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

// AuthHandler handles all authentication-related requests
type AuthHandler struct {
	db     *sql.DB
	secure bool // Set to true in production to enable HTTPS-only cookies
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(db *sql.DB, secure bool) *AuthHandler {
	return &AuthHandler{
		db:     db,
		secure: secure,
	}
}

// Register handles user registration
// POST /auth/register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate input
	req.Username = strings.TrimSpace(req.Username)
	if req.Username == "" {
		respondJSON(w, http.StatusBadRequest, AuthResponse{
			Success: false,
			Message: "Username is required",
		})
		return
	}

	if err := ValidatePassword(req.Password); err != nil {
		respondJSON(w, http.StatusBadRequest, AuthResponse{
			Success: false,
			Message: err.Error(),
		})
		return
	}

	// Hash password
	hashedPassword, err := HashPassword(req.Password)
	if err != nil {
		log.Printf("Error hashing password: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Create user
	user, err := InsertUser(req.Username, hashedPassword, h.db)
	if err != nil {
		if err == ErrUserExists {
			respondJSON(w, http.StatusConflict, AuthResponse{
				Success: false,
				Message: "Username already exists",
			})
			return
		}
		log.Printf("Error creating user: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Don't expose password hash
	user.Password = nil

	// Auto-login after registration: create session
	ipAddress := GetClientIP(r)
	userAgent := r.UserAgent()

	session, err := NewSession(user.ID, ipAddress, userAgent)
	if err != nil {
		log.Printf("Error creating session: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = CreateSession(session, h.db)
	if err != nil {
		log.Printf("Error storing session: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Set session cookie
	SetSessionCookie(w, session.ID, h.secure)

	respondJSON(w, http.StatusCreated, AuthResponse{
		Success: true,
		Message: "Registration successful",
		User:    user,
	})
}

// Login handles user login
// POST /auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get user from database
	user, err := GetUserByUsername(req.Username, h.db)
	if err != nil {
		// Return generic error to prevent username enumeration
		respondJSON(w, http.StatusUnauthorized, AuthResponse{
			Success: false,
			Message: "Invalid username or password",
		})
		return
	}

	// Compare password
	err = ComparePassword(user.Password, req.Password)
	if err != nil {
		// Return same generic error
		respondJSON(w, http.StatusUnauthorized, AuthResponse{
			Success: false,
			Message: "Invalid username or password",
		})
		return
	}

	// Create new session
	ipAddress := GetClientIP(r)
	userAgent := r.UserAgent()

	session, err := NewSession(user.ID, ipAddress, userAgent)
	if err != nil {
		log.Printf("Error creating session: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	err = CreateSession(session, h.db)
	if err != nil {
		log.Printf("Error storing session: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Set session cookie
	SetSessionCookie(w, session.ID, h.secure)

	// Don't expose password hash
	user.Password = nil

	respondJSON(w, http.StatusOK, AuthResponse{
		Success: true,
		Message: "Login successful",
		User:    user,
	})
}

// Logout handles user logout
// POST /auth/logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get session from cookie
	sessionID, err := GetSessionFromRequest(r)
	if err != nil {
		// No session cookie, just clear it anyway
		ClearSessionCookie(w)
		respondJSON(w, http.StatusOK, AuthResponse{
			Success: true,
			Message: "Logged out",
		})
		return
	}

	// Delete session from database
	err = DeleteSession(sessionID, h.db)
	if err != nil {
		log.Printf("Error deleting session: %v", err)
	}

	// Clear session cookie
	ClearSessionCookie(w)

	respondJSON(w, http.StatusOK, AuthResponse{
		Success: true,
		Message: "Logged out successfully",
	})
}

// Me returns the current user's information
// GET /auth/me
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get session from cookie
	sessionID, err := GetSessionFromRequest(r)
	if err != nil {
		respondJSON(w, http.StatusUnauthorized, AuthResponse{
			Success: false,
			Message: "Not authenticated",
		})
		return
	}

	// Validate session and get user
	_, user, err := ValidateAndRefreshSession(sessionID, h.db)
	if err != nil {
		ClearSessionCookie(w)
		respondJSON(w, http.StatusUnauthorized, AuthResponse{
			Success: false,
			Message: "Session expired",
		})
		return
	}

	respondJSON(w, http.StatusOK, AuthResponse{
		Success: true,
		User:    user,
	})
}

// respondJSON is a helper function to send JSON responses
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
