package models

import (
	"testing"
	"time"

	"github.com/DuckDHD/BuyOrBye/internal/domain"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestExpenseModel_TableName_ReturnsCorrectName(t *testing.T) {
	// Arrange
	model := ExpenseModel{}

	// Act
	tableName := model.TableName()

	// Assert
	assert.Equal(t, "expenses", tableName)
}

func TestExpenseModel_BeforeCreate_SetsIDAndTimestamps(t *testing.T) {
	// Arrange
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	model := &ExpenseModel{
		UserID:    "user-123",
		Category:  "housing",
		Name:      "Monthly Rent",
		Amount:    1200.00,
		Frequency: "monthly",
		IsFixed:   true,
		Priority:  1,
	}

	// Act
	err = model.BeforeCreate(db)

	// Assert
	assert.NoError(t, err)
	assert.NotEmpty(t, model.ID)
	assert.Contains(t, model.ID, "expense-")
	assert.False(t, model.CreatedAt.IsZero())
	assert.False(t, model.UpdatedAt.IsZero())
}

func TestExpenseModel_BeforeCreate_DoesNotOverrideExistingID(t *testing.T) {
	// Arrange
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	existingID := "expense-existing-123"
	model := &ExpenseModel{
		ID:        existingID,
		UserID:    "user-123",
		Category:  "housing",
		Name:      "Monthly Rent",
		Amount:    1200.00,
		Frequency: "monthly",
		IsFixed:   true,
		Priority:  1,
	}

	// Act
	err = model.BeforeCreate(db)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, existingID, model.ID)
}

func TestExpenseModel_BeforeUpdate_UpdatesTimestamp(t *testing.T) {
	// Arrange
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	oldTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	model := &ExpenseModel{
		ID:        "expense-123",
		UpdatedAt: oldTime,
	}

	// Act
	err = model.BeforeUpdate(db)

	// Assert
	assert.NoError(t, err)
	assert.True(t, model.UpdatedAt.After(oldTime))
}

func TestExpenseModel_ToDomain_ConvertsCorrectly(t *testing.T) {
	// Arrange
	createdAt := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	updatedAt := time.Date(2024, 1, 15, 11, 0, 0, 0, time.UTC)

	model := ExpenseModel{
		ID:        "expense-123",
		UserID:    "user-456",
		Category:  "housing",
		Name:      "Monthly Rent",
		Amount:    1200.00,
		Frequency: "monthly",
		IsFixed:   true,
		Priority:  1,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}

	// Act
	domainExpense := model.ToDomain()

	// Assert
	assert.Equal(t, model.ID, domainExpense.ID)
	assert.Equal(t, model.UserID, domainExpense.UserID)
	assert.Equal(t, model.Category, domainExpense.Category)
	assert.Equal(t, model.Name, domainExpense.Name)
	assert.Equal(t, model.Amount, domainExpense.Amount)
	assert.Equal(t, model.Frequency, domainExpense.Frequency)
	assert.Equal(t, model.IsFixed, domainExpense.IsFixed)
	assert.Equal(t, model.Priority, domainExpense.Priority)
	assert.Equal(t, model.CreatedAt, domainExpense.CreatedAt)
	assert.Equal(t, model.UpdatedAt, domainExpense.UpdatedAt)
}

func TestExpenseModel_FromDomain_ConvertsCorrectly(t *testing.T) {
	// Arrange
	domainExpense := domain.Expense{
		ID:        "expense-123",
		UserID:    "user-456",
		Category:  "housing",
		Name:      "Monthly Rent",
		Amount:    1200.00,
		Frequency: "monthly",
		IsFixed:   true,
		Priority:  1,
		CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		UpdatedAt: time.Date(2024, 1, 15, 11, 0, 0, 0, time.UTC),
	}

	model := &ExpenseModel{}

	// Act
	model.FromDomain(domainExpense)

	// Assert
	assert.Equal(t, domainExpense.ID, model.ID)
	assert.Equal(t, domainExpense.UserID, model.UserID)
	assert.Equal(t, domainExpense.Category, model.Category)
	assert.Equal(t, domainExpense.Name, model.Name)
	assert.Equal(t, domainExpense.Amount, model.Amount)
	assert.Equal(t, domainExpense.Frequency, model.Frequency)
	assert.Equal(t, domainExpense.IsFixed, model.IsFixed)
	assert.Equal(t, domainExpense.Priority, model.Priority)
	assert.Equal(t, domainExpense.CreatedAt, model.CreatedAt)
	assert.Equal(t, domainExpense.UpdatedAt, model.UpdatedAt)
}

func TestNewExpenseModelFromDomain_CreatesCorrectModel(t *testing.T) {
	// Arrange
	domainExpense := domain.Expense{
		ID:        "expense-123",
		UserID:    "user-456",
		Category:  "housing",
		Name:      "Monthly Rent",
		Amount:    1200.00,
		Frequency: "monthly",
		IsFixed:   true,
		Priority:  1,
		CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		UpdatedAt: time.Date(2024, 1, 15, 11, 0, 0, 0, time.UTC),
	}

	// Act
	model := NewExpenseModelFromDomain(domainExpense)

	// Assert
	assert.NotNil(t, model)
	assert.Equal(t, domainExpense.ID, model.ID)
	assert.Equal(t, domainExpense.UserID, model.UserID)
	assert.Equal(t, domainExpense.Category, model.Category)
	assert.Equal(t, domainExpense.Name, model.Name)
	assert.Equal(t, domainExpense.Amount, model.Amount)
	assert.Equal(t, domainExpense.Frequency, model.Frequency)
	assert.Equal(t, domainExpense.IsFixed, model.IsFixed)
	assert.Equal(t, domainExpense.Priority, model.Priority)
	assert.Equal(t, domainExpense.CreatedAt, model.CreatedAt)
	assert.Equal(t, domainExpense.UpdatedAt, model.UpdatedAt)
}

func TestExpenseModel_RoundTripConversion_PreservesData(t *testing.T) {
	// Arrange
	originalDomain := domain.Expense{
		ID:        "expense-123",
		UserID:    "user-456",
		Category:  "transport",
		Name:      "Gas Money",
		Amount:    200.00,
		Frequency: "weekly",
		IsFixed:   false,
		Priority:  2,
		CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		UpdatedAt: time.Date(2024, 1, 15, 11, 0, 0, 0, time.UTC),
	}

	// Act
	model := NewExpenseModelFromDomain(originalDomain)
	convertedDomain := model.ToDomain()

	// Assert
	assert.Equal(t, originalDomain.ID, convertedDomain.ID)
	assert.Equal(t, originalDomain.UserID, convertedDomain.UserID)
	assert.Equal(t, originalDomain.Category, convertedDomain.Category)
	assert.Equal(t, originalDomain.Name, convertedDomain.Name)
	assert.Equal(t, originalDomain.Amount, convertedDomain.Amount)
	assert.Equal(t, originalDomain.Frequency, convertedDomain.Frequency)
	assert.Equal(t, originalDomain.IsFixed, convertedDomain.IsFixed)
	assert.Equal(t, originalDomain.Priority, convertedDomain.Priority)
	assert.Equal(t, originalDomain.CreatedAt, convertedDomain.CreatedAt)
	assert.Equal(t, originalDomain.UpdatedAt, convertedDomain.UpdatedAt)
}