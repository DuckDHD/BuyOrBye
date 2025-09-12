package services

import (
	"context"
	"errors"
	"time"

	"github.com/DuckDHD/BuyOrBye/internal/domain"
)

// Common errors
var (
	ErrProfileNotFound = errors.New("profile not found")
)

// HealthService defines health management operations
type HealthService interface {
	// Profile operations
	CreateProfile(ctx context.Context, profile *domain.HealthProfile) error
	GetProfile(ctx context.Context, userID string) (*domain.HealthProfile, error)
	UpdateProfile(ctx context.Context, profile *domain.HealthProfile) error
	
	// Medical conditions
	AddCondition(ctx context.Context, condition *domain.MedicalCondition) error
	GetConditions(ctx context.Context, userID string) ([]domain.MedicalCondition, error)
	UpdateCondition(ctx context.Context, condition *domain.MedicalCondition) error
	RemoveCondition(ctx context.Context, userID, conditionID string) error
	
	// Medical expenses
	AddExpense(ctx context.Context, expense *domain.MedicalExpense) error
	GetExpenses(ctx context.Context, userID string) ([]domain.MedicalExpense, error)
	GetRecurringExpenses(ctx context.Context, userID string) ([]domain.MedicalExpense, error)
	
	// Insurance policies
	AddInsurancePolicy(ctx context.Context, policy *domain.InsurancePolicy) error
	GetActivePolicies(ctx context.Context, userID string) ([]domain.InsurancePolicy, error)
	UpdateDeductibleProgress(ctx context.Context, policyID string, amount float64) error
	
	// Calculations & Analysis
	CalculateHealthSummary(ctx context.Context, userID string) (*domain.HealthSummary, error)
}

// RiskCalculator defines health risk calculation operations
type RiskCalculator interface {
	CalculateHealthRiskScore(profile *domain.HealthProfile, conditions []domain.MedicalCondition) int
	AssessFinancialVulnerability(healthCosts, income float64) string
	RecommendEmergencyFund(riskScore int, monthlyExpenses float64) float64
	DetermineRiskLevel(score int) string
}

// MedicalCostAnalyzer defines medical cost analysis operations
type MedicalCostAnalyzer interface {
	CalculateMonthlyAverage(expenses []domain.MedicalExpense) float64
	ProjectAnnualCosts(expenses []domain.MedicalExpense, conditions []domain.MedicalCondition) float64
	IdentifyCostReductionOpportunities(expenses []domain.MedicalExpense) []CostReductionOpportunity
	AnalyzeTrends(expenses []domain.MedicalExpense) []string
}

// InsuranceEvaluator defines insurance evaluation and coverage operations
type InsuranceEvaluator interface {
	CalculateCoverage(policy *domain.InsurancePolicy, expenseAmount float64) (*CoverageResult, error)
	EvaluateCoverageGaps(policies []domain.InsurancePolicy, conditions []domain.MedicalCondition, expenses []domain.MedicalExpense) []CoverageGap
	RecommendPolicyAdjustments(policies []domain.InsurancePolicy, expenses []domain.MedicalExpense) []PolicyRecommendation
	TrackDeductibleProgress(policy *domain.InsurancePolicy, newExpenseAmount float64) (*DeductibleUpdate, error)
}

// CostReductionOpportunity represents a cost reduction opportunity
type CostReductionOpportunity struct {
	Type              string  `json:"type"`
	Description       string  `json:"description"`
	PotentialSavings  float64 `json:"potential_savings"`
	Recommendation    string  `json:"recommendation"`
}

// Repository interfaces (for dependency injection)

// HealthProfileRepository defines the interface for health profile persistence
type HealthProfileRepository interface {
	// CRUD operations
	Create(ctx context.Context, profile *domain.HealthProfile) (*domain.HealthProfile, error)
	GetByID(ctx context.Context, id uint) (*domain.HealthProfile, error)
	GetByUserID(ctx context.Context, userID string) (*domain.HealthProfile, error)
	Update(ctx context.Context, profile *domain.HealthProfile) (*domain.HealthProfile, error)
	Delete(ctx context.Context, id uint) error
	
	// Business queries
	GetWithRelations(ctx context.Context, userID string) (*domain.HealthProfile, error)
	ExistsByUserID(ctx context.Context, userID string) (bool, error)
}

// MedicalConditionRepository defines the interface for medical condition persistence
type MedicalConditionRepository interface {
	// CRUD operations
	Create(ctx context.Context, condition *domain.MedicalCondition) (*domain.MedicalCondition, error)
	GetByID(ctx context.Context, id string) (*domain.MedicalCondition, error)
	Update(ctx context.Context, condition *domain.MedicalCondition) (*domain.MedicalCondition, error)
	Delete(ctx context.Context, id string) error
	
	// Query operations
	GetByUserID(ctx context.Context, userID string, activeOnly bool) ([]*domain.MedicalCondition, error)
	GetByCategory(ctx context.Context, userID string, category string) ([]*domain.MedicalCondition, error)
	GetBySeverity(ctx context.Context, userID string, severity string) ([]*domain.MedicalCondition, error)
	GetByProfileID(ctx context.Context, profileID string) ([]*domain.MedicalCondition, error)
	
	// Aggregation operations
	GetActiveConditionCount(ctx context.Context, userID string) (int64, error)
	CalculateTotalRiskFactor(ctx context.Context, userID string) (float64, error)
	GetMedicationRequiringConditions(ctx context.Context, userID string) ([]*domain.MedicalCondition, error)
}

// MedicalExpenseRepository defines the interface for medical expense persistence
type MedicalExpenseRepository interface {
	// CRUD operations
	Create(ctx context.Context, expense *domain.MedicalExpense) (*domain.MedicalExpense, error)
	GetByID(ctx context.Context, id string) (*domain.MedicalExpense, error)
	Update(ctx context.Context, expense *domain.MedicalExpense) (*domain.MedicalExpense, error)
	Delete(ctx context.Context, id string) error
	
	// Query operations
	GetByUserID(ctx context.Context, userID string) ([]*domain.MedicalExpense, error)
	GetByDateRange(ctx context.Context, userID string, startDate, endDate time.Time) ([]*domain.MedicalExpense, error)
	GetByCategory(ctx context.Context, userID string, category string) ([]*domain.MedicalExpense, error)
	GetByFrequency(ctx context.Context, userID string, frequency string) ([]*domain.MedicalExpense, error)
	GetRecurring(ctx context.Context, userID string) ([]*domain.MedicalExpense, error)
	GetByProfileID(ctx context.Context, profileID string) ([]*domain.MedicalExpense, error)
	
	// Aggregation operations
	CalculateTotals(ctx context.Context, userID string, startDate, endDate time.Time) (*ExpenseTotals, error)
	GetMonthlyRecurringTotal(ctx context.Context, userID string) (float64, error)
	GetAnnualProjectedExpenses(ctx context.Context, userID string) (float64, error)
}

// InsurancePolicyRepository defines the interface for insurance policy persistence
type InsurancePolicyRepository interface {
	// CRUD operations
	Create(ctx context.Context, policy *domain.InsurancePolicy) (*domain.InsurancePolicy, error)
	GetByID(ctx context.Context, id string) (*domain.InsurancePolicy, error)
	Update(ctx context.Context, policy *domain.InsurancePolicy) (*domain.InsurancePolicy, error)
	Delete(ctx context.Context, id string) error
	
	// Query operations
	GetByUserID(ctx context.Context, userID string) ([]*domain.InsurancePolicy, error)
	GetByType(ctx context.Context, userID string, policyType string) ([]*domain.InsurancePolicy, error)
	GetByPolicyNumber(ctx context.Context, policyNumber string) (*domain.InsurancePolicy, error)
	GetActivePolicies(ctx context.Context, userID string) ([]*domain.InsurancePolicy, error)
	GetByProfileID(ctx context.Context, profileID string) ([]*domain.InsurancePolicy, error)
	
	// Business operations
	UpdateDeductibleProgress(ctx context.Context, policyID string, deductibleMet, outOfPocketCurrent float64) (*domain.InsurancePolicy, error)
	CalculateCoverageForExpense(ctx context.Context, policyID string, expenseAmount float64) (*CoverageCalculation, error)
	GetPoliciesByProvider(ctx context.Context, userID string, provider string) ([]*domain.InsurancePolicy, error)
}

// Supporting types for repository operations

// ExpenseTotals represents aggregated expense data
type ExpenseTotals struct {
	TotalAmount        float64 `json:"total_amount"`
	TotalInsurancePaid float64 `json:"total_insurance_paid"`
	TotalOutOfPocket   float64 `json:"total_out_of_pocket"`
	ExpenseCount       int64   `json:"expense_count"`
}

// CoverageCalculation represents insurance coverage calculation results
type CoverageCalculation struct {
	InsurancePays         float64 `json:"insurance_pays"`
	PatientPays           float64 `json:"patient_pays"`
	NewDeductibleMet      float64 `json:"new_deductible_met"`
	NewOutOfPocketUsed    float64 `json:"new_out_of_pocket_used"`
	IsDeductibleMet       bool    `json:"is_deductible_met"`
	IsOutOfPocketMaxMet   bool    `json:"is_out_of_pocket_max_met"`
	RemainingDeductible   float64 `json:"remaining_deductible"`
	RemainingOutOfPocket  float64 `json:"remaining_out_of_pocket"`
}

// HealthInsights represents aggregated health data insights
type HealthInsights struct {
	TotalActiveConditions   int64   `json:"total_active_conditions"`
	TotalRiskFactor        float64 `json:"total_risk_factor"`
	MonthlyMedicationCost  float64 `json:"monthly_medication_cost"`
	MonthlyPremiumCost     float64 `json:"monthly_premium_cost"`
	DeductibleProgress     float64 `json:"deductible_progress_percentage"`
	OutOfPocketProgress    float64 `json:"out_of_pocket_progress_percentage"`
	ProjectedAnnualCost    float64 `json:"projected_annual_cost"`
}

// Supporting types for InsuranceEvaluator operations

// CoverageResult represents the result of coverage calculation
type CoverageResult struct {
	TotalCovered       float64 `json:"total_covered"`
	PatientPays        float64 `json:"patient_pays"`
	DeductibleApplied  float64 `json:"deductible_applied"`
	CopayAmount        float64 `json:"copay_amount"`
	NewDeductibleMet   float64 `json:"new_deductible_met"`
	NewOutOfPocketUsed float64 `json:"new_out_of_pocket_used"`
}

// CoverageGap represents an identified coverage gap
type CoverageGap struct {
	Type              string  `json:"type"`
	Description       string  `json:"description"`
	RiskLevel         string  `json:"risk_level"`         // low, moderate, high, critical
	Recommendation    string  `json:"recommendation"`
	EstimatedExposure float64 `json:"estimated_exposure"`
}

// PolicyRecommendation represents a suggested policy adjustment
type PolicyRecommendation struct {
	PolicyID    string  `json:"policy_id,omitempty"`
	Type        string  `json:"type"`
	Description string  `json:"description"`
	Impact      string  `json:"impact"`
	Savings     float64 `json:"estimated_savings"`
}

// DeductibleUpdate represents the result of updating deductible progress
type DeductibleUpdate struct {
	ExpenseAmount         float64 `json:"expense_amount"`
	InsuranceCoverage     float64 `json:"insurance_coverage"`
	AmountToDeductible    float64 `json:"amount_to_deductible"`
	AmountToOutOfPocket   float64 `json:"amount_to_out_of_pocket"`
	NewDeductibleMet      float64 `json:"new_deductible_met"`
	NewOutOfPocketUsed    float64 `json:"new_out_of_pocket_used"`
	DeductibleCompleted   bool    `json:"deductible_completed"`
	OutOfPocketMaxReached bool    `json:"out_of_pocket_max_reached"`
}