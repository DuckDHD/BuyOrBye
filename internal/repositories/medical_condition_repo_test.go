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

func setupConditionTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(
		&models.HealthProfileModel{},
		&models.MedicalConditionModel{},
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

func TestMedicalConditionRepository_AddCondition_Success(t *testing.T) {
	db := setupConditionTestDB(t)
	repo := NewMedicalConditionRepository(db)
	ctx := context.Background()

	condition := &domain.MedicalCondition{
		UserID:             "test-user-123",
		ProfileID:          "1",
		Name:               "Type 2 Diabetes",
		Category:           "chronic",
		Severity:           "moderate",
		DiagnosedDate:      time.Now().AddDate(-1, 0, 0),
		IsActive:           true,
		RequiresMedication: true,
		MonthlyMedCost:     150.0,
		RiskFactor:         0.25,
	}

	result, err := repo.Create(ctx, condition)

	assert.NoError(t, err)
	assert.NotEmpty(t, result.ID)
	assert.Equal(t, "test-user-123", result.UserID)
	assert.Equal(t, "Type 2 Diabetes", result.Name)
	assert.Equal(t, "chronic", result.Category)
	assert.Equal(t, "moderate", result.Severity)
	assert.True(t, result.IsActive)
	assert.True(t, result.RequiresMedication)
	assert.Equal(t, 150.0, result.MonthlyMedCost)
	assert.Equal(t, 0.25, result.RiskFactor)
}

func TestMedicalConditionRepository_GetConditionsByUser_FilterActive(t *testing.T) {
	db := setupConditionTestDB(t)
	repo := NewMedicalConditionRepository(db)
	ctx := context.Background()

	// Create active condition
	activeCondition := &domain.MedicalCondition{
		UserID:             "test-user-123",
		ProfileID:          "1",
		Name:               "Hypertension",
		Category:           "chronic",
		Severity:           "mild",
		DiagnosedDate:      time.Now().AddDate(-2, 0, 0),
		IsActive:           true,
		RequiresMedication: true,
		MonthlyMedCost:     75.0,
		RiskFactor:         0.15,
	}

	// Create inactive condition
	inactiveCondition := &domain.MedicalCondition{
		UserID:             "test-user-123",
		ProfileID:          "1",
		Name:               "Broken Arm",
		Category:           "acute",
		Severity:           "moderate",
		DiagnosedDate:      time.Now().AddDate(0, -6, 0),
		IsActive:           false,
		RequiresMedication: false,
		MonthlyMedCost:     0.0,
		RiskFactor:         0.05,
	}

	_, err := repo.Create(ctx, activeCondition)
	require.NoError(t, err)

	_, err = repo.Create(ctx, inactiveCondition)
	require.NoError(t, err)

	// Get only active conditions
	activeConditions, err := repo.GetByUserID(ctx, "test-user-123", true)

	assert.NoError(t, err)
	assert.Len(t, activeConditions, 1)
	assert.Equal(t, "Hypertension", activeConditions[0].Name)
	assert.True(t, activeConditions[0].IsActive)

	// Get all conditions (active and inactive)
	allConditions, err := repo.GetByUserID(ctx, "test-user-123", false)

	assert.NoError(t, err)
	assert.Len(t, allConditions, 2)
}

func TestMedicalConditionRepository_GetConditionsBySeverity_Filters(t *testing.T) {
	db := setupConditionTestDB(t)
	repo := NewMedicalConditionRepository(db)
	ctx := context.Background()

	conditions := []*domain.MedicalCondition{
		{
			UserID:        "test-user-123",
			ProfileID:     "1",
			Name:          "Mild Anxiety",
			Category:      "mental_health",
			Severity:      "mild",
			DiagnosedDate: time.Now(),
			IsActive:      true,
			RiskFactor:    0.1,
		},
		{
			UserID:        "test-user-123",
			ProfileID:     "1",
			Name:          "Severe Depression",
			Category:      "mental_health",
			Severity:      "severe",
			DiagnosedDate: time.Now(),
			IsActive:      true,
			RiskFactor:    0.4,
		},
		{
			UserID:        "test-user-123",
			ProfileID:     "1",
			Name:          "Moderate Asthma",
			Category:      "chronic",
			Severity:      "moderate",
			DiagnosedDate: time.Now(),
			IsActive:      true,
			RiskFactor:    0.2,
		},
	}

	for _, condition := range conditions {
		_, err := repo.Create(ctx, condition)
		require.NoError(t, err)
	}

	// Get severe conditions
	severeConditions, err := repo.GetBySeverity(ctx, "test-user-123", "severe")

	assert.NoError(t, err)
	assert.Len(t, severeConditions, 1)
	assert.Equal(t, "Severe Depression", severeConditions[0].Name)

	// Get mild conditions
	mildConditions, err := repo.GetBySeverity(ctx, "test-user-123", "mild")

	assert.NoError(t, err)
	assert.Len(t, mildConditions, 1)
	assert.Equal(t, "Mild Anxiety", mildConditions[0].Name)
}

func TestMedicalConditionRepository_UpdateCondition_ChangesSeverity(t *testing.T) {
	db := setupConditionTestDB(t)
	repo := NewMedicalConditionRepository(db)
	ctx := context.Background()

	// Create condition
	condition := &domain.MedicalCondition{
		UserID:             "test-user-123",
		ProfileID:          "1",
		Name:               "Diabetes",
		Category:           "chronic",
		Severity:           "mild",
		DiagnosedDate:      time.Now(),
		IsActive:           true,
		RequiresMedication: false,
		MonthlyMedCost:     0.0,
		RiskFactor:         0.1,
	}

	createdCondition, err := repo.Create(ctx, condition)
	require.NoError(t, err)

	// Update severity and medication requirements
	createdCondition.Severity = "moderate"
	createdCondition.RequiresMedication = true
	createdCondition.MonthlyMedCost = 120.0
	createdCondition.RiskFactor = 0.25

	updatedCondition, err := repo.Update(ctx, createdCondition)

	assert.NoError(t, err)
	assert.Equal(t, "moderate", updatedCondition.Severity)
	assert.True(t, updatedCondition.RequiresMedication)
	assert.Equal(t, 120.0, updatedCondition.MonthlyMedCost)
	assert.Equal(t, 0.25, updatedCondition.RiskFactor)
	assert.True(t, updatedCondition.UpdatedAt.After(updatedCondition.CreatedAt))
}

func TestMedicalConditionRepository_DeleteCondition_SoftDelete(t *testing.T) {
	db := setupConditionTestDB(t)
	repo := NewMedicalConditionRepository(db)
	ctx := context.Background()

	// Create condition
	condition := &domain.MedicalCondition{
		UserID:        "test-user-123",
		ProfileID:     "1",
		Name:          "Temporary Condition",
		Category:      "acute",
		Severity:      "mild",
		DiagnosedDate: time.Now(),
		IsActive:      true,
		RiskFactor:    0.05,
	}

	createdCondition, err := repo.Create(ctx, condition)
	require.NoError(t, err)

	// Delete condition (soft delete)
	err = repo.Delete(ctx, createdCondition.ID)
	assert.NoError(t, err)

	// Verify condition is not returned in normal queries
	conditions, err := repo.GetByUserID(ctx, "test-user-123", false)
	assert.NoError(t, err)
	
	found := false
	for _, c := range conditions {
		if c.ID == createdCondition.ID {
			found = true
			break
		}
	}
	assert.False(t, found, "Deleted condition should not be returned")

	// Verify record still exists in database with deleted_at set
	var count int64
	db.Unscoped().Model(&models.MedicalConditionModel{}).Where("id = ? AND deleted_at IS NOT NULL", createdCondition.ID).Count(&count)
	assert.Equal(t, int64(1), count)
}

func TestMedicalConditionRepository_GetByCategory_Filters(t *testing.T) {
	db := setupConditionTestDB(t)
	repo := NewMedicalConditionRepository(db)
	ctx := context.Background()

	conditions := []*domain.MedicalCondition{
		{
			UserID:        "test-user-123",
			ProfileID:     "1",
			Name:          "Diabetes",
			Category:      "chronic",
			Severity:      "moderate",
			DiagnosedDate: time.Now(),
			IsActive:      true,
			RiskFactor:    0.25,
		},
		{
			UserID:        "test-user-123",
			ProfileID:     "1",
			Name:          "Broken Leg",
			Category:      "acute",
			Severity:      "severe",
			DiagnosedDate: time.Now(),
			IsActive:      true,
			RiskFactor:    0.1,
		},
		{
			UserID:        "test-user-123",
			ProfileID:     "1",
			Name:          "Annual Checkup",
			Category:      "preventive",
			Severity:      "mild",
			DiagnosedDate: time.Now(),
			IsActive:      true,
			RiskFactor:    0.0,
		},
	}

	for _, condition := range conditions {
		_, err := repo.Create(ctx, condition)
		require.NoError(t, err)
	}

	// Get chronic conditions
	chronicConditions, err := repo.GetByCategory(ctx, "test-user-123", "chronic")

	assert.NoError(t, err)
	assert.Len(t, chronicConditions, 1)
	assert.Equal(t, "Diabetes", chronicConditions[0].Name)

	// Get acute conditions
	acuteConditions, err := repo.GetByCategory(ctx, "test-user-123", "acute")

	assert.NoError(t, err)
	assert.Len(t, acuteConditions, 1)
	assert.Equal(t, "Broken Leg", acuteConditions[0].Name)
}