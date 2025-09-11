package dtos

import (
	"time"

	"github.com/DuckDHD/BuyOrBye/internal/domain"
)

/*
Response UserProfileDTO dto
User profile information for API responses with account status
*/
type UserProfileDTO struct {
	ID          string    `json:"id" example:"user-123"`
	Email       string    `json:"email" example:"user@example.com"`
	Name        string    `json:"name" example:"John Doe"`
	IsActive    bool      `json:"is_active" example:"true"`
	CreatedAt   time.Time `json:"created_at" example:"2024-01-15T10:30:00Z"`
	UpdatedAt   time.Time `json:"updated_at" example:"2024-01-15T10:30:00Z"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty" example:"2024-01-15T14:30:00Z"`
}

/*
Request UpdateUserProfileDTO dto
User profile update request with optional fields
*/
type UpdateUserProfileDTO struct {
	Name  *string `json:"name,omitempty" validate:"omitempty,min=1" example:"John Smith"`
	Email *string `json:"email,omitempty" validate:"omitempty,email" example:"johnsmith@example.com"`
}

// FromDomain converts domain.User to UserProfileDTO
func (dto *UserProfileDTO) FromDomain(user domain.User) {
	dto.ID = user.ID
	dto.Email = user.Email
	dto.Name = user.Name
	dto.IsActive = user.IsActive
	dto.CreatedAt = user.CreatedAt
	dto.UpdatedAt = user.UpdatedAt
}

// ApplyUpdates applies UpdateUserProfileDTO fields to domain.User
func (dto UpdateUserProfileDTO) ApplyUpdates(user *domain.User) {
	if dto.Name != nil {
		user.Name = *dto.Name
	}
	if dto.Email != nil {
		user.Email = *dto.Email
	}
	user.UpdatedAt = time.Now()
}