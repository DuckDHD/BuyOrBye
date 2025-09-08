package domain

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// User represents a user in the domain layer
type User struct {
	ID           string    `json:"id"`
	Email        string    `json:"email"`
	Name         string    `json:"name"`
	PasswordHash string    `json:"-"` // Never serialize password hash
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// emailRegex is a regex pattern for validating email addresses
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

// Validate validates all required fields for a User
// Returns an error if any validation fails
func (u User) Validate() error {
	var errors []string

	// Validate email
	if u.Email == "" {
		errors = append(errors, "email is required")
	} else if !emailRegex.MatchString(u.Email) {
		errors = append(errors, "invalid email format")
	}

	// Validate name
	if strings.TrimSpace(u.Name) == "" {
		errors = append(errors, "name is required")
	}

	// Validate password hash
	if u.PasswordHash == "" {
		errors = append(errors, "password hash is required")
	}

	// Validate created at
	if u.CreatedAt.IsZero() {
		errors = append(errors, "created at is required")
	}

	// Validate updated at
	if u.UpdatedAt.IsZero() {
		errors = append(errors, "updated at is required")
	}

	if len(errors) > 0 {
		return fmt.Errorf("user validation failed: %s", strings.Join(errors, "; "))
	}

	return nil
}