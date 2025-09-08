package repositories

import (
	"context"
	"fmt"

	"github.com/DuckDHD/BuyOrBye/internal/domain"
	"github.com/DuckDHD/BuyOrBye/internal/models"
	"github.com/DuckDHD/BuyOrBye/internal/services"
	"gorm.io/gorm"
)

// loanRepository implements services.LoanRepository using GORM
type loanRepository struct {
	db *gorm.DB
}

// NewLoanRepository creates a new loan repository instance
func NewLoanRepository(db *gorm.DB) services.LoanRepository {
	return &loanRepository{
		db: db,
	}
}

// SaveLoan saves or updates a loan record
func (r *loanRepository) SaveLoan(ctx context.Context, loan domain.Loan) error {
	model := models.NewLoanModelFromDomain(loan)
	
	// First try to find existing record
	var existing models.LoanModel
	result := r.db.WithContext(ctx).First(&existing, "id = ?", loan.ID)
	
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			// Create new record with explicit column selection to override defaults
			result = r.db.WithContext(ctx).Select("*").Create(model)
		} else {
			return fmt.Errorf("failed to check existing loan: %w", result.Error)
		}
	} else {
		// Update existing record with explicit selection to handle zero values
		result = r.db.WithContext(ctx).Model(&existing).Select("*").Updates(model)
	}
	
	if result.Error != nil {
		return fmt.Errorf("failed to save loan: %w", result.Error)
	}
	
	return nil
}

// GetLoanByID retrieves a loan by its ID
func (r *loanRepository) GetLoanByID(ctx context.Context, id string) (domain.Loan, error) {
	var model models.LoanModel
	
	result := r.db.WithContext(ctx).First(&model, "id = ?", id)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return domain.Loan{}, fmt.Errorf("loan with ID %s not found", id)
		}
		return domain.Loan{}, fmt.Errorf("failed to get loan by ID: %w", result.Error)
	}
	
	return model.ToDomain(), nil
}

// UpdateLoan updates an existing loan record
func (r *loanRepository) UpdateLoan(ctx context.Context, loan domain.Loan) error {
	model := models.NewLoanModelFromDomain(loan)
	
	// Use Select to explicitly update all fields including zero values
	result := r.db.WithContext(ctx).Model(&models.LoanModel{}).
		Where("id = ?", loan.ID).
		Select("*").
		Updates(model)
	
	if result.Error != nil {
		return fmt.Errorf("failed to update loan: %w", result.Error)
	}
	
	if result.RowsAffected == 0 {
		return fmt.Errorf("loan with ID %s not found", loan.ID)
	}
	
	return nil
}

// DeleteLoan soft deletes a loan record
func (r *loanRepository) DeleteLoan(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&models.LoanModel{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete loan: %w", result.Error)
	}
	
	if result.RowsAffected == 0 {
		return fmt.Errorf("loan with ID %s not found", id)
	}
	
	return nil
}

// GetUserLoans retrieves all loans for a specific user
func (r *loanRepository) GetUserLoans(ctx context.Context, userID string) ([]domain.Loan, error) {
	var models []models.LoanModel
	
	result := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&models)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get user loans: %w", result.Error)
	}
	
	loans := make([]domain.Loan, len(models))
	for i, model := range models {
		loans[i] = model.ToDomain()
	}
	
	return loans, nil
}

// GetLoansByType retrieves loans for a user filtered by loan type
func (r *loanRepository) GetLoansByType(ctx context.Context, userID string, loanType string) ([]domain.Loan, error) {
	var models []models.LoanModel
	
	result := r.db.WithContext(ctx).Where("user_id = ? AND type = ?", userID, loanType).Find(&models)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get loans by type: %w", result.Error)
	}
	
	loans := make([]domain.Loan, len(models))
	for i, model := range models {
		loans[i] = model.ToDomain()
	}
	
	return loans, nil
}

// GetLoansByInterestRateRange retrieves loans for a user within a specific interest rate range
func (r *loanRepository) GetLoansByInterestRateRange(ctx context.Context, userID string, minRate, maxRate float64) ([]domain.Loan, error) {
	var models []models.LoanModel
	
	result := r.db.WithContext(ctx).Where("user_id = ? AND interest_rate >= ? AND interest_rate <= ?", userID, minRate, maxRate).Find(&models)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get loans by interest rate range: %w", result.Error)
	}
	
	loans := make([]domain.Loan, len(models))
	for i, model := range models {
		loans[i] = model.ToDomain()
	}
	
	return loans, nil
}

// UpdateLoanBalance updates the remaining balance for a specific loan
func (r *loanRepository) UpdateLoanBalance(ctx context.Context, loanID string, newBalance float64) error {
	result := r.db.WithContext(ctx).Model(&models.LoanModel{}).
		Where("id = ?", loanID).
		Update("remaining_balance", newBalance)
	
	if result.Error != nil {
		return fmt.Errorf("failed to update loan balance: %w", result.Error)
	}
	
	if result.RowsAffected == 0 {
		return fmt.Errorf("loan with ID %s not found", loanID)
	}
	
	return nil
}

// GetNearPayoffLoans retrieves loans that are close to being paid off
func (r *loanRepository) GetNearPayoffLoans(ctx context.Context, userID string, threshold float64) ([]domain.Loan, error) {
	var models []models.LoanModel
	
	// Find loans where remaining_balance / principal_amount <= threshold
	result := r.db.WithContext(ctx).
		Where("user_id = ? AND (CAST(remaining_balance AS REAL) / CAST(principal_amount AS REAL)) <= ?", userID, threshold).
		Find(&models)
	
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get near payoff loans: %w", result.Error)
	}
	
	loans := make([]domain.Loan, len(models))
	for i, model := range models {
		loans[i] = model.ToDomain()
	}
	
	return loans, nil
}

// CalculateUserTotalDebt calculates the total remaining debt for a user
func (r *loanRepository) CalculateUserTotalDebt(ctx context.Context, userID string) (float64, error) {
	var total float64
	
	result := r.db.WithContext(ctx).Model(&models.LoanModel{}).
		Select("COALESCE(SUM(remaining_balance), 0)").
		Where("user_id = ?", userID).
		Scan(&total)
	
	if result.Error != nil {
		return 0, fmt.Errorf("failed to calculate total debt: %w", result.Error)
	}
	
	return total, nil
}

// CalculateUserMonthlyPayments calculates the total monthly loan payments for a user
func (r *loanRepository) CalculateUserMonthlyPayments(ctx context.Context, userID string) (float64, error) {
	var total float64
	
	result := r.db.WithContext(ctx).Model(&models.LoanModel{}).
		Select("COALESCE(SUM(monthly_payment), 0)").
		Where("user_id = ?", userID).
		Scan(&total)
	
	if result.Error != nil {
		return 0, fmt.Errorf("failed to calculate monthly payments: %w", result.Error)
	}
	
	return total, nil
}