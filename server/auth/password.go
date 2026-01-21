package auth

import (
	"errors"
	"golang.org/x/crypto/bcrypt"
)

const (
	// bcrypt cost factor (10-12 is recommended for production)
	// Higher values are more secure but slower
	bcryptCost = 12
)

var (
	ErrPasswordTooShort = errors.New("password must be at least 8 characters")
	ErrInvalidPassword  = errors.New("invalid password")
)

// HashPassword takes a plaintext password and returns a bcrypt hash
func HashPassword(password string) ([]byte, error) {
	if len(password) < 8 {
		return nil, ErrPasswordTooShort
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return nil, err
	}

	return hash, nil
}

// ComparePassword compares a plaintext password with a bcrypt hash
// Returns nil if the password matches, error otherwise
func ComparePassword(hashedPassword []byte, password string) error {
	err := bcrypt.CompareHashAndPassword(hashedPassword, []byte(password))
	if err != nil {
		if err == bcrypt.ErrMismatchedHashAndPassword {
			return ErrInvalidPassword
		}
		return err
	}

	return nil
}

// ValidatePassword checks if a password meets minimum requirements
func ValidatePassword(password string) error {
	if len(password) < 8 {
		return ErrPasswordTooShort
	}

	// Add additional password requirements here if needed:
	// - Must contain uppercase
	// - Must contain numbers
	// - Must contain special characters
	// etc.

	return nil
}
