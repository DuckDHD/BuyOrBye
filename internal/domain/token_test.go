package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestTokenClaims_IsExpired_NotExpired_ReturnsFalse(t *testing.T) {
	// Arrange
	futureTime := time.Now().Add(1 * time.Hour)
	claims := TokenClaims{
		UserID:    "user-123",
		Email:     "test@example.com",
		ExpiresAt: futureTime.Unix(),
	}

	// Act
	isExpired := claims.IsExpired()

	// Assert
	assert.False(t, isExpired)
}

func TestTokenClaims_IsExpired_Expired_ReturnsTrue(t *testing.T) {
	// Arrange
	pastTime := time.Now().Add(-1 * time.Hour)
	claims := TokenClaims{
		UserID:    "user-123",
		Email:     "test@example.com",
		ExpiresAt: pastTime.Unix(),
	}

	// Act
	isExpired := claims.IsExpired()

	// Assert
	assert.True(t, isExpired)
}

func TestTokenClaims_IsExpired_ExactlyNow_ReturnsTrue(t *testing.T) {
	// Arrange
	now := time.Now()
	claims := TokenClaims{
		UserID:    "user-123",
		Email:     "test@example.com",
		ExpiresAt: now.Unix(),
	}

	// Act
	time.Sleep(time.Millisecond) // Ensure we're past the exact time
	isExpired := claims.IsExpired()

	// Assert
	assert.True(t, isExpired)
}

func TestTokenClaims_IsExpired_TableDriven(t *testing.T) {
	tests := []struct {
		name           string
		timeOffset     time.Duration
		expectedResult bool
	}{
		{
			name:           "Far future",
			timeOffset:     24 * time.Hour,
			expectedResult: false,
		},
		{
			name:           "One hour future",
			timeOffset:     1 * time.Hour,
			expectedResult: false,
		},
		{
			name:           "One minute future",
			timeOffset:     1 * time.Minute,
			expectedResult: false,
		},
		{
			name:           "One second future",
			timeOffset:     1 * time.Second,
			expectedResult: false,
		},
		{
			name:           "One second past",
			timeOffset:     -1 * time.Second,
			expectedResult: true,
		},
		{
			name:           "One minute past",
			timeOffset:     -1 * time.Minute,
			expectedResult: true,
		},
		{
			name:           "One hour past",
			timeOffset:     -1 * time.Hour,
			expectedResult: true,
		},
		{
			name:           "Far past",
			timeOffset:     -24 * time.Hour,
			expectedResult: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			expiryTime := time.Now().Add(tt.timeOffset)
			claims := TokenClaims{
				UserID:    "user-123",
				Email:     "test@example.com",
				ExpiresAt: expiryTime.Unix(),
			}

			// Act
			isExpired := claims.IsExpired()

			// Assert
			assert.Equal(t, tt.expectedResult, isExpired)
		})
	}
}

func TestTokenPair_Validate_AllFieldsPresent_ReturnsNil(t *testing.T) {
	// Arrange
	tokenPair := TokenPair{
		AccessToken:  "eyJhbGciOiJIUzI1NiJ9.eyJ1c2VyX2lkIjoidXNlci0xMjMifQ.signature",
		RefreshToken: "eyJhbGciOiJIUzI1NiJ9.eyJ1c2VyX2lkIjoidXNlci0xMjMifQ.signature",
		ExpiresIn:    900, // 15 minutes
	}

	// Act
	err := tokenPair.Validate()

	// Assert
	assert.NoError(t, err)
}

func TestTokenPair_Validate_EmptyAccessToken_ReturnsError(t *testing.T) {
	// Arrange
	tokenPair := TokenPair{
		AccessToken:  "",
		RefreshToken: "eyJhbGciOiJIUzI1NiJ9.eyJ1c2VyX2lkIjoidXNlci0xMjMifQ.signature",
		ExpiresIn:    900,
	}

	// Act
	err := tokenPair.Validate()

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "access token is required")
}

func TestTokenPair_Validate_EmptyRefreshToken_ReturnsError(t *testing.T) {
	// Arrange
	tokenPair := TokenPair{
		AccessToken:  "eyJhbGciOiJIUzI1NiJ9.eyJ1c2VyX2lkIjoidXNlci0xMjMifQ.signature",
		RefreshToken: "",
		ExpiresIn:    900,
	}

	// Act
	err := tokenPair.Validate()

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "refresh token is required")
}

func TestTokenPair_Validate_ZeroExpiresIn_ReturnsError(t *testing.T) {
	// Arrange
	tokenPair := TokenPair{
		AccessToken:  "eyJhbGciOiJIUzI1NiJ9.eyJ1c2VyX2lkIjoidXNlci0xMjMifQ.signature",
		RefreshToken: "eyJhbGciOiJIUzI1NiJ9.eyJ1c2VyX2lkIjoidXNlci0xMjMifQ.signature",
		ExpiresIn:    0,
	}

	// Act
	err := tokenPair.Validate()

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expires in must be greater than 0")
}

func TestTokenPair_Validate_NegativeExpiresIn_ReturnsError(t *testing.T) {
	// Arrange
	tokenPair := TokenPair{
		AccessToken:  "eyJhbGciOiJIUzI1NiJ9.eyJ1c2VyX2lkIjoidXNlci0xMjMifQ.signature",
		RefreshToken: "eyJhbGciOiJIUzI1NiJ9.eyJ1c2VyX2lkIjoidXNlci0xMjMifQ.signature",
		ExpiresIn:    -100,
	}

	// Act
	err := tokenPair.Validate()

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expires in must be greater than 0")
}

func TestTokenPair_Validate_MultipleErrors_ReturnsAllErrors(t *testing.T) {
	// Arrange
	tokenPair := TokenPair{
		AccessToken:  "",  // Missing
		RefreshToken: "",  // Missing
		ExpiresIn:    -1,  // Invalid
	}

	// Act
	err := tokenPair.Validate()

	// Assert
	assert.Error(t, err)
	errorMsg := err.Error()
	assert.Contains(t, errorMsg, "access token is required")
	assert.Contains(t, errorMsg, "refresh token is required")
	assert.Contains(t, errorMsg, "expires in must be greater than 0")
}