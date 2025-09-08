package domain

import (
	"fmt"
	"strings"
	"time"
)

// FinanceSummary represents an aggregated view of a user's financial situation
type FinanceSummary struct {
	UserID              string
	MonthlyIncome       float64
	MonthlyExpenses     float64
	MonthlyLoanPayments float64
	DisposableIncome    float64
	DebtToIncomeRatio   float64
	SavingsRate         float64
	FinancialHealth     string
	BudgetRemaining     float64
	UpdatedAt           time.Time
}

// Financial health constants
const (
	HealthExcellent = "Excellent"
	HealthGood      = "Good"
	HealthFair      = "Fair"
	HealthPoor      = "Poor"
)

// Financial health thresholds
const (
	HealthyDebtToIncomeRatio = 0.36  // 36% is considered healthy
	ExcellentDebtToIncomeRatio = 0.28 // 28% is excellent
	PoorDebtToIncomeRatio = 0.50     // 50% is poor
	MinimumSavingsRate = 0.20        // 20% savings rate target
	GoodSavingsRate = 0.15           // 15% is decent
	FairSavingsRate = 0.10           // 10% is fair
)

// ValidHealthValues contains all valid financial health values
var ValidHealthValues = []string{
	HealthExcellent,
	HealthGood,
	HealthFair,
	HealthPoor,
}

// CalculateHealth calculates and returns the financial health rating based on various metrics
func (fs *FinanceSummary) CalculateHealth() string {
	// Poor health conditions (highest priority)
	if fs.DebtToIncomeRatio > PoorDebtToIncomeRatio {
		return HealthPoor
	}
	
	if fs.DisposableIncome < 0 {
		return HealthPoor // Overspending is poor
	}

	// Special case: zero income scenario should be poor
	if fs.MonthlyIncome == 0 && fs.DebtToIncomeRatio == 0 && fs.SavingsRate == 0 && fs.DisposableIncome == 0 {
		return HealthPoor
	}

	// Fair health conditions
	if fs.DebtToIncomeRatio > HealthyDebtToIncomeRatio {
		return HealthFair
	}
	
	if fs.SavingsRate < FairSavingsRate {
		return HealthFair // Low savings rate
	}

	// Excellent health conditions
	if fs.DebtToIncomeRatio <= ExcellentDebtToIncomeRatio && fs.SavingsRate >= MinimumSavingsRate {
		return HealthExcellent
	}

	// Default to good health
	return HealthGood
}

// GetHealthScore returns a numerical score for the financial health (4=Excellent, 3=Good, 2=Fair, 1=Poor, 0=Unknown)
func (fs *FinanceSummary) GetHealthScore() int {
	switch fs.FinancialHealth {
	case HealthExcellent:
		return 4
	case HealthGood:
		return 3
	case HealthFair:
		return 2
	case HealthPoor:
		return 1
	default:
		return 0
	}
}

// CalculateAffordability returns the maximum amount the user can afford for a purchase
// based on their financial situation
func (fs *FinanceSummary) CalculateAffordability() float64 {
	if fs.DisposableIncome <= 0 {
		return 0.0 // No affordability if overspending
	}

	// Base affordability on debt-to-income ratio and disposable income
	switch {
	case fs.DebtToIncomeRatio <= ExcellentDebtToIncomeRatio:
		// Excellent DTI: 3x disposable income
		return fs.DisposableIncome * 3.0
	case fs.DebtToIncomeRatio <= HealthyDebtToIncomeRatio:
		// Good DTI: 3x disposable income
		return fs.DisposableIncome * 3.0
	case fs.DebtToIncomeRatio <= PoorDebtToIncomeRatio:
		// Fair DTI: 2x disposable income
		return fs.DisposableIncome * 2.0
	default:
		// High DTI: 0.5x disposable income (conservative)
		return fs.DisposableIncome * 0.5
	}
}

// GetBudgetStatus returns a string describing the budget status
func (fs *FinanceSummary) GetBudgetStatus() string {
	switch {
	case fs.BudgetRemaining > 0:
		return "Surplus"
	case fs.BudgetRemaining == 0:
		return "Break Even"
	default:
		return "Deficit"
	}
}

// IsOverspending returns true if total expenses and loans exceed income
func (fs *FinanceSummary) IsOverspending() bool {
	totalOutgoings := fs.MonthlyExpenses + fs.MonthlyLoanPayments
	return totalOutgoings > fs.MonthlyIncome
}

// GetRecommendation returns personalized financial recommendations based on the summary
func (fs *FinanceSummary) GetRecommendation() string {
	recommendations := []string{}

	switch fs.FinancialHealth {
	case HealthExcellent:
		recommendations = append(recommendations, "Your finances are excellent! continue your current approach.")
		recommendations = append(recommendations, "consider investing surplus funds for long-term growth")
		
	case HealthGood:
		recommendations = append(recommendations, "Your finances are in good shape.")
		recommendations = append(recommendations, "Focus on building your emergency fund to 6 months of expenses.")
		recommendations = append(recommendations, "optimize your spending for better results.")
		
	case HealthFair:
		if fs.DebtToIncomeRatio > HealthyDebtToIncomeRatio {
			recommendations = append(recommendations, "Focus on reduce debt to improve your debt-to-income ratio below 36%.")
		}
		if fs.SavingsRate < GoodSavingsRate {
			recommendations = append(recommendations, "Work on increase savings rate to at least 15%.")
		}
		recommendations = append(recommendations, "Review your expenses to find areas for improvement.")
		
	case HealthPoor:
		if fs.DisposableIncome < 0 {
			recommendations = append(recommendations, "You're overspending. Prioritize reduce expenses immediately.")
			recommendations = append(recommendations, "increase income through side work or better employment")
		}
		if fs.DebtToIncomeRatio > PoorDebtToIncomeRatio {
			recommendations = append(recommendations, "reduce expenses to free up money for debt payments.")
			recommendations = append(recommendations, "Consider debt consolidation or payment plans.")
		}
		recommendations = append(recommendations, "Focus on essential expenses only until your situation improves.")
	}

	return strings.Join(recommendations, "; ")
}

// Validate validates the FinanceSummary struct and returns an error if validation fails
func (fs *FinanceSummary) Validate() error {
	var errors []string

	// Validate required fields
	if fs.UserID == "" {
		errors = append(errors, "user ID is required")
	}

	if fs.MonthlyIncome < 0 {
		errors = append(errors, "monthly income cannot be negative")
	}

	if fs.MonthlyExpenses < 0 {
		errors = append(errors, "monthly expenses cannot be negative")
	}

	if fs.MonthlyLoanPayments < 0 {
		errors = append(errors, "monthly loan payments cannot be negative")
	}

	if fs.FinancialHealth == "" {
		errors = append(errors, "financial health is required")
	} else if !isValidHealthValue(fs.FinancialHealth) {
		errors = append(errors, "financial health must be one of: Excellent, Good, Fair, Poor")
	}

	if fs.UpdatedAt.IsZero() {
		errors = append(errors, "updated at is required")
	}

	if len(errors) > 0 {
		return fmt.Errorf("validation failed: %s", strings.Join(errors, "; "))
	}

	return nil
}

// CalculateEmergencyFundTarget calculates the target emergency fund based on monthly expenses
func (fs *FinanceSummary) CalculateEmergencyFundTarget() float64 {
	// Target 6 months of expenses as emergency fund
	return fs.MonthlyExpenses * 6.0
}

// GetDebtToIncomeLevel returns a descriptive level for the debt-to-income ratio
func (fs *FinanceSummary) GetDebtToIncomeLevel() string {
	switch {
	case fs.DebtToIncomeRatio <= ExcellentDebtToIncomeRatio:
		return "Excellent"
	case fs.DebtToIncomeRatio <= HealthyDebtToIncomeRatio:
		return "Good"
	case fs.DebtToIncomeRatio <= PoorDebtToIncomeRatio:
		return "Fair"
	default:
		return "Poor"
	}
}

// GetSavingsRateLevel returns a descriptive level for the savings rate
func (fs *FinanceSummary) GetSavingsRateLevel() string {
	switch {
	case fs.SavingsRate >= MinimumSavingsRate:
		return "Excellent"
	case fs.SavingsRate >= GoodSavingsRate:
		return "Good"
	case fs.SavingsRate >= FairSavingsRate:
		return "Fair"
	case fs.SavingsRate >= 0:
		return "Poor"
	default:
		return "Critical" // Negative savings rate
	}
}

// isValidHealthValue checks if the provided health value is valid
func isValidHealthValue(health string) bool {
	for _, validHealth := range ValidHealthValues {
		if health == validHealth {
			return true
		}
	}
	return false
}