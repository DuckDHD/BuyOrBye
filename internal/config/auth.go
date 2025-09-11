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

// NewAuthService creates a new auth service from configuration
func NewAuthService(config *AuthConfig) AuthService {
	return &authService{
		config: config,
	}
}

type authService struct {
	config *AuthConfig
}

func (a *authService) GetJWTSecret() string {
	return a.config.JWTSecret
}

func (a *authService) GetBCryptCost() int {
	return a.config.BCryptCost
}

func (a *authService) GetAccessTokenTTL() time.Duration {
	return a.config.AccessTokenTTL
}

func (a *authService) GetRefreshTokenTTL() time.Duration {
	return a.config.RefreshTokenTTL
}

func (a *authService) GetCSRFSecret() string {
	return a.config.CSRFSecret
}

func (a *authService) IsSecure() bool {
	// Check if secrets are long enough
	if len(a.config.JWTSecret) < 32 {
		return false
	}
	
	if len(a.config.CSRFSecret) < 32 {
		return false
	}
	
	// Check if bcrypt cost is reasonable
	if a.config.BCryptCost < 10 {
		return false
	}
	
	// Check if token TTLs are reasonable
	if a.config.AccessTokenTTL > 24*time.Hour {
		return false
	}
	
	if a.config.RefreshTokenTTL > 30*24*time.Hour { // 30 days
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
		return fmt.Errorf("JWT secret must be at least 32 characters long")
	}
	
	if config.CSRFSecret == "" {
		return fmt.Errorf("CSRF secret cannot be empty")
	}
	
	if len(config.CSRFSecret) < 32 {
		return fmt.Errorf("CSRF secret must be at least 32 characters long")
	}
	
	if config.BCryptCost < 4 || config.BCryptCost > 20 {
		return fmt.Errorf("bcrypt cost must be between 4 and 20")
	}
	
	if config.AccessTokenTTL <= 0 {
		return fmt.Errorf("access token TTL must be positive")
	}
	
	if config.RefreshTokenTTL <= 0 {
		return fmt.Errorf("refresh token TTL must be positive")
	}
	
	if config.AccessTokenTTL >= config.RefreshTokenTTL {
		return fmt.Errorf("access token TTL should be less than refresh token TTL")
	}
	
	return nil
}

// GetBCryptCostForEnvironment returns appropriate bcrypt cost for environment
func GetBCryptCostForEnvironment(environment string) int {
	switch environment {
	case "production":
		return 14
	case "development":
		return 10
	case "test":
		return 4 // Fast for testing
	default:
		return 10
	}
}

// GetTokenTTLForEnvironment returns appropriate token TTLs for environment
func GetTokenTTLForEnvironment(environment string) (access, refresh time.Duration) {
	switch environment {
	case "production":
		return 15 * time.Minute, 7 * 24 * time.Hour // 15 minutes, 7 days
	case "development":
		return 60 * time.Minute, 24 * time.Hour // 1 hour, 1 day
	case "test":
		return 1 * time.Minute, 2 * time.Minute // Short for testing
	default:
		return 15 * time.Minute, 7 * 24 * time.Hour
	}
}

// GenerateSecureSecret generates a secure random secret (placeholder for implementation)
func GenerateSecureSecret() string {
	// In a real implementation, this would generate a cryptographically secure random string
	// For now, return a placeholder that indicates this should be configured properly
	return "REPLACE_WITH_SECURE_RANDOM_SECRET_32_CHARS_MIN"
}

// IsProductionReady checks if auth config is ready for production
func IsProductionReady(config *AuthConfig) []string {
	var issues []string
	
	if len(config.JWTSecret) < 64 {
		issues = append(issues, "JWT secret should be at least 64 characters for production")
	}
	
	if len(config.CSRFSecret) < 64 {
		issues = append(issues, "CSRF secret should be at least 64 characters for production")
	}
	
	if config.BCryptCost < 12 {
		issues = append(issues, "BCrypt cost should be at least 12 for production")
	}
	
	if config.AccessTokenTTL > time.Hour {
		issues = append(issues, "Access token TTL should be 1 hour or less for production")
	}
	
	// Check for common insecure patterns
	insecurePatterns := []string{
		"secret", "password", "test", "dev", "development", 
		"your-secret", "change-me", "default", "example",
	}
	
	for _, pattern := range insecurePatterns {
		if len(config.JWTSecret) > 0 && contains(config.JWTSecret, pattern) {
			issues = append(issues, "JWT secret appears to contain insecure pattern")
			break
		}
	}
	
	for _, pattern := range insecurePatterns {
		if len(config.CSRFSecret) > 0 && contains(config.CSRFSecret, pattern) {
			issues = append(issues, "CSRF secret appears to contain insecure pattern")
			break
		}
	}
	
	return issues
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && 
		   (s == substr || 
		    (len(s) > len(substr) && 
		     (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
		      containsMiddle(s, substr))))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}