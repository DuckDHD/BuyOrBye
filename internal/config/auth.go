package config

import (
	"fmt"
	"time"
)

// AuthService provides authentication configuration management
type AuthService interface {
	// GetJWTSecret returns the JWT signing secret
	GetJWTSecret() string

	// GetBCryptCost returns the bcrypt hashing cost
	GetBCryptCost() int

	// GetAccessTokenTTL returns access token time-to-live
	GetAccessTokenTTL() time.Duration

	// GetRefreshTokenTTL returns refresh token time-to-live
	GetRefreshTokenTTL() time.Duration

	// GetCSRFSecret returns the CSRF protection secret
	GetCSRFSecret() string

	// IsSecure returns true if authentication is configured securely
	IsSecure() bool
}

// NewAuthService creates a new auth service from full configuration
func NewAuthService(config *Config) AuthService {
	return &authService{
		authConfig:     &config.Auth,
		securityConfig: &config.Security,
	}
}

// NewAuthServiceFromConfig creates auth service from auth config only (for compatibility)
func NewAuthServiceFromConfig(authConfig *AuthConfig) (AuthService, error) {
	// For backward compatibility, create a minimal security config
	cfg := &Config{
		Auth: *authConfig,
		Security: SecurityConfig{
			CSRFSecret: authConfig.JWTSecret, // Use JWT secret as fallback
		},
	}
	return NewAuthService(cfg), nil
}

type authService struct {
	authConfig     *AuthConfig
	securityConfig *SecurityConfig
}

func (a *authService) GetJWTSecret() string {
	return a.authConfig.JWTSecret
}

func (a *authService) GetBCryptCost() int {
	return a.authConfig.BCryptCost
}

func (a *authService) GetAccessTokenTTL() time.Duration {
	return a.authConfig.AccessTokenTTL
}

func (a *authService) GetRefreshTokenTTL() time.Duration {
	return a.authConfig.RefreshTokenTTL
}

func (a *authService) GetCSRFSecret() string {
	return a.securityConfig.CSRFSecret
}

func (a *authService) IsSecure() bool {
	// Check if secrets are long enough
	if len(a.authConfig.JWTSecret) < 32 {
		return false
	}

	if len(a.securityConfig.CSRFSecret) < 32 {
		return false
	}

	// Check if bcrypt cost is reasonable
	if a.authConfig.BCryptCost < 10 || a.authConfig.BCryptCost > 15 {
		return false
	}

	// Check token TTLs
	if a.authConfig.AccessTokenTTL > time.Hour || a.authConfig.AccessTokenTTL < time.Minute {
		return false
	}

	if a.authConfig.RefreshTokenTTL > 30*24*time.Hour || a.authConfig.RefreshTokenTTL < time.Hour {
		return false
	}

	return true
}

// ValidateAuthConfig validates authentication configuration
func ValidateAuthConfig(config *AuthConfig) error {
	if config.JWTSecret == "" {
		return fmt.Errorf("JWT secret cannot be empty")
	}

	if len(config.JWTSecret) < 32 {
		return fmt.Errorf("JWT secret must be at least 32 characters")
	}

	if config.BCryptCost < 4 || config.BCryptCost > 20 {
		return fmt.Errorf("bcrypt cost must be between 4 and 20")
	}

	if config.AccessTokenTTL == 0 {
		return fmt.Errorf("access token TTL cannot be zero")
	}

	if config.RefreshTokenTTL == 0 {
		return fmt.Errorf("refresh token TTL cannot be zero")
	}

	return nil
}

// ValidateSecurityConfig validates security configuration
func ValidateSecurityConfig(config *SecurityConfig) error {
	if config.CSRFSecret == "" {
		return fmt.Errorf("CSRF secret cannot be empty")
	}

	if len(config.CSRFSecret) < 32 {
		return fmt.Errorf("CSRF secret must be at least 32 characters")
	}

	return nil
}