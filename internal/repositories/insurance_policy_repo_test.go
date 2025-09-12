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

func setupPolicyTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(
		&models.HealthProfileModel{},
		&models.InsurancePolicyModel{},
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

func TestInsurancePolicyRepository_AddPolicy_UniqueNumber(t *testing.T) {
	db := setupPolicyTestDB(t)
	repo := NewInsurancePolicyRepository(db)
	ctx := context.Background()

	policy1 := &domain.InsurancePolicy{
		UserID:             "test-user-123",
		ProfileID:          "1",
		Provider:           "HealthCorp",
		PolicyNumber:       "HC-12345",
		Type:               "health",
		MonthlyPremium:     250.0,
		Deductible:         1000.0,
		DeductibleMet:      200.0,
		OutOfPocketMax:     5000.0,
		OutOfPocketCurrent: 500.0,
		CoveragePercentage: 80.0,
		StartDate:          time.Now().AddDate(0, -6, 0),
		EndDate:            time.Now().AddDate(1, 0, 0),
		IsActive:           true,
	}

	policy2 := &domain.InsurancePolicy{
		UserID:             "test-user-456",
		ProfileID:          "1",
		Provider:           "DifferentCorp",
		PolicyNumber:       "HC-12345", // Same policy number
		Type:               "dental",
		MonthlyPremium:     100.0,
		Deductible:         500.0,
		OutOfPocketMax:     2000.0,
		CoveragePercentage: 70.0,
		StartDate:          time.Now(),
		EndDate:            time.Now().AddDate(1, 0, 0),
		IsActive:           true,
	}

	// First policy should succeed
	result1, err := repo.Create(ctx, policy1)
	assert.NoError(t, err)
	assert.NotEmpty(t, result1.ID)
	assert.Equal(t, "HC-12345", result1.PolicyNumber)

	// Second policy with same number should fail
	_, err = repo.Create(ctx, policy2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unique constraint")
}

func TestInsurancePolicyRepository_GetActivePolicies_FiltersByDate(t *testing.T) {
	db := setupPolicyTestDB(t)
	repo := NewInsurancePolicyRepository(db)
	ctx := context.Background()

	now := time.Now()
	pastDate := now.AddDate(-1, 0, 0)
	futureDate := now.AddDate(1, 0, 0)

	policies := []*domain.InsurancePolicy{
		{
			UserID:             "test-user-123",
			ProfileID:          "1",
			Provider:           "ActiveCorp",
			PolicyNumber:       "ACTIVE-001",
			Type:               "health",
			MonthlyPremium:     250.0,
			Deductible:         1000.0,
			OutOfPocketMax:     5000.0,
			CoveragePercentage: 80.0,
			StartDate:          pastDate,
			EndDate:            futureDate,
			IsActive:           true,
		},
		{
			UserID:             "test-user-123",
			ProfileID:          "1",
			Provider:           "ExpiredCorp",
			PolicyNumber:       "EXPIRED-001",
			Type:               "dental",
			MonthlyPremium:     100.0,
			Deductible:         500.0,
			OutOfPocketMax:     2000.0,
			CoveragePercentage: 70.0,
			StartDate:          now.AddDate(-2, 0, 0),
			EndDate:            now.AddDate(-1, 0, 0), // Expired
			IsActive:           true,
		},
		{
			UserID:             "test-user-123",
			ProfileID:          "1",
			Provider:           "FutureCorp",
			PolicyNumber:       "FUTURE-001",
			Type:               "vision",
			MonthlyPremium:     50.0,
			Deductible:         100.0,
			OutOfPocketMax:     1000.0,
			CoveragePercentage: 90.0,
			StartDate:          futureDate.AddDate(0, 1, 0), // Starts in future
			EndDate:            futureDate.AddDate(1, 0, 0),
			IsActive:           true,
		},
		{
			UserID:             "test-user-123",
			ProfileID:          "1",
			Provider:           "InactiveCorp",
			PolicyNumber:       "INACTIVE-001",
			Type:               "health",
			MonthlyPremium:     200.0,
			Deductible:         800.0,
			OutOfPocketMax:     4000.0,
			CoveragePercentage: 75.0,
			StartDate:          pastDate,
			EndDate:            futureDate,
			IsActive:           false, // Manually deactivated
		},
	}

	for _, policy := range policies {
		_, err := repo.Create(ctx, policy)
		require.NoError(t, err)
	}

	// Get currently active policies (active flag + current date range)
	activePolicies, err := repo.GetActivePolicies(ctx, "test-user-123")

	assert.NoError(t, err)
	assert.Len(t, activePolicies, 1)
	assert.Equal(t, "ACTIVE-001", activePolicies[0].PolicyNumber)
	assert.Equal(t, "ActiveCorp", activePolicies[0].Provider)
}

func TestInsurancePolicyRepository_UpdateDeductibleProgress(t *testing.T) {
	db := setupPolicyTestDB(t)
	repo := NewInsurancePolicyRepository(db)
	ctx := context.Background()

	policy := &domain.InsurancePolicy{
		UserID:             "test-user-123",
		ProfileID:          "1",
		Provider:           "HealthCorp",
		PolicyNumber:       "HC-12345",
		Type:               "health",
		MonthlyPremium:     250.0,
		Deductible:         1000.0,
		DeductibleMet:      200.0,
		OutOfPocketMax:     5000.0,
		OutOfPocketCurrent: 500.0,
		CoveragePercentage: 80.0,
		StartDate:          time.Now().AddDate(0, -6, 0),
		EndDate:            time.Now().AddDate(1, 0, 0),
		IsActive:           true,
	}

	createdPolicy, err := repo.Create(ctx, policy)
	require.NoError(t, err)

	// Update deductible progress after a new expense
	expenseAmount := 300.0
	insurancePaid := 200.0
	outOfPocketPaid := 100.0

	createdPolicy.DeductibleMet = 500.0      // 200 + 300
	createdPolicy.OutOfPocketCurrent = 600.0 // 500 + 100

	updatedPolicy, err := repo.UpdateDeductibleProgress(ctx, createdPolicy.ID, createdPolicy.DeductibleMet, createdPolicy.OutOfPocketCurrent)

	assert.NoError(t, err)
	assert.Equal(t, 500.0, updatedPolicy.DeductibleMet)
	assert.Equal(t, 600.0, updatedPolicy.OutOfPocketCurrent)
	assert.True(t, updatedPolicy.UpdatedAt.After(createdPolicy.UpdatedAt))

	// Verify values don't exceed maximums
	createdPolicy.DeductibleMet = 1200.0     // Exceeds deductible of 1000
	createdPolicy.OutOfPocketCurrent = 6000.0 // Exceeds max of 5000

	updatedPolicy, err = repo.UpdateDeductibleProgress(ctx, createdPolicy.ID, createdPolicy.DeductibleMet, createdPolicy.OutOfPocketCurrent)

	assert.NoError(t, err)
	assert.Equal(t, 1000.0, updatedPolicy.DeductibleMet)     // Capped at deductible
	assert.Equal(t, 5000.0, updatedPolicy.OutOfPocketCurrent) // Capped at max
}

func TestInsurancePolicyRepository_GetPolicyByType(t *testing.T) {
	db := setupPolicyTestDB(t)
	repo := NewInsurancePolicyRepository(db)
	ctx := context.Background()

	policies := []*domain.InsurancePolicy{
		{
			UserID:             "test-user-123",
			ProfileID:          "1",
			Provider:           "HealthCorp",
			PolicyNumber:       "HEALTH-001",
			Type:               "health",
			MonthlyPremium:     250.0,
			Deductible:         1000.0,
			OutOfPocketMax:     5000.0,
			CoveragePercentage: 80.0,
			StartDate:          time.Now(),
			EndDate:            time.Now().AddDate(1, 0, 0),
			IsActive:           true,
		},
		{
			UserID:             "test-user-123",
			ProfileID:          "1",
			Provider:           "DentalCorp",
			PolicyNumber:       "DENTAL-001",
			Type:               "dental",
			MonthlyPremium:     100.0,
			Deductible:         500.0,
			OutOfPocketMax:     2000.0,
			CoveragePercentage: 70.0,
			StartDate:          time.Now(),
			EndDate:            time.Now().AddDate(1, 0, 0),
			IsActive:           true,
		},
		{
			UserID:             "test-user-123",
			ProfileID:          "1",
			Provider:           "VisionCorp",
			PolicyNumber:       "VISION-001",
			Type:               "vision",
			MonthlyPremium:     50.0,
			Deductible:         100.0,
			OutOfPocketMax:     1000.0,
			CoveragePercentage: 90.0,
			StartDate:          time.Now(),
			EndDate:            time.Now().AddDate(1, 0, 0),
			IsActive:           true,
		},
	}

	for _, policy := range policies {
		_, err := repo.Create(ctx, policy)
		require.NoError(t, err)
	}

	// Get health policies
	healthPolicies, err := repo.GetByType(ctx, "test-user-123", "health")

	assert.NoError(t, err)
	assert.Len(t, healthPolicies, 1)
	assert.Equal(t, "health", healthPolicies[0].Type)
	assert.Equal(t, "HealthCorp", healthPolicies[0].Provider)

	// Get dental policies
	dentalPolicies, err := repo.GetByType(ctx, "test-user-123", "dental")

	assert.NoError(t, err)
	assert.Len(t, dentalPolicies, 1)
	assert.Equal(t, "dental", dentalPolicies[0].Type)
	assert.Equal(t, "DentalCorp", dentalPolicies[0].Provider)

	// Get vision policies
	visionPolicies, err := repo.GetByType(ctx, "test-user-123", "vision")

	assert.NoError(t, err)
	assert.Len(t, visionPolicies, 1)
	assert.Equal(t, "vision", visionPolicies[0].Type)
	assert.Equal(t, "VisionCorp", visionPolicies[0].Provider)
}

func TestInsurancePolicyRepository_GetPolicyByNumber(t *testing.T) {
	db := setupPolicyTestDB(t)
	repo := NewInsurancePolicyRepository(db)
	ctx := context.Background()

	policy := &domain.InsurancePolicy{
		UserID:             "test-user-123",
		ProfileID:          "1",
		Provider:           "HealthCorp",
		PolicyNumber:       "HC-12345",
		Type:               "health",
		MonthlyPremium:     250.0,
		Deductible:         1000.0,
		OutOfPocketMax:     5000.0,
		CoveragePercentage: 80.0,
		StartDate:          time.Now(),
		EndDate:            time.Now().AddDate(1, 0, 0),
		IsActive:           true,
	}

	_, err := repo.Create(ctx, policy)
	require.NoError(t, err)

	// Find policy by number
	foundPolicy, err := repo.GetByPolicyNumber(ctx, "HC-12345")

	assert.NoError(t, err)
	assert.Equal(t, "HC-12345", foundPolicy.PolicyNumber)
	assert.Equal(t, "HealthCorp", foundPolicy.Provider)
	assert.Equal(t, "test-user-123", foundPolicy.UserID)

	// Search for non-existent policy
	_, err = repo.GetByPolicyNumber(ctx, "NON-EXISTENT")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestInsurancePolicyRepository_CalculateCoverageForExpense(t *testing.T) {
	db := setupPolicyTestDB(t)
	repo := NewInsurancePolicyRepository(db)
	ctx := context.Background()

	policy := &domain.InsurancePolicy{
		UserID:             "test-user-123",
		ProfileID:          "1",
		Provider:           "HealthCorp",
		PolicyNumber:       "HC-12345",
		Type:               "health",
		MonthlyPremium:     250.0,
		Deductible:         1000.0,
		DeductibleMet:      200.0,
		OutOfPocketMax:     5000.0,
		OutOfPocketCurrent: 500.0,
		CoveragePercentage: 80.0,
		StartDate:          time.Now().AddDate(0, -6, 0),
		EndDate:            time.Now().AddDate(1, 0, 0),
		IsActive:           true,
	}

	createdPolicy, err := repo.Create(ctx, policy)
	require.NoError(t, err)

	// Test coverage calculation for a $500 expense
	expenseAmount := 500.0
	coverage, err := repo.CalculateCoverageForExpense(ctx, createdPolicy.ID, expenseAmount)

	assert.NoError(t, err)
	assert.NotNil(t, coverage)
	
	// With deductible remaining (800), and 80% coverage:
	// Remaining deductible: 1000 - 200 = 800
	// Patient pays deductible first: min(500, 800) = 500
	// No amount left for insurance coverage
	// So: Insurance pays 0, Patient pays 500
	assert.Equal(t, 0.0, coverage.InsurancePays)
	assert.Equal(t, 500.0, coverage.PatientPays)
	assert.Equal(t, 700.0, coverage.NewDeductibleMet) // 200 + 500
}

// Define the CoverageCalculation struct for the test
type CoverageCalculation struct {
	InsurancePays      float64
	PatientPays        float64
	NewDeductibleMet   float64
	NewOutOfPocketUsed float64
}