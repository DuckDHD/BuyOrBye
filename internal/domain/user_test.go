package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestUser_Validate_AllFieldsValid_ReturnsNil(t *testing.T) {
	// Arrange
	user := User{
		ID:           "user-123",
		Email:        "test@example.com",
		Name:         "Test User",
		PasswordHash: "$2a$14$hashedpassword",
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Act
	err := user.Validate()

	// Assert
	assert.NoError(t, err)
}

func TestUser_Validate_EmptyEmail_ReturnsError(t *testing.T) {
	// Arrange
	user := User{
		ID:           "user-123",
		Email:        "",
		Name:         "Test User",
		PasswordHash: "$2a$14$hashedpassword",
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Act
	err := user.Validate()

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "email is required")
}

func TestUser_Validate_InvalidEmailFormat_ReturnsError(t *testing.T) {
	// Arrange
	invalidEmails := []string{
		"invalid-email",
		"@example.com",
		"test@",
		"test.example.com",
		"test@.com",
		"test@example.",
		"test@@example.com",
		"",
	}

	for _, email := range invalidEmails {
		t.Run("email_"+email, func(t *testing.T) {
			user := User{
				ID:           "user-123",
				Email:        email,
				Name:         "Test User",
				PasswordHash: "$2a$14$hashedpassword",
				IsActive:     true,
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			}

			// Act
			err := user.Validate()

			// Assert
			assert.Error(t, err)
			if email == "" {
				assert.Contains(t, err.Error(), "email is required")
			} else {
				assert.Contains(t, err.Error(), "invalid email format")
			}
		})
	}
}

func TestUser_Validate_ValidEmailFormats_ReturnsNil(t *testing.T) {
	// Arrange
	validEmails := []string{
		"test@example.com",
		"user.name@example.com",
		"user+tag@example.co.uk",
		"test123@example-site.com",
		"user_name@example.org",
	}

	for _, email := range validEmails {
		t.Run("valid_email_"+email, func(t *testing.T) {
			user := User{
				ID:           "user-123",
				Email:        email,
				Name:         "Test User",
				PasswordHash: "$2a$14$hashedpassword",
				IsActive:     true,
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			}

			// Act
			err := user.Validate()

			// Assert
			assert.NoError(t, err)
		})
	}
}

func TestUser_Validate_EmptyName_ReturnsError(t *testing.T) {
	// Arrange
	user := User{
		ID:           "user-123",
		Email:        "test@example.com",
		Name:         "",
		PasswordHash: "$2a$14$hashedpassword",
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Act
	err := user.Validate()

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "name is required")
}

func TestUser_Validate_EmptyPasswordHash_ReturnsError(t *testing.T) {
	// Arrange
	user := User{
		ID:           "user-123",
		Email:        "test@example.com",
		Name:         "Test User",
		PasswordHash: "",
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Act
	err := user.Validate()

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "password hash is required")
}

func TestUser_Validate_ZeroCreatedAt_ReturnsError(t *testing.T) {
	// Arrange
	user := User{
		ID:           "user-123",
		Email:        "test@example.com",
		Name:         "Test User",
		PasswordHash: "$2a$14$hashedpassword",
		IsActive:     true,
		CreatedAt:    time.Time{}, // Zero time
		UpdatedAt:    time.Now(),
	}

	// Act
	err := user.Validate()

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "created at is required")
}

func TestUser_Validate_ZeroUpdatedAt_ReturnsError(t *testing.T) {
	// Arrange
	user := User{
		ID:           "user-123",
		Email:        "test@example.com",
		Name:         "Test User",
		PasswordHash: "$2a$14$hashedpassword",
		IsActive:     true,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Time{}, // Zero time
	}

	// Act
	err := user.Validate()

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "updated at is required")
}

func TestUser_Validate_MultipleErrors_ReturnsAllErrors(t *testing.T) {
	// Arrange
	user := User{
		ID:           "user-123",
		Email:        "", // Missing email
		Name:         "", // Missing name
		PasswordHash: "", // Missing password hash
		IsActive:     true,
		CreatedAt:    time.Time{}, // Zero time
		UpdatedAt:    time.Time{}, // Zero time
	}

	// Act
	err := user.Validate()

	// Assert
	assert.Error(t, err)
	errorMsg := err.Error()
	assert.Contains(t, errorMsg, "email is required")
	assert.Contains(t, errorMsg, "name is required")
	assert.Contains(t, errorMsg, "password hash is required")
	assert.Contains(t, errorMsg, "created at is required")
	assert.Contains(t, errorMsg, "updated at is required")
}