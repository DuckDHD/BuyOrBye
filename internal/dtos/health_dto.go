package dtos

import (
	"time"

	"github.com/DuckDHD/BuyOrBye/internal/domain"
)

// Health Profile DTOs

// CreateHealthProfileRequestDTO represents a request to create a health profile
type CreateHealthProfileRequestDTO struct {
	UserID               string  `json:"user_id" binding:"required"`
	Age                  int     `json:"age" binding:"required,gte=0,lte=120"`
	Gender               string  `json:"gender" binding:"required,oneof=male female other"`
	Height               float64 `json:"height" binding:"required,gt=0"`
	Weight               float64 `json:"weight" binding:"required,gt=0"`
	FamilySize           int     `json:"family_size" binding:"required,gte=1"`
	HasChronicConditions bool    `json:"has_chronic_conditions"`
	EmergencyFundHealth  float64 `json:"emergency_fund_health" binding:"gte=0"`
}

// ToDomain converts DTO to domain struct
func (dto CreateHealthProfileRequestDTO) ToDomain() *domain.HealthProfile {
	return &domain.HealthProfile{
		UserID:               dto.UserID,
		Age:                  dto.Age,
		Gender:               dto.Gender,
		Height:               dto.Height,
		Weight:               dto.Weight,
		FamilySize:           dto.FamilySize,
		HasChronicConditions: dto.HasChronicConditions,
		EmergencyFundHealth:  dto.EmergencyFundHealth,
	}
}

// UpdateHealthProfileRequestDTO represents a request to update a health profile
type UpdateHealthProfileRequestDTO struct {
	Age                  int     `json:"age" binding:"omitempty,gte=0,lte=120"`
	Gender               string  `json:"gender" binding:"omitempty,oneof=male female other"`
	Height               float64 `json:"height" binding:"omitempty,gt=0"`
	Weight               float64 `json:"weight" binding:"omitempty,gt=0"`
	FamilySize           int     `json:"family_size" binding:"omitempty,gte=1"`
	HasChronicConditions bool    `json:"has_chronic_conditions"`
	EmergencyFundHealth  float64 `json:"emergency_fund_health" binding:"omitempty,gte=0"`
}

// HealthProfileResponseDTO represents a health profile response
type HealthProfileResponseDTO struct {
	ID                   string    `json:"id"`
	UserID               string    `json:"user_id"`
	Age                  int       `json:"age"`
	Gender               string    `json:"gender"`
	Height               float64   `json:"height"`
	Weight               float64   `json:"weight"`
	BMI                  float64   `json:"bmi"`
	FamilySize           int       `json:"family_size"`
	HasChronicConditions bool      `json:"has_chronic_conditions"`
	EmergencyFundHealth  float64   `json:"emergency_fund_health"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

// FromDomain converts domain struct to DTO
func (dto *HealthProfileResponseDTO) FromDomain(profile *domain.HealthProfile) {
	dto.ID = profile.ID
	dto.UserID = profile.UserID
	dto.Age = profile.Age
	dto.Gender = profile.Gender
	dto.Height = profile.Height
	dto.Weight = profile.Weight
	dto.BMI = profile.BMI
	dto.FamilySize = profile.FamilySize
	dto.HasChronicConditions = profile.HasChronicConditions
	dto.EmergencyFundHealth = profile.EmergencyFundHealth
	dto.CreatedAt = profile.CreatedAt
	dto.UpdatedAt = profile.UpdatedAt
}

// Medical Condition DTOs

// CreateMedicalConditionRequestDTO represents a request to create a medical condition
type CreateMedicalConditionRequestDTO struct {
	UserID             string    `json:"user_id" binding:"required"`
	ProfileID          string    `json:"profile_id" binding:"required"`
	Name               string    `json:"name" binding:"required"`
	Category           string    `json:"category" binding:"required,oneof=chronic acute mental_health preventive"`
	Severity           string    `json:"severity" binding:"required,oneof=mild moderate severe critical"`
	DiagnosedDate      time.Time `json:"diagnosed_date" binding:"required"`
	RequiresMedication bool      `json:"requires_medication"`
	MonthlyMedCost     float64   `json:"monthly_med_cost" binding:"gte=0"`
	RiskFactor         float64   `json:"risk_factor" binding:"gte=0,lte=1"`
	IsActive           bool      `json:"is_active"`
}

// ToDomain converts DTO to domain struct
func (dto CreateMedicalConditionRequestDTO) ToDomain() *domain.MedicalCondition {
	return &domain.MedicalCondition{
		UserID:             dto.UserID,
		ProfileID:          dto.ProfileID,
		Name:               dto.Name,
		Category:           dto.Category,
		Severity:           dto.Severity,
		DiagnosedDate:      dto.DiagnosedDate,
		RequiresMedication: dto.RequiresMedication,
		MonthlyMedCost:     dto.MonthlyMedCost,
		RiskFactor:         dto.RiskFactor,
		IsActive:           dto.IsActive,
	}
}

// UpdateMedicalConditionRequestDTO represents a request to update a medical condition
type UpdateMedicalConditionRequestDTO struct {
	Name               string  `json:"name"`
	Category           string  `json:"category" binding:"omitempty,oneof=chronic acute mental_health preventive"`
	Severity           string  `json:"severity" binding:"omitempty,oneof=mild moderate severe critical"`
	RequiresMedication bool    `json:"requires_medication"`
	MonthlyMedCost     float64 `json:"monthly_med_cost" binding:"gte=0"`
	RiskFactor         float64 `json:"risk_factor" binding:"gte=0,lte=1"`
	IsActive           bool    `json:"is_active"`
}

// MedicalConditionResponseDTO represents a medical condition response
type MedicalConditionResponseDTO struct {
	ID                 string    `json:"id"`
	UserID             string    `json:"user_id"`
	ProfileID          string    `json:"profile_id"`
	Name               string    `json:"name"`
	Category           string    `json:"category"`
	Severity           string    `json:"severity"`
	DiagnosedDate      time.Time `json:"diagnosed_date"`
	RequiresMedication bool      `json:"requires_medication"`
	MonthlyMedCost     float64   `json:"monthly_med_cost"`
	RiskFactor         float64   `json:"risk_factor"`
	IsActive           bool      `json:"is_active"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// FromDomain converts domain struct to DTO
func (dto *MedicalConditionResponseDTO) FromDomain(condition *domain.MedicalCondition) {
	dto.ID = condition.ID
	dto.UserID = condition.UserID
	dto.ProfileID = condition.ProfileID
	dto.Name = condition.Name
	dto.Category = condition.Category
	dto.Severity = condition.Severity
	dto.DiagnosedDate = condition.DiagnosedDate
	dto.RequiresMedication = condition.RequiresMedication
	dto.MonthlyMedCost = condition.MonthlyMedCost
	dto.RiskFactor = condition.RiskFactor
	dto.IsActive = condition.IsActive
	dto.CreatedAt = condition.CreatedAt
	dto.UpdatedAt = condition.UpdatedAt
}

// Medical Expense DTOs

// CreateMedicalExpenseRequestDTO represents a request to create a medical expense
type CreateMedicalExpenseRequestDTO struct {
	UserID           string    `json:"user_id" binding:"required"`
	ProfileID        string    `json:"profile_id" binding:"required"`
	Amount           float64   `json:"amount" binding:"required,gt=0"`
	Category         string    `json:"category" binding:"required,oneof=doctor_visit medication hospital lab_test therapy equipment"`
	Description      string    `json:"description" binding:"required"`
	Date             time.Time `json:"date" binding:"required"`
	IsCovered        bool      `json:"is_covered"`
	InsurancePayment float64   `json:"insurance_payment" binding:"gte=0"`
	OutOfPocket      float64   `json:"out_of_pocket" binding:"gte=0"`
	IsRecurring      bool      `json:"is_recurring"`
	Frequency        string    `json:"frequency" binding:"required,oneof=one_time monthly quarterly annually"`
}

// ToDomain converts DTO to domain struct
func (dto CreateMedicalExpenseRequestDTO) ToDomain() *domain.MedicalExpense {
	return &domain.MedicalExpense{
		UserID:           dto.UserID,
		ProfileID:        dto.ProfileID,
		Amount:           dto.Amount,
		Category:         dto.Category,
		Description:      dto.Description,
		Date:             dto.Date,
		IsCovered:        dto.IsCovered,
		InsurancePayment: dto.InsurancePayment,
		OutOfPocket:      dto.OutOfPocket,
		IsRecurring:      dto.IsRecurring,
		Frequency:        dto.Frequency,
	}
}

// MedicalExpenseResponseDTO represents a medical expense response
type MedicalExpenseResponseDTO struct {
	ID               string    `json:"id"`
	UserID           string    `json:"user_id"`
	ProfileID        string    `json:"profile_id"`
	Amount           float64   `json:"amount"`
	Category         string    `json:"category"`
	Description      string    `json:"description"`
	Date             time.Time `json:"date"`
	IsCovered        bool      `json:"is_covered"`
	InsurancePayment float64   `json:"insurance_payment"`
	OutOfPocket      float64   `json:"out_of_pocket"`
	IsRecurring      bool      `json:"is_recurring"`
	Frequency        string    `json:"frequency"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// FromDomain converts domain struct to DTO
func (dto *MedicalExpenseResponseDTO) FromDomain(expense *domain.MedicalExpense) {
	dto.ID = expense.ID
	dto.UserID = expense.UserID
	dto.ProfileID = expense.ProfileID
	dto.Amount = expense.Amount
	dto.Category = expense.Category
	dto.Description = expense.Description
	dto.Date = expense.Date
	dto.IsCovered = expense.IsCovered
	dto.InsurancePayment = expense.InsurancePayment
	dto.OutOfPocket = expense.OutOfPocket
	dto.IsRecurring = expense.IsRecurring
	dto.Frequency = expense.Frequency
	dto.CreatedAt = expense.CreatedAt
	dto.UpdatedAt = expense.UpdatedAt
}

// Insurance Policy DTOs

// CreateInsurancePolicyRequestDTO represents a request to create an insurance policy
type CreateInsurancePolicyRequestDTO struct {
	UserID             string    `json:"user_id" binding:"required"`
	PolicyNumber       string    `json:"policy_number" binding:"required"`
	Provider           string    `json:"provider" binding:"required"`
	Type               string    `json:"type" binding:"required,oneof=health dental vision life disability"`
	CoveragePercentage float64   `json:"coverage_percentage" binding:"required,gte=0,lte=100"`
	Deductible         float64   `json:"deductible" binding:"required,gte=0"`
	OutOfPocketMax     float64   `json:"out_of_pocket_max" binding:"required,gte=0"`
	MonthlyPremium     float64   `json:"monthly_premium" binding:"required,gte=0"`
	StartDate          time.Time `json:"start_date" binding:"required"`
	EndDate            time.Time `json:"end_date" binding:"required"`
	IsActive           bool      `json:"is_active"`
}

// ToDomain converts DTO to domain struct
func (dto CreateInsurancePolicyRequestDTO) ToDomain() *domain.InsurancePolicy {
	return &domain.InsurancePolicy{
		UserID:             dto.UserID,
		PolicyNumber:       dto.PolicyNumber,
		Provider:           dto.Provider,
		Type:               dto.Type,
		CoveragePercentage: dto.CoveragePercentage,
		Deductible:         dto.Deductible,
		OutOfPocketMax:     dto.OutOfPocketMax,
		MonthlyPremium:     dto.MonthlyPremium,
		StartDate:          dto.StartDate,
		EndDate:            dto.EndDate,
		IsActive:           dto.IsActive,
	}
}

// UpdateInsurancePolicyRequestDTO represents a request to update an insurance policy
type UpdateInsurancePolicyRequestDTO struct {
	CoveragePercentage float64   `json:"coverage_percentage" binding:"gte=0,lte=100"`
	Deductible         float64   `json:"deductible" binding:"gte=0"`
	OutOfPocketMax     float64   `json:"out_of_pocket_max" binding:"gte=0"`
	MonthlyPremium     float64   `json:"monthly_premium" binding:"gte=0"`
	EndDate            time.Time `json:"end_date"`
	IsActive           bool      `json:"is_active"`
}

// InsurancePolicyResponseDTO represents an insurance policy response
type InsurancePolicyResponseDTO struct {
	ID                 string    `json:"id"`
	UserID             string    `json:"user_id"`
	PolicyNumber       string    `json:"policy_number"`
	Provider           string    `json:"provider"`
	Type               string    `json:"type"`
	CoveragePercentage float64   `json:"coverage_percentage"`
	Deductible         float64   `json:"deductible"`
	DeductibleMet      float64   `json:"deductible_met"`
	OutOfPocketMax     float64   `json:"out_of_pocket_max"`
	OutOfPocketCurrent float64   `json:"out_of_pocket_current"`
	MonthlyPremium     float64   `json:"monthly_premium"`
	StartDate          time.Time `json:"start_date"`
	EndDate            time.Time `json:"end_date"`
	IsActive           bool      `json:"is_active"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// FromDomain converts domain struct to DTO
func (dto *InsurancePolicyResponseDTO) FromDomain(policy *domain.InsurancePolicy) {
	dto.ID = policy.ID
	dto.UserID = policy.UserID
	dto.PolicyNumber = policy.PolicyNumber
	dto.Provider = policy.Provider
	dto.Type = policy.Type
	dto.CoveragePercentage = policy.CoveragePercentage
	dto.Deductible = policy.Deductible
	dto.DeductibleMet = policy.DeductibleMet
	dto.OutOfPocketMax = policy.OutOfPocketMax
	dto.OutOfPocketCurrent = policy.OutOfPocketCurrent
	dto.MonthlyPremium = policy.MonthlyPremium
	dto.StartDate = policy.StartDate
	dto.EndDate = policy.EndDate
	dto.IsActive = policy.IsActive
	dto.CreatedAt = policy.CreatedAt
	dto.UpdatedAt = policy.UpdatedAt
}

// UpdateDeductibleRequestDTO represents a request to update deductible progress
type UpdateDeductibleRequestDTO struct {
	Amount float64 `json:"amount" binding:"required,gt=0"`
}

// Health Summary DTOs

// HealthSummaryResponseDTO represents a health summary response
type HealthSummaryResponseDTO struct {
	UserID                    string    `json:"user_id"`
	HealthRiskScore           int       `json:"health_risk_score"`
	HealthRiskLevel           string    `json:"health_risk_level"`
	MonthlyMedicalExpenses    float64   `json:"monthly_medical_expenses"`
	MonthlyInsurancePremiums  float64   `json:"monthly_insurance_premiums"`
	AnnualDeductibleRemaining float64   `json:"annual_deductible_remaining"`
	OutOfPocketRemaining      float64   `json:"out_of_pocket_remaining"`
	TotalHealthCosts          float64   `json:"total_health_costs"`
	CoverageGapRisk           float64   `json:"coverage_gap_risk"`
	RecommendedEmergencyFund  float64   `json:"recommended_emergency_fund"`
	FinancialVulnerability    string    `json:"financial_vulnerability"`
	PriorityAdjustment        float64   `json:"priority_adjustment"`
	UpdatedAt                 time.Time `json:"updated_at"`
}

// FromDomain converts domain struct to DTO
func (dto *HealthSummaryResponseDTO) FromDomain(summary *domain.HealthSummary) {
	dto.UserID = summary.UserID
	dto.HealthRiskScore = summary.HealthRiskScore
	dto.HealthRiskLevel = summary.HealthRiskLevel
	dto.MonthlyMedicalExpenses = summary.MonthlyMedicalExpenses
	dto.MonthlyInsurancePremiums = summary.MonthlyInsurancePremiums
	dto.AnnualDeductibleRemaining = summary.AnnualDeductibleRemaining
	dto.OutOfPocketRemaining = summary.OutOfPocketRemaining
	dto.TotalHealthCosts = summary.TotalHealthCosts
	dto.CoverageGapRisk = summary.CoverageGapRisk
	dto.RecommendedEmergencyFund = summary.RecommendedEmergencyFund
	dto.FinancialVulnerability = summary.FinancialVulnerability
	dto.PriorityAdjustment = summary.PriorityAdjustment
	dto.UpdatedAt = summary.UpdatedAt
}

// List Response DTOs for collections

// MedicalConditionListResponseDTO represents a list of medical conditions
type MedicalConditionListResponseDTO struct {
	Conditions []MedicalConditionResponseDTO `json:"conditions"`
	Total      int                           `json:"total"`
}

// MedicalExpenseListResponseDTO represents a list of medical expenses
type MedicalExpenseListResponseDTO struct {
	Expenses []MedicalExpenseResponseDTO `json:"expenses"`
	Total    int                         `json:"total"`
}

// InsurancePolicyListResponseDTO represents a list of insurance policies
type InsurancePolicyListResponseDTO struct {
	Policies []InsurancePolicyResponseDTO `json:"policies"`
	Total    int                          `json:"total"`
}