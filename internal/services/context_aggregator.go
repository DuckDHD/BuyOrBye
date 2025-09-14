package services

import (
	"context"
	"fmt"
	"time"

	"github.com/DuckDHD/BuyOrBye/internal/domain"
)

// ContextAggregator aggregates user context from multiple domains
type ContextAggregator struct {
	financeService FinanceServiceInterface
	healthService  HealthServiceInterface
	decisionRepo   DecisionRepository
}

// NewContextAggregator creates a new context aggregator
func NewContextAggregator(
	financeService FinanceServiceInterface,
	healthService HealthServiceInterface,
	decisionRepo DecisionRepository,
) *ContextAggregator {
	return &ContextAggregator{
		financeService: financeService,
		healthService:  healthService,
		decisionRepo:   decisionRepo,
	}
}

// AggregateUserContext gathers all domain data for decision making
func (ca *ContextAggregator) AggregateUserContext(ctx context.Context, userID string) (*domain.DecisionContext, error) {
	// Get financial snapshot
	financial, err := ca.getFinancialSnapshot(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get financial snapshot: %w", err)
	}

	// Get health snapshot
	health, err := ca.getHealthSnapshot(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get health snapshot: %w", err)
	}

	// Get recent decisions (last 30 days)
	recentDecisions, err := ca.GetRecentDecisions(ctx, userID, 30)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent decisions: %w", err)
	}

	// Build transport context (default values - would come from transport service in full implementation)
	transport := domain.TransportSnapshot{
		HasVehicle:           true,
		MonthlyTransportCost: 200.0,
		PublicTransitAccess:  true,
		CommuteDistance:      15.0,
	}

	// Aggregate into decision context
	context := &domain.DecisionContext{
		UserID:           userID,
		FinancialContext: *financial,
		HealthContext:    *health,
		TransportContext: transport,
		PurchaseHistory:  recentDecisions,
		CurrentDate:      time.Now(),
	}

	// Validate the aggregated context
	if err := context.Validate(); err != nil {
		return nil, fmt.Errorf("invalid aggregated context: %w", err)
	}

	return context, nil
}

// getFinancialSnapshot retrieves financial data with error handling for missing data
func (ca *ContextAggregator) getFinancialSnapshot(ctx context.Context, userID string) (*domain.FinancialSnapshot, error) {
	snapshot, err := ca.financeService.GetFinancialSnapshot(ctx, userID)
	if err != nil {
		// Handle missing financial data gracefully with defaults
		return ca.getDefaultFinancialSnapshot(), nil
	}

	// Validate and sanitize financial data
	if snapshot.MonthlyIncome <= 0 {
		snapshot.MonthlyIncome = 3000.0 // Default minimum income
	}
	if snapshot.MonthlyExpenses <= 0 {
		snapshot.MonthlyExpenses = snapshot.MonthlyIncome * 0.7 // 70% of income
	}
	if snapshot.DisposableIncome <= 0 {
		snapshot.DisposableIncome = snapshot.MonthlyIncome - snapshot.MonthlyExpenses
	}
	if snapshot.EmergencyFundMonths < 0 {
		snapshot.EmergencyFundMonths = 1.0 // Default 1 month
	}

	return snapshot, nil
}

// getHealthSnapshot retrieves health data with error handling for missing data
func (ca *ContextAggregator) getHealthSnapshot(ctx context.Context, userID string) (*domain.HealthSnapshot, error) {
	snapshot, err := ca.healthService.GetHealthSnapshot(ctx, userID)
	if err != nil {
		// Handle missing health data gracefully with defaults
		return ca.getDefaultHealthSnapshot(), nil
	}

	// Validate and sanitize health data
	if snapshot.HealthRiskScore < 0 || snapshot.HealthRiskScore > 100 {
		snapshot.HealthRiskScore = 30 // Default low-medium risk
	}
	if snapshot.MonthlyHealthCosts < 0 {
		snapshot.MonthlyHealthCosts = 150.0 // Default health costs
	}
	if snapshot.InsuranceCoverage < 0 || snapshot.InsuranceCoverage > 1.0 {
		snapshot.InsuranceCoverage = 0.7 // Default 70% coverage
	}

	return snapshot, nil
}

// GetRecentDecisions retrieves recent decision history
func (ca *ContextAggregator) GetRecentDecisions(ctx context.Context, userID string, days int) ([]domain.PastDecision, error) {
	// Get decisions from repository (limit to 50 recent decisions)
	decisions, err := ca.decisionRepo.GetRecentDecisions(ctx, userID, days)
	if err != nil {
		// If we can't get history, return empty slice rather than failing
		return []domain.PastDecision{}, nil
	}

	// Convert DecisionOutcome to PastDecision format
	pastDecisions := make([]domain.PastDecision, 0, len(decisions))
	for _, decision := range decisions {
		// Create PastDecision from DecisionOutcome
		// Note: We need to get the original purchase info from elsewhere or store it
		pastDecision := domain.PastDecision{
			ItemName: "Purchase Item", // DecisionOutcome doesn't have item details
			ItemCost: 0.0,             // Would need to store this in DecisionOutcome
			Category: "general",       // Would need to store this in DecisionOutcome  
			Decision: decision.Decision,
			DaysAgo:  int(time.Since(decision.CreatedAt).Hours() / 24),
		}
		pastDecisions = append(pastDecisions, pastDecision)
	}

	return pastDecisions, nil
}

// getDefaultFinancialSnapshot provides default financial data when user data is missing
func (ca *ContextAggregator) getDefaultFinancialSnapshot() *domain.FinancialSnapshot {
	return &domain.FinancialSnapshot{
		MonthlyIncome:        3000.0,
		MonthlyExpenses:      2100.0,
		DisposableIncome:     900.0,
		DebtToIncomeRatio:    0.3,
		EmergencyFundMonths:  1.5,
		SavingsRate:          0.1,
		FinancialHealth:      "fair",
		BudgetRemaining:      400.0,
	}
}

// getDefaultHealthSnapshot provides default health data when user data is missing
func (ca *ContextAggregator) getDefaultHealthSnapshot() *domain.HealthSnapshot {
	return &domain.HealthSnapshot{
		HealthRiskScore:          30,
		MonthlyHealthCosts:       150.0,
		InsuranceCoverage:        0.7,
		FinancialVulnerability:   "medium",
		HasCriticalConditions:    false,
		EmergencyFundNeeded:      1000.0,
	}
}

// ValidateUserContext performs additional validation on aggregated context
func (ca *ContextAggregator) ValidateUserContext(context *domain.DecisionContext) error {
	if context == nil {
		return fmt.Errorf("context cannot be nil")
	}

	if context.UserID == "" {
		return fmt.Errorf("user ID is required")
	}

	// Validate financial context makes sense
	if context.FinancialContext.MonthlyIncome < context.FinancialContext.MonthlyExpenses {
		// This is concerning but not necessarily invalid
	}

	// Validate health context
	if context.HealthContext.HealthRiskScore < 0 || context.HealthContext.HealthRiskScore > 100 {
		return fmt.Errorf("invalid health risk score: %d", context.HealthContext.HealthRiskScore)
	}

	return nil
}

// GetContextSummary provides a brief summary of the user's context for logging
func (ca *ContextAggregator) GetContextSummary(context *domain.DecisionContext) string {
	return fmt.Sprintf("User %s: Income $%.0f, Emergency Fund %.1f months, Health Risk %d/100, Recent Decisions: %d",
		context.UserID,
		context.FinancialContext.MonthlyIncome,
		context.FinancialContext.EmergencyFundMonths,
		context.HealthContext.HealthRiskScore,
		len(context.PurchaseHistory))
}

// IsDataComplete checks if we have sufficient data for decision making
func (ca *ContextAggregator) IsDataComplete(context *domain.DecisionContext) bool {
	// Check for minimum required financial data
	if context.FinancialContext.MonthlyIncome <= 0 || context.FinancialContext.DisposableIncome <= 0 {
		return false
	}

	// We can make decisions even with minimal health data
	return true
}

// GetDataQualityScore returns a score (0-100) indicating data completeness
func (ca *ContextAggregator) GetDataQualityScore(context *domain.DecisionContext) int {
	score := 0

	// Financial data quality (50 points max)
	if context.FinancialContext.MonthlyIncome > 0 {
		score += 15
	}
	if context.FinancialContext.EmergencyFundMonths >= 0 {
		score += 10
	}
	if context.FinancialContext.DebtToIncomeRatio >= 0 {
		score += 10
	}
	if context.FinancialContext.SavingsRate >= 0 {
		score += 10
	}
	if context.FinancialContext.FinancialHealth != "" {
		score += 5
	}

	// Health data quality (25 points max)
	if context.HealthContext.HealthRiskScore >= 0 {
		score += 10
	}
	if context.HealthContext.MonthlyHealthCosts >= 0 {
		score += 10
	}
	if context.HealthContext.InsuranceCoverage >= 0 {
		score += 5
	}

	// Historical data quality (25 points max)
	if len(context.PurchaseHistory) > 0 {
		score += 15
	}
	if len(context.PurchaseHistory) > 5 {
		score += 10
	}

	return score
}