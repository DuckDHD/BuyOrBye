package services

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/DuckDHD/BuyOrBye/internal/domain"
)


// debtCalculator implements the DebtCalculator interface
type debtCalculator struct {
	financeService FinanceService
}

// NewDebtCalculator creates a new DebtCalculator instance
func NewDebtCalculator(financeService FinanceService) DebtCalculator {
	return &debtCalculator{
		financeService: financeService,
	}
}

// CalculateTotalDebt sums all loan balances for a user
func (dc *debtCalculator) CalculateTotalDebt(ctx context.Context, userID string) (float64, error) {
	loans, err := dc.financeService.GetUserLoans(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to get user loans: %w", err)
	}

	totalDebt := 0.0
	for _, loan := range loans {
		totalDebt += loan.RemainingBalance
	}

	return totalDebt, nil
}

// ProjectDebtFreeDate estimates when the user will be debt-free based on current payments
func (dc *debtCalculator) ProjectDebtFreeDate(ctx context.Context, userID string) (time.Time, error) {
	loans, err := dc.financeService.GetUserLoans(ctx, userID)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to get user loans: %w", err)
	}

	if len(loans) == 0 {
		return time.Now(), nil // Already debt-free
	}

	// Calculate payoff time for each loan
	maxMonths := 0
	for _, loan := range loans {
		if loan.MonthlyPayment <= 0 {
			// If no monthly payment, assume minimum payment based on rate and end date
			if loan.EndDate.After(time.Now()) {
				monthsRemaining := int(time.Until(loan.EndDate).Hours() / (24 * 30))
				if monthsRemaining > maxMonths {
					maxMonths = monthsRemaining
				}
			}
			continue
		}

		months := dc.calculatePayoffMonths(loan.RemainingBalance, loan.InterestRate, loan.MonthlyPayment)
		if months > maxMonths {
			maxMonths = months
		}
	}

	// Add buffer for safety
	if maxMonths == 0 {
		maxMonths = 360 // Default to 30 years if calculations fail
	}

	projectedDate := time.Now().AddDate(0, maxMonths, 0)
	return projectedDate, nil
}

// SuggestPaymentStrategy recommends avalanche vs snowball approach
func (dc *debtCalculator) SuggestPaymentStrategy(ctx context.Context, userID string, extraPayment float64) (PaymentStrategy, error) {
	loans, err := dc.financeService.GetUserLoans(ctx, userID)
	if err != nil {
		return PaymentStrategy{}, fmt.Errorf("failed to get user loans: %w", err)
	}

	if len(loans) == 0 {
		return PaymentStrategy{
			UserID:       userID,
			StrategyType: "No Debt",
		}, nil
	}

	// Calculate both strategies
	avalanche := dc.calculateAvalancheStrategy(loans, extraPayment)
	snowball := dc.calculateSnowballStrategy(loans, extraPayment)

	// Compare strategies
	var recommendedStrategy PaymentStrategy
	var reason string

	interestDifference := avalanche.TotalInterestSaved - snowball.TotalInterestSaved
	timeDifference := avalanche.MonthsSaved - snowball.MonthsSaved

	// Recommend avalanche if significant interest savings
	if interestDifference > 500 || timeDifference > 6 {
		recommendedStrategy = avalanche
		reason = fmt.Sprintf("Avalanche method saves $%.2f in interest and %d months compared to snowball", 
			interestDifference, timeDifference)
	} else {
		// If difference is small, recommend snowball for psychological benefits
		recommendedStrategy = snowball
		reason = "Snowball method provides psychological benefits with similar financial outcomes"
	}

	recommendedStrategy.UserID = userID
	recommendedStrategy.ExtraPaymentAmount = extraPayment
	recommendedStrategy.RecommendedReason = reason

	return recommendedStrategy, nil
}

// CalculateInterestSavings calculates savings from making extra payments
func (dc *debtCalculator) CalculateInterestSavings(ctx context.Context, userID string, extraPayment float64) (InterestSavings, error) {
	loans, err := dc.financeService.GetUserLoans(ctx, userID)
	if err != nil {
		return InterestSavings{}, fmt.Errorf("failed to get user loans: %w", err)
	}

	// Calculate current scenario (no extra payment)
	currentInterest, currentMonths := dc.calculateTotalInterestAndTime(loans, 0)
	currentDebtFreeDate := time.Now().AddDate(0, currentMonths, 0)

	// Calculate with extra payment (distribute proportionally by balance)
	newInterest, newMonths := dc.calculateTotalInterestAndTime(loans, extraPayment)
	newDebtFreeDate := time.Now().AddDate(0, newMonths, 0)

	interestSaved := currentInterest - newInterest
	monthsSaved := currentMonths - newMonths

	// Calculate break-even point (when total extra payments equal interest saved)
	breakEvenMonths := 0
	if extraPayment > 0 {
		breakEvenMonths = int(math.Ceil(interestSaved / extraPayment))
	}

	// Recommend optimal extra payment (10-15% of total monthly payment)
	totalMinPayments := 0.0
	for _, loan := range loans {
		totalMinPayments += loan.MonthlyPayment
	}
	recommendedExtra := totalMinPayments * 0.125 // 12.5% of total payments

	savings := InterestSavings{
		UserID:                    userID,
		ExtraPaymentAmount:        extraPayment,
		CurrentTotalInterest:      currentInterest,
		NewTotalInterest:          newInterest,
		InterestSaved:             interestSaved,
		MonthsSaved:               monthsSaved,
		CurrentDebtFreeDate:       currentDebtFreeDate,
		NewDebtFreeDate:           newDebtFreeDate,
		BreakEvenMonths:           breakEvenMonths,
		RecommendedExtraPayment:   recommendedExtra,
	}

	return savings, nil
}

// GetDebtAnalysis provides comprehensive debt analysis
func (dc *debtCalculator) GetDebtAnalysis(ctx context.Context, userID string) (DebtAnalysis, error) {
	loans, err := dc.financeService.GetUserLoans(ctx, userID)
	if err != nil {
		return DebtAnalysis{}, fmt.Errorf("failed to get user loans: %w", err)
	}

	if len(loans) == 0 {
		return DebtAnalysis{
			UserID:           userID,
			DebtHealthStatus: "Excellent",
			Recommendations:  []string{"Great job! You have no debt."},
		}, nil
	}

	// Calculate totals and find extremes
	totalDebt := 0.0
	totalPayments := 0.0
	weightedRateSum := 0.0
	
	var highest, lowest, largest, smallest domain.Loan
	highest.InterestRate = -1 // Initialize to impossible value
	lowest.InterestRate = math.MaxFloat64
	largest.RemainingBalance = -1
	smallest.RemainingBalance = math.MaxFloat64

	for _, loan := range loans {
		totalDebt += loan.RemainingBalance
		totalPayments += loan.MonthlyPayment
		weightedRateSum += loan.InterestRate * loan.RemainingBalance

		if loan.InterestRate > highest.InterestRate {
			highest = loan
		}
		if loan.InterestRate < lowest.InterestRate {
			lowest = loan
		}
		if loan.RemainingBalance > largest.RemainingBalance {
			largest = loan
		}
		if loan.RemainingBalance < smallest.RemainingBalance {
			smallest = loan
		}
	}

	var weightedAvgRate float64
	if totalDebt > 0 {
		weightedAvgRate = weightedRateSum / totalDebt
	}

	// Get financial summary for debt-to-income ratio
	summary, err := dc.financeService.CalculateFinanceSummary(ctx, userID)
	if err != nil {
		return DebtAnalysis{}, fmt.Errorf("failed to get financial summary: %w", err)
	}

	debtToIncomeRatio := summary.DebtToIncomeRatio

	// Calculate total interest remaining and payoff time
	totalInterest, totalMonths := dc.calculateTotalInterestAndTime(loans, 0)

	// Create payoff projections
	var payoffProjections []LoanPaymentPlan
	for i, loan := range loans {
		months := dc.calculatePayoffMonths(loan.RemainingBalance, loan.InterestRate, loan.MonthlyPayment)
		interest := dc.calculateTotalInterest(loan.RemainingBalance, loan.InterestRate, loan.MonthlyPayment)
		
		projection := LoanPaymentPlan{
			LoanID:             loan.ID,
			Lender:             loan.Lender,
			CurrentBalance:     loan.RemainingBalance,
			InterestRate:       loan.InterestRate,
			MinimumPayment:     loan.MonthlyPayment,
			RecommendedPayment: loan.MonthlyPayment,
			PayoffOrder:        i + 1,
			MonthsToPayoff:     months,
			TotalInterest:      interest,
			PayoffDate:         time.Now().AddDate(0, months, 0),
		}
		payoffProjections = append(payoffProjections, projection)
	}

	// Determine debt health status
	debtHealthStatus := dc.calculateDebtHealthStatus(debtToIncomeRatio, weightedAvgRate, totalDebt, summary.MonthlyIncome)

	// Generate recommendations
	recommendations := dc.generateDebtRecommendations(debtHealthStatus, debtToIncomeRatio, weightedAvgRate, loans, summary)

	analysis := DebtAnalysis{
		UserID:                  userID,
		TotalDebt:               totalDebt,
		TotalMonthlyPayments:    totalPayments,
		WeightedAverageRate:     weightedAvgRate,
		HighestRateLoan:        LoanSummary{
			ID: highest.ID,
			Lender: highest.Lender,
			Type: highest.Type,
			RemainingBalance: highest.RemainingBalance,
			InterestRate: highest.InterestRate,
			MonthlyPayment: highest.MonthlyPayment,
		},
		LowestRateLoan:         LoanSummary{
			ID: lowest.ID,
			Lender: lowest.Lender,
			Type: lowest.Type,
			RemainingBalance: lowest.RemainingBalance,
			InterestRate: lowest.InterestRate,
			MonthlyPayment: lowest.MonthlyPayment,
		},
		LargestBalanceLoan:     LoanSummary{
			ID: largest.ID,
			Lender: largest.Lender,
			Type: largest.Type,
			RemainingBalance: largest.RemainingBalance,
			InterestRate: largest.InterestRate,
			MonthlyPayment: largest.MonthlyPayment,
		},
		SmallestBalanceLoan:    LoanSummary{
			ID: smallest.ID,
			Lender: smallest.Lender,
			Type: smallest.Type,
			RemainingBalance: smallest.RemainingBalance,
			InterestRate: smallest.InterestRate,
			MonthlyPayment: smallest.MonthlyPayment,
		},
		DebtToIncomeRatio:      debtToIncomeRatio * 100, // Convert to percentage
		MonthsToPayoff:         totalMonths,
		TotalInterestRemaining: totalInterest,
		DebtHealthStatus:       debtHealthStatus,
		Recommendations:        recommendations,
		PayoffProjections:      payoffProjections,
	}

	return analysis, nil
}

// CalculateMinimumPayment calculates the minimum required payment for a loan
func (dc *debtCalculator) CalculateMinimumPayment(principalAmount, interestRate float64, termMonths int) float64 {
	if principalAmount <= 0 || termMonths <= 0 {
		return 0
	}

	if interestRate <= 0 {
		// No interest, just divide principal by term
		return principalAmount / float64(termMonths)
	}

	// Convert annual rate to monthly
	monthlyRate := interestRate / 12 / 100
	
	// Calculate using standard loan payment formula
	// PMT = P * (r * (1 + r)^n) / ((1 + r)^n - 1)
	factor := math.Pow(1+monthlyRate, float64(termMonths))
	payment := principalAmount * (monthlyRate * factor) / (factor - 1)

	return payment
}

// Helper methods

func (dc *debtCalculator) calculatePayoffMonths(balance, interestRate, payment float64) int {
	if payment <= 0 || balance <= 0 {
		return 0
	}

	monthlyRate := interestRate / 12 / 100
	if monthlyRate <= 0 {
		// No interest case
		return int(math.Ceil(balance / payment))
	}

	// Check if payment covers interest
	monthlyInterest := balance * monthlyRate
	if payment <= monthlyInterest {
		return 999 // Will never pay off with current payment
	}

	// Calculate months using loan payoff formula
	// n = -log(1 - (P * r / PMT)) / log(1 + r)
	months := -math.Log(1-(balance*monthlyRate/payment)) / math.Log(1+monthlyRate)
	return int(math.Ceil(months))
}

func (dc *debtCalculator) calculateTotalInterest(balance, interestRate, payment float64) float64 {
	if payment <= 0 || balance <= 0 {
		return 0
	}

	months := dc.calculatePayoffMonths(balance, interestRate, payment)
	totalPayments := float64(months) * payment
	return totalPayments - balance
}

func (dc *debtCalculator) calculateTotalInterestAndTime(loans []domain.Loan, extraPayment float64) (float64, int) {
	totalInterest := 0.0
	maxMonths := 0

	// Distribute extra payment proportionally by balance
	totalBalance := 0.0
	for _, loan := range loans {
		totalBalance += loan.RemainingBalance
	}

	for _, loan := range loans {
		effectivePayment := loan.MonthlyPayment
		if extraPayment > 0 && totalBalance > 0 {
			proportion := loan.RemainingBalance / totalBalance
			effectivePayment += extraPayment * proportion
		}

		months := dc.calculatePayoffMonths(loan.RemainingBalance, loan.InterestRate, effectivePayment)
		interest := dc.calculateTotalInterest(loan.RemainingBalance, loan.InterestRate, effectivePayment)

		totalInterest += interest
		if months > maxMonths {
			maxMonths = months
		}
	}

	return totalInterest, maxMonths
}

func (dc *debtCalculator) calculateAvalancheStrategy(loans []domain.Loan, extraPayment float64) PaymentStrategy {
	// Sort by interest rate (highest first)
	sortedLoans := make([]domain.Loan, len(loans))
	copy(sortedLoans, loans)
	sort.Slice(sortedLoans, func(i, j int) bool {
		return sortedLoans[i].InterestRate > sortedLoans[j].InterestRate
	})

	totalInterest, totalMonths := dc.calculateStrategyPayoff(sortedLoans, extraPayment)

	var plans []LoanPaymentPlan
	for i, loan := range sortedLoans {
		payment := loan.MonthlyPayment
		if i == 0 { // First loan gets extra payment
			payment += extraPayment
		}

		months := dc.calculatePayoffMonths(loan.RemainingBalance, loan.InterestRate, payment)
		interest := dc.calculateTotalInterest(loan.RemainingBalance, loan.InterestRate, payment)

		plan := LoanPaymentPlan{
			LoanID:              loan.ID,
			Lender:              loan.Lender,
			CurrentBalance:      loan.RemainingBalance,
			InterestRate:        loan.InterestRate,
			MinimumPayment:      loan.MonthlyPayment,
			RecommendedPayment:  payment,
			PayoffOrder:         i + 1,
			MonthsToPayoff:      months,
			TotalInterest:       interest,
			PayoffDate:          time.Now().AddDate(0, months, 0),
		}
		plans = append(plans, plan)
	}

	// Calculate savings compared to no extra payment
	baseInterest, baseMonths := dc.calculateTotalInterestAndTime(loans, 0)

	return PaymentStrategy{
		StrategyType:          "Avalanche",
		PrioritizedLoans:      plans,
		TotalInterestSaved:    baseInterest - totalInterest,
		MonthsSaved:           baseMonths - totalMonths,
		MonthlyPaymentPlan:    dc.getTotalMonthlyPayments(loans) + extraPayment,
		ProjectedDebtFreeDate: time.Now().AddDate(0, totalMonths, 0),
	}
}

func (dc *debtCalculator) calculateSnowballStrategy(loans []domain.Loan, extraPayment float64) PaymentStrategy {
	// Sort by balance (smallest first)
	sortedLoans := make([]domain.Loan, len(loans))
	copy(sortedLoans, loans)
	sort.Slice(sortedLoans, func(i, j int) bool {
		return sortedLoans[i].RemainingBalance < sortedLoans[j].RemainingBalance
	})

	totalInterest, totalMonths := dc.calculateStrategyPayoff(sortedLoans, extraPayment)

	var plans []LoanPaymentPlan
	for i, loan := range sortedLoans {
		payment := loan.MonthlyPayment
		if i == 0 { // First loan gets extra payment
			payment += extraPayment
		}

		months := dc.calculatePayoffMonths(loan.RemainingBalance, loan.InterestRate, payment)
		interest := dc.calculateTotalInterest(loan.RemainingBalance, loan.InterestRate, payment)

		plan := LoanPaymentPlan{
			LoanID:              loan.ID,
			Lender:              loan.Lender,
			CurrentBalance:      loan.RemainingBalance,
			InterestRate:        loan.InterestRate,
			MinimumPayment:      loan.MonthlyPayment,
			RecommendedPayment:  payment,
			PayoffOrder:         i + 1,
			MonthsToPayoff:      months,
			TotalInterest:       interest,
			PayoffDate:          time.Now().AddDate(0, months, 0),
		}
		plans = append(plans, plan)
	}

	// Calculate savings compared to no extra payment
	baseInterest, baseMonths := dc.calculateTotalInterestAndTime(loans, 0)

	return PaymentStrategy{
		StrategyType:          "Snowball",
		PrioritizedLoans:      plans,
		TotalInterestSaved:    baseInterest - totalInterest,
		MonthsSaved:           baseMonths - totalMonths,
		MonthlyPaymentPlan:    dc.getTotalMonthlyPayments(loans) + extraPayment,
		ProjectedDebtFreeDate: time.Now().AddDate(0, totalMonths, 0),
	}
}

func (dc *debtCalculator) calculateStrategyPayoff(sortedLoans []domain.Loan, extraPayment float64) (float64, int) {
	// Simplified calculation - apply extra payment to first loan until paid off
	totalInterest := 0.0
	totalMonths := 0

	remainingExtra := extraPayment
	for _, loan := range sortedLoans {
		effectivePayment := loan.MonthlyPayment + remainingExtra
		months := dc.calculatePayoffMonths(loan.RemainingBalance, loan.InterestRate, effectivePayment)
		interest := dc.calculateTotalInterest(loan.RemainingBalance, loan.InterestRate, effectivePayment)

		totalInterest += interest
		if months > totalMonths {
			totalMonths = months
		}

		// After first loan is paid off, extra payment goes to next loan
		remainingExtra = 0
	}

	return totalInterest, totalMonths
}

func (dc *debtCalculator) getTotalMonthlyPayments(loans []domain.Loan) float64 {
	total := 0.0
	for _, loan := range loans {
		total += loan.MonthlyPayment
	}
	return total
}

func (dc *debtCalculator) calculateDebtHealthStatus(debtToIncomeRatio, avgRate, totalDebt, monthlyIncome float64) string {
	// Poor health indicators
	if debtToIncomeRatio > 0.50 { // > 50% DTI
		return "Poor"
	}
	if avgRate > 20.0 { // > 20% average rate (high-interest debt)
		return "Poor"
	}
	if monthlyIncome > 0 && totalDebt > monthlyIncome*10 { // Debt > 10x monthly income
		return "Poor"
	}

	// Fair health indicators
	if debtToIncomeRatio > 0.36 { // > 36% DTI
		return "Fair"
	}
	if avgRate > 10.0 { // > 10% average rate
		return "Fair"
	}

	// Good health indicators
	if debtToIncomeRatio > 0.20 { // > 20% DTI
		return "Good"
	}
	if avgRate > 6.0 { // > 6% average rate
		return "Good"
	}

	// Excellent health
	return "Excellent"
}

func (dc *debtCalculator) generateDebtRecommendations(healthStatus string, dtiRatio, avgRate float64, loans []domain.Loan, summary domain.FinanceSummary) []string {
	var recommendations []string

	switch healthStatus {
	case "Poor":
		recommendations = append(recommendations, "Your debt levels are concerning. Immediate action required.")
		if dtiRatio > 0.50 {
			recommendations = append(recommendations, "Your debt-to-income ratio exceeds 50%. Focus on debt reduction before new purchases.")
		}
		if avgRate > 15.0 {
			recommendations = append(recommendations, "Consider debt consolidation to reduce high interest rates.")
		}
		recommendations = append(recommendations, "Avoid taking on any new debt.")
		recommendations = append(recommendations, "Consider credit counseling services.")

	case "Fair":
		recommendations = append(recommendations, "Your debt is manageable but needs attention.")
		if dtiRatio > 0.36 {
			recommendations = append(recommendations, "Work to reduce debt-to-income ratio below 36%.")
		}
		recommendations = append(recommendations, "Make extra payments when possible to reduce interest.")
		recommendations = append(recommendations, "Avoid new debt until current debt is reduced.")

	case "Good":
		recommendations = append(recommendations, "Your debt levels are reasonable.")
		recommendations = append(recommendations, "Consider making extra payments to save on interest.")
		recommendations = append(recommendations, "Maintain current payment discipline.")

	case "Excellent":
		recommendations = append(recommendations, "Excellent debt management!")
		recommendations = append(recommendations, "Consider using surplus for investments after maintaining emergency fund.")
	}

	// Add strategy-specific recommendations
	if len(loans) > 1 {
		highestRate := 0.0
		for _, loan := range loans {
			if loan.InterestRate > highestRate {
				highestRate = loan.InterestRate
			}
		}

		if highestRate > avgRate*1.5 {
			recommendations = append(recommendations, "Focus extra payments on highest interest rate debt first (avalanche method).")
		} else {
			recommendations = append(recommendations, "Consider snowball method to build momentum by paying smallest balances first.")
		}
	}

	// Add payment amount recommendations
	if summary.DisposableIncome > 100 {
		extraPayment := summary.DisposableIncome * 0.5 // Suggest using 50% of disposable income
		recommendations = append(recommendations, fmt.Sprintf("Consider making an extra $%.2f monthly payment to accelerate debt payoff.", extraPayment))
	}

	return recommendations
}