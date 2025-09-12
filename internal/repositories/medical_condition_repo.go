package repositories

import (
	"context"
	"fmt"
	"strconv"

	"gorm.io/gorm"

	"github.com/DuckDHD/BuyOrBye/internal/domain"
	"github.com/DuckDHD/BuyOrBye/internal/models"
	"github.com/DuckDHD/BuyOrBye/internal/services"
)

// medicalConditionRepository implements services.MedicalConditionRepository
type medicalConditionRepository struct {
	db *gorm.DB
}

// NewMedicalConditionRepository creates a new medical condition repository
func NewMedicalConditionRepository(db *gorm.DB) services.MedicalConditionRepository {
	return &medicalConditionRepository{db: db}
}

// Create creates a new medical condition
func (r *medicalConditionRepository) Create(ctx context.Context, condition *domain.MedicalCondition) (*domain.MedicalCondition, error) {
	// Convert ProfileID from string to uint
	profileID, err := strconv.ParseUint(condition.ProfileID, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid profile ID: %w", err)
	}

	model := &models.MedicalConditionModel{}
	model.FromDomain(condition, uint(profileID))

	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return nil, fmt.Errorf("failed to create medical condition: %w", err)
	}

	return model.ToDomain(), nil
}

// GetByID retrieves a medical condition by ID
func (r *medicalConditionRepository) GetByID(ctx context.Context, id string) (*domain.MedicalCondition, error) {
	idUint, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid condition ID: %w", err)
	}

	var model models.MedicalConditionModel
	
	if err := r.db.WithContext(ctx).First(&model, uint(idUint)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("medical condition with ID %s not found", id)
		}
		return nil, fmt.Errorf("failed to get medical condition: %w", err)
	}

	return model.ToDomain(), nil
}

// Update updates a medical condition
func (r *medicalConditionRepository) Update(ctx context.Context, condition *domain.MedicalCondition) (*domain.MedicalCondition, error) {
	idUint, err := strconv.ParseUint(condition.ID, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid condition ID: %w", err)
	}

	profileID, err := strconv.ParseUint(condition.ProfileID, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid profile ID: %w", err)
	}

	var model models.MedicalConditionModel
	if err := r.db.WithContext(ctx).First(&model, uint(idUint)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("medical condition with ID %s not found", condition.ID)
		}
		return nil, fmt.Errorf("failed to find medical condition for update: %w", err)
	}

	// Update fields from domain
	model.FromDomain(condition, uint(profileID))
	model.ID = uint(idUint) // Preserve ID

	if err := r.db.WithContext(ctx).Save(&model).Error; err != nil {
		return nil, fmt.Errorf("failed to update medical condition: %w", err)
	}

	return model.ToDomain(), nil
}

// Delete performs soft delete on a medical condition
func (r *medicalConditionRepository) Delete(ctx context.Context, id string) error {
	idUint, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		return fmt.Errorf("invalid condition ID: %w", err)
	}

	result := r.db.WithContext(ctx).Delete(&models.MedicalConditionModel{}, uint(idUint))
	if result.Error != nil {
		return fmt.Errorf("failed to delete medical condition: %w", result.Error)
	}
	
	if result.RowsAffected == 0 {
		return fmt.Errorf("medical condition with ID %s not found", id)
	}

	return nil
}

// GetByUserID retrieves medical conditions by user ID with optional active filter
func (r *medicalConditionRepository) GetByUserID(ctx context.Context, userID string, activeOnly bool) ([]*domain.MedicalCondition, error) {
	var models []models.MedicalConditionModel
	
	query := r.db.WithContext(ctx).Where("user_id = ?", userID)
	if activeOnly {
		query = query.Where("is_active = ?", true)
	}

	if err := query.Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to get medical conditions: %w", err)
	}

	conditions := make([]*domain.MedicalCondition, len(models))
	for i, model := range models {
		conditions[i] = model.ToDomain()
	}

	return conditions, nil
}

// GetByCategory retrieves medical conditions by category
func (r *medicalConditionRepository) GetByCategory(ctx context.Context, userID string, category string) ([]*domain.MedicalCondition, error) {
	var models []models.MedicalConditionModel
	
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND category = ?", userID, category).
		Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to get medical conditions by category: %w", err)
	}

	conditions := make([]*domain.MedicalCondition, len(models))
	for i, model := range models {
		conditions[i] = model.ToDomain()
	}

	return conditions, nil
}

// GetBySeverity retrieves medical conditions by severity
func (r *medicalConditionRepository) GetBySeverity(ctx context.Context, userID string, severity string) ([]*domain.MedicalCondition, error) {
	var models []models.MedicalConditionModel
	
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND severity = ?", userID, severity).
		Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to get medical conditions by severity: %w", err)
	}

	conditions := make([]*domain.MedicalCondition, len(models))
	for i, model := range models {
		conditions[i] = model.ToDomain()
	}

	return conditions, nil
}

// GetByProfileID retrieves medical conditions by profile ID
func (r *medicalConditionRepository) GetByProfileID(ctx context.Context, profileID string) ([]*domain.MedicalCondition, error) {
	profileIDUint, err := strconv.ParseUint(profileID, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid profile ID: %w", err)
	}

	var models []models.MedicalConditionModel
	
	if err := r.db.WithContext(ctx).
		Where("profile_id = ?", uint(profileIDUint)).
		Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to get medical conditions by profile: %w", err)
	}

	conditions := make([]*domain.MedicalCondition, len(models))
	for i, model := range models {
		conditions[i] = model.ToDomain()
	}

	return conditions, nil
}

// GetActiveConditionCount returns count of active conditions for a user
func (r *medicalConditionRepository) GetActiveConditionCount(ctx context.Context, userID string) (int64, error) {
	var count int64
	
	if err := r.db.WithContext(ctx).
		Model(&models.MedicalConditionModel{}).
		Where("user_id = ? AND is_active = ?", userID, true).
		Count(&count).Error; err != nil {
		return 0, fmt.Errorf("failed to count active conditions: %w", err)
	}

	return count, nil
}

// CalculateTotalRiskFactor calculates sum of risk factors for active conditions
func (r *medicalConditionRepository) CalculateTotalRiskFactor(ctx context.Context, userID string) (float64, error) {
	var result struct {
		TotalRisk float64
	}
	
	if err := r.db.WithContext(ctx).
		Model(&models.MedicalConditionModel{}).
		Select("COALESCE(SUM(risk_factor), 0) as total_risk").
		Where("user_id = ? AND is_active = ?", userID, true).
		Scan(&result).Error; err != nil {
		return 0, fmt.Errorf("failed to calculate total risk factor: %w", err)
	}

	return result.TotalRisk, nil
}

// GetMedicationRequiringConditions returns conditions that require medication
func (r *medicalConditionRepository) GetMedicationRequiringConditions(ctx context.Context, userID string) ([]*domain.MedicalCondition, error) {
	var models []models.MedicalConditionModel
	
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND requires_medication = ? AND is_active = ?", userID, true, true).
		Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to get medication-requiring conditions: %w", err)
	}

	conditions := make([]*domain.MedicalCondition, len(models))
	for i, model := range models {
		conditions[i] = model.ToDomain()
	}

	return conditions, nil
}