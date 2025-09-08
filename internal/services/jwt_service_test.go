package services

import (
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJWTService_GenerateTokenPair_ReturnsValidTokens(t *testing.T) {
	// Arrange
	os.Setenv("JWT_SECRET", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET")
	
	service, err := NewJWTService()
	require.NoError(t, err)
	
	userID := "user-123"
	email := "test@example.com"

	// Act
	tokenPair, err := service.GenerateTokenPair(userID, email)

	// Assert
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenPair.AccessToken)
	assert.NotEmpty(t, tokenPair.RefreshToken)
	assert.Equal(t, int64(900), tokenPair.ExpiresIn) // 15 minutes = 900 seconds

	// Verify tokens are valid JWT format
	_, err = jwt.Parse(tokenPair.AccessToken, func(token *jwt.Token) (interface{}, error) {
		return []byte("test-secret-key"), nil
	})
	assert.NoError(t, err)

	_, err = jwt.Parse(tokenPair.RefreshToken, func(token *jwt.Token) (interface{}, error) {
		return []byte("test-secret-key"), nil
	})
	assert.NoError(t, err)
}

func TestJWTService_GenerateTokenPair_CorrectTTL(t *testing.T) {
	// Arrange
	os.Setenv("JWT_SECRET", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET")
	
	service, err := NewJWTService()
	require.NoError(t, err)
	
	userID := "user-123"
	email := "test@example.com"
	startTime := time.Now()

	// Act
	tokenPair, err := service.GenerateTokenPair(userID, email)
	require.NoError(t, err)

	// Assert access token expiry (15 minutes)
	accessClaims, err := service.ValidateAccessToken(tokenPair.AccessToken)
	require.NoError(t, err)
	
	expectedAccessExpiry := startTime.Add(15 * time.Minute)
	actualAccessExpiry := time.Unix(accessClaims.ExpiresAt, 0)
	assert.WithinDuration(t, expectedAccessExpiry, actualAccessExpiry, time.Second*10)

	// Assert refresh token expiry (7 days)
	refreshClaims, err := service.ValidateRefreshToken(tokenPair.RefreshToken)
	require.NoError(t, err)
	
	expectedRefreshExpiry := startTime.Add(7 * 24 * time.Hour)
	actualRefreshExpiry := time.Unix(refreshClaims.ExpiresAt, 0)
	assert.WithinDuration(t, expectedRefreshExpiry, actualRefreshExpiry, time.Second*10)
}

func TestJWTService_ValidateAccessToken_ValidToken_ReturnsClaims(t *testing.T) {
	// Arrange
	os.Setenv("JWT_SECRET", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET")
	
	service, err := NewJWTService()
	require.NoError(t, err)
	
	userID := "user-123"
	email := "test@example.com"
	
	tokenPair, err := service.GenerateTokenPair(userID, email)
	require.NoError(t, err)

	// Act
	claims, err := service.ValidateAccessToken(tokenPair.AccessToken)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, email, claims.Email)
	assert.True(t, claims.ExpiresAt > time.Now().Unix())
}

func TestJWTService_ValidateAccessToken_ExpiredToken_ReturnsError(t *testing.T) {
	// Arrange
	os.Setenv("JWT_SECRET", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET")
	
	service, err := NewJWTService()
	require.NoError(t, err)

	// Create an already expired token by mocking time
	originalNow := time.Now()
	pastTime := originalNow.Add(-1 * time.Hour)
	
	// Create token with past expiry
	claims := jwt.MapClaims{
		"user_id": "user-123",
		"email":   "test@example.com",
		"exp":     pastTime.Unix(),
		"iat":     pastTime.Add(-15 * time.Minute).Unix(),
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	expiredToken, err := token.SignedString([]byte("test-secret-key"))
	require.NoError(t, err)

	// Act
	claims2, err := service.ValidateAccessToken(expiredToken)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, claims2)
	assert.Contains(t, err.Error(), "token is expired")
}

func TestJWTService_ValidateAccessToken_InvalidSignature_ReturnsError(t *testing.T) {
	// Arrange
	os.Setenv("JWT_SECRET", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET")
	
	service, err := NewJWTService()
	require.NoError(t, err)

	// Create token with wrong secret
	claims := jwt.MapClaims{
		"user_id": "user-123",
		"email":   "test@example.com",
		"exp":     time.Now().Add(15 * time.Minute).Unix(),
		"iat":     time.Now().Unix(),
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	invalidToken, err := token.SignedString([]byte("wrong-secret-key"))
	require.NoError(t, err)

	// Act
	claims2, err := service.ValidateAccessToken(invalidToken)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, claims2)
	assert.Contains(t, err.Error(), "signature is invalid")
}

func TestJWTService_ValidateRefreshToken_ValidToken_ReturnsClaims(t *testing.T) {
	// Arrange
	os.Setenv("JWT_SECRET", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET")
	
	service, err := NewJWTService()
	require.NoError(t, err)
	
	userID := "user-123"
	email := "test@example.com"
	
	tokenPair, err := service.GenerateTokenPair(userID, email)
	require.NoError(t, err)

	// Act
	claims, err := service.ValidateRefreshToken(tokenPair.RefreshToken)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, userID, claims.UserID)
	assert.Equal(t, email, claims.Email)
	assert.True(t, claims.ExpiresAt > time.Now().Unix())
}

func TestJWTService_ValidateRefreshToken_ExpiredToken_ReturnsError(t *testing.T) {
	// Arrange
	os.Setenv("JWT_SECRET", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET")
	
	service, err := NewJWTService()
	require.NoError(t, err)

	// Create an expired refresh token
	pastTime := time.Now().Add(-8 * 24 * time.Hour) // 8 days ago
	
	claims := jwt.MapClaims{
		"user_id": "user-123",
		"email":   "test@example.com",
		"exp":     pastTime.Unix(),
		"iat":     pastTime.Add(-7 * 24 * time.Hour).Unix(),
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	expiredToken, err := token.SignedString([]byte("test-secret-key"))
	require.NoError(t, err)

	// Act
	claims2, err := service.ValidateRefreshToken(expiredToken)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, claims2)
	assert.Contains(t, err.Error(), "token is expired")
}

func TestJWTService_ValidateTokens_InvalidFormat_ReturnsError(t *testing.T) {
	// Arrange
	os.Setenv("JWT_SECRET", "test-secret-key")
	defer os.Unsetenv("JWT_SECRET")
	
	service, err := NewJWTService()
	require.NoError(t, err)

	invalidTokens := []string{
		"invalid.token.format",
		"",
		"not.a.jwt",
		"header.payload", // Missing signature
		"too.many.parts.here.invalid",
	}

	for _, invalidToken := range invalidTokens {
		t.Run("invalid_token_"+invalidToken, func(t *testing.T) {
			// Act & Assert access token
			claims, err := service.ValidateAccessToken(invalidToken)
			assert.Error(t, err)
			assert.Nil(t, claims)

			// Act & Assert refresh token
			claims, err = service.ValidateRefreshToken(invalidToken)
			assert.Error(t, err)
			assert.Nil(t, claims)
		})
	}
}

func TestJWTService_NewJWTService_MissingSecret_ReturnsError(t *testing.T) {
	// Arrange
	os.Unsetenv("JWT_SECRET")

	// Act
	service, err := NewJWTService()

	// Assert
	assert.Error(t, err)
	assert.Nil(t, service)
	assert.Contains(t, err.Error(), "JWT_SECRET environment variable is required")
}

func TestJWTService_NewJWTService_EmptySecret_ReturnsError(t *testing.T) {
	// Arrange
	os.Setenv("JWT_SECRET", "")
	defer os.Unsetenv("JWT_SECRET")

	// Act
	service, err := NewJWTService()

	// Assert
	assert.Error(t, err)
	assert.Nil(t, service)
	assert.Contains(t, err.Error(), "JWT_SECRET environment variable is required")
}