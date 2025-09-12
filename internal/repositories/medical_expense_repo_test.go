package repositories

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/DuckDHD/BuyOrBye/internal/domain"
	"github.com/DuckDHD/BuyOrBye/internal/models"
)

func setupMedicalExpenseTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(
		&models.HealthProfileModel{},
		&models.MedicalExpenseModel{},
	)
	require.NoError(t, err)

	// Create a test health profile
	profile := &models.HealthProfileModel{
		UserID:     "test-user-123",
		Age:        30,
		Gender:     "male",
		Height:     180.0,
		Weight:     75.0,
		FamilySize: 2,
	}
	db.Create(profile)

	return db
}

func TestMedicalExpenseRepository_AddExpense_CalculatesOutOfPocket(t *testing.T) {
	db := setupExpenseTestDB(t)
	repo := NewMedicalExpenseRepository(db)
	ctx := context.Background()

	expense := &domain.MedicalExpense{
		UserID:           "test-user-123",
		ProfileID:        "1",
		Amount:           200.0,
		Category:         "doctor_visit",
		Description:      "Annual physical checkup",
		IsRecurring:      false,
		Frequency:        "one_time",
		IsCovered:        true,
		InsurancePayment: 160.0,
		Date:             time.Now(),
	}

	result, err := repo.Create(ctx, expense)

	assert.NoError(t, err)
	assert.NotEmpty(t, result.ID)
	assert.Equal(t, "test-user-123", result.UserID)
	assert.Equal(t, 200.0, result.Amount)
	assert.Equal(t, "doctor_visit", result.Category)
	assert.Equal(t, "Annual physical checkup", result.Description)
	assert.False(t, result.IsRecurring)
	assert.Equal(t, "one_time", result.Frequency)
	assert.True(t, result.IsCovered)
	assert.Equal(t, 160.0, result.InsurancePayment)
	// Out of pocket should be calculated as Amount - InsurancePayment
	assert.Equal(t, 40.0, result.OutOfPocket)
}

func TestMedicalExpenseRepository_GetExpensesByDateRange(t *testing.T) {
	db := setupExpenseTestDB(t)
	repo := NewMedicalExpenseRepository(db)
	ctx := context.Background()

	now := time.Now()
	lastMonth := now.AddDate(0, -1, 0)
	twoMonthsAgo := now.AddDate(0, -2, 0)
	nextMonth := now.AddDate(0, 1, 0)

	expenses := []*domain.MedicalExpense{
		{
			UserID:      "test-user-123",
			ProfileID:   "1",
			Amount:      100.0,
			Category:    "medication",
			Description: "Current month medication",
			Date:        now,
		},
		{
			UserID:      "test-user-123",
			ProfileID:   "1",
			Amount:      150.0,
			Category:    "doctor_visit",
			Description: "Last month visit",
			Date:        lastMonth,
		},
		{
			UserID:      "test-user-123",
			ProfileID:   "1",
			Amount:      75.0,
			Category:    "lab_test",
			Description: "Two months ago test",
			Date:        twoMonthsAgo,
		},
		{
			UserID:      "test-user-123",
			ProfileID:   "1",
			Amount:      200.0,
			Category:    "therapy",
			Description: "Future appointment",
			Date:        nextMonth,
		},
	}

	for _, expense := range expenses {
		_, err := repo.Create(ctx, expense)
		require.NoError(t, err)
	}

	// Get expenses from last month to current month
	startDate := lastMonth.AddDate(0, 0, -5) // 5 days before last month
	endDate := now.AddDate(0, 0, 5)          // 5 days after current month

	result, err := repo.GetByDateRange(ctx, "test-user-123", startDate, endDate)

	assert.NoError(t, err)
	assert.Len(t, result, 2)

	// Results should be ordered by date (most recent first)
	descriptions := make([]string, len(result))
	for i, expense := range result {
		descriptions[i] = expense.Description
	}

	assert.Contains(t, descriptions, "Current month medication")
	assert.Contains(t, descriptions, "Last month visit")
	assert.NotContains(t, descriptions, "Two months ago test")
	assert.NotContains(t, descriptions, "Future appointment")
}

func TestMedicalExpenseRepository_GetRecurringExpenses_OnlyRecurring(t *testing.T) {
	db := setupExpenseTestDB(t)
	repo := NewMedicalExpenseRepository(db)
	ctx := context.Background()

	expenses := []*domain.MedicalExpense{
		{
			UserID:      "test-user-123",
			ProfileID:   "1",
			Amount:      120.0,
			Category:    "medication",
			Description: "Monthly insulin",
			IsRecurring: true,
			Frequency:   "monthly",
			Date:        time.Now(),
		},
		{
			UserID:      "test-user-123",
			ProfileID:   "1",
			Amount:      200.0,
			Category:    "therapy",
			Description: "Weekly therapy",
			IsRecurring: true,
			Frequency:   "weekly",
			Date:        time.Now(),
		},
		{
			UserID:      "test-user-123",
			ProfileID:   "1",
			Amount:      80.0,
			Category:    "doctor_visit",
			Description: "Emergency visit",
			IsRecurring: false,
			Frequency:   "one_time",
			Date:        time.Now(),
		},
		{
			UserID:      "test-user-123",
			ProfileID:   "1",
			Amount:      300.0,
			Category:    "equipment",
			Description: "Blood glucose monitor",
			IsRecurring: false,
			Frequency:   "one_time",
			Date:        time.Now(),
		},
	}

	for _, expense := range expenses {
		_, err := repo.Create(ctx, expense)
		require.NoError(t, err)
	}

	recurringExpenses, err := repo.GetRecurring(ctx, "test-user-123")

	assert.NoError(t, err)
	assert.Len(t, recurringExpenses, 2)

	descriptions := make([]string, len(recurringExpenses))
	for i, expense := range recurringExpenses {
		descriptions[i] = expense.Description
		assert.True(t, expense.IsRecurring)
	}

	assert.Contains(t, descriptions, "Monthly insulin")
	assert.Contains(t, descriptions, "Weekly therapy")
	assert.NotContains(t, descriptions, "Emergency visit")
	assert.NotContains(t, descriptions, "Blood glucose monitor")
}

func TestMedicalExpenseRepository_GetExpensesByCategory(t *testing.T) {
	db := setupExpenseTestDB(t)
	repo := NewMedicalExpenseRepository(db)
	ctx := context.Background()

	expenses := []*domain.MedicalExpense{
		{
			UserID:      "test-user-123",
			ProfileID:   "1",
			Amount:      120.0,
			Category:    "medication",
			Description: "Prescription drugs",
			Date:        time.Now(),
		},
		{
			UserID:      "test-user-123",
			ProfileID:   "1",
			Amount:      80.0,
			Category:    "medication",
			Description: "Over-the-counter medicine",
			Date:        time.Now(),
		},
		{
			UserID:      "test-user-123",
			ProfileID:   "1",
			Amount:      200.0,
			Category:    "doctor_visit",
			Description: "Specialist consultation",
			Date:        time.Now(),
		},
		{
			UserID:      "test-user-123",
			ProfileID:   "1",
			Amount:      150.0,
			Category:    "lab_test",
			Description: "Blood work",
			Date:        time.Now(),
		},
	}

	for _, expense := range expenses {
		_, err := repo.Create(ctx, expense)
		require.NoError(t, err)
	}

	// Get medication expenses
	medicationExpenses, err := repo.GetByCategory(ctx, "test-user-123", "medication")

	assert.NoError(t, err)
	assert.Len(t, medicationExpenses, 2)

	for _, expense := range medicationExpenses {
		assert.Equal(t, "medication", expense.Category)
	}

	// Get doctor visit expenses
	doctorExpenses, err := repo.GetByCategory(ctx, "test-user-123", "doctor_visit")

	assert.NoError(t, err)
	assert.Len(t, doctorExpenses, 1)
	assert.Equal(t, "Specialist consultation", doctorExpenses[0].Description)
}

func TestMedicalExpenseRepository_CalculateTotalExpenses(t *testing.T) {
	db := setupExpenseTestDB(t)
	repo := NewMedicalExpenseRepository(db)
	ctx := context.Background()

	expenses := []*domain.MedicalExpense{
		{
			UserID:           "test-user-123",
			ProfileID:        "1",
			Amount:           200.0,
			Category:         "doctor_visit",
			Description:      "Visit 1",
			InsurancePayment: 160.0,
			Date:             time.Now(),
		},
		{
			UserID:           "test-user-123",
			ProfileID:        "1",
			Amount:           150.0,
			Category:         "medication",
			Description:      "Medication 1",
			InsurancePayment: 120.0,
			Date:             time.Now(),
		},
		{
			UserID:           "test-user-123",
			ProfileID:        "1",
			Amount:           100.0,
			Category:         "lab_test",
			Description:      "Test 1",
			InsurancePayment: 80.0,
			Date:             time.Now(),
		},
	}

	for _, expense := range expenses {
		_, err := repo.Create(ctx, expense)
		require.NoError(t, err)
	}

	// Calculate totals for date range
	startDate := time.Now().AddDate(0, 0, -1) // Yesterday
	endDate := time.Now().AddDate(0, 0, 1)    // Tomorrow

	totals, err := repo.CalculateTotals(ctx, "test-user-123", startDate, endDate)

	assert.NoError(t, err)
	assert.Equal(t, 450.0, totals.TotalAmount)        // 200 + 150 + 100
	assert.Equal(t, 360.0, totals.TotalInsurancePaid) // 160 + 120 + 80
	assert.Equal(t, 90.0, totals.TotalOutOfPocket)    // 40 + 30 + 20
}

func TestMedicalExpenseRepository_GetExpensesByFrequency(t *testing.T) {
	db := setupExpenseTestDB(t)
	repo := NewMedicalExpenseRepository(db)
	ctx := context.Background()

	expenses := []*domain.MedicalExpense{
		{
			UserID:      "test-user-123",
			ProfileID:   "1",
			Amount:      100.0,
			Category:    "medication",
			Description: "Daily medication",
			IsRecurring: true,
			Frequency:   "daily",
			Date:        time.Now(),
		},
		{
			UserID:      "test-user-123",
			ProfileID:   "1",
			Amount:      200.0,
			Category:    "therapy",
			Description: "Monthly therapy",
			IsRecurring: true,
			Frequency:   "monthly",
			Date:        time.Now(),
		},
		{
			UserID:      "test-user-123",
			ProfileID:   "1",
			Amount:      300.0,
			Category:    "doctor_visit",
			Description: "Annual checkup",
			IsRecurring: true,
			Frequency:   "annually",
			Date:        time.Now(),
		},
	}

	for _, expense := range expenses {
		_, err := repo.Create(ctx, expense)
		require.NoError(t, err)
	}

	// Get monthly frequency expenses
	monthlyExpenses, err := repo.GetByFrequency(ctx, "test-user-123", "monthly")

	assert.NoError(t, err)
	assert.Len(t, monthlyExpenses, 1)
	assert.Equal(t, "Monthly therapy", monthlyExpenses[0].Description)
	assert.Equal(t, "monthly", monthlyExpenses[0].Frequency)

	// Get annual frequency expenses
	annualExpenses, err := repo.GetByFrequency(ctx, "test-user-123", "annually")

	assert.NoError(t, err)
	assert.Len(t, annualExpenses, 1)
	assert.Equal(t, "Annual checkup", annualExpenses[0].Description)
}
