package types

import (
	"github.com/DuckDHD/BuyOrBye/internal/domain"
)

/*
Response ErrorResponseDTO dto
Standard error response format for HTTP endpoints
*/
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

/*
Response ValidationErrorResponseDTO dto
Validation error response with field-specific error details
*/
type ValidationErrorResponseDTO struct {
	Error   string                 `json:"error"`
	Message string                 `json:"message"`
	Code    int                    `json:"code"`
	Fields  map[string]any `json:"fields,omitempty"`
}

// NewValidationErrorResponse creates a new ValidationErrorResponseDTO
func NewValidationErrorResponse(message string, fields map[string]any) *ValidationErrorResponseDTO {
	return &ValidationErrorResponseDTO{
		Code:    400,
		Error:   "validation_error",
		Message: message,
		Fields:  fields,
	}
}

/*
UI UserDTO dto
Simple user data transfer object for UI templates
*/
type UserDTO struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

/*
Layout LayoutDTO dto
Layout data transfer object for base template rendering with CSRF protection
*/
type LayoutDTO struct {
	Title     string            `json:"title"`
	CSRFToken string            `json:"csrf_token"`
	User      *UserResponseDTO  `json:"user,omitempty"`
}

// FromDomainUser converts domain.User to UserDTO
func (dto *UserDTO) FromDomainUser(user domain.User) {
	dto.ID = user.ID
	dto.Email = user.Email
	dto.Name = user.Name
}

// ToDomain converts UserDTO to domain.User
func (dto *UserDTO) ToDomain() *domain.User {
	return &domain.User{
		ID:    dto.ID,
		Email: dto.Email,
		Name:  dto.Name,
	}
}

// FromDomainUserList converts slice of domain.User to slice of UserDTO
func FromDomainUserListToCommon(users []domain.User) []UserDTO {
	dtos := make([]UserDTO, len(users))
	for i, user := range users {
		dtos[i].FromDomainUser(user)
	}
	return dtos
}

// NewLayoutDTO creates a new LayoutDTO with basic fields
func NewLayoutDTO(title, csrfToken string, user *UserResponseDTO) LayoutDTO {
	return LayoutDTO{
		Title:     title,
		CSRFToken: csrfToken,
		User:      user,
	}
}