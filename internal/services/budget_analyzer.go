package services

import (
	"context"
	"fmt"
	"sort"

	"github.com/DuckDHD/BuyOrBye/internal/domain"
)


// budgetAnalyzer implements the BudgetAnalyzer interface
type budgetAnalyzer struct {
	financeService FinanceService
}

// NewBudgetAnalyzer creates a new BudgetAnalyzer instance
func NewBudgetAnalyzer(financeService FinanceService) BudgetAnalyzer {
	return &budgetAnalyzer{
		financeService: financeService,
	}
}

// AnalyzeBudget identifies overspending categories and budget issues
func (ba *budgetAnalyzer) AnalyzeBudget(ctx context.Context, userID string) (BudgetAnalysis, error) {
	// Get financial summary
	summary, err := ba.financeService.CalculateFinanceSummary(ctx, userID)
	if err != nil {
		return BudgetAnalysis{}, fmt.Errorf("failed to get financial summary: %w", err)
	}

	// Get spending insights for category analysis
	insights, err := ba.GetSpendingInsights(ctx, userID)
	if err != nil {
		return BudgetAnalysis{}, fmt.Errorf("failed to get spending insights: %w", err)
	}

	// Determine budget status
	var budgetStatus string
	var overspendingAmount float64

	if summary.DisposableIncome > 0 {
		budgetStatus = "Surplus"
	} else if summary.DisposableIncome == 0 {
		budgetStatus = "Balanced"
	} else {
		budgetStatus = "Deficit"
		overspendingAmount = -summary.DisposableIncome
	}

	// Identify overspending categories using general guidelines
	var overspendingCategories []CategoryOverspending
	totalIncome := summary.MonthlyIncome

	// Common category spending guidelines (as percentage of income)
	categoryLimits := map[string]float64{
		"Housing":       0.30, // 30% of income
		"Transportation": 0.15, // 15% of income
		"Food":          0.12, // 12% of income
		"Groceries":     0.12, // 12% of income
		"Utilities":     0.08, // 8% of income
		"Insurance":     0.05, // 5% of income
		"Entertainment": 0.05, // 5% of income
		"Dining Out":    0.05, // 5% of income
		"Shopping":      0.05, // 5% of income
		"Healthcare":    0.05, // 5% of income
		"Other":         0.05, // 5% of income
	}

	for _, spending := range insights.CategoryBreakdown {
		if limit, exists := categoryLimits[spending.Category]; exists {
			recommendedMax := totalIncome * limit
			if spending.MonthlyAmount > recommendedMax {
				overspending := CategoryOverspending{
					Category:        spending.Category,
					MonthlySpent:    spending.MonthlyAmount,
					RecommendedMax:  recommendedMax,
					OverspendAmount: spending.MonthlyAmount - recommendedMax,
					Percentage:      spending.PercentageOfIncome,
				}
				overspendingCategories = append(overspendingCategories, overspending)
			}
		}
	}

	// Sort overspending categories by amount
	sort.Slice(overspendingCategories, func(i, j int) bool {
		return overspendingCategories[i].OverspendAmount > overspendingCategories[j].OverspendAmount
	})

	// Generate recommended actions
	var recommendedActions []string
	
	if budgetStatus == "Deficit" {
		recommendedActions = append(recommendedActions, "Reduce expenses immediately to avoid debt accumulation")
		
		if len(overspendingCategories) > 0 {
			topCategory := overspendingCategories[0]
			recommendedActions = append(recommendedActions, 
				fmt.Sprintf("Focus on reducing %s expenses by $%.2f", topCategory.Category, topCategory.OverspendAmount))
		}
		
		recommendedActions = append(recommendedActions, "Consider increasing income through side work or better employment")
	} else if budgetStatus == "Balanced" {
		recommendedActions = append(recommendedActions, "Build an emergency fund with surplus funds")
		recommendedActions = append(recommendedActions, "Look for small optimizations to create savings")
	} else {
		recommendedActions = append(recommendedActions, "Allocate surplus to savings and investments")
		
		if len(overspendingCategories) > 0 {
			recommendedActions = append(recommendedActions, "Optimize overspending categories to increase savings")
		}
	}

	// Calculate budget health score (1-10)
	healthScore := calculateBudgetHealthScore(summary, len(overspendingCategories))

	analysis := BudgetAnalysis{
		UserID:                 userID,
		TotalMonthlyIncome:     summary.MonthlyIncome,
		TotalMonthlyExpenses:   summary.MonthlyExpenses,
		MonthlyLoanPayments:    summary.MonthlyLoanPayments,
		BudgetStatus:           budgetStatus,
		OverspendingAmount:     overspendingAmount,
		OverspendingCategories: overspendingCategories,
		RecommendedActions:     recommendedActions,
		BudgetHealthScore:      healthScore,
	}

	return analysis, nil
}

// GetSpendingInsights provides category-wise spending analysis
func (ba *budgetAnalyzer) GetSpendingInsights(ctx context.Context, userID string) (SpendingInsights, error) {
	// Get expenses and income data
	expenses, err := ba.financeService.GetUserExpenses(ctx, userID)
	if err != nil {
		return SpendingInsights{}, fmt.Errorf("failed to get expenses: %w", err)
	}

	summary, err := ba.financeService.CalculateFinanceSummary(ctx, userID)
	if err != nil {
		return SpendingInsights{}, fmt.Errorf("failed to get financial summary: %w", err)
	}

	// Group expenses by category
	categoryMap := make(map[string]*CategorySpending)

	for _, expense := range expenses {
		// Normalize expense to monthly
		monthlyAmount, err := ba.financeService.NormalizeToMonthly(expense.Amount, expense.Frequency)
		if err != nil {
			continue // Skip invalid frequencies
		}

		if category, exists := categoryMap[expense.Category]; exists {
			category.MonthlyAmount += monthlyAmount
			category.ExpenseCount++
		} else {
			categoryMap[expense.Category] = &CategorySpending{
				Category:      expense.Category,
				MonthlyAmount: monthlyAmount,
				ExpenseCount:  1,
				IsFixed:       expense.IsFixed,
			}
		}
	}

	// Convert map to slice and calculate percentages
	var categoryBreakdown []CategorySpending
	totalSpending := summary.MonthlyExpenses
	totalIncome := summary.MonthlyIncome

	for _, category := range categoryMap {
		category.AverageAmount = category.MonthlyAmount / float64(category.ExpenseCount)
		
		if totalSpending > 0 {
			category.PercentageOfTotal = (category.MonthlyAmount / totalSpending) * 100
		}
		
		if totalIncome > 0 {
			category.PercentageOfIncome = (category.MonthlyAmount / totalIncome) * 100
		}
		
		categoryBreakdown = append(categoryBreakdown, *category)
	}

	// Sort by monthly amount
	sort.Slice(categoryBreakdown, func(i, j int) bool {
		return categoryBreakdown[i].MonthlyAmount > categoryBreakdown[j].MonthlyAmount
	})

	// Find highest and lowest categories
	var highest, lowest CategorySpending
	if len(categoryBreakdown) > 0 {
		highest = categoryBreakdown[0]
		lowest = categoryBreakdown[len(categoryBreakdown)-1]
	}

	// Calculate variable vs fixed ratio
	var variableTotal, fixedTotal float64
	for _, category := range categoryBreakdown {
		if category.IsFixed {
			fixedTotal += category.MonthlyAmount
		} else {
			variableTotal += category.MonthlyAmount
		}
	}

	var variableVsFixedRatio float64
	if fixedTotal > 0 {
		variableVsFixedRatio = variableTotal / fixedTotal
	}

	// Determine spending efficiency
	spendingEfficiency := determineSpendingEfficiency(summary, variableVsFixedRatio)

	insights := SpendingInsights{
		UserID:               userID,
		TotalMonthlySpending: totalSpending,
		CategoryBreakdown:    categoryBreakdown,
		HighestCategory:      highest,
		LowestCategory:       lowest,
		VariableVsFixedRatio: variableVsFixedRatio,
		SpendingEfficiency:   spendingEfficiency,
	}

	return insights, nil
}

// RecommendSavings applies the 50/30/20 rule and provides savings recommendations
func (ba *budgetAnalyzer) RecommendSavings(ctx context.Context, userID string) (SavingsRecommendation, error) {
	summary, err := ba.financeService.CalculateFinanceSummary(ctx, userID)
	if err != nil {
		return SavingsRecommendation{}, fmt.Errorf("failed to get financial summary: %w", err)
	}

	monthlyIncome := summary.MonthlyIncome

	// Calculate ideal 50/30/20 breakdown
	idealNeeds := monthlyIncome * 0.50
	idealWants := monthlyIncome * 0.30
	idealSavings := monthlyIncome * 0.20

	rule5030020 := FiftyThirtyTwentyBreakdown{
		Needs:         idealNeeds,
		Wants:         idealWants,
		Savings:       idealSavings,
		NeedsPercent:  50.0,
		WantsPercent:  30.0,
		SavingsPercent: 20.0,
	}

	// Calculate current allocation (approximation)
	// Needs = Fixed expenses + Loan payments
	// Wants = Variable expenses  
	// Savings = Disposable income (if positive)

	expenses, err := ba.financeService.GetUserExpenses(ctx, userID)
	if err != nil {
		return SavingsRecommendation{}, fmt.Errorf("failed to get expenses: %w", err)
	}

	var currentNeeds, currentWants float64
	for _, expense := range expenses {
		monthlyAmount, err := ba.financeService.NormalizeToMonthly(expense.Amount, expense.Frequency)
		if err != nil {
			continue
		}

		if expense.IsFixed {
			currentNeeds += monthlyAmount
		} else {
			currentWants += monthlyAmount
		}
	}

	// Add loan payments to needs
	currentNeeds += summary.MonthlyLoanPayments

	// Current savings is disposable income if positive
	currentSavings := summary.DisposableIncome
	if currentSavings < 0 {
		currentSavings = 0
	}

	// Calculate percentages
	var currentNeedsPercent, currentWantsPercent, currentSavingsPercent float64
	if monthlyIncome > 0 {
		currentNeedsPercent = (currentNeeds / monthlyIncome) * 100
		currentWantsPercent = (currentWants / monthlyIncome) * 100
		currentSavingsPercent = (currentSavings / monthlyIncome) * 100
	}

	currentAllocation := FiftyThirtyTwentyBreakdown{
		Needs:          currentNeeds,
		Wants:          currentWants,
		Savings:        currentSavings,
		NeedsPercent:   currentNeedsPercent,
		WantsPercent:   currentWantsPercent,
		SavingsPercent: currentSavingsPercent,
	}

	// Calculate savings gap
	savingsGap := idealSavings - currentSavings

	// Generate recommendations
	var recommendedActions []string
	
	if savingsGap > 0 {
		recommendedActions = append(recommendedActions, 
			fmt.Sprintf("Increase savings by $%.2f per month to reach 20%% target", savingsGap))

		if currentWantsPercent > 30 {
			excessWants := currentWants - idealWants
			recommendedActions = append(recommendedActions, 
				fmt.Sprintf("Reduce discretionary spending by $%.2f to align with 30%% guideline", excessWants))
		}

		if currentNeedsPercent > 50 {
			excessNeeds := currentNeeds - idealNeeds
			recommendedActions = append(recommendedActions, 
				fmt.Sprintf("Consider reducing essential expenses by $%.2f or increasing income", excessNeeds))
		}
	} else {
		recommendedActions = append(recommendedActions, "Great job! You're meeting the 20% savings target")
		recommendedActions = append(recommendedActions, "Consider investing surplus savings for long-term growth")
	}

	// Calculate achievability score
	achievabilityScore := calculateAchievabilityScore(summary, savingsGap)

	recommendation := SavingsRecommendation{
		UserID:             userID,
		MonthlyIncome:      monthlyIncome,
		Rule5030020:        rule5030020,
		CurrentAllocation:  currentAllocation,
		SavingsGap:         savingsGap,
		RecommendedActions: recommendedActions,
		AchievabilityScore: achievabilityScore,
	}

	return recommendation, nil
}

// IdentifyUnnecessaryExpenses finds expenses that can be optimized
func (ba *budgetAnalyzer) IdentifyUnnecessaryExpenses(ctx context.Context, userID string) ([]ExpenseOptimization, error) {
	expenses, err := ba.financeService.GetUserExpenses(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get expenses: %w", err)
	}

	insights, err := ba.GetSpendingInsights(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get spending insights: %w", err)
	}

	var optimizations []ExpenseOptimization

	// Define optimization rules
	for _, expense := range expenses {
		monthlyAmount, err := ba.financeService.NormalizeToMonthly(expense.Amount, expense.Frequency)
		if err != nil {
			continue
		}

		// Skip if amount is too small to bother with
		if monthlyAmount < 5.0 {
			continue
		}

		optimization := ExpenseOptimization{
			ExpenseID:     expense.ID,
			Category:      expense.Category,
			Description:   expense.Name,
			CurrentAmount: monthlyAmount,
		}

		// Apply optimization rules by category
		switch expense.Category {
		case "Entertainment", "Dining Out":
			// Suggest 25% reduction for entertainment and dining
			optimization.RecommendedAmount = monthlyAmount * 0.75
			optimization.PotentialSavings = monthlyAmount * 0.25
			optimization.OptimizationType = "Reduce"
			optimization.Priority = 3
			optimization.Reasoning = "Entertainment and dining expenses can often be reduced without major lifestyle changes"

		case "Shopping":
			// Suggest 30% reduction for shopping
			optimization.RecommendedAmount = monthlyAmount * 0.70
			optimization.PotentialSavings = monthlyAmount * 0.30
			optimization.OptimizationType = "Reduce"
			optimization.Priority = 4
			optimization.Reasoning = "Non-essential shopping can be reduced by being more selective with purchases"

		case "Subscriptions", "Streaming":
			// Suggest eliminating or reducing subscriptions
			if monthlyAmount < 50 {
				optimization.RecommendedAmount = 0
				optimization.PotentialSavings = monthlyAmount
				optimization.OptimizationType = "Eliminate"
			} else {
				optimization.RecommendedAmount = monthlyAmount * 0.50
				optimization.PotentialSavings = monthlyAmount * 0.50
				optimization.OptimizationType = "Reduce"
			}
			optimization.Priority = 5
			optimization.Reasoning = "Review active subscriptions and cancel unused services"

		case "Transportation":
			// Suggest optimization for high transportation costs
			if monthlyAmount > 300 {
				optimization.RecommendedAmount = monthlyAmount * 0.85
				optimization.PotentialSavings = monthlyAmount * 0.15
				optimization.OptimizationType = "Substitute"
				optimization.Priority = 2
				optimization.Reasoning = "Consider carpooling, public transit, or more fuel-efficient transportation"
			}

		case "Utilities":
			// Suggest small reduction in utilities through efficiency
			if monthlyAmount > 100 {
				optimization.RecommendedAmount = monthlyAmount * 0.90
				optimization.PotentialSavings = monthlyAmount * 0.10
				optimization.OptimizationType = "Reduce"
				optimization.Priority = 2
				optimization.Reasoning = "Reduce utility costs through energy-efficient practices"
			}
		}

		// Only include if there are potential savings
		if optimization.PotentialSavings > 0 {
			optimizations = append(optimizations, optimization)
		}

		// Check for high-frequency expenses that might indicate waste
		categorySpending := getCategorySpending(insights.CategoryBreakdown, expense.Category)
		if categorySpending != nil && categorySpending.ExpenseCount > 10 {
			// High frequency might indicate inefficient spending
			optimization.Priority++
			optimization.Reasoning += "; High frequency suggests potential for consolidation or reduction"
		}
	}

	// Sort by priority (highest first) then by potential savings
	sort.Slice(optimizations, func(i, j int) bool {
		if optimizations[i].Priority == optimizations[j].Priority {
			return optimizations[i].PotentialSavings > optimizations[j].PotentialSavings
		}
		return optimizations[i].Priority > optimizations[j].Priority
	})

	return optimizations, nil
}

// Helper functions

func calculateBudgetHealthScore(summary domain.FinanceSummary, overspendingCategories int) int {
	score := 10

	// Deduct points for negative disposable income
	if summary.DisposableIncome < 0 {
		score -= 4 // Major issue
	} else if summary.DisposableIncome < summary.MonthlyIncome*0.05 { // Less than 5% disposable
		score -= 2
	}

	// Deduct points for high debt-to-income ratio
	if summary.DebtToIncomeRatio > 0.50 {
		score -= 3
	} else if summary.DebtToIncomeRatio > 0.36 {
		score -= 1
	}

	// Deduct points for overspending categories
	score -= overspendingCategories / 2

	// Ensure minimum score of 1
	if score < 1 {
		score = 1
	}

	return score
}

func determineSpendingEfficiency(summary domain.FinanceSummary, variableFixedRatio float64) string {
	// Good efficiency: Low variable expenses relative to fixed, good savings rate
	if summary.SavingsRate >= 0.15 && variableFixedRatio < 0.5 {
		return "Efficient"
	}
	
	// Poor efficiency: High variable expenses, negative disposable income
	if summary.DisposableIncome < 0 || variableFixedRatio > 1.5 {
		return "Wasteful"
	}

	return "Moderate"
}

func calculateAchievabilityScore(summary domain.FinanceSummary, savingsGap float64) int {
	// Base score of 5
	score := 5

	// If already meeting savings target
	if savingsGap <= 0 {
		return 10
	}

	// Adjust based on gap relative to income
	if summary.MonthlyIncome > 0 {
		gapPercentage := savingsGap / summary.MonthlyIncome

		if gapPercentage < 0.05 { // Less than 5% of income
			score += 3
		} else if gapPercentage < 0.10 { // Less than 10% of income
			score += 1
		} else if gapPercentage > 0.20 { // More than 20% of income
			score -= 3
		}
	}

	// Adjust based on current financial health
	if summary.DisposableIncome > 0 {
		score += 2
	} else {
		score -= 3
	}

	// Ensure score is between 1 and 10
	if score > 10 {
		score = 10
	}
	if score < 1 {
		score = 1
	}

	return score
}

func getCategorySpending(categories []CategorySpending, categoryName string) *CategorySpending {
	for i := range categories {
		if categories[i].Category == categoryName {
			return &categories[i]
		}
	}
	return nil
}