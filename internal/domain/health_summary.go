package domain

import "time"

// HealthSummary represents aggregated health and financial data
type HealthSummary struct {
	UserID                    string    `json:"user_id"`
	HealthRiskScore           int       `json:"health_risk_score"`           // 0-100 (0=excellent, 100=critical)
	HealthRiskLevel           string    `json:"health_risk_level"`           // "low", "moderate", "high", "critical"
	MonthlyMedicalExpenses    float64   `json:"monthly_medical_expenses"`
	MonthlyInsurancePremiums  float64   `json:"monthly_insurance_premiums"`
	AnnualDeductibleRemaining float64   `json:"annual_deductible_remaining"`
	OutOfPocketRemaining      float64   `json:"out_of_pocket_remaining"`
	TotalHealthCosts          float64   `json:"total_health_costs"`          // premiums + out-of-pocket
	CoverageGapRisk           float64   `json:"coverage_gap_risk"`           // uncovered potential expenses
	RecommendedEmergencyFund  float64   `json:"recommended_emergency_fund"`  // based on health risks
	FinancialVulnerability    string    `json:"financial_vulnerability"`     // "secure", "moderate", "vulnerable", "critical"
	PriorityAdjustment        float64   `json:"priority_adjustment"`         // multiplier for purchase decisions
	UpdatedAt                 time.Time `json:"updated_at"`
}

// DetermineHealthLevel determines the health risk level based on score
func (h *HealthSummary) DetermineHealthLevel() string {
	switch {
	case h.HealthRiskScore <= 25:
		return "low"
	case h.HealthRiskScore <= 50:
		return "moderate"
	case h.HealthRiskScore <= 75:
		return "high"
	default:
		return "critical"
	}
}

// DetermineFinancialVulnerability determines financial vulnerability level
func (h *HealthSummary) DetermineFinancialVulnerability() string {
	return h.FinancialVulnerability
}

// CalculateTotalHealthCosts calculates total monthly health costs
func (h *HealthSummary) CalculateTotalHealthCosts() float64 {
	return h.MonthlyMedicalExpenses + h.MonthlyInsurancePremiums
}

// GetRiskMultiplier returns a multiplier for purchase decisions based on health risk
func (h *HealthSummary) GetRiskMultiplier() float64 {
	switch h.HealthRiskLevel {
	case "low":
		return 1.0
	case "moderate":
		return 1.2
	case "high":
		return 1.5
	case "critical":
		return 2.0
	default:
		return 1.0
	}
}