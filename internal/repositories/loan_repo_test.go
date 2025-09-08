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

func setupLoanTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Create tables
	err = db.AutoMigrate(&models.LoanModel{})
	require.NoError(t, err)

	return db
}

func createTestLoan(userID, lender, loanType string, principal, remaining, payment, rate float64) domain.Loan {
	return domain.Loan{
		ID:               "loan-" + lender + "-123",
		UserID:           userID,
		Lender:           lender,
		Type:             loanType,
		PrincipalAmount:  principal,
		RemainingBalance: remaining,
		MonthlyPayment:   payment,
		InterestRate:     rate,
		EndDate:          time.Date(2050, 12, 31, 0, 0, 0, 0, time.UTC),
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
}

func TestLoanRepository_SaveLoan_Success_CreatesNewRecord(t *testing.T) {
	// Arrange
	db := setupLoanTestDB(t)
	repo := NewLoanRepository(db)
	ctx := context.Background()

	loan := createTestLoan("user-123", "Chase Bank", "mortgage", 250000.00, 240000.00, 1266.71, 4.5)

	// Act
	err := repo.SaveLoan(ctx, loan)

	// Assert
	assert.NoError(t, err)

	// Verify in database
	var saved models.LoanModel
	result := db.First(&saved, "id = ?", loan.ID)
	assert.NoError(t, result.Error)
	assert.Equal(t, loan.UserID, saved.UserID)
	assert.Equal(t, loan.Lender, saved.Lender)
	assert.Equal(t, loan.Type, saved.Type)
	assert.Equal(t, loan.PrincipalAmount, saved.PrincipalAmount)
	assert.Equal(t, loan.RemainingBalance, saved.RemainingBalance)
	assert.Equal(t, loan.MonthlyPayment, saved.MonthlyPayment)
	assert.Equal(t, loan.InterestRate, saved.InterestRate)
}

func TestLoanRepository_SaveLoan_ExistingRecord_UpdatesRecord(t *testing.T) {
	// Arrange
	db := setupLoanTestDB(t)
	repo := NewLoanRepository(db)
	ctx := context.Background()

	loan := createTestLoan("user-123", "Chase", "mortgage", 250000.00, 240000.00, 1266.71, 4.5)

	// Save initial record
	err := repo.SaveLoan(ctx, loan)
	require.NoError(t, err)

	// Modify and save again
	loan.RemainingBalance = 235000.00
	loan.Lender = "Updated Chase"
	loan.UpdatedAt = time.Now().Add(time.Hour)

	// Act
	err = repo.SaveLoan(ctx, loan)

	// Assert
	assert.NoError(t, err)

	// Verify update
	var saved models.LoanModel
	result := db.First(&saved, "id = ?", loan.ID)
	assert.NoError(t, result.Error)
	assert.Equal(t, 235000.00, saved.RemainingBalance)
	assert.Equal(t, "Updated Chase", saved.Lender)
}

func TestLoanRepository_GetUserLoans_ReturnsActive_ForValidUser(t *testing.T) {
	// Arrange
	db := setupLoanTestDB(t)
	repo := NewLoanRepository(db)
	ctx := context.Background()

	userID := "user-123"
	otherUserID := "user-456"

	// Create test loans
	loan1 := createTestLoan(userID, "Chase", "mortgage", 250000.00, 240000.00, 1266.71, 4.5)
	loan2 := createTestLoan(userID, "Ford Credit", "auto", 30000.00, 25000.00, 550.00, 6.0)
	loan3 := createTestLoan(otherUserID, "Wells Fargo", "mortgage", 300000.00, 280000.00, 1500.00, 5.0)

	// Save test data
	require.NoError(t, repo.SaveLoan(ctx, loan1))
	require.NoError(t, repo.SaveLoan(ctx, loan2))
	require.NoError(t, repo.SaveLoan(ctx, loan3))

	// Act
	loans, err := repo.GetUserLoans(ctx, userID)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, loans, 2)

	// Verify only user's loans are returned
	lenders := make([]string, len(loans))
	for i, loan := range loans {
		assert.Equal(t, userID, loan.UserID)
		lenders[i] = loan.Lender
	}
	assert.Contains(t, lenders, "Chase")
	assert.Contains(t, lenders, "Ford Credit")
	assert.NotContains(t, lenders, "Wells Fargo")
}

func TestLoanRepository_GetUserLoans_EmptyUser_ReturnsEmpty(t *testing.T) {
	// Arrange
	db := setupLoanTestDB(t)
	repo := NewLoanRepository(db)
	ctx := context.Background()

	// Act
	loans, err := repo.GetUserLoans(ctx, "nonexistent-user")

	// Assert
	assert.NoError(t, err)
	assert.Empty(t, loans)
}

func TestLoanRepository_UpdateLoanBalance_Success_UpdatesBalance(t *testing.T) {
	// Arrange
	db := setupLoanTestDB(t)
	repo := NewLoanRepository(db)
	ctx := context.Background()

	loan := createTestLoan("user-123", "Chase", "mortgage", 250000.00, 240000.00, 1266.71, 4.5)
	require.NoError(t, repo.SaveLoan(ctx, loan))

	newBalance := 230000.00

	// Act
	err := repo.UpdateLoanBalance(ctx, loan.ID, newBalance)

	// Assert
	assert.NoError(t, err)

	// Verify balance update
	var updated models.LoanModel
	result := db.First(&updated, "id = ?", loan.ID)
	assert.NoError(t, result.Error)
	assert.Equal(t, newBalance, updated.RemainingBalance)
	
	// Verify other fields unchanged
	assert.Equal(t, loan.PrincipalAmount, updated.PrincipalAmount)
	assert.Equal(t, loan.MonthlyPayment, updated.MonthlyPayment)
	assert.Equal(t, loan.InterestRate, updated.InterestRate)
}

func TestLoanRepository_UpdateLoanBalance_NonexistentLoan_ReturnsError(t *testing.T) {
	// Arrange
	db := setupLoanTestDB(t)
	repo := NewLoanRepository(db)
	ctx := context.Background()

	// Act
	err := repo.UpdateLoanBalance(ctx, "nonexistent-id", 100000.00)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestLoanRepository_GetLoansByType_Filters_ReturnsCorrectType(t *testing.T) {
	// Arrange
	db := setupLoanTestDB(t)
	repo := NewLoanRepository(db)
	ctx := context.Background()

	userID := "user-123"

	// Create loans of different types
	mortgageLoan1 := createTestLoan(userID, "Chase", "mortgage", 250000.00, 240000.00, 1266.71, 4.5)
	mortgageLoan2 := createTestLoan(userID, "Wells Fargo", "mortgage", 300000.00, 280000.00, 1500.00, 5.0)
	autoLoan := createTestLoan(userID, "Ford Credit", "auto", 30000.00, 25000.00, 550.00, 6.0)
	personalLoan := createTestLoan(userID, "Bank of America", "personal", 10000.00, 8000.00, 200.00, 12.0)

	require.NoError(t, repo.SaveLoan(ctx, mortgageLoan1))
	require.NoError(t, repo.SaveLoan(ctx, mortgageLoan2))
	require.NoError(t, repo.SaveLoan(ctx, autoLoan))
	require.NoError(t, repo.SaveLoan(ctx, personalLoan))

	// Act
	mortgageLoans, err := repo.GetLoansByType(ctx, userID, "mortgage")

	// Assert
	assert.NoError(t, err)
	assert.Len(t, mortgageLoans, 2)

	for _, loan := range mortgageLoans {
		assert.Equal(t, "mortgage", loan.Type)
		assert.Equal(t, userID, loan.UserID)
	}

	lenders := []string{mortgageLoans[0].Lender, mortgageLoans[1].Lender}
	assert.Contains(t, lenders, "Chase")
	assert.Contains(t, lenders, "Wells Fargo")
}

func TestLoanRepository_GetLoansByType_NoMatches_ReturnsEmpty(t *testing.T) {
	// Arrange
	db := setupLoanTestDB(t)
	repo := NewLoanRepository(db)
	ctx := context.Background()

	userID := "user-123"

	autoLoan := createTestLoan(userID, "Ford Credit", "auto", 30000.00, 25000.00, 550.00, 6.0)
	require.NoError(t, repo.SaveLoan(ctx, autoLoan))

	// Act
	loans, err := repo.GetLoansByType(ctx, userID, "student")

	// Assert
	assert.NoError(t, err)
	assert.Empty(t, loans)
}

func TestLoanRepository_UpdateLoan_Success_UpdatesExistingRecord(t *testing.T) {
	// Arrange
	db := setupLoanTestDB(t)
	repo := NewLoanRepository(db)
	ctx := context.Background()

	loan := createTestLoan("user-123", "Chase", "mortgage", 250000.00, 240000.00, 1266.71, 4.5)
	require.NoError(t, repo.SaveLoan(ctx, loan))

	// Modify loan
	loan.RemainingBalance = 230000.00
	loan.Lender = "Updated Chase Bank"
	loan.MonthlyPayment = 1300.00
	loan.InterestRate = 4.0

	// Act
	err := repo.UpdateLoan(ctx, loan)

	// Assert
	assert.NoError(t, err)

	// Verify update
	var updated models.LoanModel
	result := db.First(&updated, "id = ?", loan.ID)
	assert.NoError(t, result.Error)
	assert.Equal(t, 230000.00, updated.RemainingBalance)
	assert.Equal(t, "Updated Chase Bank", updated.Lender)
	assert.Equal(t, 1300.00, updated.MonthlyPayment)
	assert.Equal(t, 4.0, updated.InterestRate)
}

func TestLoanRepository_UpdateLoan_NonexistentRecord_ReturnsError(t *testing.T) {
	// Arrange
	db := setupLoanTestDB(t)
	repo := NewLoanRepository(db)
	ctx := context.Background()

	loan := createTestLoan("user-123", "Nonexistent", "mortgage", 250000.00, 240000.00, 1266.71, 4.5)

	// Act
	err := repo.UpdateLoan(ctx, loan)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestLoanRepository_DeleteLoan_SoftDelete_MarksAsDeleted(t *testing.T) {
	// Arrange
	db := setupLoanTestDB(t)
	repo := NewLoanRepository(db)
	ctx := context.Background()

	loan := createTestLoan("user-123", "Chase", "mortgage", 250000.00, 240000.00, 1266.71, 4.5)
	require.NoError(t, repo.SaveLoan(ctx, loan))

	// Act
	err := repo.DeleteLoan(ctx, loan.ID)

	// Assert
	assert.NoError(t, err)

	// Verify soft delete - record should not be found in normal queries
	loans, err := repo.GetUserLoans(ctx, loan.UserID)
	assert.NoError(t, err)
	assert.Empty(t, loans)

	// Verify record still exists but with deleted_at set
	var deleted models.LoanModel
	result := db.Unscoped().First(&deleted, "id = ?", loan.ID)
	assert.NoError(t, result.Error)
	assert.NotNil(t, deleted.DeletedAt)
}

func TestLoanRepository_DeleteLoan_NonexistentRecord_ReturnsError(t *testing.T) {
	// Arrange
	db := setupLoanTestDB(t)
	repo := NewLoanRepository(db)
	ctx := context.Background()

	// Act
	err := repo.DeleteLoan(ctx, "nonexistent-id")

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestLoanRepository_GetLoanByID_Success_ReturnsCorrectLoan(t *testing.T) {
	// Arrange
	db := setupLoanTestDB(t)
	repo := NewLoanRepository(db)
	ctx := context.Background()

	loan := createTestLoan("user-123", "Chase", "mortgage", 250000.00, 240000.00, 1266.71, 4.5)
	require.NoError(t, repo.SaveLoan(ctx, loan))

	// Act
	found, err := repo.GetLoanByID(ctx, loan.ID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, loan.ID, found.ID)
	assert.Equal(t, loan.UserID, found.UserID)
	assert.Equal(t, loan.Lender, found.Lender)
	assert.Equal(t, loan.Type, found.Type)
	assert.Equal(t, loan.PrincipalAmount, found.PrincipalAmount)
}

func TestLoanRepository_GetLoanByID_NotFound_ReturnsError(t *testing.T) {
	// Arrange
	db := setupLoanTestDB(t)
	repo := NewLoanRepository(db)
	ctx := context.Background()

	// Act
	_, err := repo.GetLoanByID(ctx, "nonexistent-id")

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestLoanRepository_GetLoansByInterestRateRange_FiltersCorrectly(t *testing.T) {
	// Arrange
	db := setupLoanTestDB(t)
	repo := NewLoanRepository(db)
	ctx := context.Background()

	userID := "user-123"

	// Create loans with different interest rates
	lowRateLoan := createTestLoan(userID, "Chase", "mortgage", 250000.00, 240000.00, 1266.71, 3.5)
	mediumRateLoan := createTestLoan(userID, "Ford Credit", "auto", 30000.00, 25000.00, 550.00, 6.0)
	highRateLoan := createTestLoan(userID, "Bank of America", "personal", 10000.00, 8000.00, 200.00, 15.0)

	require.NoError(t, repo.SaveLoan(ctx, lowRateLoan))
	require.NoError(t, repo.SaveLoan(ctx, mediumRateLoan))
	require.NoError(t, repo.SaveLoan(ctx, highRateLoan))

	// Act - get loans with interest rate between 5% and 10%
	loans, err := repo.GetLoansByInterestRateRange(ctx, userID, 5.0, 10.0)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, loans, 1)
	assert.Equal(t, "Ford Credit", loans[0].Lender)
	assert.Equal(t, 6.0, loans[0].InterestRate)
}

func TestLoanRepository_CalculateUserTotalDebt_ReturnsCorrectSum(t *testing.T) {
	// Arrange
	db := setupLoanTestDB(t)
	repo := NewLoanRepository(db)
	ctx := context.Background()

	userID := "user-123"
	otherUserID := "user-456"

	// Create test loans for user
	loan1 := createTestLoan(userID, "Chase", "mortgage", 250000.00, 240000.00, 1266.71, 4.5)
	loan2 := createTestLoan(userID, "Ford Credit", "auto", 30000.00, 25000.00, 550.00, 6.0)
	
	// Create loan for other user
	otherLoan := createTestLoan(otherUserID, "Wells Fargo", "mortgage", 300000.00, 280000.00, 1500.00, 5.0)

	require.NoError(t, repo.SaveLoan(ctx, loan1))
	require.NoError(t, repo.SaveLoan(ctx, loan2))
	require.NoError(t, repo.SaveLoan(ctx, otherLoan))

	// Act
	totalDebt, err := repo.CalculateUserTotalDebt(ctx, userID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 265000.00, totalDebt) // 240000 + 25000, excluding other user's loan
}

func TestLoanRepository_CalculateUserMonthlyPayments_ReturnsCorrectSum(t *testing.T) {
	// Arrange
	db := setupLoanTestDB(t)
	repo := NewLoanRepository(db)
	ctx := context.Background()

	userID := "user-123"

	// Create test loans
	loan1 := createTestLoan(userID, "Chase", "mortgage", 250000.00, 240000.00, 1266.71, 4.5)
	loan2 := createTestLoan(userID, "Ford Credit", "auto", 30000.00, 25000.00, 550.00, 6.0)

	require.NoError(t, repo.SaveLoan(ctx, loan1))
	require.NoError(t, repo.SaveLoan(ctx, loan2))

	// Act
	totalPayments, err := repo.CalculateUserMonthlyPayments(ctx, userID)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 1816.71, totalPayments) // 1266.71 + 550.00
}

func TestLoanRepository_GetNearPayoffLoans_ReturnsLoansNearCompletion(t *testing.T) {
	// Arrange
	db := setupLoanTestDB(t)
	repo := NewLoanRepository(db)
	ctx := context.Background()

	userID := "user-123"
	
	// Create loans with different remaining balances
	nearPayoffLoan := createTestLoan(userID, "Chase", "auto", 25000.00, 1000.00, 500.00, 6.0) // Low remaining balance
	regularLoan := createTestLoan(userID, "Wells Fargo", "mortgage", 250000.00, 200000.00, 1200.00, 4.5) // High remaining balance

	require.NoError(t, repo.SaveLoan(ctx, nearPayoffLoan))
	require.NoError(t, repo.SaveLoan(ctx, regularLoan))

	// Act - get loans with remaining balance less than 5% of principal
	nearPayoffLoans, err := repo.GetNearPayoffLoans(ctx, userID, 0.05) // 5% threshold

	// Assert
	assert.NoError(t, err)
	assert.Len(t, nearPayoffLoans, 1)
	assert.Equal(t, "Chase", nearPayoffLoans[0].Lender)
	
	// Verify the loan is indeed near payoff (1000/25000 = 4% < 5%)
	payoffPercentage := nearPayoffLoans[0].RemainingBalance / nearPayoffLoans[0].PrincipalAmount
	assert.True(t, payoffPercentage < 0.05)
}