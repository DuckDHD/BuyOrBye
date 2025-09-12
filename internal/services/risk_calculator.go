package services

import (
	"github.com/DuckDHD/BuyOrBye/internal/domain"
)

// riskCalculator implements the RiskCalculator interface
type riskCalculator struct{}

// NewRiskCalculator creates a new risk calculator instance
func NewRiskCalculator() RiskCalculator {
	return &riskCalculator{}
}

// CalculateHealthRiskScore calculates comprehensive health risk score
// Based on DOMAIN_HEALTH.md specifications:
// - Age scoring: <30 (0pts), 30-40 (5pts), 40-50 (10pts), 50-60 (15pts), 60+ (20pts)
// - BMI scoring: 18.5-25 (0pts), 25-30 (8pts), <18.5 or >30 (15pts)
// - Conditions: mild (2pts), moderate (5pts), severe (10pts), critical (15pts each)
// - Family size: 1-2 (0pts), 3-4 (5pts), 5+ (10pts)
// - Cap total at 100
func (r *riskCalculator) CalculateHealthRiskScore(profile *domain.HealthProfile, conditions []domain.MedicalCondition) int {
	score := 0

	// Age scoring
	score += r.calculateAgePoints(profile.Age)

	// BMI scoring - use profile BMI if available, otherwise calculate
	bmi := profile.BMI
	if bmi == 0 {
		bmi, _ = profile.CalculateBMI()
	}
	score += r.calculateBMIPoints(bmi)

	// Conditions scoring
	for _, condition := range conditions {
		if condition.IsActive {
			score += r.calculateConditionSeverityPoints(condition.Severity)
		}
	}

	// Family size scoring
	score += r.calculateFamilySizePoints(profile.FamilySize)

	// Cap at 100
	if score > 100 {
		score = 100
	}

	return score
}

// AssessFinancialVulnerability evaluates financial vulnerability based on health costs to income ratio
// Vulnerability levels: <5% (secure), 5-10% (moderate), 10-20% (vulnerable), >20% (critical)
func (r *riskCalculator) AssessFinancialVulnerability(healthCosts, income float64) string {
	if income <= 0 {
		return "critical" // No income data
	}

	ratio := (healthCosts / income) * 100

	switch {
	case ratio < 5:
		return "secure"
	case ratio < 10:
		return "moderate"
	case ratio < 20:
		return "vulnerable"
	default:
		return "critical"
	}
}

// RecommendEmergencyFund calculates recommended emergency fund
// Base 6 months + risk adjustment: base * (1 + riskScore/100)
func (r *riskCalculator) RecommendEmergencyFund(riskScore int, monthlyExpenses float64) float64 {
	baseMonths := 6.0
	riskMultiplier := 1.0 + (float64(riskScore)/100.0)
	return baseMonths * monthlyExpenses * riskMultiplier
}

// DetermineRiskLevel converts risk score to risk level
// Risk levels: 0-25 (low), 26-50 (moderate), 51-75 (high), 76-100 (critical)
func (r *riskCalculator) DetermineRiskLevel(score int) string {
	switch {
	case score <= 25:
		return "low"
	case score <= 50:
		return "moderate"
	case score <= 75:
		return "high"
	default:
		return "critical"
	}
}

// Helper methods for scoring calculations

// calculateAgePoints calculates points based on age
// <30 (0pts), 30-40 (5pts), 40-50 (10pts), 50-60 (15pts), 60+ (20pts)
func (r *riskCalculator) calculateAgePoints(age int) int {
	switch {
	case age < 30:
		return 0
	case age < 40:
		return 5
	case age < 50:
		return 10
	case age < 60:
		return 15
	default:
		return 20
	}
}

// calculateBMIPoints calculates points based on BMI
// 18.5-25 (0pts), 25-30 (8pts), <18.5 or >30 (15pts)
func (r *riskCalculator) calculateBMIPoints(bmi float64) int {
	switch {
	case bmi >= 18.5 && bmi <= 25:
		return 0
	case bmi > 25 && bmi <= 30:
		return 8
	default: // <18.5 or >30
		return 15
	}
}

// calculateConditionSeverityPoints calculates points based on condition severity
// mild (2pts), moderate (5pts), severe (10pts), critical (15pts)
func (r *riskCalculator) calculateConditionSeverityPoints(severity string) int {
	switch severity {
	case "mild":
		return 2
	case "moderate":
		return 5
	case "severe":
		return 10
	case "critical":
		return 15
	default:
		return 2 // Default to mild
	}
}

// calculateFamilySizePoints calculates points based on family size
// 1-2 (0pts), 3-4 (5pts), 5+ (10pts)
func (r *riskCalculator) calculateFamilySizePoints(familySize int) int {
	switch {
	case familySize <= 2:
		return 0
	case familySize <= 4:
		return 5
	default: // 5+
		return 10
	}
}