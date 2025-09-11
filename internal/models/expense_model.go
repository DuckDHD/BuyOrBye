package models

import (
	"time"

	"github.com/DuckDHD/BuyOrBye/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ExpenseModel represents the expense table structure in the database
type ExpenseModel struct {
	ID        string         `gorm:"primaryKey;type:varchar(256)" json:"id"`
	UserID    string         `gorm:"not null;index;type:varchar(256)" json:"user_id"`
	Category  string         `gorm:"not null;type:varchar(50)" json:"category"`
	Name      string         `gorm:"not null;type:varchar(255)" json:"name"`
	Amount    float64        `gorm:"not null;type:decimal(10,2)" json:"amount"`
	Frequency string         `gorm:"not null;type:varchar(20)" json:"frequency"`
	IsFixed   bool           `gorm:"not null;default:false" json:"is_fixed"`
	Priority  int            `gorm:"not null;type:tinyint" json:"priority"`
	CreatedAt time.Time      `gorm:"not null" json:"created_at"`
	UpdatedAt time.Time      `gorm:"not null" json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`

	// Foreign key relationship
	User UserModel `gorm:"foreignKey:UserID;references:ID" json:"user,omitempty"`
}

// TableName returns the table name for GORM
func (ExpenseModel) TableName() string {
	return "expenses"
}

// BeforeCreate sets the ID if not provided
func (e *ExpenseModel) BeforeCreate(tx *gorm.DB) error {
	if e.ID == "" {
		e.ID = "expense-" + uuid.New().String()
	}
	if e.CreatedAt.IsZero() {
		e.CreatedAt = time.Now()
	}
	if e.UpdatedAt.IsZero() {
		e.UpdatedAt = time.Now()
	}
	return nil
}

// BeforeUpdate updates the UpdatedAt timestamp
func (e *ExpenseModel) BeforeUpdate(tx *gorm.DB) error {
	e.UpdatedAt = time.Now()
	return nil
}

// ToDomain converts ExpenseModel to domain.Expense
func (e ExpenseModel) ToDomain() domain.Expense {
	return domain.Expense{
		ID:        e.ID,
		UserID:    e.UserID,
		Category:  e.Category,
		Name:      e.Name,
		Amount:    e.Amount,
		Frequency: e.Frequency,
		IsFixed:   e.IsFixed,
		Priority:  e.Priority,
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
	}
}

// FromDomain creates ExpenseModel from domain.Expense
func (e *ExpenseModel) FromDomain(expense domain.Expense) {
	e.ID = expense.ID
	e.UserID = expense.UserID
	e.Category = expense.Category
	e.Name = expense.Name
	e.Amount = expense.Amount
	e.Frequency = expense.Frequency
	e.IsFixed = expense.IsFixed
	e.Priority = expense.Priority
	e.CreatedAt = expense.CreatedAt
	e.UpdatedAt = expense.UpdatedAt
}

// NewExpenseModelFromDomain creates a new ExpenseModel from domain.Expense
func NewExpenseModelFromDomain(expense domain.Expense) *ExpenseModel {
	model := &ExpenseModel{}
	model.FromDomain(expense)
	return model
}
