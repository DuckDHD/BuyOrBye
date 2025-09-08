package models

import (
	"testing"
	"time"

	"github.com/DuckDHD/BuyOrBye/internal/domain"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestIncomeModel_TableName_ReturnsCorrectName(t *testing.T) {
	// Arrange
	model := IncomeModel{}

	// Act
	tableName := model.TableName()

	// Assert
	assert.Equal(t, "incomes", tableName)
}

func TestIncomeModel_BeforeCreate_SetsIDAndTimestamps(t *testing.T) {
	// Arrange
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	model := &IncomeModel{
		UserID:    "user-123",
		Source:    "Test Income",
		Amount:    5000.00,
		Frequency: "monthly",
		IsActive:  true,
	}

	// Act
	err = model.BeforeCreate(db)

	// Assert
	assert.NoError(t, err)
	assert.NotEmpty(t, model.ID)
	assert.Contains(t, model.ID, "income-")
	assert.False(t, model.CreatedAt.IsZero())
	assert.False(t, model.UpdatedAt.IsZero())
}

func TestIncomeModel_BeforeCreate_DoesNotOverrideExistingID(t *testing.T) {
	// Arrange
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	existingID := "income-existing-123"
	model := &IncomeModel{
		ID:        existingID,
		UserID:    "user-123",
		Source:    "Test Income",
		Amount:    5000.00,
		Frequency: "monthly",
		IsActive:  true,
	}

	// Act
	err = model.BeforeCreate(db)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, existingID, model.ID)
}

func TestIncomeModel_BeforeUpdate_UpdatesTimestamp(t *testing.T) {
	// Arrange
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	assert.NoError(t, err)

	oldTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	model := &IncomeModel{
		ID:        "income-123",
		UpdatedAt: oldTime,
	}

	// Act
	err = model.BeforeUpdate(db)

	// Assert
	assert.NoError(t, err)
	assert.True(t, model.UpdatedAt.After(oldTime))
}

func TestIncomeModel_ToDomain_ConvertsCorrectly(t *testing.T) {
	// Arrange
	createdAt := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	updatedAt := time.Date(2024, 1, 15, 11, 0, 0, 0, time.UTC)

	model := IncomeModel{
		ID:        "income-123",
		UserID:    "user-456",
		Source:    "Software Engineer Salary",
		Amount:    5000.00,
		Frequency: "monthly",
		IsActive:  true,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}

	// Act
	domainIncome := model.ToDomain()

	// Assert
	assert.Equal(t, model.ID, domainIncome.ID)
	assert.Equal(t, model.UserID, domainIncome.UserID)
	assert.Equal(t, model.Source, domainIncome.Source)
	assert.Equal(t, model.Amount, domainIncome.Amount)
	assert.Equal(t, model.Frequency, domainIncome.Frequency)
	assert.Equal(t, model.IsActive, domainIncome.IsActive)
	assert.Equal(t, model.CreatedAt, domainIncome.CreatedAt)
	assert.Equal(t, model.UpdatedAt, domainIncome.UpdatedAt)
}

func TestIncomeModel_FromDomain_ConvertsCorrectly(t *testing.T) {
	// Arrange
	domainIncome := domain.Income{
		ID:        "income-123",
		UserID:    "user-456",
		Source:    "Software Engineer Salary",
		Amount:    5000.00,
		Frequency: "monthly",
		IsActive:  true,
		CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		UpdatedAt: time.Date(2024, 1, 15, 11, 0, 0, 0, time.UTC),
	}

	model := &IncomeModel{}

	// Act
	model.FromDomain(domainIncome)

	// Assert
	assert.Equal(t, domainIncome.ID, model.ID)
	assert.Equal(t, domainIncome.UserID, model.UserID)
	assert.Equal(t, domainIncome.Source, model.Source)
	assert.Equal(t, domainIncome.Amount, model.Amount)
	assert.Equal(t, domainIncome.Frequency, model.Frequency)
	assert.Equal(t, domainIncome.IsActive, model.IsActive)
	assert.Equal(t, domainIncome.CreatedAt, model.CreatedAt)
	assert.Equal(t, domainIncome.UpdatedAt, model.UpdatedAt)
}

func TestNewIncomeModelFromDomain_CreatesCorrectModel(t *testing.T) {
	// Arrange
	domainIncome := domain.Income{
		ID:        "income-123",
		UserID:    "user-456",
		Source:    "Software Engineer Salary",
		Amount:    5000.00,
		Frequency: "monthly",
		IsActive:  true,
		CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		UpdatedAt: time.Date(2024, 1, 15, 11, 0, 0, 0, time.UTC),
	}

	// Act
	model := NewIncomeModelFromDomain(domainIncome)

	// Assert
	assert.NotNil(t, model)
	assert.Equal(t, domainIncome.ID, model.ID)
	assert.Equal(t, domainIncome.UserID, model.UserID)
	assert.Equal(t, domainIncome.Source, model.Source)
	assert.Equal(t, domainIncome.Amount, model.Amount)
	assert.Equal(t, domainIncome.Frequency, model.Frequency)
	assert.Equal(t, domainIncome.IsActive, model.IsActive)
	assert.Equal(t, domainIncome.CreatedAt, model.CreatedAt)
	assert.Equal(t, domainIncome.UpdatedAt, model.UpdatedAt)
}

func TestIncomeModel_RoundTripConversion_PreservesData(t *testing.T) {
	// Arrange
	originalDomain := domain.Income{
		ID:        "income-123",
		UserID:    "user-456",
		Source:    "Software Engineer Salary",
		Amount:    5000.00,
		Frequency: "monthly",
		IsActive:  true,
		CreatedAt: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		UpdatedAt: time.Date(2024, 1, 15, 11, 0, 0, 0, time.UTC),
	}

	// Act
	model := NewIncomeModelFromDomain(originalDomain)
	convertedDomain := model.ToDomain()

	// Assert
	assert.Equal(t, originalDomain.ID, convertedDomain.ID)
	assert.Equal(t, originalDomain.UserID, convertedDomain.UserID)
	assert.Equal(t, originalDomain.Source, convertedDomain.Source)
	assert.Equal(t, originalDomain.Amount, convertedDomain.Amount)
	assert.Equal(t, originalDomain.Frequency, convertedDomain.Frequency)
	assert.Equal(t, originalDomain.IsActive, convertedDomain.IsActive)
	assert.Equal(t, originalDomain.CreatedAt, convertedDomain.CreatedAt)
	assert.Equal(t, originalDomain.UpdatedAt, convertedDomain.UpdatedAt)
}

func TestGenerateID_CreatesUniqueIDs(t *testing.T) {
	// Act
	id1 := generateID("income")
	id2 := generateID("income")

	// Assert
	assert.NotEqual(t, id1, id2)
	assert.Contains(t, id1, "income-")
	assert.Contains(t, id2, "income-")
	assert.True(t, len(id1) > 10) // Should be longer than just "income-"
	assert.True(t, len(id2) > 10)
}