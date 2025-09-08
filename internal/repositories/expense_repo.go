package repositories

import (
	"context"
	"fmt"

	"github.com/DuckDHD/BuyOrBye/internal/domain"
	"github.com/DuckDHD/BuyOrBye/internal/models"
	"github.com/DuckDHD/BuyOrBye/internal/services"
	"gorm.io/gorm"
)

// expenseRepository implements services.ExpenseRepository using GORM
type expenseRepository struct {
	db *gorm.DB
}

// NewExpenseRepository creates a new expense repository instance
func NewExpenseRepository(db *gorm.DB) services.ExpenseRepository {
	return &expenseRepository{
		db: db,
	}
}

// SaveExpense saves or updates an expense record
func (r *expenseRepository) SaveExpense(ctx context.Context, expense domain.Expense) error {
	model := models.NewExpenseModelFromDomain(expense)
	
	// First try to find existing record
	var existing models.ExpenseModel
	result := r.db.WithContext(ctx).First(&existing, "id = ?", expense.ID)
	
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			// Create new record with explicit column selection to override defaults
			result = r.db.WithContext(ctx).Select("*").Create(model)
		} else {
			return fmt.Errorf("failed to check existing expense: %w", result.Error)
		}
	} else {
		// Update existing record with explicit selection to handle zero values
		result = r.db.WithContext(ctx).Model(&existing).Select("*").Updates(model)
	}
	
	if result.Error != nil {
		return fmt.Errorf("failed to save expense: %w", result.Error)
	}
	
	return nil
}

// GetExpenseByID retrieves an expense by its ID
func (r *expenseRepository) GetExpenseByID(ctx context.Context, id string) (domain.Expense, error) {
	var model models.ExpenseModel
	
	result := r.db.WithContext(ctx).First(&model, "id = ?", id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return domain.Expense{}, fmt.Errorf("expense with ID %s not found", id)
		}
		return domain.Expense{}, fmt.Errorf("failed to get expense by ID: %w", result.Error)
	}
	
	return model.ToDomain(), nil
}

// UpdateExpense updates an existing expense record
func (r *expenseRepository) UpdateExpense(ctx context.Context, expense domain.Expense) error {
	model := models.NewExpenseModelFromDomain(expense)
	
	// Use Select to explicitly update all fields including zero values
	result := r.db.WithContext(ctx).Model(&models.ExpenseModel{}).
		Where("id = ?", expense.ID).
		Select("*").
		Updates(model)
	
	if result.Error != nil {
		return fmt.Errorf("failed to update expense: %w", result.Error)
	}
	
	if result.RowsAffected == 0 {
		return fmt.Errorf("expense with ID %s not found", expense.ID)
	}
	
	return nil
}

// DeleteExpense soft deletes an expense record
func (r *expenseRepository) DeleteExpense(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&models.ExpenseModel{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete expense: %w", result.Error)
	}
	
	if result.RowsAffected == 0 {
		return fmt.Errorf("expense with ID %s not found", id)
	}
	
	return nil
}

// GetUserExpenses retrieves all expenses for a specific user
func (r *expenseRepository) GetUserExpenses(ctx context.Context, userID string) ([]domain.Expense, error) {
	var models []models.ExpenseModel
	
	result := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&models)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get user expenses: %w", result.Error)
	}
	
	expenses := make([]domain.Expense, len(models))
	for i, model := range models {
		expenses[i] = model.ToDomain()
	}
	
	return expenses, nil
}

// GetExpensesByCategory retrieves expenses for a user filtered by category
func (r *expenseRepository) GetExpensesByCategory(ctx context.Context, userID string, category string) ([]domain.Expense, error) {
	var models []models.ExpenseModel
	
	result := r.db.WithContext(ctx).Where("user_id = ? AND category = ?", userID, category).Find(&models)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get expenses by category: %w", result.Error)
	}
	
	expenses := make([]domain.Expense, len(models))
	for i, model := range models {
		expenses[i] = model.ToDomain()
	}
	
	return expenses, nil
}

// GetExpensesByFrequency retrieves expenses for a user filtered by frequency
func (r *expenseRepository) GetExpensesByFrequency(ctx context.Context, userID string, frequency string) ([]domain.Expense, error) {
	var models []models.ExpenseModel
	
	result := r.db.WithContext(ctx).Where("user_id = ? AND frequency = ?", userID, frequency).Find(&models)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get expenses by frequency: %w", result.Error)
	}
	
	expenses := make([]domain.Expense, len(models))
	for i, model := range models {
		expenses[i] = model.ToDomain()
	}
	
	return expenses, nil
}

// GetExpensesByPriority retrieves expenses for a user filtered by priority
func (r *expenseRepository) GetExpensesByPriority(ctx context.Context, userID string, priority int) ([]domain.Expense, error) {
	var models []models.ExpenseModel
	
	result := r.db.WithContext(ctx).Where("user_id = ? AND priority = ?", userID, priority).Find(&models)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get expenses by priority: %w", result.Error)
	}
	
	expenses := make([]domain.Expense, len(models))
	for i, model := range models {
		expenses[i] = model.ToDomain()
	}
	
	return expenses, nil
}

// GetFixedExpenses retrieves only fixed expenses for a specific user
func (r *expenseRepository) GetFixedExpenses(ctx context.Context, userID string) ([]domain.Expense, error) {
	var models []models.ExpenseModel
	
	result := r.db.WithContext(ctx).Where("user_id = ? AND is_fixed = ?", userID, true).Find(&models)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get fixed expenses: %w", result.Error)
	}
	
	expenses := make([]domain.Expense, len(models))
	for i, model := range models {
		expenses[i] = model.ToDomain()
	}
	
	return expenses, nil
}

// GetVariableExpenses retrieves only variable expenses for a specific user
func (r *expenseRepository) GetVariableExpenses(ctx context.Context, userID string) ([]domain.Expense, error) {
	var models []models.ExpenseModel
	
	result := r.db.WithContext(ctx).Where("user_id = ? AND is_fixed = ?", userID, false).Find(&models)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get variable expenses: %w", result.Error)
	}
	
	expenses := make([]domain.Expense, len(models))
	for i, model := range models {
		expenses[i] = model.ToDomain()
	}
	
	return expenses, nil
}

// CalculateUserTotalExpenses calculates the total expenses for a user
func (r *expenseRepository) CalculateUserTotalExpenses(ctx context.Context, userID string) (float64, error) {
	var total float64
	
	result := r.db.WithContext(ctx).Model(&models.ExpenseModel{}).
		Select("COALESCE(SUM(amount), 0)").
		Where("user_id = ?", userID).
		Scan(&total)
	
	if result.Error != nil {
		return 0, fmt.Errorf("failed to calculate total expenses: %w", result.Error)
	}
	
	return total, nil
}

// CalculateTotalByCategory calculates the total expenses for a user in a specific category
func (r *expenseRepository) CalculateTotalByCategory(ctx context.Context, userID string, category string) (float64, error) {
	var total float64
	
	result := r.db.WithContext(ctx).Model(&models.ExpenseModel{}).
		Select("COALESCE(SUM(amount), 0)").
		Where("user_id = ? AND category = ?", userID, category).
		Scan(&total)
	
	if result.Error != nil {
		return 0, fmt.Errorf("failed to calculate total by category: %w", result.Error)
	}
	
	return total, nil
}