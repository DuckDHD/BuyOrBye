package services

import (
	"context"
	"fmt"

	"github.com/DuckDHD/BuyOrBye/internal/domain"
)

// healthService implements the HealthService interface
type healthService struct {
	profileRepo   HealthProfileRepository
	conditionRepo MedicalConditionRepository
	expenseRepo   MedicalExpenseRepository
	policyRepo    InsurancePolicyRepository
	riskCalc      RiskCalculator
	costAnalyzer  MedicalCostAnalyzer
}

// NewHealthService creates a new health service instance
func NewHealthService(
	profileRepo HealthProfileRepository,
	conditionRepo MedicalConditionRepository,
	expenseRepo MedicalExpenseRepository,
	policyRepo InsurancePolicyRepository,
	riskCalc RiskCalculator,
	costAnalyzer MedicalCostAnalyzer,
) HealthService {
	return &healthService{
		profileRepo:   profileRepo,
		conditionRepo: conditionRepo,
		expenseRepo:   expenseRepo,
		policyRepo:    policyRepo,
		riskCalc:      riskCalc,
		costAnalyzer:  costAnalyzer,
	}
}

// Profile operations
func (h *healthService) CreateProfile(ctx context.Context, profile *domain.HealthProfile) error {
	// Check if user already has a profile (one per user constraint)
	exists, err := h.profileRepo.ExistsByUserID(ctx, profile.UserID)
	if err != nil {
		return fmt.Errorf("error checking existing profile: %w", err)
	}

	if exists {
		return fmt.Errorf("user already has a health profile")
	}

	// Validate the profile
	if err := profile.Validate(); err != nil {
		return fmt.Errorf("profile validation failed: %w", err)
	}

	// Calculate BMI before saving
	bmi, err := profile.CalculateBMI()
	if err != nil {
		return fmt.Errorf("BMI calculation failed: %w", err)
	}
	profile.BMI = bmi

	_, err = h.profileRepo.Create(ctx, profile)
	return err
}

func (h *healthService) GetProfile(ctx context.Context, userID string) (*domain.HealthProfile, error) {
	return h.profileRepo.GetByUserID(ctx, userID)
}

func (h *healthService) UpdateProfile(ctx context.Context, profile *domain.HealthProfile) error {
	if err := profile.Validate(); err != nil {
		return fmt.Errorf("profile validation failed: %w", err)
	}

	// Recalculate BMI before updating
	bmi, err := profile.CalculateBMI()
	if err != nil {
		return fmt.Errorf("BMI calculation failed: %w", err)
	}
	profile.BMI = bmi

	_, err = h.profileRepo.Update(ctx, profile)
	return err
}

// Medical conditions
func (h *healthService) AddCondition(ctx context.Context, condition *domain.MedicalCondition) error {
	if err := condition.Validate(); err != nil {
		return fmt.Errorf("condition validation failed: %w", err)
	}

	// Calculate risk factor based on severity if not provided
	if condition.RiskFactor == 0 {
		condition.RiskFactor = h.calculateRiskFactorBySeverity(condition.Severity)
	}

	_, err := h.conditionRepo.Create(ctx, condition)
	return err
}

func (h *healthService) GetConditions(ctx context.Context, userID string) ([]domain.MedicalCondition, error) {
	conditions, err := h.conditionRepo.GetByUserID(ctx, userID, false) // Get all conditions
	if err != nil {
		return nil, err
	}

	// Convert from []*domain.MedicalCondition to []domain.MedicalCondition
	result := make([]domain.MedicalCondition, len(conditions))
	for i, condition := range conditions {
		result[i] = *condition
	}
	return result, nil
}

func (h *healthService) UpdateCondition(ctx context.Context, condition *domain.MedicalCondition) error {
	if err := condition.Validate(); err != nil {
		return fmt.Errorf("condition validation failed: %w", err)
	}

	_, err := h.conditionRepo.Update(ctx, condition)
	return err
}

func (h *healthService) RemoveCondition(ctx context.Context, userID, conditionID string) error {
	return h.conditionRepo.Delete(ctx, conditionID)
}

// Medical expenses
func (h *healthService) AddExpense(ctx context.Context, expense *domain.MedicalExpense) error {
	if err := expense.Validate(); err != nil {
		return fmt.Errorf("expense validation failed: %w", err)
	}

	// Calculate out-of-pocket after insurance if covered
	if expense.IsCovered && expense.InsurancePayment > 0 {
		expense.OutOfPocket = expense.Amount - expense.InsurancePayment
		if expense.OutOfPocket < 0 {
			expense.OutOfPocket = 0
		}
	} else {
		expense.OutOfPocket = expense.Amount
	}

	_, err := h.expenseRepo.Create(ctx, expense)
	return err
}

func (h *healthService) GetExpenses(ctx context.Context, userID string) ([]domain.MedicalExpense, error) {
	expenses, err := h.expenseRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Convert from []*domain.MedicalExpense to []domain.MedicalExpense
	result := make([]domain.MedicalExpense, len(expenses))
	for i, expense := range expenses {
		result[i] = *expense
	}
	return result, nil
}

func (h *healthService) GetRecurringExpenses(ctx context.Context, userID string) ([]domain.MedicalExpense, error) {
	expenses, err := h.expenseRepo.GetRecurring(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Convert from []*domain.MedicalExpense to []domain.MedicalExpense
	result := make([]domain.MedicalExpense, len(expenses))
	for i, expense := range expenses {
		result[i] = *expense
	}
	return result, nil
}

// Insurance policies
func (h *healthService) AddInsurancePolicy(ctx context.Context, policy *domain.InsurancePolicy) error {
	if err := policy.Validate(); err != nil {
		return fmt.Errorf("policy validation failed: %w", err)
	}

	// Validate no overlapping policies of the same type
	existingPolicies, err := h.policyRepo.GetByType(ctx, policy.UserID, policy.Type)
	if err != nil {
		return fmt.Errorf("failed to check existing policies: %w", err)
	}

	for _, existing := range existingPolicies {
		if existing.IsActive && existing.ID != policy.ID {
			// Check for date overlap
			if policy.StartDate.Before(existing.EndDate) && policy.EndDate.After(existing.StartDate) {
				return fmt.Errorf("policy overlaps with existing active %s policy %s", policy.Type, existing.PolicyNumber)
			}
		}
	}

	_, err = h.policyRepo.Create(ctx, policy)
	return err
}

func (h *healthService) GetActivePolicies(ctx context.Context, userID string) ([]domain.InsurancePolicy, error) {
	policies, err := h.policyRepo.GetActivePolicies(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Convert from []*domain.InsurancePolicy to []domain.InsurancePolicy
	result := make([]domain.InsurancePolicy, len(policies))
	for i, policy := range policies {
		result[i] = *policy
	}
	return result, nil
}

func (h *healthService) UpdateDeductibleProgress(ctx context.Context, policyID string, amount float64) error {
	// Get current policy
	policy, err := h.policyRepo.GetByID(ctx, policyID)
	if err != nil {
		return fmt.Errorf("failed to get policy: %w", err)
	}

	// Update deductible progress
	newDeductibleMet := policy.DeductibleMet + amount
	newOutOfPocketCurrent := policy.OutOfPocketCurrent + amount

	_, err = h.policyRepo.UpdateDeductibleProgress(ctx, policyID, newDeductibleMet, newOutOfPocketCurrent)
	return err
}

// Calculations & Analysis
func (h *healthService) CalculateHealthSummary(ctx context.Context, userID string) (*domain.HealthSummary, error) {
	// Get user's health profile
	profile, err := h.profileRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user profile: %w", err)
	}

	// Get medical conditions (active only for calculations)
	conditionPtrs, err := h.conditionRepo.GetByUserID(ctx, userID, true)
	if err != nil {
		return nil, fmt.Errorf("failed to get conditions: %w", err)
	}

	// Convert to value slice for compatibility
	conditions := make([]domain.MedicalCondition, len(conditionPtrs))
	for i, condition := range conditionPtrs {
		conditions[i] = *condition
	}

	// Get medical expenses
	expensePtrs, err := h.expenseRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get expenses: %w", err)
	}

	// Convert to value slice for compatibility
	expenses := make([]domain.MedicalExpense, len(expensePtrs))
	for i, expense := range expensePtrs {
		expenses[i] = *expense
	}

	// Get active insurance policies
	policyPtrs, err := h.policyRepo.GetActivePolicies(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get policies: %w", err)
	}

	// Convert to value slice for compatibility
	policies := make([]domain.InsurancePolicy, len(policyPtrs))
	for i, policy := range policyPtrs {
		policies[i] = *policy
	}

	// Calculate risk score
	riskScore := h.riskCalc.CalculateHealthRiskScore(profile, conditions)

	// Calculate costs
	monthlyAverage := h.costAnalyzer.CalculateMonthlyAverage(expenses)
	projectedAnnual := h.costAnalyzer.ProjectAnnualCosts(expenses, conditions)

	// Calculate out-of-pocket costs
	totalOutOfPocket := 0.0
	for _, expense := range expenses {
		totalOutOfPocket += expense.OutOfPocket
	}

	// Assess financial vulnerability (assuming some income - this would come from user data)
	assumedIncome := 60000.0 // This should come from user financial profile
	financialVulnerability := h.riskCalc.AssessFinancialVulnerability(projectedAnnual, assumedIncome)

	// Recommend emergency fund
	emergencyFund := h.riskCalc.RecommendEmergencyFund(riskScore, monthlyAverage)

	// Calculate insurance premiums and deductible info from policies
	monthlyPremiums := 0.0
	totalDeductibleRemaining := 0.0
	for _, policy := range policies {
		monthlyPremiums += policy.MonthlyPremium
		totalDeductibleRemaining += policy.GetRemainingDeductible()
	}

	// Calculate priority adjustment based on health risk
	priorityAdjustment := 1.0
	if riskScore > 75 {
		priorityAdjustment = 1.5 // Critical risk
	} else if riskScore > 50 {
		priorityAdjustment = 1.3 // High risk
	} else if riskScore > 25 {
		priorityAdjustment = 1.1 // Moderate risk
	}

	summary := &domain.HealthSummary{
		UserID:                    userID,
		HealthRiskScore:           riskScore,
		HealthRiskLevel:           h.riskCalc.DetermineRiskLevel(riskScore),
		MonthlyMedicalExpenses:    monthlyAverage,
		MonthlyInsurancePremiums:  monthlyPremiums,
		AnnualDeductibleRemaining: totalDeductibleRemaining,
		OutOfPocketRemaining:      totalOutOfPocket,
		TotalHealthCosts:          monthlyAverage + monthlyPremiums,
		CoverageGapRisk:           projectedAnnual - totalOutOfPocket,
		RecommendedEmergencyFund:  emergencyFund,
		FinancialVulnerability:    financialVulnerability,
		PriorityAdjustment:        priorityAdjustment,
		UpdatedAt:                 profile.UpdatedAt,
	}

	return summary, nil
}

// GetHealthContext prepares health data for Decision domain
// TODO: Implement when domain.HealthContext is available
/*
func (h *healthService) GetHealthContext(ctx context.Context, userID string) (*domain.HealthContext, error) {
	// Get health summary with all calculations
	summary, err := h.CalculateHealthSummary(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get health summary: %w", err)
	}

	// Get profile for BMI and basic info
	profile, err := h.profileRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get health profile: %w", err)
	}

	// Get active conditions count
	activeConditionCount, err := h.conditionRepo.GetActiveConditionCount(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active condition count: %w", err)
	}

	// Get total risk factor
	totalRiskFactor, err := h.conditionRepo.CalculateTotalRiskFactor(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate total risk factor: %w", err)
	}

	return &domain.HealthContext{
		UserID:                   userID,
		HealthRiskScore:          summary.HealthRiskScore,
		HealthRiskLevel:          summary.HealthRiskLevel,
		BMI:                      profile.BMI,
		Age:                      profile.Age,
		ActiveConditionCount:     int(activeConditionCount),
		TotalRiskFactor:          totalRiskFactor,
		MonthlyHealthCosts:       summary.TotalHealthCosts,
		AnnualHealthProjection:   summary.MonthlyMedicalExpenses * 12,
		CoverageGapRisk:          summary.CoverageGapRisk,
		RecommendedEmergencyFund: summary.RecommendedEmergencyFund,
		FinancialVulnerability:   summary.FinancialVulnerability,
		PriorityMultiplier:       summary.PriorityAdjustment,
	}, nil
}
*/

// Helper methods
func (h *healthService) calculateRiskFactorBySeverity(severity string) float64 {
	switch severity {
	case "mild":
		return 0.1
	case "moderate":
		return 0.25
	case "severe":
		return 0.4
	case "critical":
		return 0.6
	default:
		return 0.1
	}
}

func (h *healthService) countActiveConditions(conditions []domain.MedicalCondition) int {
	count := 0
	for _, condition := range conditions {
		if condition.IsActive {
			count++
		}
	}
	return count
}

func (h *healthService) countHighRiskConditions(conditions []domain.MedicalCondition) int {
	count := 0
	for _, condition := range conditions {
		if condition.IsActive && (condition.Severity == "severe" || condition.Severity == "critical") {
			count++
		}
	}
	return count
}
