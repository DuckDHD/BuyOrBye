package domain

import (
	"fmt"
	"strings"
	"time"
)

// TokenPair represents a pair of access and refresh tokens
type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"` // Expiration time in seconds
}

// TokenClaims represents the claims stored in JWT tokens
type TokenClaims struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	ExpiresAt int64  `json:"expires_at"` // Unix timestamp
}

// IsExpired checks if the token claims have expired
func (tc TokenClaims) IsExpired() bool {
	return time.Now().Unix() >= tc.ExpiresAt
}

// Validate validates all required fields for a TokenPair
// Returns an error if any validation fails
func (tp TokenPair) Validate() error {
	var errors []string

	// Validate access token
	if strings.TrimSpace(tp.AccessToken) == "" {
		errors = append(errors, "access token is required")
	}

	// Validate refresh token
	if strings.TrimSpace(tp.RefreshToken) == "" {
		errors = append(errors, "refresh token is required")
	}

	// Validate expires in
	if tp.ExpiresIn <= 0 {
		errors = append(errors, "expires in must be greater than 0")
	}

	if len(errors) > 0 {
		return fmt.Errorf("token pair validation failed: %s", strings.Join(errors, "; "))
	}

	return nil
}