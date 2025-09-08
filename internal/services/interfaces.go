package services

import (
	"context"
	"time"

	"github.com/DuckDHD/BuyOrBye/internal/domain"
)

// UserRepository defines the interface for user persistence operations
// This interface is defined in the services layer following the dependency inversion principle
// The repository layer implements this interface using GORM
type UserRepository interface {
	// Create saves a new user to the database
	// Returns an error if the user already exists or if there's a database error
	Create(ctx context.Context, user *domain.User) error

	// GetByEmail retrieves a user by their email address
	// Returns domain.ErrUserNotFound if the user doesn't exist
	GetByEmail(ctx context.Context, email string) (*domain.User, error)

	// GetByID retrieves a user by their ID
	// Returns domain.ErrUserNotFound if the user doesn't exist
	GetByID(ctx context.Context, userID string) (*domain.User, error)

	// Update modifies an existing user's data
	// Returns domain.ErrUserNotFound if the user doesn't exist
	Update(ctx context.Context, user *domain.User) error

	// UpdateLastLogin updates the last login timestamp for a user
	// Returns domain.ErrUserNotFound if the user doesn't exist
	UpdateLastLogin(ctx context.Context, userID string, loginTime time.Time) error
}

// TokenRepository defines the interface for refresh token persistence operations
// This interface is defined in the services layer following the dependency inversion principle
// The repository layer implements this interface using GORM
type TokenRepository interface {
	// SaveRefreshToken stores a refresh token for a user
	// If a token already exists for the user, it should be replaced
	SaveRefreshToken(ctx context.Context, userID, token string, expiresAt time.Time) error

	// GetRefreshToken retrieves a refresh token by the token string
	// Returns domain.ErrTokenNotFound if the token doesn't exist
	// Returns domain.ErrTokenExpired if the token has expired
	// Returns domain.ErrTokenRevoked if the token has been revoked
	GetRefreshToken(ctx context.Context, token string) (userID string, err error)

	// RevokeToken marks a refresh token as revoked
	// Returns domain.ErrTokenNotFound if the token doesn't exist
	RevokeToken(ctx context.Context, token string) error

	// RevokeAllUserTokens marks all refresh tokens for a user as revoked
	// This is useful for logout from all devices functionality
	RevokeAllUserTokens(ctx context.Context, userID string) error

	// CleanupExpiredTokens removes expired tokens from the database
	// This method should be called periodically for maintenance
	CleanupExpiredTokens(ctx context.Context) error
}