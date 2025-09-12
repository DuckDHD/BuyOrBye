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

// healthProfileRepository implements services.HealthProfileRepository
type healthProfileRepository struct {
	db *gorm.DB
}

// NewHealthProfileRepository creates a new health profile repository
func NewHealthProfileRepository(db *gorm.DB) services.HealthProfileRepository {
	return &healthProfileRepository{db: db}
}

// Create creates a new health profile
func (r *healthProfileRepository) Create(ctx context.Context, profile *domain.HealthProfile) (*domain.HealthProfile, error) {
	model := &models.HealthProfileModel{}
	model.FromDomain(profile)

	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		if isHealthProfileDuplicateKeyError(err) {
			return nil, fmt.Errorf("health profile already exists for user %s: unique constraint violation", profile.UserID)
		}
		return nil, fmt.Errorf("failed to create health profile: %w", err)
	}

	return model.ToDomain(), nil
}

// GetByID retrieves a health profile by ID
func (r *healthProfileRepository) GetByID(ctx context.Context, id uint) (*domain.HealthProfile, error) {
	var model models.HealthProfileModel
	
	if err := r.db.WithContext(ctx).First(&model, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("health profile with ID %d not found", id)
		}
		return nil, fmt.Errorf("failed to get health profile: %w", err)
	}

	return model.ToDomain(), nil
}

// GetByUserID retrieves a health profile by user ID
func (r *healthProfileRepository) GetByUserID(ctx context.Context, userID string) (*domain.HealthProfile, error) {
	var model models.HealthProfileModel
	
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("health profile not found for user %s", userID)
		}
		return nil, fmt.Errorf("failed to get health profile: %w", err)
	}

	return model.ToDomain(), nil
}

// Update updates a health profile
func (r *healthProfileRepository) Update(ctx context.Context, profile *domain.HealthProfile) (*domain.HealthProfile, error) {
	id, err := strconv.ParseUint(profile.ID, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid profile ID: %w", err)
	}

	var model models.HealthProfileModel
	if err := r.db.WithContext(ctx).First(&model, uint(id)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("health profile with ID %s not found", profile.ID)
		}
		return nil, fmt.Errorf("failed to find health profile for update: %w", err)
	}

	// Update fields from domain
	model.FromDomain(profile)
	model.ID = uint(id) // Preserve ID

	if err := r.db.WithContext(ctx).Save(&model).Error; err != nil {
		return nil, fmt.Errorf("failed to update health profile: %w", err)
	}

	return model.ToDomain(), nil
}

// Delete deletes a health profile (and cascades to related records)
func (r *healthProfileRepository) Delete(ctx context.Context, id uint) error {
	result := r.db.WithContext(ctx).Delete(&models.HealthProfileModel{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete health profile: %w", result.Error)
	}
	
	if result.RowsAffected == 0 {
		return fmt.Errorf("health profile with ID %d not found", id)
	}

	return nil
}

// GetWithRelations retrieves a health profile with all related entities preloaded
func (r *healthProfileRepository) GetWithRelations(ctx context.Context, userID string) (*domain.HealthProfile, error) {
	var model models.HealthProfileModel
	
	if err := r.db.WithContext(ctx).
		Preload("Conditions").
		Preload("Expenses").
		Preload("Policies").
		Where("user_id = ?", userID).
		First(&model).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("health profile not found for user %s", userID)
		}
		return nil, fmt.Errorf("failed to get health profile with relations: %w", err)
	}

	return model.ToDomain(), nil
}

// ExistsByUserID checks if a health profile exists for the given user ID
func (r *healthProfileRepository) ExistsByUserID(ctx context.Context, userID string) (bool, error) {
	var count int64
	
	if err := r.db.WithContext(ctx).
		Model(&models.HealthProfileModel{}).
		Where("user_id = ?", userID).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check health profile existence: %w", err)
	}

	return count > 0, nil
}

// isHealthProfileDuplicateKeyError checks if the error is a duplicate key/unique constraint violation
func isHealthProfileDuplicateKeyError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return containsStr(errStr, "UNIQUE constraint failed") ||
		containsStr(errStr, "Duplicate entry") ||
		containsStr(errStr, "unique constraint")
}

// containsStr checks if a string contains a substring (case-sensitive helper)
func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}