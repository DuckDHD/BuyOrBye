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

func setupIncomeTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Create tables
	err = db.AutoMigrate(&models.IncomeModel{})
	require.NoError(t, err)

	return db
}

func createTestIncome(userID, source string, amount float64, frequency string) domain.Income {
	return domain.Income{
		ID:        "income-" + source + "-123",
		UserID:    userID,
		Source:    source,
		Amount:    amount,
		Frequency: frequency,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func createTestIncomeWithActiveStatus(userID, source string, amount float64, frequency string, isActive bool) domain.Income {
	return domain.Income{
		ID:        "income-" + source + "-123",
		UserID:    userID,
		Source:    source,
		Amount:    amount,
		Frequency: frequency,
		IsActive:  isActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func TestIncomeRepository_SaveIncome_Success_CreatesNewRecord(t *testing.T) {
	// Arrange
	db := setupIncomeTestDB(t)
	repo := NewIncomeRepository(db)
	ctx := context.Background()

	income := createTestIncome("user-123", "Salary", 5000.00, "monthly")

	// Act
	err := repo.SaveIncome(ctx, income)

	// Assert
	assert.NoError(t, err)

	// Verify in database
	var saved models.IncomeModel
	result := db.First(&saved, "id = ?", income.ID)
	assert.NoError(t, result.Error)
	assert.Equal(t, income.UserID, saved.UserID)
	assert.Equal(t, income.Source, saved.Source)
	assert.Equal(t, income.Amount, saved.Amount)
	assert.Equal(t, income.Frequency, saved.Frequency)
	assert.True(t, saved.IsActive)
}

func TestIncomeRepository_SaveIncome_ExistingRecord_UpdatesRecord(t *testing.T) {
	// Arrange
	db := setupIncomeTestDB(t)
	repo := NewIncomeRepository(db)
	ctx := context.Background()

	income := createTestIncome("user-123", "Salary", 5000.00, "monthly")

	// Save initial record
	err := repo.SaveIncome(ctx, income)
	require.NoError(t, err)

	// Modify and save again
	income.Amount = 5500.00
	income.Source = "Updated Salary"
	income.UpdatedAt = time.Now().Add(time.Hour)

	// Act
	err = repo.SaveIncome(ctx, income)

	// Assert
	assert.NoError(t, err)

	// Verify update
	var saved models.IncomeModel
	result := db.First(&saved, "id = ?", income.ID)
	assert.NoError(t, result.Error)
	assert.Equal(t, 5500.00, saved.Amount)
	assert.Equal(t, "Updated Salary", saved.Source)
}

func TestIncomeRepository_GetUserIncomes_ReturnsAll_ForValidUser(t *testing.T) {
	// Arrange
	db := setupIncomeTestDB(t)
	repo := NewIncomeRepository(db)
	ctx := context.Background()

	userID := "user-123"
	otherUserID := "user-456"

	// Create test incomes
	income1 := createTestIncome(userID, "Salary", 5000.00, "monthly")
	income2 := createTestIncome(userID, "Freelance", 1000.00, "weekly")
	income3 := createTestIncome(otherUserID, "Other Salary", 6000.00, "monthly")

	// Save test data
	require.NoError(t, repo.SaveIncome(ctx, income1))
	require.NoError(t, repo.SaveIncome(ctx, income2))
	require.NoError(t, repo.SaveIncome(ctx, income3))

	// Act
	incomes, err := repo.GetUserIncomes(ctx, userID)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, incomes, 2)

	// Verify only user's incomes are returned
	sources := make([]string, len(incomes))
	for i, income := range incomes {
		assert.Equal(t, userID, income.UserID)
		sources[i] = income.Source
	}
	assert.Contains(t, sources, "Salary")
	assert.Contains(t, sources, "Freelance")
	assert.NotContains(t, sources, "Other Salary")
}

func TestIncomeRepository_GetUserIncomes_EmptyUser_ReturnsEmpty(t *testing.T) {
	// Arrange
	db := setupIncomeTestDB(t)
	repo := NewIncomeRepository(db)
	ctx := context.Background()

	// Act
	incomes, err := repo.GetUserIncomes(ctx, "nonexistent-user")

	// Assert
	assert.NoError(t, err)
	assert.Empty(t, incomes)
}

func TestIncomeRepository_UpdateIncome_Success_UpdatesExistingRecord(t *testing.T) {
	// Arrange
	db := setupIncomeTestDB(t)
	repo := NewIncomeRepository(db)
	ctx := context.Background()

	income := createTestIncome("user-123", "Salary", 5000.00, "monthly")
	require.NoError(t, repo.SaveIncome(ctx, income))

	// Modify income
	income.Amount = 6000.00
	income.Source = "Senior Salary"
	income.Frequency = "weekly"

	// Act
	err := repo.UpdateIncome(ctx, income)

	// Assert
	assert.NoError(t, err)

	// Verify update
	var updated models.IncomeModel
	result := db.First(&updated, "id = ?", income.ID)
	assert.NoError(t, result.Error)
	assert.Equal(t, 6000.00, updated.Amount)
	assert.Equal(t, "Senior Salary", updated.Source)
	assert.Equal(t, "weekly", updated.Frequency)
}

func TestIncomeRepository_UpdateIncome_NonexistentRecord_ReturnsError(t *testing.T) {
	// Arrange
	db := setupIncomeTestDB(t)
	repo := NewIncomeRepository(db)
	ctx := context.Background()

	income := createTestIncome("user-123", "Nonexistent", 5000.00, "monthly")

	// Act
	err := repo.UpdateIncome(ctx, income)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestIncomeRepository_DeleteIncome_SoftDelete_MarksAsDeleted(t *testing.T) {
	// Arrange
	db := setupIncomeTestDB(t)
	repo := NewIncomeRepository(db)
	ctx := context.Background()

	income := createTestIncome("user-123", "Salary", 5000.00, "monthly")
	require.NoError(t, repo.SaveIncome(ctx, income))

	// Act
	err := repo.DeleteIncome(ctx, income.ID)

	// Assert
	assert.NoError(t, err)

	// Verify soft delete - record should not be found in normal queries
	incomes, err := repo.GetUserIncomes(ctx, income.UserID)
	assert.NoError(t, err)
	assert.Empty(t, incomes)

	// Verify record still exists but with deleted_at set
	var deleted models.IncomeModel
	result := db.Unscoped().First(&deleted, "id = ?", income.ID)
	assert.NoError(t, result.Error)
	assert.NotNil(t, deleted.DeletedAt)
}

func TestIncomeRepository_DeleteIncome_NonexistentRecord_ReturnsError(t *testing.T) {
	// Arrange
	db := setupIncomeTestDB(t)
	repo := NewIncomeRepository(db)
	ctx := context.Background()

	// Act
	err := repo.DeleteIncome(ctx, "nonexistent-id")

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestIncomeRepository_GetActiveIncomes_FiltersInactive_ReturnsOnlyActive(t *testing.T) {
	// Arrange
	db := setupIncomeTestDB(t)
	repo := NewIncomeRepository(db)
	ctx := context.Background()

	userID := "user-123"

	// Create active and inactive incomes
	activeIncome := createTestIncomeWithActiveStatus(userID, "Active Salary", 5000.00, "monthly", true)
	inactiveIncome := createTestIncomeWithActiveStatus(userID, "Inactive Freelance", 1000.00, "weekly", false)


	require.NoError(t, repo.SaveIncome(ctx, activeIncome))
	require.NoError(t, repo.SaveIncome(ctx, inactiveIncome))

	// Act
	incomes, err := repo.GetActiveIncomes(ctx, userID)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, incomes, 1)
	assert.Equal(t, "Active Salary", incomes[0].Source)
	assert.True(t, incomes[0].IsActive)
}

func TestIncomeRepository_GetActiveIncomes_NoActiveIncomes_ReturnsEmpty(t *testing.T) {
	// Arrange
	db := setupIncomeTestDB(t)
	repo := NewIncomeRepository(db)
	ctx := context.Background()

	userID := "user-123"

	// Create only inactive income
	inactiveIncome := createTestIncomeWithActiveStatus(userID, "Inactive", 1000.00, "weekly", false)
	require.NoError(t, repo.SaveIncome(ctx, inactiveIncome))

	// Act
	incomes, err := repo.GetActiveIncomes(ctx, userID)

	// Assert
	assert.NoError(t, err)
	assert.Empty(t, incomes)
}

func TestIncomeRepository_GetIncomeByID_Success_ReturnsCorrectIncome(t *testing.T) {
	// Arrange
	db := setupIncomeTestDB(t)
	repo := NewIncomeRepository(db)
	ctx := context.Background()

	income := createTestIncome("user-123", "Salary", 5000.00, "monthly")
	require.NoError(t, repo.SaveIncome(ctx, income))

	// Act
	found, err := repo.GetIncomeByID(ctx, income.ID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, income.ID, found.ID)
	assert.Equal(t, income.UserID, found.UserID)
	assert.Equal(t, income.Source, found.Source)
	assert.Equal(t, income.Amount, found.Amount)
}

func TestIncomeRepository_GetIncomeByID_NotFound_ReturnsError(t *testing.T) {
	// Arrange
	db := setupIncomeTestDB(t)
	repo := NewIncomeRepository(db)
	ctx := context.Background()

	// Act
	_, err := repo.GetIncomeByID(ctx, "nonexistent-id")

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestIncomeRepository_GetUserIncomesByFrequency_FiltersCorrectly(t *testing.T) {
	// Arrange
	db := setupIncomeTestDB(t)
	repo := NewIncomeRepository(db)
	ctx := context.Background()

	userID := "user-123"

	// Create incomes with different frequencies
	monthly1 := createTestIncome(userID, "Salary", 5000.00, "monthly")
	monthly2 := createTestIncome(userID, "Bonus", 1000.00, "monthly")
	weekly := createTestIncome(userID, "Freelance", 500.00, "weekly")

	require.NoError(t, repo.SaveIncome(ctx, monthly1))
	require.NoError(t, repo.SaveIncome(ctx, monthly2))
	require.NoError(t, repo.SaveIncome(ctx, weekly))

	// Act
	monthlyIncomes, err := repo.GetUserIncomesByFrequency(ctx, userID, "monthly")

	// Assert
	assert.NoError(t, err)
	assert.Len(t, monthlyIncomes, 2)

	for _, income := range monthlyIncomes {
		assert.Equal(t, "monthly", income.Frequency)
		assert.Equal(t, userID, income.UserID)
	}
}

func TestIncomeRepository_CalculateUserTotalIncome_ReturnsCorrectSum(t *testing.T) {
	// Arrange
	db := setupIncomeTestDB(t)
	repo := NewIncomeRepository(db)
	ctx := context.Background()

	userID := "user-123"

	// Create test incomes
	income1 := createTestIncome(userID, "Salary", 5000.00, "monthly")
	income2 := createTestIncome(userID, "Freelance", 1500.00, "monthly")
	inactiveIncome := createTestIncomeWithActiveStatus(userID, "Old Job", 3000.00, "monthly", false)

	require.NoError(t, repo.SaveIncome(ctx, income1))
	require.NoError(t, repo.SaveIncome(ctx, income2))
	require.NoError(t, repo.SaveIncome(ctx, inactiveIncome))

	// Act
	total, err := repo.CalculateUserTotalIncome(ctx, userID, true) // Active only

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 6500.00, total) // 5000 + 1500, excluding inactive
}

func TestIncomeRepository_CalculateUserTotalIncome_IncludeInactive_ReturnsAllIncomes(t *testing.T) {
	// Arrange
	db := setupIncomeTestDB(t)
	repo := NewIncomeRepository(db)
	ctx := context.Background()

	userID := "user-123"

	income1 := createTestIncome(userID, "Salary", 5000.00, "monthly")
	inactiveIncome := createTestIncomeWithActiveStatus(userID, "Old Job", 3000.00, "monthly", false)

	require.NoError(t, repo.SaveIncome(ctx, income1))
	require.NoError(t, repo.SaveIncome(ctx, inactiveIncome))

	// Act
	total, err := repo.CalculateUserTotalIncome(ctx, userID, false) // Include inactive

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 8000.00, total) // 5000 + 3000, including inactive
}