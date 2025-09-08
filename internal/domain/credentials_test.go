package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCredentials_Validate_AllFieldsValid_ReturnsNil(t *testing.T) {
	// Arrange
	credentials := Credentials{
		Email:    "test@example.com",
		Password: "ValidPassword123!",
	}

	// Act
	err := credentials.Validate()

	// Assert
	assert.NoError(t, err)
}

func TestCredentials_Validate_EmptyEmail_ReturnsError(t *testing.T) {
	// Arrange
	credentials := Credentials{
		Email:    "",
		Password: "ValidPassword123!",
	}

	// Act
	err := credentials.Validate()

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "email is required")
}

func TestCredentials_Validate_InvalidEmailFormat_ReturnsError(t *testing.T) {
	// Arrange
	invalidEmails := []string{
		"invalid-email",
		"@example.com",
		"test@",
		"test.example.com",
		"test@.com",
		"test@example.",
		"test@@example.com",
	}

	for _, email := range invalidEmails {
		t.Run("invalid_email_"+email, func(t *testing.T) {
			credentials := Credentials{
				Email:    email,
				Password: "ValidPassword123!",
			}

			// Act
			err := credentials.Validate()

			// Assert
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "invalid email format")
		})
	}
}

func TestCredentials_Validate_ValidEmailFormats_ReturnsNil(t *testing.T) {
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
			credentials := Credentials{
				Email:    email,
				Password: "ValidPassword123!",
			}

			// Act
			err := credentials.Validate()

			// Assert
			assert.NoError(t, err)
		})
	}
}

func TestCredentials_Validate_EmptyPassword_ReturnsError(t *testing.T) {
	// Arrange
	credentials := Credentials{
		Email:    "test@example.com",
		Password: "",
	}

	// Act
	err := credentials.Validate()

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "password is required")
}

func TestCredentials_Validate_PasswordTooShort_ReturnsError(t *testing.T) {
	// Arrange
	shortPasswords := []string{
		"1234567",    // 7 characters
		"Ab1!",       // 4 characters
		"Pass1",      // 6 characters
		"1234567",    // Exactly 7 characters
	}

	for _, password := range shortPasswords {
		t.Run("short_password_"+password, func(t *testing.T) {
			credentials := Credentials{
				Email:    "test@example.com",
				Password: password,
			}

			// Act
			err := credentials.Validate()

			// Assert
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "password must be at least 8 characters")
		})
	}
}

func TestCredentials_Validate_ValidPasswordLengths_ReturnsNil(t *testing.T) {
	// Arrange
	validPasswords := []string{
		"12345678",           // Exactly 8 characters
		"ValidPassword123!",  // Normal length
		"VeryLongPasswordWithManyCharacters123!", // Long password
	}

	for _, password := range validPasswords {
		t.Run("valid_password_length_"+password, func(t *testing.T) {
			credentials := Credentials{
				Email:    "test@example.com",
				Password: password,
			}

			// Act
			err := credentials.Validate()

			// Assert
			assert.NoError(t, err)
		})
	}
}

func TestCredentials_Validate_MultipleErrors_ReturnsAllErrors(t *testing.T) {
	// Arrange
	credentials := Credentials{
		Email:    "invalid-email", // Invalid email format
		Password: "short",         // Too short password
	}

	// Act
	err := credentials.Validate()

	// Assert
	assert.Error(t, err)
	errorMsg := err.Error()
	assert.Contains(t, errorMsg, "invalid email format")
	assert.Contains(t, errorMsg, "password must be at least 8 characters")
}

func TestCredentials_Validate_BothFieldsEmpty_ReturnsAllErrors(t *testing.T) {
	// Arrange
	credentials := Credentials{
		Email:    "",
		Password: "",
	}

	// Act
	err := credentials.Validate()

	// Assert
	assert.Error(t, err)
	errorMsg := err.Error()
	assert.Contains(t, errorMsg, "email is required")
	assert.Contains(t, errorMsg, "password is required")
}