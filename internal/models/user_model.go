package models

import (
	"strconv"
	"time"

	"github.com/DuckDHD/BuyOrBye/internal/domain"
	"gorm.io/gorm"
)

// UserModel represents the GORM model for users table
// This struct defines the database schema and should only be used in the repository layer
type UserModel struct {
	gorm.Model
	Email        string     `gorm:"uniqueIndex;not null"`
	Name         string     `gorm:"not null"`
	PasswordHash string     `gorm:"not null"`
	IsActive     bool       `gorm:"default:true"`
	LastLoginAt  *time.Time `gorm:"default:null"`
}

// TableName returns the table name for GORM
func (UserModel) TableName() string {
	return "users"
}

// ToDomain converts the GORM model to a domain entity
// This method should only be called in the repository layer
func (m UserModel) ToDomain() domain.User {
	return domain.User{
		ID:           strconv.FormatUint(uint64(m.ID), 10),
		Email:        m.Email,
		Name:         m.Name,
		PasswordHash: m.PasswordHash,
		IsActive:     m.IsActive,
		CreatedAt:    m.CreatedAt,
		UpdatedAt:    m.UpdatedAt,
	}
}

// UserFromDomain converts a domain entity to a GORM model
// This function should only be called in the repository layer
func UserFromDomain(d domain.User) UserModel {
	var id uint
	if d.ID != "" {
		if parsed, err := strconv.ParseUint(d.ID, 10, 32); err == nil {
			id = uint(parsed)
		}
	}

	model := UserModel{
		Email:        d.Email,
		Name:         d.Name,
		PasswordHash: d.PasswordHash,
		IsActive:     d.IsActive,
	}

	// Set ID if it exists (for updates)
	if id > 0 {
		model.ID = id
	}

	// Set timestamps if they exist
	if !d.CreatedAt.IsZero() {
		model.CreatedAt = d.CreatedAt
	}
	if !d.UpdatedAt.IsZero() {
		model.UpdatedAt = d.UpdatedAt
	}

	return model
}