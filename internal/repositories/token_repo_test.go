package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/DuckDHD/BuyOrBye/internal/domain"
	"github.com/DuckDHD/BuyOrBye/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// createTestUserInDB creates a test user in the database and returns the user
func createTestUserInDB(t *testing.T, db *gorm.DB) *domain.User {
	t.Helper()
	
	repo := NewUserRepository(db)
	ctx := context.Background()
	
	user := createTestUser()
	err := repo.Create(ctx, user)
	require.NoError(t, err)
	
	return user
}

func TestTokenRepository_SaveRefreshToken_Success(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	repo := NewTokenRepository(db)
	ctx := context.Background()
	
	user := createTestUserInDB(t, db)
	token := "refresh_token_123"
	expiresAt := time.Now().Add(24 * time.Hour)

	// Act
	err := repo.SaveRefreshToken(ctx, user.ID, token, expiresAt)

	// Assert
	assert.NoError(t, err)

	// Verify in database
	var model models.RefreshTokenModel
	err = db.Where("token = ?", token).First(&model).Error
	assert.NoError(t, err)
	assert.Equal(t, token, model.Token)
	assert.Equal(t, user.ID, model.ToUserID())
	assert.Equal(t, expiresAt.Unix(), model.ExpiresAt.Unix())
	assert.False(t, model.IsRevoked)
}

func TestTokenRepository_SaveRefreshToken_ReplaceExisting_Success(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	repo := NewTokenRepository(db)
	ctx := context.Background()
	
	user := createTestUserInDB(t, db)
	oldToken := "old_refresh_token"
	newToken := "new_refresh_token"
	expiresAt := time.Now().Add(24 * time.Hour)

	// Save first token
	err := repo.SaveRefreshToken(ctx, user.ID, oldToken, expiresAt)
	require.NoError(t, err)

	// Act - Save new token (should replace the old one)
	err = repo.SaveRefreshToken(ctx, user.ID, newToken, expiresAt)

	// Assert
	assert.NoError(t, err)

	// Verify only the new token exists for this user
	var count int64
	db.Model(&models.RefreshTokenModel{}).Where("user_id = ? AND is_revoked = false", user.ID).Count(&count)
	assert.Equal(t, int64(1), count)

	// Verify the new token is in database
	var model models.RefreshTokenModel
	err = db.Where("token = ?", newToken).First(&model).Error
	assert.NoError(t, err)
	assert.Equal(t, newToken, model.Token)
}

func TestTokenRepository_SaveRefreshToken_InvalidUserID_ReturnsError(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	repo := NewTokenRepository(db)
	ctx := context.Background()
	
	token := "refresh_token_123"
	expiresAt := time.Now().Add(24 * time.Hour)

	// Act
	err := repo.SaveRefreshToken(ctx, "999999", token, expiresAt)

	// Assert
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrUserNotFound)
}

func TestTokenRepository_GetRefreshToken_Success(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	repo := NewTokenRepository(db)
	ctx := context.Background()
	
	user := createTestUserInDB(t, db)
	token := "refresh_token_123"
	expiresAt := time.Now().Add(24 * time.Hour)

	err := repo.SaveRefreshToken(ctx, user.ID, token, expiresAt)
	require.NoError(t, err)

	// Act
	userID, err := repo.GetRefreshToken(ctx, token)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, user.ID, userID)
}

func TestTokenRepository_GetRefreshToken_NotFound_ReturnsError(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	repo := NewTokenRepository(db)
	ctx := context.Background()

	// Act
	userID, err := repo.GetRefreshToken(ctx, "nonexistent_token")

	// Assert
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrTokenNotFound)
	assert.Empty(t, userID)
}

func TestTokenRepository_GetRefreshToken_ExpiredToken_ReturnsError(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	repo := NewTokenRepository(db)
	ctx := context.Background()
	
	user := createTestUserInDB(t, db)
	token := "expired_token"
	expiredTime := time.Now().Add(-1 * time.Hour) // Expired 1 hour ago

	err := repo.SaveRefreshToken(ctx, user.ID, token, expiredTime)
	require.NoError(t, err)

	// Act
	userID, err := repo.GetRefreshToken(ctx, token)

	// Assert
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrTokenExpired)
	assert.Empty(t, userID)
}

func TestTokenRepository_GetRefreshToken_RevokedToken_ReturnsError(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	repo := NewTokenRepository(db)
	ctx := context.Background()
	
	user := createTestUserInDB(t, db)
	token := "revoked_token"
	expiresAt := time.Now().Add(24 * time.Hour)

	err := repo.SaveRefreshToken(ctx, user.ID, token, expiresAt)
	require.NoError(t, err)

	// Revoke the token
	err = repo.RevokeToken(ctx, token)
	require.NoError(t, err)

	// Act
	userID, err := repo.GetRefreshToken(ctx, token)

	// Assert
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrTokenRevoked)
	assert.Empty(t, userID)
}

func TestTokenRepository_RevokeToken_Success(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	repo := NewTokenRepository(db)
	ctx := context.Background()
	
	user := createTestUserInDB(t, db)
	token := "token_to_revoke"
	expiresAt := time.Now().Add(24 * time.Hour)

	err := repo.SaveRefreshToken(ctx, user.ID, token, expiresAt)
	require.NoError(t, err)

	// Act
	err = repo.RevokeToken(ctx, token)

	// Assert
	assert.NoError(t, err)

	// Verify token is revoked in database
	var model models.RefreshTokenModel
	err = db.Where("token = ?", token).First(&model).Error
	require.NoError(t, err)
	assert.True(t, model.IsRevoked)
	assert.NotNil(t, model.RevokedAt)
}

func TestTokenRepository_RevokeToken_NotFound_ReturnsError(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	repo := NewTokenRepository(db)
	ctx := context.Background()

	// Act
	err := repo.RevokeToken(ctx, "nonexistent_token")

	// Assert
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrTokenNotFound)
}

func TestTokenRepository_RevokeAllUserTokens_Success(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	repo := NewTokenRepository(db)
	ctx := context.Background()
	
	user := createTestUserInDB(t, db)
	expiresAt := time.Now().Add(24 * time.Hour)

	// Create multiple tokens for the user
	tokens := []string{"token1", "token2", "token3"}
	for _, token := range tokens {
		err := repo.SaveRefreshToken(ctx, user.ID, token, expiresAt)
		require.NoError(t, err)
	}

	// Act
	err := repo.RevokeAllUserTokens(ctx, user.ID)

	// Assert
	assert.NoError(t, err)

	// Verify all tokens are revoked
	var count int64
	db.Model(&models.RefreshTokenModel{}).
		Where("user_id = ? AND is_revoked = false", user.ID).
		Count(&count)
	assert.Equal(t, int64(0), count)

	// Verify all tokens are marked as revoked
	var revokedCount int64
	db.Model(&models.RefreshTokenModel{}).
		Where("user_id = ? AND is_revoked = true", user.ID).
		Count(&revokedCount)
	assert.Equal(t, int64(3), revokedCount)
}

func TestTokenRepository_RevokeAllUserTokens_UserNotFound_Success(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	repo := NewTokenRepository(db)
	ctx := context.Background()

	// Act - Should not error even if user doesn't exist
	err := repo.RevokeAllUserTokens(ctx, "999999")

	// Assert
	assert.NoError(t, err)
}

func TestTokenRepository_CleanupExpiredTokens_Success(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	repo := NewTokenRepository(db)
	ctx := context.Background()
	
	user1 := createTestUserInDB(t, db)
	user2 := createTestUser()
	user2.Email = "user2@example.com"
	userRepo := NewUserRepository(db)
	err := userRepo.Create(ctx, user2)
	require.NoError(t, err)

	// Create expired tokens for different users (they won't replace each other)
	expiredTime := time.Now().Add(-1 * time.Hour)
	err = repo.SaveRefreshToken(ctx, user1.ID, "expired_token_1", expiredTime)
	require.NoError(t, err)
	err = repo.SaveRefreshToken(ctx, user2.ID, "expired_token_2", expiredTime)
	require.NoError(t, err)

	// Create valid tokens for different users
	validTime := time.Now().Add(24 * time.Hour)
	err = repo.SaveRefreshToken(ctx, user1.ID, "valid_token_1", validTime)
	require.NoError(t, err)
	err = repo.SaveRefreshToken(ctx, user2.ID, "valid_token_2", validTime)
	require.NoError(t, err)

	// Count initial tokens (should be 2 valid tokens, expired ones are revoked but not deleted)
	var initialCount int64
	db.Model(&models.RefreshTokenModel{}).Count(&initialCount)
	assert.Equal(t, int64(4), initialCount) // 2 expired (revoked) + 2 valid

	// Act
	err = repo.CleanupExpiredTokens(ctx)

	// Assert
	assert.NoError(t, err)

	// Verify only non-expired tokens remain
	var remainingCount int64
	db.Model(&models.RefreshTokenModel{}).Count(&remainingCount)
	assert.Equal(t, int64(2), remainingCount) // Should only have the 2 valid tokens
}

func TestTokenRepository_CleanupExpiredTokens_NoExpiredTokens_Success(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	repo := NewTokenRepository(db)
	ctx := context.Background()
	
	user := createTestUserInDB(t, db)

	// Create only valid tokens
	validTime := time.Now().Add(24 * time.Hour)
	err := repo.SaveRefreshToken(ctx, user.ID, "valid_token", validTime)
	require.NoError(t, err)

	// Act
	err = repo.CleanupExpiredTokens(ctx)

	// Assert
	assert.NoError(t, err)

	// Verify token is still there
	var count int64
	db.Model(&models.RefreshTokenModel{}).Count(&count)
	assert.Equal(t, int64(1), count)
}

func TestTokenRepository_SaveRefreshToken_EmptyUserID_ReturnsError(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	repo := NewTokenRepository(db)
	ctx := context.Background()
	
	token := "refresh_token_123"
	expiresAt := time.Now().Add(24 * time.Hour)

	// Act
	err := repo.SaveRefreshToken(ctx, "", token, expiresAt)

	// Assert
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvalidUserData)
}

func TestTokenRepository_SaveRefreshToken_EmptyToken_ReturnsError(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	repo := NewTokenRepository(db)
	ctx := context.Background()
	
	user := createTestUserInDB(t, db)
	expiresAt := time.Now().Add(24 * time.Hour)

	// Act
	err := repo.SaveRefreshToken(ctx, user.ID, "", expiresAt)

	// Assert
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvalidUserData)
}

func TestTokenRepository_GetRefreshToken_EmptyToken_ReturnsError(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	repo := NewTokenRepository(db)
	ctx := context.Background()

	// Act
	userID, err := repo.GetRefreshToken(ctx, "")

	// Assert
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvalidToken)
	assert.Empty(t, userID)
}

func TestTokenRepository_RevokeToken_EmptyToken_ReturnsError(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	repo := NewTokenRepository(db)
	ctx := context.Background()

	// Act
	err := repo.RevokeToken(ctx, "")

	// Assert
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvalidToken)
}

func TestTokenRepository_RevokeAllUserTokens_EmptyUserID_ReturnsError(t *testing.T) {
	// Arrange
	db := setupTestDB(t)
	repo := NewTokenRepository(db)
	ctx := context.Background()

	// Act
	err := repo.RevokeAllUserTokens(ctx, "")

	// Assert
	assert.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrInvalidUserData)
}