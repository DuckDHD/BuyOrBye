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

/*
Chat ChatPageDTO dto
Chat page data transfer object for chat landing page
*/
type ChatPageDTO struct {
	CSRFToken string           `json:"csrf_token"`
	User      *UserResponseDTO `json:"user,omitempty"`
}

/*
Chat ChatMessageDTO dto
Chat message data transfer object for chat interactions
*/
type ChatMessageDTO struct {
	Message   string `json:"message" binding:"required,max=500"`
	SessionID string `json:"session_id,omitempty"`
}

/*
Chat ChatResponseDTO dto
Chat response data transfer object for AI responses
*/
type ChatResponseDTO struct {
	Response  string          `json:"response"`
	Status    string          `json:"status"` // processing, need_info, decision
	Decision  *DecisionResult `json:"decision,omitempty"`
	Questions []string        `json:"questions,omitempty"`
	SessionID string          `json:"session_id"`
}

/*
AI DecisionResult dto
AI decision result data transfer object
*/
type DecisionResult struct {
	Recommendation string  `json:"recommendation"` // buy, wait, bye
	Confidence     float64 `json:"confidence"`
	Reasoning      string  `json:"reasoning"`
	WaitPeriod     *int    `json:"wait_period,omitempty"` // days
}

// NewLayoutDTO creates a new LayoutDTO with basic fields
func NewLayoutDTO(title, csrfToken string, user *UserResponseDTO) LayoutDTO {
	return LayoutDTO{
		Title:     title,
		CSRFToken: csrfToken,
		User:      user,
	}
}

// NewChatPageDTO creates a new ChatPageDTO
func NewChatPageDTO(csrfToken string, user *UserResponseDTO) ChatPageDTO {
	return ChatPageDTO{
		CSRFToken: csrfToken,
		User:      user,
	}
}