package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/DuckDHD/BuyOrBye/internal/domain"
	"github.com/DuckDHD/BuyOrBye/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupExpenseTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Create tables
	err = db.AutoMigrate(&models.ExpenseModel{})
	require.NoError(t, err)

	return db
}

func createTestExpense(userID, category, name string, amount float64, frequency string, isFixed bool, priority int) domain.Expense {
	return domain.Expense{
		ID:        "expense-" + name + "-123",
		UserID:    userID,
		Category:  category,
		Name:      name,
		Amount:    amount,
		Frequency: frequency,
		IsFixed:   isFixed,
		Priority:  priority,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func TestExpenseRepository_SaveExpense_Success_CreatesNewRecord(t *testing.T) {
	// Arrange
	db := setupExpenseTestDB(t)
	repo := NewExpenseRepository(db)
	ctx := context.Background()

	expense := createTestExpense("user-123", "housing", "Monthly Rent", 1200.00, "monthly", true, 1)

	// Act
	err := repo.SaveExpense(ctx, expense)

	// Assert
	assert.NoError(t, err)

	// Verify in database
	var saved models.ExpenseModel
	result := db.First(&saved, "id = ?", expense.ID)
	assert.NoError(t, result.Error)
	assert.Equal(t, expense.UserID, saved.UserID)
	assert.Equal(t, expense.Category, saved.Category)
	assert.Equal(t, expense.Name, saved.Name)
	assert.Equal(t, expense.Amount, saved.Amount)
	assert.Equal(t, expense.Frequency, saved.Frequency)
	assert.Equal(t, expense.IsFixed, saved.IsFixed)
	assert.Equal(t, expense.Priority, saved.Priority)
}

func TestExpenseRepository_SaveExpense_ExistingRecord_UpdatesRecord(t *testing.T) {
	// Arrange
	db := setupExpenseTestDB(t)
	repo := NewExpenseRepository(db)
	ctx := context.Background()

	expense := createTestExpense("user-123", "housing", "Rent", 1200.00, "monthly", true, 1)

	// Save initial record
	err := repo.SaveExpense(ctx, expense)
	require.NoError(t, err)

	// Modify and save again
	expense.Amount = 1300.00
	expense.Name = "Updated Rent"
	expense.UpdatedAt = time.Now().Add(time.Hour)

	// Act
	err = repo.SaveExpense(ctx, expense)

	// Assert
	assert.NoError(t, err)

	// Verify update
	var saved models.ExpenseModel
	result := db.First(&saved, "id = ?", expense.ID)
	assert.NoError(t, result.Error)
	assert.Equal(t, 1300.00, saved.Amount)
	assert.Equal(t, "Updated Rent", saved.Name)
}

func TestExpenseRepository_GetUserExpenses_ReturnsAll_ForValidUser(t *testing.T) {
	// Arrange
	db := setupExpenseTestDB(t)
	repo := NewExpenseRepository(db)
	ctx := context.Background()

	userID := "user-123"
	otherUserID := "user-456"

	// Create test expenses
	expense1 := createTestExpense(userID, "housing", "Rent", 1200.00, "monthly", true, 1)
	expense2 := createTestExpense(userID, "food", "Groceries", 500.00, "weekly", false, 2)
	expense3 := createTestExpense(otherUserID, "housing", "Other Rent", 1500.00, "monthly", true, 1)

	// Save test data
	require.NoError(t, repo.SaveExpense(ctx, expense1))
	require.NoError(t, repo.SaveExpense(ctx, expense2))
	require.NoError(t, repo.SaveExpense(ctx, expense3))

	// Act
	expenses, err := repo.GetUserExpenses(ctx, userID)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, expenses, 2)

	// Verify only user's expenses are returned
	names := make([]string, len(expenses))
	for i, expense := range expenses {
		assert.Equal(t, userID, expense.UserID)
		names[i] = expense.Name
	}
	assert.Contains(t, names, "Rent")
	assert.Contains(t, names, "Groceries")
	assert.NotContains(t, names, "Other Rent")
}

func TestExpenseRepository_GetUserExpenses_EmptyUser_ReturnsEmpty(t *testing.T) {
	// Arrange
	db := setupExpenseTestDB(t)
	repo := NewExpenseRepository(db)
	ctx := context.Background()

	// Act
	expenses, err := repo.GetUserExpenses(ctx, "nonexistent-user")

	// Assert
	assert.NoError(t, err)
	assert.Empty(t, expenses)
}

func TestExpenseRepository_GetExpensesByCategory_Filters_ReturnsCorrectCategory(t *testing.T) {
	// Arrange
	db := setupExpenseTestDB(t)
	repo := NewExpenseRepository(db)
	ctx := context.Background()

	userID := "user-123"

	// Create expenses in different categories
	housingExpense := createTestExpense(userID, "housing", "Rent", 1200.00, "monthly", true, 1)
	foodExpense1 := createTestExpense(userID, "food", "Groceries", 500.00, "weekly", false, 2)
	foodExpense2 := createTestExpense(userID, "food", "Dining Out", 200.00, "weekly", false, 3)
	transportExpense := createTestExpense(userID, "transport", "Gas", 300.00, "monthly", false, 2)

	require.NoError(t, repo.SaveExpense(ctx, housingExpense))
	require.NoError(t, repo.SaveExpense(ctx, foodExpense1))
	require.NoError(t, repo.SaveExpense(ctx, foodExpense2))
	require.NoError(t, repo.SaveExpense(ctx, transportExpense))

	// Act
	foodExpenses, err := repo.GetExpensesByCategory(ctx, userID, "food")

	// Assert
	assert.NoError(t, err)
	assert.Len(t, foodExpenses, 2)

	for _, expense := range foodExpenses {
		assert.Equal(t, "food", expense.Category)
		assert.Equal(t, userID, expense.UserID)
	}

	names := []string{foodExpenses[0].Name, foodExpenses[1].Name}
	assert.Contains(t, names, "Groceries")
	assert.Contains(t, names, "Dining Out")
}

func TestExpenseRepository_GetExpensesByCategory_NoMatches_ReturnsEmpty(t *testing.T) {
	// Arrange
	db := setupExpenseTestDB(t)
	repo := NewExpenseRepository(db)
	ctx := context.Background()

	userID := "user-123"

	housingExpense := createTestExpense(userID, "housing", "Rent", 1200.00, "monthly", true, 1)
	require.NoError(t, repo.SaveExpense(ctx, housingExpense))

	// Act
	expenses, err := repo.GetExpensesByCategory(ctx, userID, "entertainment")

	// Assert
	assert.NoError(t, err)
	assert.Empty(t, expenses)
}

func TestExpenseRepository_GetFixedExpenses_OnlyFixed_ReturnsOnlyFixedExpenses(t *testing.T) {
	// Arrange
	db := setupExpenseTestDB(t)
	repo := NewExpenseRepository(db)
	ctx := context.Background()

	userID := "user-123"

	// Create fixed and variable expenses
	fixedExpense1 := createTestExpense(userID, "housing", "Rent", 1200.00, "monthly", true, 1)
	fixedExpense2 := createTestExpense(userID, "utilities", "Internet", 80.00, "monthly", true, 1)
	variableExpense := createTestExpense(userID, "food", "Groceries", 500.00, "weekly", false, 2)

	require.NoError(t, repo.SaveExpense(ctx, fixedExpense1))
	require.NoError(t, repo.SaveExpense(ctx, fixedExpense2))
	require.NoError(t, repo.SaveExpense(ctx, variableExpense))

	// Act
	fixedExpenses, err := repo.GetFixedExpenses(ctx, userID)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, fixedExpenses, 2)

	for _, expense := range fixedExpenses {
		assert.True(t, expense.IsFixed)
		assert.Equal(t, userID, expense.UserID)
	}

	names := []string{fixedExpenses[0].Name, fixedExpenses[1].Name}
	assert.Contains(t, names, "Rent")
	assert.Contains(t, names, "Internet")
	assert.NotContains(t, names, "Groceries")
}

func TestExpenseRepository_GetFixedExpenses_NoFixedExpenses_ReturnsEmpty(t *testing.T) {
	// Arrange
	db := setupExpenseTestDB(t)
	repo := NewExpenseRepository(db)
	ctx := context.Background()

	userID := "user-123"

	// Create only variable expenses
	variableExpense := createTestExpense(userID, "food", "Groceries", 500.00, "weekly", false, 2)
	require.NoError(t, repo.SaveExpense(ctx, variableExpense))

	// Act
	expenses, err := repo.GetFixedExpenses(ctx, userID)

	// Assert
	assert.NoError(t, err)
	assert.Empty(t, expenses)
}

func TestExpenseRepository_UpdateExpense_Success_UpdatesExistingRecord(t *testing.T) {
	// Arrange
	db := setupExpenseTestDB(t)
	repo := NewExpenseRepository(db)
	ctx := context.Background()

	expense := createTestExpense("user-123", "housing", "Rent", 1200.00, "monthly", true, 1)
	require.NoError(t, repo.SaveExpense(ctx, expense))

	// Modify expense
	expense.Amount = 1300.00
	expense.Name = "Updated Rent"
	expense.Category = "utilities"
	expense.IsFixed = false
	expense.Priority = 2

	// Act
	err := repo.UpdateExpense(ctx, expense)

	// Assert
	assert.NoError(t, err)

	// Verify update
	var updated models.ExpenseModel
	result := db.First(&updated, "id = ?", expense.ID)
	assert.NoError(t, result.Error)
	assert.Equal(t, 1300.00, updated.Amount)
	assert.Equal(t, "Updated Rent", updated.Name)
	assert.Equal(t, "utilities", updated.Category)
	assert.False(t, updated.IsFixed)
	assert.Equal(t, 2, updated.Priority)
}

func TestExpenseRepository_UpdateExpense_NonexistentRecord_ReturnsError(t *testing.T) {
	// Arrange
	db := setupExpenseTestDB(t)
	repo := NewExpenseRepository(db)
	ctx := context.Background()

	expense := createTestExpense("user-123", "housing", "Nonexistent", 1200.00, "monthly", true, 1)

	// Act
	err := repo.UpdateExpense(ctx, expense)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestExpenseRepository_DeleteExpense_SoftDelete_MarksAsDeleted(t *testing.T) {
	// Arrange
	db := setupExpenseTestDB(t)
	repo := NewExpenseRepository(db)
	ctx := context.Background()

	expense := createTestExpense("user-123", "housing", "Rent", 1200.00, "monthly", true, 1)
	require.NoError(t, repo.SaveExpense(ctx, expense))

	// Act
	err := repo.DeleteExpense(ctx, expense.ID)

	// Assert
	assert.NoError(t, err)

	// Verify soft delete - record should not be found in normal queries
	expenses, err := repo.GetUserExpenses(ctx, expense.UserID)
	assert.NoError(t, err)
	assert.Empty(t, expenses)

	// Verify record still exists but with deleted_at set
	var deleted models.ExpenseModel
	result := db.Unscoped().First(&deleted, "id = ?", expense.ID)
	assert.NoError(t, result.Error)
	assert.NotNil(t, deleted.DeletedAt)
}

func TestExpenseRepository_DeleteExpense_NonexistentRecord_ReturnsError(t *testing.T) {
	// Arrange
	db := setupExpenseTestDB(t)
	repo := NewExpenseRepository(db)
	ctx := context.Background()

	// Act
	err := repo.DeleteExpense(ctx, "nonexistent-id")

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestExpenseRepository_GetExpenseByID_Success_ReturnsCorrectExpense(t *testing.T) {
	// Arrange
	db := setupExpenseTestDB(t)
	repo := NewExpenseRepository(db)
	ctx := context.Background()

	expense := createTestExpense("user-123", "housing", "Rent", 1200.00, "monthly", true, 1)
	require.NoError(t, repo.SaveExpense(ctx, expense))

	// Act
	found, err := repo.GetExpenseByID(ctx, expense.ID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, expense.ID, found.ID)
	assert.Equal(t, expense.UserID, found.UserID)
	assert.Equal(t, expense.Category, found.Category)
	assert.Equal(t, expense.Name, found.Name)
}

func TestExpenseRepository_GetExpenseByID_NotFound_ReturnsError(t *testing.T) {
	// Arrange
	db := setupExpenseTestDB(t)
	repo := NewExpenseRepository(db)
	ctx := context.Background()

	// Act
	_, err := repo.GetExpenseByID(ctx, "nonexistent-id")

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestExpenseRepository_GetExpensesByPriority_FiltersCorrectly(t *testing.T) {
	// Arrange
	db := setupExpenseTestDB(t)
	repo := NewExpenseRepository(db)
	ctx := context.Background()

	userID := "user-123"

	// Create expenses with different priorities
	essential1 := createTestExpense(userID, "housing", "Rent", 1200.00, "monthly", true, 1)
	essential2 := createTestExpense(userID, "utilities", "Electric", 150.00, "monthly", true, 1)
	important := createTestExpense(userID, "food", "Groceries", 500.00, "weekly", false, 2)
	niceToHave := createTestExpense(userID, "entertainment", "Movies", 50.00, "monthly", false, 3)

	require.NoError(t, repo.SaveExpense(ctx, essential1))
	require.NoError(t, repo.SaveExpense(ctx, essential2))
	require.NoError(t, repo.SaveExpense(ctx, important))
	require.NoError(t, repo.SaveExpense(ctx, niceToHave))

	// Act
	essentialExpenses, err := repo.GetExpensesByPriority(ctx, userID, 1)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, essentialExpenses, 2)

	for _, expense := range essentialExpenses {
		assert.Equal(t, 1, expense.Priority)
		assert.Equal(t, userID, expense.UserID)
	}
}

func TestExpenseRepository_CalculateUserTotalExpenses_ReturnsCorrectSum(t *testing.T) {
	// Arrange
	db := setupExpenseTestDB(t)
	repo := NewExpenseRepository(db)
	ctx := context.Background()

	userID := "user-123"
	otherUserID := "user-456"

	// Create test expenses for user
	expense1 := createTestExpense(userID, "housing", "Rent", 1200.00, "monthly", true, 1)
	expense2 := createTestExpense(userID, "food", "Groceries", 500.00, "monthly", false, 2)

	// Create expense for other user
	otherExpense := createTestExpense(otherUserID, "housing", "Other Rent", 1500.00, "monthly", true, 1)

	require.NoError(t, repo.SaveExpense(ctx, expense1))
	require.NoError(t, repo.SaveExpense(ctx, expense2))
	require.NoError(t, repo.SaveExpense(ctx, otherExpense))

	// Act
	total, err := repo.CalculateUserTotalExpenses(ctx, userID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 1700.00, total) // 1200 + 500, excluding other user's expense
}

func TestExpenseRepository_GetExpensesByFrequency_FiltersCorrectly(t *testing.T) {
	// Arrange
	db := setupExpenseTestDB(t)
	repo := NewExpenseRepository(db)
	ctx := context.Background()

	userID := "user-123"

	// Create expenses with different frequencies
	monthly1 := createTestExpense(userID, "housing", "Rent", 1200.00, "monthly", true, 1)
	monthly2 := createTestExpense(userID, "utilities", "Electric", 150.00, "monthly", true, 1)
	weekly := createTestExpense(userID, "food", "Groceries", 100.00, "weekly", false, 2)

	require.NoError(t, repo.SaveExpense(ctx, monthly1))
	require.NoError(t, repo.SaveExpense(ctx, monthly2))
	require.NoError(t, repo.SaveExpense(ctx, weekly))

	// Act
	monthlyExpenses, err := repo.GetExpensesByFrequency(ctx, userID, "monthly")

	// Assert
	assert.NoError(t, err)
	assert.Len(t, monthlyExpenses, 2)

	for _, expense := range monthlyExpenses {
		assert.Equal(t, "monthly", expense.Frequency)
		assert.Equal(t, userID, expense.UserID)
	}
}
