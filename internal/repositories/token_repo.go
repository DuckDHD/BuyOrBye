package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/DuckDHD/BuyOrBye/internal/domain"
	"github.com/DuckDHD/BuyOrBye/internal/models"
	"github.com/DuckDHD/BuyOrBye/internal/services"
	"gorm.io/gorm"
)

// tokenRepository implements the TokenRepository interface using GORM
type tokenRepository struct {
	db *gorm.DB
}

// NewTokenRepository creates a new instance of TokenRepository
func NewTokenRepository(db *gorm.DB) services.TokenRepository {
	return &tokenRepository{
		db: db,
	}
}

// SaveRefreshToken stores a refresh token for a user
// If a token already exists for the user, it should be replaced
func (r *tokenRepository) SaveRefreshToken(ctx context.Context, userID, token string, expiresAt time.Time) error {
	if userID == "" {
		return fmt.Errorf("userID cannot be empty: %w", domain.ErrInvalidUserData)
	}
	if token == "" {
		return fmt.Errorf("token cannot be empty: %w", domain.ErrInvalidUserData)
	}

	// Verify user exists
	var userModel models.UserModel
	if err := r.db.WithContext(ctx).Where("id = ?", userID).First(&userModel).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("user with ID %s not found: %w", userID, domain.ErrUserNotFound)
		}
		return fmt.Errorf("failed to verify user existence: %w", err)
	}

	// Start transaction to ensure atomicity
	tx := r.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("failed to start transaction: %w", tx.Error)
	}
	defer tx.Rollback()

	// Revoke all existing tokens for this user
	if err := tx.Model(&models.RefreshTokenModel{}).
		Where("user_id = ? AND is_revoked = false", userID).
		Updates(map[string]interface{}{
			"is_revoked": true,
			"revoked_at": time.Now(),
		}).Error; err != nil {
		return fmt.Errorf("failed to revoke existing tokens: %w", err)
	}

	// Create new token
	tokenModel := models.RefreshTokenFromDomain(userID, token, expiresAt)
	if err := tx.Create(&tokenModel).Error; err != nil {
		return fmt.Errorf("failed to create refresh token: %w", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetRefreshToken retrieves a refresh token by the token string
// Returns domain.ErrTokenNotFound if the token doesn't exist
// Returns domain.ErrTokenExpired if the token has expired
// Returns domain.ErrTokenRevoked if the token has been revoked
func (r *tokenRepository) GetRefreshToken(ctx context.Context, token string) (userID string, err error) {
	if token == "" {
		return "", fmt.Errorf("token cannot be empty: %w", domain.ErrInvalidToken)
	}

	var tokenModel models.RefreshTokenModel
	if err := r.db.WithContext(ctx).Where("token = ?", token).First(&tokenModel).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", fmt.Errorf("token not found: %w", domain.ErrTokenNotFound)
		}
		return "", fmt.Errorf("failed to get refresh token: %w", err)
	}

	// Check if token is revoked
	if tokenModel.IsRevoked {
		return "", fmt.Errorf("token has been revoked: %w", domain.ErrTokenRevoked)
	}

	// Check if token is expired
	if tokenModel.IsExpired() {
		return "", fmt.Errorf("token has expired: %w", domain.ErrTokenExpired)
	}

	return tokenModel.ToUserID(), nil
}

// RevokeToken marks a refresh token as revoked
// Returns domain.ErrTokenNotFound if the token doesn't exist
func (r *tokenRepository) RevokeToken(ctx context.Context, token string) error {
	if token == "" {
		return fmt.Errorf("token cannot be empty: %w", domain.ErrInvalidToken)
	}

	now := time.Now()
	result := r.db.WithContext(ctx).Model(&models.RefreshTokenModel{}).
		Where("token = ?", token).
		Updates(map[string]interface{}{
			"is_revoked": true,
			"revoked_at": now,
		})

	if result.Error != nil {
		return fmt.Errorf("failed to revoke token: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("token not found: %w", domain.ErrTokenNotFound)
	}

	return nil
}

// RevokeAllUserTokens marks all refresh tokens for a user as revoked
// This is useful for logout from all devices functionality
func (r *tokenRepository) RevokeAllUserTokens(ctx context.Context, userID string) error {
	if userID == "" {
		return fmt.Errorf("userID cannot be empty: %w", domain.ErrInvalidUserData)
	}

	now := time.Now()
	result := r.db.WithContext(ctx).Model(&models.RefreshTokenModel{}).
		Where("user_id = ? AND is_revoked = false", userID).
		Updates(map[string]interface{}{
			"is_revoked": true,
			"revoked_at": now,
		})

	if result.Error != nil {
		return fmt.Errorf("failed to revoke user tokens: %w", result.Error)
	}

	// Note: We don't return an error if no rows were affected
	// because it's valid to call this method even if the user has no active tokens

	return nil
}

// CleanupExpiredTokens removes expired tokens from the database
// This method should be called periodically for maintenance
func (r *tokenRepository) CleanupExpiredTokens(ctx context.Context) error {
	result := r.db.WithContext(ctx).
		Where("expires_at < ?", time.Now()).
		Delete(&models.RefreshTokenModel{})

	if result.Error != nil {
		return fmt.Errorf("failed to cleanup expired tokens: %w", result.Error)
	}

	return nil
}