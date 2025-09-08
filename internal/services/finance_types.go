package services

import "time"

// Budget Analysis Types

// BudgetAnalysis represents the result of budget analysis
type BudgetAnalysis struct {
	UserID              string
	TotalMonthlyIncome  float64
	TotalMonthlyExpenses float64
	MonthlyLoanPayments float64
	BudgetStatus        string // "Surplus", "Balanced", "Deficit"
	OverspendingAmount  float64
	OverspendingCategories []CategoryOverspending
	RecommendedActions  []string
	BudgetHealthScore   int // 1-10 scale
}

// CategoryOverspending represents overspending in a specific category
type CategoryOverspending struct {
	Category        string
	MonthlySpent    float64
	RecommendedMax  float64
	OverspendAmount float64
	Percentage      float64 // Percentage of total income
}

// SpendingInsights provides detailed spending analysis
type SpendingInsights struct {
	UserID                string
	TotalMonthlySpending  float64
	CategoryBreakdown     []CategorySpending
	HighestCategory       CategorySpending
	LowestCategory        CategorySpending
	VariableVsFixedRatio  float64 // Variable expenses / Fixed expenses
	SpendingEfficiency    string  // "Efficient", "Moderate", "Wasteful"
}

// CategorySpending represents spending in a specific category
type CategorySpending struct {
	Category           string
	MonthlyAmount      float64
	PercentageOfIncome float64
	PercentageOfTotal  float64
	ExpenseCount       int
	AverageAmount      float64
	IsFixed            bool
}

// SavingsRecommendation based on 50/30/20 rule
type SavingsRecommendation struct {
	UserID                 string
	MonthlyIncome          float64
	Rule5030020            FiftyThirtyTwentyBreakdown
	CurrentAllocation      FiftyThirtyTwentyBreakdown
	SavingsGap             float64
	RecommendedActions     []string
	AchievabilityScore     int // 1-10, how achievable the recommendations are
}

// FiftyThirtyTwentyBreakdown represents the 50/30/20 budget rule
type FiftyThirtyTwentyBreakdown struct {
	Needs       float64 // 50% - Essential expenses
	Wants       float64 // 30% - Discretionary spending
	Savings     float64 // 20% - Savings and debt repayment
	NeedsPercent float64
	WantsPercent float64
	SavingsPercent float64
}

// ExpenseOptimization represents an optimization opportunity
type ExpenseOptimization struct {
	ExpenseID          string
	Category           string
	Description        string
	CurrentAmount      float64
	RecommendedAmount  float64
	PotentialSavings   float64
	OptimizationType   string // "Reduce", "Eliminate", "Substitute"
	Priority           int    // 1-5, 5 being highest priority
	Reasoning          string
}

// Debt Calculator Types

// PaymentStrategy represents a debt payment strategy recommendation
type PaymentStrategy struct {
	UserID                string
	StrategyType          string // "Avalanche", "Snowball", "Custom"
	ExtraPaymentAmount    float64
	PrioritizedLoans      []LoanPaymentPlan
	TotalInterestSaved    float64
	MonthsSaved           int
	RecommendedReason     string
	MonthlyPaymentPlan    float64
	ProjectedDebtFreeDate time.Time
}

// LoanPaymentPlan represents a payment plan for a specific loan
type LoanPaymentPlan struct {
	LoanID              string
	Lender              string
	CurrentBalance      float64
	InterestRate        float64
	MinimumPayment      float64
	RecommendedPayment  float64
	PayoffOrder         int
	MonthsToPayoff      int
	TotalInterest       float64
	PayoffDate          time.Time
}

// InterestSavings represents the savings from making extra payments
type InterestSavings struct {
	UserID                    string
	ExtraPaymentAmount        float64
	CurrentTotalInterest      float64
	NewTotalInterest          float64
	InterestSaved             float64
	MonthsSaved               int
	CurrentDebtFreeDate       time.Time
	NewDebtFreeDate           time.Time
	BreakEvenMonths           int // Months to break even on extra payment
	RecommendedExtraPayment   float64
}

// DebtAnalysis provides comprehensive debt analysis
type DebtAnalysis struct {
	UserID                    string
	TotalDebt                 float64
	TotalMonthlyPayments      float64
	WeightedAverageRate       float64
	HighestRateLoan          LoanSummary
	LowestRateLoan           LoanSummary
	LargestBalanceLoan       LoanSummary
	SmallestBalanceLoan      LoanSummary
	DebtToIncomeRatio        float64
	MonthsToPayoff           int
	TotalInterestRemaining   float64
	DebtHealthStatus         string // "Excellent", "Good", "Fair", "Poor"
	Recommendations          []string
	PayoffProjections        []LoanPaymentPlan
}

// LoanSummary represents a simplified loan for analysis purposes
type LoanSummary struct {
	ID               string
	Lender           string
	Type             string
	RemainingBalance float64
	InterestRate     float64
	MonthlyPayment   float64
}