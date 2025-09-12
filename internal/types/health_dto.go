package types

import (
	"time"

	"github.com/DuckDHD/BuyOrBye/internal/domain"
)

// CreateHealthProfileDTO represents the request to create a health profile
type CreateHealthProfileDTO struct {
	Age        int     `json:"age" validate:"required,min=0,max=150"`
	Gender     string  `json:"gender" validate:"required,oneof=male female other"`
	Height     float64 `json:"height" validate:"required,gt=0"` // in cm
	Weight     float64 `json:"weight" validate:"required,gt=0"` // in kg
	FamilySize int     `json:"family_size" validate:"required,min=1,max=20"`
}

// UpdateProfileDTO represents partial updates to a health profile
type UpdateProfileDTO struct {
	Age        *int     `json:"age,omitempty" validate:"omitempty,min=0,max=150"`
	Gender     *string  `json:"gender,omitempty" validate:"omitempty,oneof=male female other"`
	Height     *float64 `json:"height,omitempty" validate:"omitempty,gt=0"`
	Weight     *float64 `json:"weight,omitempty" validate:"omitempty,gt=0"`
	FamilySize *int     `json:"family_size,omitempty" validate:"omitempty,min=1,max=20"`
}

// HealthProfileResponseDTO represents a health profile response
type HealthProfileResponseDTO struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	Age        int       `json:"age"`
	Gender     string    `json:"gender"`
	Height     float64   `json:"height"`
	Weight     float64   `json:"weight"`
	BMI        float64   `json:"bmi"`
	FamilySize int       `json:"family_size"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// AddMedicalConditionDTO represents the request to add a medical condition
type AddMedicalConditionDTO struct {
	Name               string  `json:"name" validate:"required,min=2,max=100"`
	Category           string  `json:"category" validate:"required,oneof=chronic acute mental_health preventive"`
	Severity           string  `json:"severity" validate:"required,oneof=mild moderate severe critical"`
	DiagnosedDate      string  `json:"diagnosed_date" validate:"required"` // ISO 8601 date
	IsActive           *bool   `json:"is_active,omitempty"`                // defaults to true
	RequiresMedication *bool   `json:"requires_medication,omitempty"`      // defaults to false
	MonthlyMedCost     float64 `json:"monthly_med_cost" validate:"omitempty,gte=0"`
	RiskFactor         float64 `json:"risk_factor" validate:"omitempty,gte=0,lte=1"` // defaults to 0.1
}

// UpdateConditionDTO represents partial updates to a medical condition
type UpdateConditionDTO struct {
	Name               *string  `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
	Category           *string  `json:"category,omitempty" validate:"omitempty,oneof=chronic acute mental_health preventive"`
	Severity           *string  `json:"severity,omitempty" validate:"omitempty,oneof=mild moderate severe critical"`
	IsActive           *bool    `json:"is_active,omitempty"`
	RequiresMedication *bool    `json:"requires_medication,omitempty"`
	MonthlyMedCost     *float64 `json:"monthly_med_cost,omitempty" validate:"omitempty,gte=0"`
	RiskFactor         *float64 `json:"risk_factor,omitempty" validate:"omitempty,gte=0,lte=1"`
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
	IsActive           bool      `json:"is_active"`
	RequiresMedication bool      `json:"requires_medication"`
	MonthlyMedCost     float64   `json:"monthly_med_cost"`
	RiskFactor         float64   `json:"risk_factor"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// AddMedicalExpenseDTO represents the request to add a medical expense
type AddMedicalExpenseDTO struct {
	Amount           float64 `json:"amount" validate:"required,gt=0"`
	Category         string  `json:"category" validate:"required,oneof=doctor_visit medication hospital lab_test therapy equipment"`
	Description      string  `json:"description" validate:"required,min=2,max=200"`
	IsRecurring      bool    `json:"is_recurring"`
	Frequency        string  `json:"frequency" validate:"omitempty,oneof=daily weekly bi-weekly monthly quarterly semi-annually annually one_time"`
	IsCovered        bool    `json:"is_covered"`
	InsurancePayment float64 `json:"insurance_payment" validate:"omitempty,gte=0"`
	Date             string  `json:"date" validate:"required"` // ISO 8601 date
}

// MedicalExpenseResponseDTO represents a medical expense response
type MedicalExpenseResponseDTO struct {
	ID               string    `json:"id"`
	UserID           string    `json:"user_id"`
	ProfileID        string    `json:"profile_id"`
	Amount           float64   `json:"amount"`
	Category         string    `json:"category"`
	Description      string    `json:"description"`
	IsRecurring      bool      `json:"is_recurring"`
	Frequency        string    `json:"frequency"`
	IsCovered        bool      `json:"is_covered"`
	InsurancePayment float64   `json:"insurance_payment"`
	OutOfPocket      float64   `json:"out_of_pocket"`
	Date             time.Time `json:"date"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// AddInsurancePolicyDTO represents the request to add an insurance policy
type AddInsurancePolicyDTO struct {
	Provider            string  `json:"provider" validate:"required,min=2,max=100"`
	PolicyNumber        string  `json:"policy_number" validate:"required,min=5,max=50"`
	Type                string  `json:"type" validate:"required,oneof=health dental vision"`
	MonthlyPremium      float64 `json:"monthly_premium" validate:"required,gt=0"`
	Deductible          float64 `json:"deductible" validate:"required,gte=0"`
	OutOfPocketMax      float64 `json:"out_of_pocket_max" validate:"required,gt=0"`
	CoveragePercentage  int     `json:"coverage_percentage" validate:"required,min=0,max=100"`
	StartDate           string  `json:"start_date" validate:"required"`    // ISO 8601 date
	EndDate             string  `json:"end_date" validate:"required"`      // ISO 8601 date
	IsActive            *bool   `json:"is_active,omitempty"`               // defaults to true
	DeductibleMet       float64 `json:"deductible_met" validate:"omitempty,gte=0"`
	OutOfPocketCurrent  float64 `json:"out_of_pocket_current" validate:"omitempty,gte=0"`
}

// InsurancePolicyResponseDTO represents an insurance policy response
type InsurancePolicyResponseDTO struct {
	ID                 string    `json:"id"`
	UserID             string    `json:"user_id"`
	ProfileID          string    `json:"profile_id"`
	Provider           string    `json:"provider"`
	PolicyNumber       string    `json:"policy_number"`
	Type               string    `json:"type"`
	MonthlyPremium     float64   `json:"monthly_premium"`
	Deductible         float64   `json:"deductible"`
	OutOfPocketMax     float64   `json:"out_of_pocket_max"`
	CoveragePercentage int       `json:"coverage_percentage"`
	StartDate          time.Time `json:"start_date"`
	EndDate            time.Time `json:"end_date"`
	IsActive           bool      `json:"is_active"`
	DeductibleMet      float64   `json:"deductible_met"`
	OutOfPocketCurrent float64   `json:"out_of_pocket_current"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// HealthSummaryDTO represents the comprehensive health summary response
type HealthSummaryDTO struct {
	UserID                    string    `json:"user_id"`
	HealthRiskScore           int       `json:"health_risk_score"`           // 0-100
	HealthRiskLevel           string    `json:"health_risk_level"`           // low/moderate/high/critical
	MonthlyMedicalExpenses    float64   `json:"monthly_medical_expenses"`
	MonthlyInsurancePremiums  float64   `json:"monthly_insurance_premiums"`
	AnnualDeductibleRemaining float64   `json:"annual_deductible_remaining"`
	OutOfPocketRemaining      float64   `json:"out_of_pocket_remaining"`
	TotalHealthCosts          float64   `json:"total_health_costs"`
	CoverageGapRisk           float64   `json:"coverage_gap_risk"`
	RecommendedEmergencyFund  float64   `json:"recommended_emergency_fund"`
	FinancialVulnerability    string    `json:"financial_vulnerability"`     // secure/moderate/vulnerable/critical
	PriorityAdjustment        float64   `json:"priority_adjustment"`         // multiplier for purchase decisions
	ActiveConditionsCount     int       `json:"active_conditions_count"`
	HighRiskConditionsCount   int       `json:"high_risk_conditions_count"`
	CostReductionOpportunities int      `json:"cost_reduction_opportunities"`
	UpdatedAt                 time.Time `json:"updated_at"`
}

// ToDomain conversion methods

// ToDomain converts CreateHealthProfileDTO to domain.HealthProfile
func (dto CreateHealthProfileDTO) ToDomain(userID string) *domain.HealthProfile {
	profile := &domain.HealthProfile{
		UserID:     userID,
		Age:        dto.Age,
		Gender:     dto.Gender,
		Height:     dto.Height,
		Weight:     dto.Weight,
		FamilySize: dto.FamilySize,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	
	// Calculate BMI
	bmi, _ := profile.CalculateBMI()
	profile.BMI = bmi
	
	return profile
}

// ToDomain converts AddMedicalConditionDTO to domain.MedicalCondition
func (dto AddMedicalConditionDTO) ToDomain(userID, profileID string) (*domain.MedicalCondition, error) {
	diagnosedDate, err := time.Parse("2006-01-02", dto.DiagnosedDate)
	if err != nil {
		return nil, err
	}
	
	isActive := true
	if dto.IsActive != nil {
		isActive = *dto.IsActive
	}
	
	requiresMedication := false
	if dto.RequiresMedication != nil {
		requiresMedication = *dto.RequiresMedication
	}
	
	riskFactor := 0.1
	if dto.RiskFactor > 0 {
		riskFactor = dto.RiskFactor
	}
	
	return &domain.MedicalCondition{
		UserID:             userID,
		ProfileID:          profileID,
		Name:               dto.Name,
		Category:           dto.Category,
		Severity:           dto.Severity,
		DiagnosedDate:      diagnosedDate,
		IsActive:           isActive,
		RequiresMedication: requiresMedication,
		MonthlyMedCost:     dto.MonthlyMedCost,
		RiskFactor:         riskFactor,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}, nil
}

// ToDomain converts AddMedicalExpenseDTO to domain.MedicalExpense
func (dto AddMedicalExpenseDTO) ToDomain(userID, profileID string) (*domain.MedicalExpense, error) {
	date, err := time.Parse("2006-01-02", dto.Date)
	if err != nil {
		return nil, err
	}
	
	frequency := dto.Frequency
	if frequency == "" {
		if dto.IsRecurring {
			frequency = "monthly" // default for recurring
		} else {
			frequency = "one_time"
		}
	}
	
	return &domain.MedicalExpense{
		UserID:           userID,
		ProfileID:        profileID,
		Amount:           dto.Amount,
		Category:         dto.Category,
		Description:      dto.Description,
		IsRecurring:      dto.IsRecurring,
		Frequency:        frequency,
		IsCovered:        dto.IsCovered,
		InsurancePayment: dto.InsurancePayment,
		OutOfPocket:      dto.Amount - dto.InsurancePayment, // Calculate out-of-pocket
		Date:             date,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}, nil
}

// ToDomain converts AddInsurancePolicyDTO to domain.InsurancePolicy
func (dto AddInsurancePolicyDTO) ToDomain(userID, profileID string) (*domain.InsurancePolicy, error) {
	startDate, err := time.Parse("2006-01-02", dto.StartDate)
	if err != nil {
		return nil, err
	}
	
	endDate, err := time.Parse("2006-01-02", dto.EndDate)
	if err != nil {
		return nil, err
	}
	
	isActive := true
	if dto.IsActive != nil {
		isActive = *dto.IsActive
	}
	
	return &domain.InsurancePolicy{
		UserID:             userID,
		ProfileID:          profileID,
		Provider:           dto.Provider,
		PolicyNumber:       dto.PolicyNumber,
		Type:               dto.Type,
		MonthlyPremium:     dto.MonthlyPremium,
		Deductible:         dto.Deductible,
		OutOfPocketMax:     dto.OutOfPocketMax,
		CoveragePercentage: float64(dto.CoveragePercentage),
		StartDate:          startDate,
		EndDate:            endDate,
		IsActive:           isActive,
		DeductibleMet:      dto.DeductibleMet,
		OutOfPocketCurrent: dto.OutOfPocketCurrent,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}, nil
}

// FromDomain conversion methods

// FromDomain converts domain.HealthProfile to HealthProfileResponseDTO
func FromHealthProfile(profile *domain.HealthProfile) HealthProfileResponseDTO {
	return HealthProfileResponseDTO{
		ID:         profile.ID,
		UserID:     profile.UserID,
		Age:        profile.Age,
		Gender:     profile.Gender,
		Height:     profile.Height,
		Weight:     profile.Weight,
		BMI:        profile.BMI,
		FamilySize: profile.FamilySize,
		CreatedAt:  profile.CreatedAt,
		UpdatedAt:  profile.UpdatedAt,
	}
}

// FromDomain converts domain.MedicalCondition to MedicalConditionResponseDTO
func FromMedicalCondition(condition *domain.MedicalCondition) MedicalConditionResponseDTO {
	return MedicalConditionResponseDTO{
		ID:                 condition.ID,
		UserID:             condition.UserID,
		ProfileID:          condition.ProfileID,
		Name:               condition.Name,
		Category:           condition.Category,
		Severity:           condition.Severity,
		DiagnosedDate:      condition.DiagnosedDate,
		IsActive:           condition.IsActive,
		RequiresMedication: condition.RequiresMedication,
		MonthlyMedCost:     condition.MonthlyMedCost,
		RiskFactor:         condition.RiskFactor,
		CreatedAt:          condition.CreatedAt,
		UpdatedAt:          condition.UpdatedAt,
	}
}

// FromDomain converts domain.MedicalExpense to MedicalExpenseResponseDTO
func FromMedicalExpense(expense *domain.MedicalExpense) MedicalExpenseResponseDTO {
	return MedicalExpenseResponseDTO{
		ID:               expense.ID,
		UserID:           expense.UserID,
		ProfileID:        expense.ProfileID,
		Amount:           expense.Amount,
		Category:         expense.Category,
		Description:      expense.Description,
		IsRecurring:      expense.IsRecurring,
		Frequency:        expense.Frequency,
		IsCovered:        expense.IsCovered,
		InsurancePayment: expense.InsurancePayment,
		OutOfPocket:      expense.OutOfPocket,
		Date:             expense.Date,
		CreatedAt:        expense.CreatedAt,
		UpdatedAt:        expense.UpdatedAt,
	}
}

// FromDomain converts domain.InsurancePolicy to InsurancePolicyResponseDTO
func FromInsurancePolicy(policy *domain.InsurancePolicy) InsurancePolicyResponseDTO {
	return InsurancePolicyResponseDTO{
		ID:                 policy.ID,
		UserID:             policy.UserID,
		ProfileID:          policy.ProfileID,
		Provider:           policy.Provider,
		PolicyNumber:       policy.PolicyNumber,
		Type:               policy.Type,
		MonthlyPremium:     policy.MonthlyPremium,
		Deductible:         policy.Deductible,
		OutOfPocketMax:     policy.OutOfPocketMax,
		CoveragePercentage: int(policy.CoveragePercentage),
		StartDate:          policy.StartDate,
		EndDate:            policy.EndDate,
		IsActive:           policy.IsActive,
		DeductibleMet:      policy.DeductibleMet,
		OutOfPocketCurrent: policy.OutOfPocketCurrent,
		CreatedAt:          policy.CreatedAt,
		UpdatedAt:          policy.UpdatedAt,
	}
}

// FromDomain converts domain.HealthSummary to HealthSummaryDTO
func FromHealthSummary(summary *domain.HealthSummary, activeConditions, highRiskConditions, costReductionOpportunities int) HealthSummaryDTO {
	return HealthSummaryDTO{
		UserID:                    summary.UserID,
		HealthRiskScore:           summary.HealthRiskScore,
		HealthRiskLevel:           summary.HealthRiskLevel,
		MonthlyMedicalExpenses:    summary.MonthlyMedicalExpenses,
		MonthlyInsurancePremiums:  summary.MonthlyInsurancePremiums,
		AnnualDeductibleRemaining: summary.AnnualDeductibleRemaining,
		OutOfPocketRemaining:      summary.OutOfPocketRemaining,
		TotalHealthCosts:          summary.TotalHealthCosts,
		CoverageGapRisk:           summary.CoverageGapRisk,
		RecommendedEmergencyFund:  summary.RecommendedEmergencyFund,
		FinancialVulnerability:    summary.FinancialVulnerability,
		PriorityAdjustment:        summary.PriorityAdjustment,
		ActiveConditionsCount:     activeConditions,
		HighRiskConditionsCount:   highRiskConditions,
		CostReductionOpportunities: costReductionOpportunities,
		UpdatedAt:                 summary.UpdatedAt,
	}
}