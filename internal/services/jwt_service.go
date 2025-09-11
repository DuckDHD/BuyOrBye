package services

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	
	"github.com/DuckDHD/BuyOrBye/internal/config"
	"github.com/DuckDHD/BuyOrBye/internal/domain"
)

// JWTService defines the interface for JWT token operations
type JWTService interface {
	// GenerateTokenPair creates both access and refresh tokens for a user
	GenerateTokenPair(userID, email string) (*domain.TokenPair, error)
	
	// ValidateAccessToken validates an access token and returns its claims
	ValidateAccessToken(tokenString string) (*domain.TokenClaims, error)
	
	// ValidateRefreshToken validates a refresh token and returns its claims
	ValidateRefreshToken(tokenString string) (*domain.TokenClaims, error)
}

// jwtService implements JWTService using github.com/golang-jwt/jwt/v5
type jwtService struct {
	secret            []byte
	accessTokenTTL    time.Duration // 15 minutes
	refreshTokenTTL   time.Duration // 7 days
}

// NewJWTService creates a new JWT service instance
// Requires JWT_SECRET environment variable to be set
func NewJWTService() (JWTService, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return nil, fmt.Errorf("JWT_SECRET environment variable is required")
	}

	return &jwtService{
		secret:            []byte(secret),
		accessTokenTTL:    15 * time.Minute, // 15 minutes as specified
		refreshTokenTTL:   7 * 24 * time.Hour, // 7 days as specified
	}, nil
}

// NewJWTServiceFromConfig creates a new JWT service instance from configuration
func NewJWTServiceFromConfig(authConfig *config.AuthConfig) (JWTService, error) {
	if authConfig == nil {
		return nil, fmt.Errorf("auth configuration cannot be nil")
	}
	
	if authConfig.JWTSecret == "" {
		return nil, fmt.Errorf("JWT secret cannot be empty")
	}
	
	if len(authConfig.JWTSecret) < 32 {
		return nil, fmt.Errorf("JWT secret must be at least 32 characters long")
	}

	return &jwtService{
		secret:            []byte(authConfig.JWTSecret),
		accessTokenTTL:    authConfig.AccessTokenTTL,
		refreshTokenTTL:   authConfig.RefreshTokenTTL,
	}, nil
}

// GenerateTokenPair creates both access and refresh tokens for a user
func (js *jwtService) GenerateTokenPair(userID, email string) (*domain.TokenPair, error) {
	if userID == "" {
		return nil, fmt.Errorf("user ID cannot be empty")
	}
	if email == "" {
		return nil, fmt.Errorf("email cannot be empty")
	}

	now := time.Now()

	// Create access token (15 minutes)
	accessClaims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"exp":     now.Add(js.accessTokenTTL).Unix(),
		"iat":     now.Unix(),
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString(js.secret)
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	// Create refresh token (7 days)
	refreshClaims := jwt.MapClaims{
		"user_id": userID,
		"email":   email,
		"exp":     now.Add(js.refreshTokenTTL).Unix(),
		"iat":     now.Unix(),
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString(js.secret)
	if err != nil {
		return nil, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return &domain.TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		ExpiresIn:    int64(js.accessTokenTTL.Seconds()), // 900 seconds = 15 minutes
	}, nil
}

// ValidateAccessToken validates an access token and returns its claims
func (js *jwtService) ValidateAccessToken(tokenString string) (*domain.TokenClaims, error) {
	return js.validateToken(tokenString, "access token")
}

// ValidateRefreshToken validates a refresh token and returns its claims
func (js *jwtService) ValidateRefreshToken(tokenString string) (*domain.TokenClaims, error) {
	return js.validateToken(tokenString, "refresh token")
}

// validateToken validates a token string and returns its claims
func (js *jwtService) validateToken(tokenString, tokenType string) (*domain.TokenClaims, error) {
	if tokenString == "" {
		return nil, fmt.Errorf("%s cannot be empty", tokenType)
	}

	// Parse the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Verify the signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return js.secret, nil
	})

	if err != nil {
		// Check for specific error types using errors.Is (v5 approach)
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, fmt.Errorf("%s is expired", tokenType)
		}
		if errors.Is(err, jwt.ErrTokenSignatureInvalid) {
			return nil, fmt.Errorf("%s signature is invalid", tokenType)
		}
		return nil, fmt.Errorf("failed to parse %s: %w", tokenType, err)
	}

	// Check if token is valid
	if !token.Valid {
		return nil, fmt.Errorf("%s is invalid", tokenType)
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("failed to extract claims from %s", tokenType)
	}

	// Extract user ID
	userID, ok := claims["user_id"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid user_id in %s claims", tokenType)
	}

	// Extract email
	email, ok := claims["email"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid email in %s claims", tokenType)
	}

	// Extract expiry time
	exp, ok := claims["exp"].(float64)
	if !ok {
		return nil, fmt.Errorf("invalid exp in %s claims", tokenType)
	}

	return &domain.TokenClaims{
		UserID:    userID,
		Email:     email,
		ExpiresAt: int64(exp),
	}, nil
}