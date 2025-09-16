package types

import (
	"time"
	"github.com/DuckDHD/BuyOrBye/internal/domain"
)

// UserResponseDTO represents user data for template rendering
// This DTO is specifically for templates and UI rendering
type UserResponseDTO struct {
	ID          string     `json:"id" validate:"required"`
	Email       string     `json:"email" validate:"required,email"`
	Name        string     `json:"name" validate:"required,min=1"`
	IsActive    bool       `json:"is_active"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
}

// FromDomain converts domain.User to UserResponseDTO
func (dto *UserResponseDTO) FromDomain(user *domain.User) {
	dto.ID = user.ID
	dto.Email = user.Email
	dto.Name = user.Name
	dto.IsActive = user.IsActive
	dto.CreatedAt = user.CreatedAt
	dto.UpdatedAt = user.UpdatedAt
	dto.LastLoginAt = nil // Not available in domain
}

// FromDomainUser creates UserResponseDTO from domain.User
func FromDomainUser(user *domain.User) UserResponseDTO {
	return UserResponseDTO{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		IsActive:  user.IsActive,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		LastLoginAt: nil, // Not available in domain
	}
}

// FromDomainUserList creates slice of UserResponseDTO from slice of domain.User
func FromDomainUserList(users []*domain.User) []UserResponseDTO {
	dtos := make([]UserResponseDTO, len(users))
	for i, user := range users {
		dtos[i] = FromDomainUser(user)
	}
	return dtos
}

// ToDomain converts UserResponseDTO to domain.User
func (dto UserResponseDTO) ToDomain() *domain.User {
	return &domain.User{
		ID:        dto.ID,
		Email:     dto.Email,
		Name:      dto.Name,
		IsActive:  dto.IsActive,
		CreatedAt: dto.CreatedAt,
		UpdatedAt: dto.UpdatedAt,
	}
}

/*
Request UserUpdateDTO dto
User profile update request
*/
type UserUpdateDTO struct {
	Name  string `json:"name" validate:"required,min=1"`
	Email string `json:"email" validate:"required,email"`
}

// ToDomain converts UserUpdateDTO to domain.User
func (dto UserUpdateDTO) ToDomain() *domain.User {
	return &domain.User{
		Name:  dto.Name,
		Email: dto.Email,
	}
}

// UserProfileDTO represents user profile data (alias for UserResponseDTO)
type UserProfileDTO = UserResponseDTO