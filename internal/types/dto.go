package types

import (
	"github.com/DuckDHD/BuyOrBye/internal/domain"
)

// LoginRequestDTO represents the login request payload from HTTP clients
type LoginRequestDTO struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// ToDomain converts LoginRequestDTO to domain.Credentials
func (dto LoginRequestDTO) ToDomain() domain.Credentials {
	return domain.Credentials{
		Email:    dto.Email,
		Password: dto.Password,
	}
}

// RegisterRequestDTO represents the registration request payload from HTTP clients
type RegisterRequestDTO struct {
	Email    string `json:"email" validate:"required,email"`
	Name     string `json:"name" validate:"required,min=1"`
	Password string `json:"password" validate:"required,min=8"`
}

// ToDomain converts RegisterRequestDTO to domain.User
func (dto RegisterRequestDTO) ToDomain() *domain.User {
	return &domain.User{
		Email: dto.Email,
		Name:  dto.Name,
		// PasswordHash and timestamps will be set by the service layer
		IsActive: true,
	}
}

// RefreshTokenRequestDTO represents the refresh token request payload from HTTP clients
type RefreshTokenRequestDTO struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// TokenResponseDTO represents the successful authentication response to HTTP clients
type TokenResponseDTO struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// FromDomain converts domain.TokenPair to TokenResponseDTO
func (dto *TokenResponseDTO) FromDomain(tokenPair *domain.TokenPair) {
	dto.AccessToken = tokenPair.AccessToken
	dto.RefreshToken = tokenPair.RefreshToken
	dto.ExpiresIn = tokenPair.ExpiresIn
	dto.TokenType = "Bearer"
}

// ErrorResponseDTO represents error responses to HTTP clients
type ErrorResponseDTO struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    int    `json:"code"`
}

// NewErrorResponse creates a new ErrorResponseDTO
func NewErrorResponse(code int, error string, message string) *ErrorResponseDTO {
	return &ErrorResponseDTO{
		Code:    code,
		Error:   error,
		Message: message,
	}
}

// ValidationErrorResponseDTO represents validation error responses with field-specific errors
type ValidationErrorResponseDTO struct {
	Error   string                 `json:"error"`
	Message string                 `json:"message"`
	Code    int                    `json:"code"`
	Fields  map[string]interface{} `json:"fields,omitempty"`
}

// NewValidationErrorResponse creates a new ValidationErrorResponseDTO
func NewValidationErrorResponse(message string, fields map[string]interface{}) *ValidationErrorResponseDTO {
	return &ValidationErrorResponseDTO{
		Code:    400,
		Error:   "validation_error",
		Message: message,
		Fields:  fields,
	}
}