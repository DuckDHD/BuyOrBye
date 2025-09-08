package repositories

import (
	"context"
	"fmt"

	"github.com/DuckDHD/BuyOrBye/internal/domain"
	"github.com/DuckDHD/BuyOrBye/internal/models"
	"github.com/DuckDHD/BuyOrBye/internal/services"
	"gorm.io/gorm"
)

// incomeRepository implements services.IncomeRepository using GORM
type incomeRepository struct {
	db *gorm.DB
}

// NewIncomeRepository creates a new income repository instance
func NewIncomeRepository(db *gorm.DB) services.IncomeRepository {
	return &incomeRepository{
		db: db,
	}
}

// SaveIncome saves or updates an income record
func (r *incomeRepository) SaveIncome(ctx context.Context, income domain.Income) error {
	model := models.NewIncomeModelFromDomain(income)
	
	// First try to find existing record
	var existing models.IncomeModel
	result := r.db.WithContext(ctx).First(&existing, "id = ?", income.ID)
	
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			// Create new record with explicit column selection to override defaults
			result = r.db.WithContext(ctx).Select("*").Create(model)
		} else {
			return fmt.Errorf("failed to check existing income: %w", result.Error)
		}
	} else {
		// Update existing record with explicit selection to handle zero values
		result = r.db.WithContext(ctx).Model(&existing).Select("*").Updates(model)
	}
	
	if result.Error != nil {
		return fmt.Errorf("failed to save income: %w", result.Error)
	}
	
	return nil
}

// GetIncomeByID retrieves an income by its ID
func (r *incomeRepository) GetIncomeByID(ctx context.Context, id string) (domain.Income, error) {
	var model models.IncomeModel
	
	result := r.db.WithContext(ctx).First(&model, "id = ?", id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return domain.Income{}, fmt.Errorf("income with ID %s not found", id)
		}
		return domain.Income{}, fmt.Errorf("failed to get income by ID: %w", result.Error)
	}
	
	return model.ToDomain(), nil
}

// UpdateIncome updates an existing income record
func (r *incomeRepository) UpdateIncome(ctx context.Context, income domain.Income) error {
	model := models.NewIncomeModelFromDomain(income)
	
	// Use Select to explicitly update all fields including zero values
	result := r.db.WithContext(ctx).Model(&models.IncomeModel{}).
		Where("id = ?", income.ID).
		Select("*").
		Updates(model)
	
	if result.Error != nil {
		return fmt.Errorf("failed to update income: %w", result.Error)
	}
	
	if result.RowsAffected == 0 {
		return fmt.Errorf("income with ID %s not found", income.ID)
	}
	
	return nil
}

// DeleteIncome soft deletes an income record
func (r *incomeRepository) DeleteIncome(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&models.IncomeModel{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete income: %w", result.Error)
	}
	
	if result.RowsAffected == 0 {
		return fmt.Errorf("income with ID %s not found", id)
	}
	
	return nil
}

// GetUserIncomes retrieves all incomes for a specific user
func (r *incomeRepository) GetUserIncomes(ctx context.Context, userID string) ([]domain.Income, error) {
	var models []models.IncomeModel
	
	result := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&models)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get user incomes: %w", result.Error)
	}
	
	incomes := make([]domain.Income, len(models))
	for i, model := range models {
		incomes[i] = model.ToDomain()
	}
	
	return incomes, nil
}

// GetActiveIncomes retrieves only active incomes for a specific user
func (r *incomeRepository) GetActiveIncomes(ctx context.Context, userID string) ([]domain.Income, error) {
	var models []models.IncomeModel
	
	result := r.db.WithContext(ctx).Where("user_id = ? AND is_active = ?", userID, true).Find(&models)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get active incomes: %w", result.Error)
	}
	
	incomes := make([]domain.Income, len(models))
	for i, model := range models {
		incomes[i] = model.ToDomain()
	}
	
	return incomes, nil
}

// GetUserIncomesByFrequency retrieves user incomes filtered by frequency
func (r *incomeRepository) GetUserIncomesByFrequency(ctx context.Context, userID string, frequency string) ([]domain.Income, error) {
	var models []models.IncomeModel
	
	result := r.db.WithContext(ctx).Where("user_id = ? AND frequency = ?", userID, frequency).Find(&models)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get incomes by frequency: %w", result.Error)
	}
	
	incomes := make([]domain.Income, len(models))
	for i, model := range models {
		incomes[i] = model.ToDomain()
	}
	
	return incomes, nil
}

// CalculateUserTotalIncome calculates the total income for a user
func (r *incomeRepository) CalculateUserTotalIncome(ctx context.Context, userID string, activeOnly bool) (float64, error) {
	var total float64
	
	query := r.db.WithContext(ctx).Model(&models.IncomeModel{}).
		Select("COALESCE(SUM(amount), 0)").
		Where("user_id = ?", userID)
	
	if activeOnly {
		query = query.Where("is_active = ?", true)
	}
	
	result := query.Scan(&total)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to calculate total income: %w", result.Error)
	}
	
	return total, nil
}