package handlers

import (
	"context"

	"github.com/DuckDHD/BuyOrBye/internal/domain"
)

// AuthService interface is defined in handlers package following consumer-defined principle
// This interface is consumed by AuthHandler in this package
type AuthService interface {
	// Login authenticates a user with email and password
	// Returns a token pair on successful authentication
	// Returns domain.ErrInvalidCredentials if credentials are invalid
	// Returns domain.ErrAccountInactive if account is inactive
	Login(ctx context.Context, credentials domain.Credentials) (*domain.TokenPair, error)

	// Register creates a new user account and returns authentication tokens
	// Returns domain.ErrUserAlreadyExists if user already exists
	// Returns domain.ErrInvalidUserData if user data validation fails
	Register(ctx context.Context, user *domain.User, password string) (*domain.TokenPair, error)

	// RefreshToken generates a new token pair using a valid refresh token
	// Returns domain.ErrInvalidToken if token is invalid
	// Returns domain.ErrTokenExpired if token has expired
	// Returns domain.ErrTokenRevoked if token has been revoked
	RefreshToken(ctx context.Context, refreshToken string) (*domain.TokenPair, error)

	// Logout revokes a user's refresh token
	// Returns domain.ErrInvalidToken if token is invalid
	Logout(ctx context.Context, refreshToken string) error
}