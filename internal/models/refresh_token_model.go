package models

import (
	"strconv"
	"time"

	"gorm.io/gorm"
)

// RefreshTokenModel represents the GORM model for refresh tokens table
// This struct defines the database schema and should only be used in the repository layer
type RefreshTokenModel struct {
	gorm.Model
	UserID    uint      `gorm:"not null;index"`
	Token     string    `gorm:"uniqueIndex;not null;size:255"`
	ExpiresAt time.Time `gorm:"not null"`
	IsRevoked bool      `gorm:"default:false"`
	RevokedAt *time.Time
	User      UserModel `gorm:"foreignKey:UserID;references:ID;onDelete:CASCADE"`
}

// TableName returns the table name for GORM
func (RefreshTokenModel) TableName() string {
	return "refresh_tokens"
}

// IsExpired checks if the refresh token has expired
func (r RefreshTokenModel) IsExpired() bool {
	return time.Now().After(r.ExpiresAt)
}

// IsValid checks if the refresh token is valid (not expired and not revoked)
func (r RefreshTokenModel) IsValid() bool {
	return !r.IsExpired() && !r.IsRevoked
}

// ToUserID converts the model's UserID to a string format used by domain entities
func (r RefreshTokenModel) ToUserID() string {
	return strconv.FormatUint(uint64(r.UserID), 10)
}

// RefreshTokenFromDomain creates a RefreshTokenModel from domain data
// userID should be a string representation of the user ID
// token is the actual refresh token string
// expiresAt is when the token expires
func RefreshTokenFromDomain(userID, token string, expiresAt time.Time) RefreshTokenModel {
	var uid uint
	if parsed, err := strconv.ParseUint(userID, 10, 32); err == nil {
		uid = uint(parsed)
	}

	return RefreshTokenModel{
		UserID:    uid,
		Token:     token,
		ExpiresAt: expiresAt,
		IsRevoked: false,
	}
}