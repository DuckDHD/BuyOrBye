package domain

import (
	"fmt"
	"time"
)

// DecisionContext aggregates all user context needed for making a purchase decision
type DecisionContext struct {
	UserID            string             `json:"user_id"`
	FinancialContext  FinancialSnapshot  `json:"financial_context"`
	HealthContext     HealthSnapshot     `json:"health_context"`
	TransportContext  TransportSnapshot  `json:"transport_context"`
	PurchaseHistory   []PastDecision     `json:"purchase_history"`
	CurrentDate       time.Time          `json:"current_date"`
}

// FinancialSnapshot represents a user's current financial situation
type FinancialSnapshot struct {
	MonthlyIncome        float64 `json:"monthly_income"`
	MonthlyExpenses      float64 `json:"monthly_expenses"`
	DisposableIncome     float64 `json:"disposable_income"`
	DebtToIncomeRatio    float64 `json:"debt_to_income_ratio"`
	EmergencyFundMonths  float64 `json:"emergency_fund_months"`
	SavingsRate          float64 `json:"savings_rate"`
	FinancialHealth      string  `json:"financial_health"`      // "poor", "fair", "good", "excellent"
	BudgetRemaining      float64 `json:"budget_remaining"`
}

// HealthSnapshot represents a user's current health and medical financial situation
type HealthSnapshot struct {
	HealthRiskScore          int     `json:"health_risk_score"`          // 0-100, higher is riskier
	MonthlyHealthCosts       float64 `json:"monthly_health_costs"`
	InsuranceCoverage        float64 `json:"insurance_coverage"`         // 0.0-1.0
	FinancialVulnerability   string  `json:"financial_vulnerability"`    // "low", "medium", "high"
	HasCriticalConditions    bool    `json:"has_critical_conditions"`
	EmergencyFundNeeded      float64 `json:"emergency_fund_needed"`
}

// TransportSnapshot represents a user's transportation situation
type TransportSnapshot struct {
	HasVehicle           bool    `json:"has_vehicle"`
	MonthlyTransportCost float64 `json:"monthly_transport_cost"`
	PublicTransitAccess  bool    `json:"public_transit_access"`
	CommuteDistance      float64 `json:"commute_distance"`      // miles
}

// PastDecision represents a previous purchase decision
type PastDecision struct {
	ItemName    string  `json:"item_name"`
	ItemCost    float64 `json:"item_cost"`
	Decision    string  `json:"decision"`    // "BUY", "WAIT", "BYE"
	DaysAgo     int     `json:"days_ago"`
	Category    string  `json:"category"`
}

// Valid financial health levels
var validFinancialHealth = map[string]bool{
	"poor":      true,
	"fair":      true,
	"good":      true,
	"excellent": true,
}

// Valid vulnerability levels
var validVulnerability = map[string]bool{
	"low":    true,
	"medium": true,
	"high":   true,
}

// Validate performs comprehensive validation of the decision context
func (dc *DecisionContext) Validate() error {
	// Validate required fields
	if dc.UserID == "" {
		return fmt.Errorf("user ID is required")
	}

	// Validate financial context
	if err := dc.FinancialContext.Validate(); err != nil {
		return fmt.Errorf("financial snapshot validation failed: %w", err)
	}

	// Validate health context
	if err := dc.HealthContext.Validate(); err != nil {
		return fmt.Errorf("health snapshot validation failed: %w", err)
	}

	// Validate transport context
	if err := dc.TransportContext.Validate(); err != nil {
		return fmt.Errorf("transport snapshot validation failed: %w", err)
	}

	// Validate past decisions
	for i, decision := range dc.PurchaseHistory {
		if err := decision.Validate(); err != nil {
			return fmt.Errorf("past decision validation failed at index %d: %w", i, err)
		}
	}

	return nil
}

// Validate performs validation on FinancialSnapshot
func (fs *FinancialSnapshot) Validate() error {
	if fs.MonthlyIncome < 0 {
		return fmt.Errorf("monthly income must be non-negative")
	}

	if fs.MonthlyExpenses < 0 {
		return fmt.Errorf("monthly expenses must be non-negative")
	}

	if fs.DisposableIncome < 0 {
		return fmt.Errorf("disposable income must be non-negative")
	}

	if fs.DebtToIncomeRatio < 0 || fs.DebtToIncomeRatio > 10.0 {
		return fmt.Errorf("debt to income ratio must be between 0.0 and 10.0")
	}

	if fs.EmergencyFundMonths < 0 {
		return fmt.Errorf("emergency fund months must be non-negative")
	}

	if fs.SavingsRate < 0 || fs.SavingsRate > 1.0 {
		return fmt.Errorf("savings rate must be between 0.0 and 1.0")
	}

	if fs.FinancialHealth != "" && !validFinancialHealth[fs.FinancialHealth] {
		return fmt.Errorf("invalid financial health: %s. Must be one of: poor, fair, good, excellent", fs.FinancialHealth)
	}

	if fs.BudgetRemaining < 0 {
		return fmt.Errorf("budget remaining must be non-negative")
	}

	return nil
}

// Validate performs validation on HealthSnapshot
func (hs *HealthSnapshot) Validate() error {
	if hs.HealthRiskScore < 0 || hs.HealthRiskScore > 100 {
		return fmt.Errorf("health risk score must be between 0 and 100")
	}

	if hs.MonthlyHealthCosts < 0 {
		return fmt.Errorf("monthly health costs must be non-negative")
	}

	if hs.InsuranceCoverage < 0 || hs.InsuranceCoverage > 1.0 {
		return fmt.Errorf("insurance coverage must be between 0.0 and 1.0")
	}

	if hs.FinancialVulnerability != "" && !validVulnerability[hs.FinancialVulnerability] {
		return fmt.Errorf("invalid financial vulnerability: %s. Must be one of: low, medium, high", hs.FinancialVulnerability)
	}

	if hs.EmergencyFundNeeded < 0 {
		return fmt.Errorf("emergency fund needed must be non-negative")
	}

	return nil
}

// Validate performs validation on TransportSnapshot
func (ts *TransportSnapshot) Validate() error {
	if ts.MonthlyTransportCost < 0 {
		return fmt.Errorf("monthly transport cost must be non-negative")
	}

	if ts.CommuteDistance < 0 {
		return fmt.Errorf("commute distance must be non-negative")
	}

	return nil
}

// Validate performs validation on PastDecision
func (pd *PastDecision) Validate() error {
	if pd.ItemName == "" {
		return fmt.Errorf("item name is required")
	}

	if pd.ItemCost < 0 {
		return fmt.Errorf("item cost must be non-negative")
	}

	if !validDecisions[pd.Decision] {
		return fmt.Errorf("invalid decision: %s. Must be one of: BUY, WAIT, BYE", pd.Decision)
	}

	if pd.DaysAgo < 0 {
		return fmt.Errorf("days ago must be non-negative")
	}

	if pd.Category != "" && !validCategories[pd.Category] {
		return fmt.Errorf("invalid category: %s", pd.Category)
	}

	return nil
}

// GetTotalRecentSpending calculates total spending from BUY decisions within specified days
func (dc *DecisionContext) GetTotalRecentSpending(days int) float64 {
	total := 0.0
	for _, decision := range dc.PurchaseHistory {
		if decision.Decision == "BUY" && decision.DaysAgo <= days {
			total += decision.ItemCost
		}
	}
	return total
}

// GetRecentDecisionsByCategory returns recent decisions filtered by category and timeframe
func (dc *DecisionContext) GetRecentDecisionsByCategory(category string, days int) []PastDecision {
	var filtered []PastDecision
	for _, decision := range dc.PurchaseHistory {
		if decision.Category == category && decision.DaysAgo <= days {
			filtered = append(filtered, decision)
		}
	}
	return filtered
}

// HasRecentPurchaseInCategory checks if there's a recent BUY decision in the specified category
func (dc *DecisionContext) HasRecentPurchaseInCategory(category string, days int) bool {
	for _, decision := range dc.PurchaseHistory {
		if decision.Decision == "BUY" && decision.Category == category && decision.DaysAgo <= days {
			return true
		}
	}
	return false
}

// IsFinanciallyStressed returns true if user shows signs of financial stress
func (dc *DecisionContext) IsFinanciallyStressed() bool {
	return dc.FinancialContext.DebtToIncomeRatio > 0.5 || // High debt ratio
		dc.FinancialContext.EmergencyFundMonths < 3.0 // Low emergency fund
}

// HasHealthRisks returns true if user has elevated health risks
func (dc *DecisionContext) HasHealthRisks() bool {
	return dc.HealthContext.HealthRiskScore > 70 || // High risk score
		dc.HealthContext.HasCriticalConditions // Critical conditions
}

// GetAffordabilityRatio calculates how much of disposable income a purchase would consume
func (dc *DecisionContext) GetAffordabilityRatio(purchaseCost float64) float64 {
	if dc.FinancialContext.DisposableIncome <= 0 {
		return 999.0 // Effectively infinite ratio if no disposable income
	}
	return purchaseCost / dc.FinancialContext.DisposableIncome
}

// CanAfford checks if user can afford a purchase based on disposable income percentage
func (dc *DecisionContext) CanAfford(purchaseCost float64, maxPercentage float64) bool {
	ratio := dc.GetAffordabilityRatio(purchaseCost)
	return ratio <= maxPercentage
}

// GetRecentBuyDecisions returns recent BUY decisions within specified days
func (dc *DecisionContext) GetRecentBuyDecisions(days int) []PastDecision {
	var buyDecisions []PastDecision
	for _, decision := range dc.PurchaseHistory {
		if decision.Decision == "BUY" && decision.DaysAgo <= days {
			buyDecisions = append(buyDecisions, decision)
		}
	}
	return buyDecisions
}

// GetRecentWaitDecisions returns recent WAIT decisions within specified days
func (dc *DecisionContext) GetRecentWaitDecisions(days int) []PastDecision {
	var waitDecisions []PastDecision
	for _, decision := range dc.PurchaseHistory {
		if decision.Decision == "WAIT" && decision.DaysAgo <= days {
			waitDecisions = append(waitDecisions, decision)
		}
	}
	return waitDecisions
}

// GetDecisionPattern analyzes recent decision patterns
func (dc *DecisionContext) GetDecisionPattern(days int) map[string]int {
	pattern := map[string]int{
		"BUY":  0,
		"WAIT": 0,
		"BYE":  0,
	}

	for _, decision := range dc.PurchaseHistory {
		if decision.DaysAgo <= days {
			pattern[decision.Decision]++
		}
	}

	return pattern
}

// String returns a string representation of the decision context
func (dc *DecisionContext) String() string {
	return fmt.Sprintf("DecisionContext{UserID: %s, Income: %.2f, EmergencyFund: %.1f months, Debt Ratio: %.2f}",
		dc.UserID, dc.FinancialContext.MonthlyIncome, dc.FinancialContext.EmergencyFundMonths, dc.FinancialContext.DebtToIncomeRatio)
}