package domain

import (
	"fmt"
	"regexp"
	"strings"
)

// Credentials represents user login credentials in the domain layer
type Credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// credentialsEmailRegex is a regex pattern for validating email addresses in credentials
var credentialsEmailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

const minPasswordLength = 8

// Validate validates the credentials fields
// Returns an error if any validation fails
func (c Credentials) Validate() error {
	var errors []string

	// Validate email
	if c.Email == "" {
		errors = append(errors, "email is required")
	} else if !credentialsEmailRegex.MatchString(c.Email) {
		errors = append(errors, "invalid email format")
	}

	// Validate password
	if c.Password == "" {
		errors = append(errors, "password is required")
	} else if len(strings.TrimSpace(c.Password)) < minPasswordLength {
		errors = append(errors, fmt.Sprintf("password must be at least %d characters", minPasswordLength))
	}

	if len(errors) > 0 {
		return fmt.Errorf("credentials validation failed: %s", strings.Join(errors, "; "))
	}

	return nil
}