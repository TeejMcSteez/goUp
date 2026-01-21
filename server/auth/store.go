package auth

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"time"
)

var (
	ErrUserNotFound     = errors.New("user not found")
	ErrSessionNotFound  = errors.New("session not found")
	ErrSessionExpired   = errors.New("session expired")
	ErrUserExists       = errors.New("user already exists")
)

// SetupStore initializes the auth database with users and sessions tables
func SetupStore() (*sql.DB, error) {
	db, err := sql.Open("sqlite", "./.ad.db")
	if err != nil {
		log.Printf("Error setting up auth database: %v", err)
		return nil, err
	}
	db.SetMaxOpenConns(1)

	// Create users table
	createUsersTableSQL := `CREATE TABLE IF NOT EXISTS users (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT UNIQUE NOT NULL,
		password BLOB NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	_, err = db.ExecContext(context.Background(), createUsersTableSQL)
	if err != nil {
		return nil, err
	}

	// Create sessions table
	createSessionsTableSQL := `CREATE TABLE IF NOT EXISTS sessions (
		session_id TEXT PRIMARY KEY,
		user_id INTEGER NOT NULL,
		created_at TIMESTAMP NOT NULL,
		expires_at TIMESTAMP NOT NULL,
		last_activity TIMESTAMP NOT NULL,
		ip_address TEXT,
		user_agent TEXT,
		FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
	);`

	_, err = db.ExecContext(context.Background(), createSessionsTableSQL)
	if err != nil {
		return nil, err
	}

	// Create indexes for better performance
	createIndexesSQL := []string{
		`CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON sessions(user_id);`,
		`CREATE INDEX IF NOT EXISTS idx_sessions_expires_at ON sessions(expires_at);`,
	}

	for _, sql := range createIndexesSQL {
		_, err = db.ExecContext(context.Background(), sql)
		if err != nil {
			return nil, err
		}
	}

	log.Println("Successfully setup auth database")
	return db, nil
}

// InsertUser creates a new user with hashed password
func InsertUser(username string, hashedPassword []byte, db *sql.DB) (*User, error) {
	insertUserSQL := `INSERT INTO users (username, password) VALUES (?, ?);`

	result, err := db.Exec(insertUserSQL, username, hashedPassword)
	if err != nil {
		// Check for unique constraint violation
		if err.Error() == "UNIQUE constraint failed: users.username" {
			return nil, ErrUserExists
		}
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &User{
		ID:       id,
		Username: username,
		Password: hashedPassword,
	}, nil
}

// GetUserByUsername retrieves a user by username
func GetUserByUsername(username string, db *sql.DB) (*User, error) {
	query := `SELECT id, username, password FROM users WHERE username = ? LIMIT 1;`

	var user User
	err := db.QueryRow(query, username).Scan(&user.ID, &user.Username, &user.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

// GetUserByID retrieves a user by ID
func GetUserByID(userID int64, db *sql.DB) (*User, error) {
	query := `SELECT id, username, password FROM users WHERE id = ? LIMIT 1;`

	var user User
	err := db.QueryRow(query, userID).Scan(&user.ID, &user.Username, &user.Password)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

// CreateSession stores a new session in the database
func CreateSession(session *Session, db *sql.DB) error {
	insertSessionSQL := `INSERT INTO sessions
		(session_id, user_id, created_at, expires_at, last_activity, ip_address, user_agent)
		VALUES (?, ?, ?, ?, ?, ?, ?);`

	_, err := db.Exec(insertSessionSQL,
		session.ID,
		session.UserID,
		session.CreatedAt,
		session.ExpiresAt,
		session.LastActivity,
		session.IPAddress,
		session.UserAgent,
	)

	return err
}

// GetSession retrieves a session by session ID and validates expiration
func GetSession(sessionID string, db *sql.DB) (*Session, error) {
	query := `SELECT session_id, user_id, created_at, expires_at, last_activity, ip_address, user_agent
		FROM sessions WHERE session_id = ? LIMIT 1;`

	var session Session
	err := db.QueryRow(query, sessionID).Scan(
		&session.ID,
		&session.UserID,
		&session.CreatedAt,
		&session.ExpiresAt,
		&session.LastActivity,
		&session.IPAddress,
		&session.UserAgent,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrSessionNotFound
		}
		return nil, err
	}

	// Check if session is expired
	if time.Now().After(session.ExpiresAt) {
		// Delete expired session
		DeleteSession(sessionID, db)
		return nil, ErrSessionExpired
	}

	return &session, nil
}

// UpdateSessionActivity updates the last_activity timestamp for a session
func UpdateSessionActivity(sessionID string, db *sql.DB) error {
	updateSQL := `UPDATE sessions SET last_activity = ? WHERE session_id = ?;`
	_, err := db.Exec(updateSQL, time.Now(), sessionID)
	return err
}

// DeleteSession removes a session from the database
func DeleteSession(sessionID string, db *sql.DB) error {
	deleteSQL := `DELETE FROM sessions WHERE session_id = ?;`
	_, err := db.Exec(deleteSQL, sessionID)
	return err
}

// DeleteUserSessions removes all sessions for a specific user (logout all devices)
func DeleteUserSessions(userID int64, db *sql.DB) error {
	deleteSQL := `DELETE FROM sessions WHERE user_id = ?;`
	_, err := db.Exec(deleteSQL, userID)
	return err
}

// CleanupExpiredSessions removes all expired sessions from the database
// This should be called periodically (e.g., via a background job)
func CleanupExpiredSessions(db *sql.DB) error {
	deleteSQL := `DELETE FROM sessions WHERE expires_at < ?;`
	result, err := db.Exec(deleteSQL, time.Now())
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		log.Printf("Cleaned up %d expired sessions", rowsAffected)
	}

	return nil
}
