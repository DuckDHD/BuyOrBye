package services

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestPasswordService_HashPassword_ValidPassword_ReturnsBcryptHash(t *testing.T) {
	// Arrange
	service := NewPasswordService()
	password := "ValidPassword123!"

	// Act
	hash, err := service.HashPassword(password)

	// Assert
	assert.NoError(t, err)
	assert.NotEmpty(t, hash)
	assert.NotEqual(t, password, hash)
	
	// Verify it's a valid bcrypt hash
	cost, err := bcrypt.Cost([]byte(hash))
	assert.NoError(t, err)
	assert.Equal(t, 14, cost, "Expected bcrypt cost of 14")
}

func TestPasswordService_HashPassword_EmptyPassword_ReturnsError(t *testing.T) {
	// Arrange
	service := NewPasswordService()
	password := ""

	// Act
	hash, err := service.HashPassword(password)

	// Assert
	assert.Error(t, err)
	assert.Empty(t, hash)
	assert.Contains(t, err.Error(), "password cannot be empty")
}

func TestPasswordService_HashPassword_TableDriven(t *testing.T) {
	tests := []struct {
		name        string
		password    string
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Valid password",
			password:    "ValidPassword123!",
			expectError: false,
		},
		{
			name:        "Empty password",
			password:    "",
			expectError: true,
			errorMsg:    "password cannot be empty",
		},
		{
			name:        "Very long password",
			password:    string(make([]byte, 100)),
			expectError: false,
		},
		{
			name:        "Password with special characters",
			password:    "P@ssw0rd!@#$%^&*()",
			expectError: false,
		},
	}

	service := NewPasswordService()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			hash, err := service.HashPassword(tt.password)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Empty(t, hash)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, hash)
				assert.NotEqual(t, tt.password, hash)
			}
		})
	}
}

func TestPasswordService_CheckPassword_CorrectPassword_ReturnsNil(t *testing.T) {
	// Arrange
	service := NewPasswordService()
	password := "CorrectPassword123!"
	hash, err := service.HashPassword(password)
	require.NoError(t, err)

	// Act
	err = service.CheckPassword(hash, password)

	// Assert
	assert.NoError(t, err)
}

func TestPasswordService_CheckPassword_WrongPassword_ReturnsError(t *testing.T) {
	// Arrange
	service := NewPasswordService()
	correctPassword := "CorrectPassword123!"
	wrongPassword := "WrongPassword123!"
	hash, err := service.HashPassword(correctPassword)
	require.NoError(t, err)

	// Act
	err = service.CheckPassword(hash, wrongPassword)

	// Assert
	assert.Error(t, err)
	assert.True(t, errors.Is(err, bcrypt.ErrMismatchedHashAndPassword))
}

func TestPasswordService_CheckPassword_PreventTimingAttacks(t *testing.T) {
	// Arrange
	service := NewPasswordService()
	password := "TestPassword123!"
	hash, err := service.HashPassword(password)
	require.NoError(t, err)

	// Test with correct password multiple times
	correctTimes := make([]time.Duration, 10)
	for i := 0; i < 10; i++ {
		start := time.Now()
		service.CheckPassword(hash, password)
		correctTimes[i] = time.Since(start)
	}

	// Test with wrong password multiple times
	wrongTimes := make([]time.Duration, 10)
	for i := 0; i < 10; i++ {
		start := time.Now()
		service.CheckPassword(hash, "WrongPassword123!")
		wrongTimes[i] = time.Since(start)
	}

	// Assert that timing differences are within reasonable bounds
	// This is a basic timing attack prevention test
	// In practice, bcrypt should handle this, but we test to ensure consistency
	avgCorrect := averageDuration(correctTimes)
	avgWrong := averageDuration(wrongTimes)
	
	// The difference should be minimal (within 50% of each other)
	// This is a loose test since exact timing depends on system load
	ratio := float64(avgCorrect) / float64(avgWrong)
	assert.True(t, ratio > 0.5 && ratio < 2.0, 
		"Timing difference too significant: correct=%v, wrong=%v, ratio=%f", 
		avgCorrect, avgWrong, ratio)
}

func TestPasswordService_CheckPassword_EmptyHash_ReturnsError(t *testing.T) {
	// Arrange
	service := NewPasswordService()
	password := "TestPassword123!"

	// Act
	err := service.CheckPassword("", password)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "hash cannot be empty")
}

func TestPasswordService_CheckPassword_EmptyPassword_ReturnsError(t *testing.T) {
	// Arrange
	service := NewPasswordService()
	hash := "$2a$14$test.hash.here"

	// Act
	err := service.CheckPassword(hash, "")

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "password cannot be empty")
}

func TestPasswordService_CheckPassword_InvalidHash_ReturnsError(t *testing.T) {
	// Arrange
	service := NewPasswordService()
	invalidHash := "invalid.hash.format"
	password := "TestPassword123!"

	// Act
	err := service.CheckPassword(invalidHash, password)

	// Assert
	assert.Error(t, err)
}

// Helper function to calculate average duration
func averageDuration(durations []time.Duration) time.Duration {
	var total time.Duration
	for _, d := range durations {
		total += d
	}
	return total / time.Duration(len(durations))
}