package repositories

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"gorm.io/gorm"

	"github.com/DuckDHD/BuyOrBye/internal/domain"
	"github.com/DuckDHD/BuyOrBye/internal/models"
	"github.com/DuckDHD/BuyOrBye/internal/services"
)

// medicalExpenseRepository implements services.MedicalExpenseRepository
type medicalExpenseRepository struct {
	db *gorm.DB
}

// NewMedicalExpenseRepository creates a new medical expense repository
func NewMedicalExpenseRepository(db *gorm.DB) services.MedicalExpenseRepository {
	return &medicalExpenseRepository{db: db}
}

// Create creates a new medical expense
func (r *medicalExpenseRepository) Create(ctx context.Context, expense *domain.MedicalExpense) (*domain.MedicalExpense, error) {
	// Convert ProfileID from string to uint
	profileID, err := strconv.ParseUint(expense.ProfileID, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid profile ID: %w", err)
	}

	model := &models.MedicalExpenseModel{}
	model.FromDomain(expense, uint(profileID))

	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return nil, fmt.Errorf("failed to create medical expense: %w", err)
	}

	return model.ToDomain(), nil
}

// GetByID retrieves a medical expense by ID
func (r *medicalExpenseRepository) GetByID(ctx context.Context, id string) (*domain.MedicalExpense, error) {
	idUint, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid expense ID: %w", err)
	}

	var model models.MedicalExpenseModel
	
	if err := r.db.WithContext(ctx).First(&model, uint(idUint)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("medical expense with ID %s not found", id)
		}
		return nil, fmt.Errorf("failed to get medical expense: %w", err)
	}

	return model.ToDomain(), nil
}

// Update updates a medical expense
func (r *medicalExpenseRepository) Update(ctx context.Context, expense *domain.MedicalExpense) (*domain.MedicalExpense, error) {
	idUint, err := strconv.ParseUint(expense.ID, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid expense ID: %w", err)
	}

	profileID, err := strconv.ParseUint(expense.ProfileID, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid profile ID: %w", err)
	}

	var model models.MedicalExpenseModel
	if err := r.db.WithContext(ctx).First(&model, uint(idUint)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("medical expense with ID %s not found", expense.ID)
		}
		return nil, fmt.Errorf("failed to find medical expense for update: %w", err)
	}

	// Update fields from domain
	model.FromDomain(expense, uint(profileID))
	model.ID = uint(idUint) // Preserve ID

	if err := r.db.WithContext(ctx).Save(&model).Error; err != nil {
		return nil, fmt.Errorf("failed to update medical expense: %w", err)
	}

	return model.ToDomain(), nil
}

// Delete performs soft delete on a medical expense
func (r *medicalExpenseRepository) Delete(ctx context.Context, id string) error {
	idUint, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		return fmt.Errorf("invalid expense ID: %w", err)
	}

	result := r.db.WithContext(ctx).Delete(&models.MedicalExpenseModel{}, uint(idUint))
	if result.Error != nil {
		return fmt.Errorf("failed to delete medical expense: %w", result.Error)
	}
	
	if result.RowsAffected == 0 {
		return fmt.Errorf("medical expense with ID %s not found", id)
	}

	return nil
}

// GetByUserID retrieves medical expenses by user ID
func (r *medicalExpenseRepository) GetByUserID(ctx context.Context, userID string) ([]*domain.MedicalExpense, error) {
	var models []models.MedicalExpenseModel
	
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("date DESC").
		Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to get medical expenses: %w", err)
	}

	expenses := make([]*domain.MedicalExpense, len(models))
	for i, model := range models {
		expenses[i] = model.ToDomain()
	}

	return expenses, nil
}

// GetByDateRange retrieves medical expenses within a date range
func (r *medicalExpenseRepository) GetByDateRange(ctx context.Context, userID string, startDate, endDate time.Time) ([]*domain.MedicalExpense, error) {
	var models []models.MedicalExpenseModel
	
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND date >= ? AND date <= ?", userID, startDate, endDate).
		Order("date DESC").
		Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to get medical expenses by date range: %w", err)
	}

	expenses := make([]*domain.MedicalExpense, len(models))
	for i, model := range models {
		expenses[i] = model.ToDomain()
	}

	return expenses, nil
}

// GetByCategory retrieves medical expenses by category
func (r *medicalExpenseRepository) GetByCategory(ctx context.Context, userID string, category string) ([]*domain.MedicalExpense, error) {
	var models []models.MedicalExpenseModel
	
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND category = ?", userID, category).
		Order("date DESC").
		Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to get medical expenses by category: %w", err)
	}

	expenses := make([]*domain.MedicalExpense, len(models))
	for i, model := range models {
		expenses[i] = model.ToDomain()
	}

	return expenses, nil
}

// GetByFrequency retrieves medical expenses by frequency
func (r *medicalExpenseRepository) GetByFrequency(ctx context.Context, userID string, frequency string) ([]*domain.MedicalExpense, error) {
	var models []models.MedicalExpenseModel
	
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND frequency = ?", userID, frequency).
		Order("date DESC").
		Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to get medical expenses by frequency: %w", err)
	}

	expenses := make([]*domain.MedicalExpense, len(models))
	for i, model := range models {
		expenses[i] = model.ToDomain()
	}

	return expenses, nil
}

// GetRecurring retrieves recurring medical expenses
func (r *medicalExpenseRepository) GetRecurring(ctx context.Context, userID string) ([]*domain.MedicalExpense, error) {
	var models []models.MedicalExpenseModel
	
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND is_recurring = ?", userID, true).
		Order("date DESC").
		Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to get recurring medical expenses: %w", err)
	}

	expenses := make([]*domain.MedicalExpense, len(models))
	for i, model := range models {
		expenses[i] = model.ToDomain()
	}

	return expenses, nil
}

// GetByProfileID retrieves medical expenses by profile ID
func (r *medicalExpenseRepository) GetByProfileID(ctx context.Context, profileID string) ([]*domain.MedicalExpense, error) {
	profileIDUint, err := strconv.ParseUint(profileID, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid profile ID: %w", err)
	}

	var models []models.MedicalExpenseModel
	
	if err := r.db.WithContext(ctx).
		Where("profile_id = ?", uint(profileIDUint)).
		Order("date DESC").
		Find(&models).Error; err != nil {
		return nil, fmt.Errorf("failed to get medical expenses by profile: %w", err)
	}

	expenses := make([]*domain.MedicalExpense, len(models))
	for i, model := range models {
		expenses[i] = model.ToDomain()
	}

	return expenses, nil
}

// CalculateTotals calculates expense totals within a date range
func (r *medicalExpenseRepository) CalculateTotals(ctx context.Context, userID string, startDate, endDate time.Time) (*services.ExpenseTotals, error) {
	var result struct {
		TotalAmount        float64
		TotalInsurancePaid float64
		TotalOutOfPocket   float64
		ExpenseCount       int64
	}
	
	if err := r.db.WithContext(ctx).
		Model(&models.MedicalExpenseModel{}).
		Select(`
			COALESCE(SUM(amount), 0) as total_amount,
			COALESCE(SUM(insurance_payment), 0) as total_insurance_paid,
			COALESCE(SUM(out_of_pocket), 0) as total_out_of_pocket,
			COUNT(*) as expense_count
		`).
		Where("user_id = ? AND date >= ? AND date <= ?", userID, startDate, endDate).
		Scan(&result).Error; err != nil {
		return nil, fmt.Errorf("failed to calculate expense totals: %w", err)
	}

	return &services.ExpenseTotals{
		TotalAmount:        result.TotalAmount,
		TotalInsurancePaid: result.TotalInsurancePaid,
		TotalOutOfPocket:   result.TotalOutOfPocket,
		ExpenseCount:       result.ExpenseCount,
	}, nil
}

// GetMonthlyRecurringTotal calculates total monthly recurring expenses
func (r *medicalExpenseRepository) GetMonthlyRecurringTotal(ctx context.Context, userID string) (float64, error) {
	// Get all recurring expenses and convert to monthly
	var expenses []models.MedicalExpenseModel
	if err := r.db.WithContext(ctx).
		Where("user_id = ? AND is_recurring = ?", userID, true).
		Find(&expenses).Error; err != nil {
		return 0, fmt.Errorf("failed to get recurring expenses: %w", err)
	}

	monthlyTotal := 0.0
	for _, expense := range expenses {
		monthlyAmount := convertToMonthlyAmount(expense.Amount, expense.Frequency)
		monthlyTotal += monthlyAmount
	}

	return monthlyTotal, nil
}

// GetAnnualProjectedExpenses calculates projected annual expenses
func (r *medicalExpenseRepository) GetAnnualProjectedExpenses(ctx context.Context, userID string) (float64, error) {
	monthlyTotal, err := r.GetMonthlyRecurringTotal(ctx, userID)
	if err != nil {
		return 0, err
	}

	// Project annual based on monthly recurring expenses
	annualProjected := monthlyTotal * 12

	// Add one-time expenses from the last 12 months as a baseline
	oneYear := time.Now().AddDate(-1, 0, 0)
	now := time.Now()
	
	var oneTimeTotal struct {
		Total float64
	}
	
	if err := r.db.WithContext(ctx).
		Model(&models.MedicalExpenseModel{}).
		Select("COALESCE(SUM(amount), 0) as total").
		Where("user_id = ? AND is_recurring = ? AND date >= ? AND date <= ?", userID, false, oneYear, now).
		Scan(&oneTimeTotal).Error; err != nil {
		return 0, fmt.Errorf("failed to calculate one-time expenses: %w", err)
	}

	return annualProjected + oneTimeTotal.Total, nil
}

// convertToMonthlyAmount converts expense amount to monthly based on frequency
func convertToMonthlyAmount(amount float64, frequency string) float64 {
	switch frequency {
	case "daily":
		return amount * 30 // Approximate month
	case "weekly":
		return amount * 4.33 // Approximate weeks per month
	case "bi-weekly":
		return amount * 2.17 // Approximate bi-weeks per month
	case "monthly":
		return amount
	case "quarterly":
		return amount / 3
	case "semi-annually":
		return amount / 6
	case "annually":
		return amount / 12
	default: // "one_time" and others
		return 0 // One-time expenses don't contribute to monthly recurring
	}
}