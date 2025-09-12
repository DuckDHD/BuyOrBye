package repositories

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/DuckDHD/BuyOrBye/internal/domain"
	"github.com/DuckDHD/BuyOrBye/internal/models"
)

func setupHealthProfileTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate health models
	err = db.AutoMigrate(
		&models.HealthProfileModel{},
		&models.MedicalConditionModel{},
		&models.MedicalExpenseModel{},
		&models.InsurancePolicyModel{},
	)
	require.NoError(t, err)

	return db
}

func TestHealthProfileRepository_CreateProfile_Success(t *testing.T) {
	db := setupTestDB(t)
	repo := NewHealthProfileRepository(db)
	ctx := context.Background()

	profile := &domain.HealthProfile{
		UserID:     "test-user-123",
		Age:        30,
		Gender:     "male",
		Height:     180.0,
		Weight:     75.0,
		FamilySize: 2,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	result, err := repo.Create(ctx, profile)

	assert.NoError(t, err)
	assert.NotEmpty(t, result.ID)
	assert.Equal(t, "test-user-123", result.UserID)
	assert.Equal(t, 30, result.Age)
	assert.Equal(t, "male", result.Gender)
	assert.Equal(t, 180.0, result.Height)
	assert.Equal(t, 75.0, result.Weight)
	assert.Equal(t, 2, result.FamilySize)
}

func TestHealthProfileRepository_CreateProfile_DuplicateUserFails(t *testing.T) {
	db := setupTestDB(t)
	repo := NewHealthProfileRepository(db)
	ctx := context.Background()

	profile1 := &domain.HealthProfile{
		UserID:     "test-user-123",
		Age:        30,
		Gender:     "male",
		Height:     180.0,
		Weight:     75.0,
		FamilySize: 2,
	}

	profile2 := &domain.HealthProfile{
		UserID:     "test-user-123", // Same user ID
		Age:        25,
		Gender:     "female",
		Height:     165.0,
		Weight:     60.0,
		FamilySize: 1,
	}

	// First creation should succeed
	_, err := repo.Create(ctx, profile1)
	assert.NoError(t, err)

	// Second creation with same user ID should fail
	_, err = repo.Create(ctx, profile2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unique constraint")
}

func TestHealthProfileRepository_GetProfile_WithRelations(t *testing.T) {
	db := setupTestDB(t)
	repo := NewHealthProfileRepository(db)
	ctx := context.Background()

	// Create health profile
	profile := &domain.HealthProfile{
		UserID:     "test-user-123",
		Age:        30,
		Gender:     "male",
		Height:     180.0,
		Weight:     75.0,
		FamilySize: 2,
	}

	createdProfile, err := repo.Create(ctx, profile)
	require.NoError(t, err)

	// Convert profile ID from string to uint
	profileIDUint, err := strconv.ParseUint(createdProfile.ID, 10, 32)
	require.NoError(t, err)

	// Add related medical condition
	conditionModel := &models.MedicalConditionModel{
		UserID:             "test-user-123",
		ProfileID:          uint(profileIDUint),
		Name:               "Diabetes",
		Category:           "chronic",
		Severity:           "moderate",
		DiagnosedDate:      time.Now(),
		IsActive:           true,
		RequiresMedication: true,
		MonthlyMedCost:     100.0,
		RiskFactor:         0.3,
	}
	db.Create(conditionModel)

	// Add related medical expense
	expenseModel := &models.MedicalExpenseModel{
		UserID:           "test-user-123",
		ProfileID:        uint(profileIDUint),
		Amount:           150.0,
		Category:         "medication",
		Description:      "Monthly insulin",
		IsRecurring:      true,
		Frequency:        "monthly",
		IsCovered:        true,
		InsurancePayment: 120.0,
		OutOfPocket:      30.0,
		Date:             time.Now(),
	}
	db.Create(expenseModel)

	// Add related insurance policy
	policyModel := &models.InsurancePolicyModel{
		UserID:             "test-user-123",
		ProfileID:          uint(profileIDUint),
		Provider:           "Health Corp",
		PolicyNumber:       "HC-12345",
		Type:               "health",
		MonthlyPremium:     250.0,
		AnnualDeductible:   1000.0,
		OutOfPocketMax:     5000.0,
		CoveragePercentage: 80,
		StartDate:          time.Now().AddDate(0, -6, 0),
		EndDate:            time.Now().AddDate(1, 0, 0),
		IsActive:           true,
		DeductibleMet:      200.0,
		OutOfPocketCurrent: 500.0,
	}
	db.Create(policyModel)

	// Get profile with relations
	result, err := repo.GetByUserID(ctx, "test-user-123")

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "test-user-123", result.UserID)

	// Verify relations are loaded (this will be checked after implementation)
	// The specific assertion will depend on how we structure the domain model
	// to include related entities
}

func TestHealthProfileRepository_UpdateProfile_Success(t *testing.T) {
	db := setupTestDB(t)
	repo := NewHealthProfileRepository(db)
	ctx := context.Background()

	// Create initial profile
	profile := &domain.HealthProfile{
		UserID:     "test-user-123",
		Age:        30,
		Gender:     "male",
		Height:     180.0,
		Weight:     75.0,
		FamilySize: 2,
	}

	createdProfile, err := repo.Create(ctx, profile)
	require.NoError(t, err)

	// Update profile
	createdProfile.Age = 31
	createdProfile.Weight = 77.0
	createdProfile.FamilySize = 3

	updatedProfile, err := repo.Update(ctx, createdProfile)

	assert.NoError(t, err)
	assert.Equal(t, 31, updatedProfile.Age)
	assert.Equal(t, 77.0, updatedProfile.Weight)
	assert.Equal(t, 3, updatedProfile.FamilySize)
	assert.True(t, updatedProfile.UpdatedAt.After(updatedProfile.CreatedAt))
}

func TestHealthProfileRepository_DeleteProfile_CascadeDeletes(t *testing.T) {
	db := setupTestDB(t)
	repo := NewHealthProfileRepository(db)
	ctx := context.Background()

	// Create health profile
	profile := &domain.HealthProfile{
		UserID:     "test-user-123",
		Age:        30,
		Gender:     "male",
		Height:     180.0,
		Weight:     75.0,
		FamilySize: 2,
	}

	createdProfile, err := repo.Create(ctx, profile)
	require.NoError(t, err)

	// Convert profile ID from string to uint
	profileIDUint, err := strconv.ParseUint(createdProfile.ID, 10, 32)
	require.NoError(t, err)

	// Add related records
	conditionModel := &models.MedicalConditionModel{
		UserID:        "test-user-123",
		ProfileID:     uint(profileIDUint),
		Name:          "Test Condition",
		Category:      "chronic",
		Severity:      "mild",
		DiagnosedDate: time.Now(),
		IsActive:      true,
	}
	db.Create(conditionModel)

	expenseModel := &models.MedicalExpenseModel{
		UserID:      "test-user-123",
		ProfileID:   uint(profileIDUint),
		Amount:      100.0,
		Category:    "medication",
		Description: "Test expense",
		Date:        time.Now(),
	}
	db.Create(expenseModel)

	policyModel := &models.InsurancePolicyModel{
		UserID:             "test-user-123",
		ProfileID:          uint(profileIDUint),
		Provider:           "Test Insurance",
		PolicyNumber:       "TEST-123",
		Type:               "health",
		MonthlyPremium:     200.0,
		AnnualDeductible:   1000.0,
		OutOfPocketMax:     5000.0,
		CoveragePercentage: 80,
		StartDate:          time.Now(),
		EndDate:            time.Now().AddDate(1, 0, 0),
		IsActive:           true,
	}
	db.Create(policyModel)

	// Delete profile
	err = repo.Delete(ctx, uint(profileIDUint))
	assert.NoError(t, err)

	// Verify profile is deleted
	_, err = repo.GetByID(ctx, uint(profileIDUint))
	assert.Error(t, err)

	// Verify related records are cascade deleted
	var conditionCount int64
	db.Model(&models.MedicalConditionModel{}).Where("profile_id = ?", uint(profileIDUint)).Count(&conditionCount)
	assert.Equal(t, int64(0), conditionCount)

	var expenseCount int64
	db.Model(&models.MedicalExpenseModel{}).Where("profile_id = ?", uint(profileIDUint)).Count(&expenseCount)
	assert.Equal(t, int64(0), expenseCount)

	var policyCount int64
	db.Model(&models.InsurancePolicyModel{}).Where("profile_id = ?", uint(profileIDUint)).Count(&policyCount)
	assert.Equal(t, int64(0), policyCount)
}

func TestHealthProfileRepository_GetByID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewHealthProfileRepository(db)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, uint(999))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestHealthProfileRepository_GetByUserID_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewHealthProfileRepository(db)
	ctx := context.Background()

	_, err := repo.GetByUserID(ctx, "non-existent-user")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}
