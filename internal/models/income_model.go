package models

import (
	"time"

	"github.com/DuckDHD/BuyOrBye/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// IncomeModel represents the income table structure in the database
type IncomeModel struct {
	ID        string         `gorm:"primaryKey;type:varchar(36)" json:"id"`
	UserID    string         `gorm:"not null;index;type:varchar(36)" json:"user_id"`
	Source    string         `gorm:"not null;type:varchar(255)" json:"source"`
	Amount    float64        `gorm:"not null;type:decimal(10,2)" json:"amount"`
	Frequency string         `gorm:"not null;type:varchar(20)" json:"frequency"`
	IsActive  bool           `gorm:"not null" json:"is_active"`
	CreatedAt time.Time      `gorm:"not null" json:"created_at"`
	UpdatedAt time.Time      `gorm:"not null" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	// Foreign key relationship
	User UserModel `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
}

// TableName returns the table name for GORM
func (IncomeModel) TableName() string {
	return "incomes"
}

// generateID generates a unique ID with the given prefix
func generateID(prefix string) string {
	return prefix + "-" + uuid.New().String()
}

// BeforeCreate sets the ID if not provided
func (i *IncomeModel) BeforeCreate(tx *gorm.DB) error {
	if i.ID == "" {
		i.ID = generateID("income")
	}
	if i.CreatedAt.IsZero() {
		i.CreatedAt = time.Now()
	}
	if i.UpdatedAt.IsZero() {
		i.UpdatedAt = time.Now()
	}
	return nil
}

// BeforeUpdate updates the UpdatedAt timestamp
func (i *IncomeModel) BeforeUpdate(tx *gorm.DB) error {
	i.UpdatedAt = time.Now()
	return nil
}

// ToDomain converts IncomeModel to domain.Income
func (i IncomeModel) ToDomain() domain.Income {
	return domain.Income{
		ID:        i.ID,
		UserID:    i.UserID,
		Source:    i.Source,
		Amount:    i.Amount,
		Frequency: i.Frequency,
		IsActive:  i.IsActive,
		CreatedAt: i.CreatedAt,
		UpdatedAt: i.UpdatedAt,
	}
}

// FromDomain creates IncomeModel from domain.Income
func (i *IncomeModel) FromDomain(income domain.Income) {
	i.ID = income.ID
	i.UserID = income.UserID
	i.Source = income.Source
	i.Amount = income.Amount
	i.Frequency = income.Frequency
	i.IsActive = income.IsActive
	i.CreatedAt = income.CreatedAt
	i.UpdatedAt = income.UpdatedAt
}

// NewIncomeModelFromDomain creates a new IncomeModel from domain.Income
func NewIncomeModelFromDomain(income domain.Income) *IncomeModel {
	model := &IncomeModel{}
	model.FromDomain(income)
	return model
}