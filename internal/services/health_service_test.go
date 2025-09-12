package services

import (
	"context"
	"testing"
	"time"

	"github.com/DuckDHD/BuyOrBye/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Mock repository interfaces
type MockHealthProfileRepository struct {
	mock.Mock
}

func (m *MockHealthProfileRepository) Create(ctx context.Context, profile *domain.HealthProfile) (*domain.HealthProfile, error) {
	args := m.Called(ctx, profile)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.HealthProfile), args.Error(1)
}

func (m *MockHealthProfileRepository) GetByID(ctx context.Context, id uint) (*domain.HealthProfile, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.HealthProfile), args.Error(1)
}

func (m *MockHealthProfileRepository) GetByUserID(ctx context.Context, userID string) (*domain.HealthProfile, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.HealthProfile), args.Error(1)
}

func (m *MockHealthProfileRepository) Update(ctx context.Context, profile *domain.HealthProfile) (*domain.HealthProfile, error) {
	args := m.Called(ctx, profile)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.HealthProfile), args.Error(1)
}

func (m *MockHealthProfileRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockHealthProfileRepository) GetWithRelations(ctx context.Context, userID string) (*domain.HealthProfile, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.HealthProfile), args.Error(1)
}

func (m *MockHealthProfileRepository) ExistsByUserID(ctx context.Context, userID string) (bool, error) {
	args := m.Called(ctx, userID)
	return args.Bool(0), args.Error(1)
}

type MockMedicalConditionRepository struct {
	mock.Mock
}

func (m *MockMedicalConditionRepository) Create(ctx context.Context, condition *domain.MedicalCondition) (*domain.MedicalCondition, error) {
	args := m.Called(ctx, condition)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.MedicalCondition), args.Error(1)
}

func (m *MockMedicalConditionRepository) GetByID(ctx context.Context, id string) (*domain.MedicalCondition, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.MedicalCondition), args.Error(1)
}

func (m *MockMedicalConditionRepository) Update(ctx context.Context, condition *domain.MedicalCondition) (*domain.MedicalCondition, error) {
	args := m.Called(ctx, condition)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.MedicalCondition), args.Error(1)
}

func (m *MockMedicalConditionRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockMedicalConditionRepository) GetByUserID(ctx context.Context, userID string, activeOnly bool) ([]*domain.MedicalCondition, error) {
	args := m.Called(ctx, userID, activeOnly)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.MedicalCondition), args.Error(1)
}

func (m *MockMedicalConditionRepository) GetByCategory(ctx context.Context, userID string, category string) ([]*domain.MedicalCondition, error) {
	args := m.Called(ctx, userID, category)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.MedicalCondition), args.Error(1)
}

func (m *MockMedicalConditionRepository) GetBySeverity(ctx context.Context, userID string, severity string) ([]*domain.MedicalCondition, error) {
	args := m.Called(ctx, userID, severity)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.MedicalCondition), args.Error(1)
}

func (m *MockMedicalConditionRepository) GetByProfileID(ctx context.Context, profileID string) ([]*domain.MedicalCondition, error) {
	args := m.Called(ctx, profileID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.MedicalCondition), args.Error(1)
}

func (m *MockMedicalConditionRepository) GetActiveConditionCount(ctx context.Context, userID string) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockMedicalConditionRepository) CalculateTotalRiskFactor(ctx context.Context, userID string) (float64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockMedicalConditionRepository) GetMedicationRequiringConditions(ctx context.Context, userID string) ([]*domain.MedicalCondition, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.MedicalCondition), args.Error(1)
}

type MockMedicalExpenseRepository struct {
	mock.Mock
}

func (m *MockMedicalExpenseRepository) Create(ctx context.Context, expense *domain.MedicalExpense) (*domain.MedicalExpense, error) {
	args := m.Called(ctx, expense)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.MedicalExpense), args.Error(1)
}

func (m *MockMedicalExpenseRepository) GetByID(ctx context.Context, id string) (*domain.MedicalExpense, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.MedicalExpense), args.Error(1)
}

func (m *MockMedicalExpenseRepository) Update(ctx context.Context, expense *domain.MedicalExpense) (*domain.MedicalExpense, error) {
	args := m.Called(ctx, expense)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.MedicalExpense), args.Error(1)
}

func (m *MockMedicalExpenseRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockMedicalExpenseRepository) GetByUserID(ctx context.Context, userID string) ([]*domain.MedicalExpense, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.MedicalExpense), args.Error(1)
}

func (m *MockMedicalExpenseRepository) GetByDateRange(ctx context.Context, userID string, startDate, endDate time.Time) ([]*domain.MedicalExpense, error) {
	args := m.Called(ctx, userID, startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.MedicalExpense), args.Error(1)
}

func (m *MockMedicalExpenseRepository) GetByCategory(ctx context.Context, userID string, category string) ([]*domain.MedicalExpense, error) {
	args := m.Called(ctx, userID, category)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.MedicalExpense), args.Error(1)
}

func (m *MockMedicalExpenseRepository) GetByFrequency(ctx context.Context, userID string, frequency string) ([]*domain.MedicalExpense, error) {
	args := m.Called(ctx, userID, frequency)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.MedicalExpense), args.Error(1)
}

func (m *MockMedicalExpenseRepository) GetRecurring(ctx context.Context, userID string) ([]*domain.MedicalExpense, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.MedicalExpense), args.Error(1)
}

func (m *MockMedicalExpenseRepository) GetByProfileID(ctx context.Context, profileID string) ([]*domain.MedicalExpense, error) {
	args := m.Called(ctx, profileID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.MedicalExpense), args.Error(1)
}

func (m *MockMedicalExpenseRepository) CalculateTotals(ctx context.Context, userID string, startDate, endDate time.Time) (*ExpenseTotals, error) {
	args := m.Called(ctx, userID, startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*ExpenseTotals), args.Error(1)
}

func (m *MockMedicalExpenseRepository) GetMonthlyRecurringTotal(ctx context.Context, userID string) (float64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(float64), args.Error(1)
}

func (m *MockMedicalExpenseRepository) GetAnnualProjectedExpenses(ctx context.Context, userID string) (float64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(float64), args.Error(1)
}

type MockInsurancePolicyRepository struct {
	mock.Mock
}

func (m *MockInsurancePolicyRepository) Create(ctx context.Context, policy *domain.InsurancePolicy) (*domain.InsurancePolicy, error) {
	args := m.Called(ctx, policy)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.InsurancePolicy), args.Error(1)
}

func (m *MockInsurancePolicyRepository) GetByID(ctx context.Context, id string) (*domain.InsurancePolicy, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.InsurancePolicy), args.Error(1)
}

func (m *MockInsurancePolicyRepository) Update(ctx context.Context, policy *domain.InsurancePolicy) (*domain.InsurancePolicy, error) {
	args := m.Called(ctx, policy)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.InsurancePolicy), args.Error(1)
}

func (m *MockInsurancePolicyRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockInsurancePolicyRepository) GetByUserID(ctx context.Context, userID string) ([]*domain.InsurancePolicy, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.InsurancePolicy), args.Error(1)
}

func (m *MockInsurancePolicyRepository) GetByType(ctx context.Context, userID string, policyType string) ([]*domain.InsurancePolicy, error) {
	args := m.Called(ctx, userID, policyType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.InsurancePolicy), args.Error(1)
}

func (m *MockInsurancePolicyRepository) GetByPolicyNumber(ctx context.Context, policyNumber string) (*domain.InsurancePolicy, error) {
	args := m.Called(ctx, policyNumber)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.InsurancePolicy), args.Error(1)
}

func (m *MockInsurancePolicyRepository) GetActivePolicies(ctx context.Context, userID string) ([]*domain.InsurancePolicy, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.InsurancePolicy), args.Error(1)
}

func (m *MockInsurancePolicyRepository) GetByProfileID(ctx context.Context, profileID string) ([]*domain.InsurancePolicy, error) {
	args := m.Called(ctx, profileID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.InsurancePolicy), args.Error(1)
}

func (m *MockInsurancePolicyRepository) UpdateDeductibleProgress(ctx context.Context, policyID string, deductibleMet, outOfPocketCurrent float64) (*domain.InsurancePolicy, error) {
	args := m.Called(ctx, policyID, deductibleMet, outOfPocketCurrent)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.InsurancePolicy), args.Error(1)
}

func (m *MockInsurancePolicyRepository) CalculateCoverageForExpense(ctx context.Context, policyID string, expenseAmount float64) (*CoverageCalculation, error) {
	args := m.Called(ctx, policyID, expenseAmount)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*CoverageCalculation), args.Error(1)
}

func (m *MockInsurancePolicyRepository) GetPoliciesByProvider(ctx context.Context, userID string, provider string) ([]*domain.InsurancePolicy, error) {
	args := m.Called(ctx, userID, provider)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.InsurancePolicy), args.Error(1)
}

type MockRiskCalculator struct {
	mock.Mock
}

func (m *MockRiskCalculator) CalculateHealthRiskScore(profile *domain.HealthProfile, conditions []domain.MedicalCondition) int {
	args := m.Called(profile, conditions)
	return args.Int(0)
}

func (m *MockRiskCalculator) AssessFinancialVulnerability(healthCosts, income float64) string {
	args := m.Called(healthCosts, income)
	return args.String(0)
}

func (m *MockRiskCalculator) RecommendEmergencyFund(riskScore int, monthlyExpenses float64) float64 {
	args := m.Called(riskScore, monthlyExpenses)
	return args.Get(0).(float64)
}

func (m *MockRiskCalculator) DetermineRiskLevel(score int) string {
	args := m.Called(score)
	return args.String(0)
}

type MockMedicalCostAnalyzer struct {
	mock.Mock
}

func (m *MockMedicalCostAnalyzer) CalculateMonthlyAverage(expenses []domain.MedicalExpense) float64 {
	args := m.Called(expenses)
	return args.Get(0).(float64)
}

func (m *MockMedicalCostAnalyzer) ProjectAnnualCosts(expenses []domain.MedicalExpense, conditions []domain.MedicalCondition) float64 {
	args := m.Called(expenses, conditions)
	return args.Get(0).(float64)
}

func (m *MockMedicalCostAnalyzer) IdentifyCostReductionOpportunities(expenses []domain.MedicalExpense) []CostReductionOpportunity {
	args := m.Called(expenses)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]CostReductionOpportunity)
}

func (m *MockMedicalCostAnalyzer) AnalyzeTrends(expenses []domain.MedicalExpense) []string {
	args := m.Called(expenses)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]string)
}

// Test cases
func TestHealthService_CreateProfile_Success(t *testing.T) {
	// Arrange
	mockProfileRepo := &MockHealthProfileRepository{}
	mockConditionRepo := &MockMedicalConditionRepository{}
	mockExpenseRepo := &MockMedicalExpenseRepository{}
	mockPolicyRepo := &MockInsurancePolicyRepository{}
	mockRiskCalc := &MockRiskCalculator{}
	mockCostAnalyzer := &MockMedicalCostAnalyzer{}

	service := NewHealthService(
		mockProfileRepo,
		mockConditionRepo,
		mockExpenseRepo,
		mockPolicyRepo,
		mockRiskCalc,
		mockCostAnalyzer,
	)

	profile := &domain.HealthProfile{
		UserID:     "user123",
		Age:        30,
		Gender:     "male",
		Height:     175.0,
		Weight:     70.0,
		FamilySize: 2,
	}

	// Set expectations
	mockProfileRepo.On("ExistsByUserID", mock.Anything, "user123").Return(false, nil)
	mockProfileRepo.On("Create", mock.Anything, mock.MatchedBy(func(p *domain.HealthProfile) bool {
		return p.UserID == "user123" && p.BMI > 0
	})).Return(profile, nil)

	// Act
	err := service.CreateProfile(context.Background(), profile)

	// Assert
	assert.NoError(t, err)
	assert.Greater(t, profile.BMI, 0.0, "BMI should be calculated")
	mockProfileRepo.AssertExpectations(t)
}

func TestHealthService_CreateProfile_UserAlreadyHasProfile(t *testing.T) {
	// Arrange
	mockProfileRepo := &MockHealthProfileRepository{}
	mockConditionRepo := &MockMedicalConditionRepository{}
	mockExpenseRepo := &MockMedicalExpenseRepository{}
	mockPolicyRepo := &MockInsurancePolicyRepository{}
	mockRiskCalc := &MockRiskCalculator{}
	mockCostAnalyzer := &MockMedicalCostAnalyzer{}

	service := NewHealthService(
		mockProfileRepo,
		mockConditionRepo,
		mockExpenseRepo,
		mockPolicyRepo,
		mockRiskCalc,
		mockCostAnalyzer,
	)

	profile := &domain.HealthProfile{
		UserID:     "user123",
		Age:        30,
		Gender:     "male",
		Height:     175.0,
		Weight:     70.0,
		FamilySize: 2,
	}

	// Set expectations
	mockProfileRepo.On("ExistsByUserID", mock.Anything, "user123").Return(true, nil)

	// Act
	err := service.CreateProfile(context.Background(), profile)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "user already has a health profile")
	mockProfileRepo.AssertExpectations(t)
}

func TestHealthService_CalculateHealthSummary_Success(t *testing.T) {
	// Arrange
	mockProfileRepo := &MockHealthProfileRepository{}
	mockConditionRepo := &MockMedicalConditionRepository{}
	mockExpenseRepo := &MockMedicalExpenseRepository{}
	mockPolicyRepo := &MockInsurancePolicyRepository{}
	mockRiskCalc := &MockRiskCalculator{}
	mockCostAnalyzer := &MockMedicalCostAnalyzer{}

	service := NewHealthService(
		mockProfileRepo,
		mockConditionRepo,
		mockExpenseRepo,
		mockPolicyRepo,
		mockRiskCalc,
		mockCostAnalyzer,
	)

	userID := "user123"
	profile := &domain.HealthProfile{
		UserID:     userID,
		Age:        45,
		Gender:     "female",
		Height:     165.0,
		Weight:     65.0,
		BMI:        23.9,
		FamilySize: 3,
		UpdatedAt:  time.Now(),
	}

	conditions := []*domain.MedicalCondition{
		{
			ID:       "cond1",
			UserID:   userID,
			Name:     "Hypertension",
			Category: "chronic",
			Severity: "moderate",
			IsActive: true,
		},
	}

	expenses := []*domain.MedicalExpense{
		{
			ID:          "exp1",
			UserID:      userID,
			Amount:      100.0,
			OutOfPocket: 100.0,
			Category:    "medication",
			IsRecurring: true,
			Frequency:   "monthly",
		},
	}

	policies := []*domain.InsurancePolicy{
		{
			ID:             "pol1",
			UserID:         userID,
			MonthlyPremium: 300.0,
			Deductible:     1500.0,
			DeductibleMet:  500.0,
		},
	}

	// Set expectations
	mockProfileRepo.On("GetByUserID", mock.Anything, userID).Return(profile, nil)
	mockConditionRepo.On("GetByUserID", mock.Anything, userID, true).Return(conditions, nil)
	mockExpenseRepo.On("GetByUserID", mock.Anything, userID).Return(expenses, nil)
	mockPolicyRepo.On("GetActivePolicies", mock.Anything, userID).Return(policies, nil)

	mockRiskCalc.On("CalculateHealthRiskScore", profile, mock.AnythingOfType("[]domain.MedicalCondition")).Return(35)
	mockRiskCalc.On("DetermineRiskLevel", 35).Return("moderate")
	mockRiskCalc.On("AssessFinancialVulnerability", mock.AnythingOfType("float64"), mock.AnythingOfType("float64")).Return("moderate")
	mockRiskCalc.On("RecommendEmergencyFund", 35, mock.AnythingOfType("float64")).Return(15000.0)

	mockCostAnalyzer.On("CalculateMonthlyAverage", mock.AnythingOfType("[]domain.MedicalExpense")).Return(100.0)
	mockCostAnalyzer.On("ProjectAnnualCosts", mock.AnythingOfType("[]domain.MedicalExpense"), mock.AnythingOfType("[]domain.MedicalCondition")).Return(1200.0)

	// Act
	summary, err := service.CalculateHealthSummary(context.Background(), userID)

	// Assert
	assert.NoError(t, err)
	require.NotNil(t, summary)
	assert.Equal(t, userID, summary.UserID)
	assert.Equal(t, 35, summary.HealthRiskScore)
	assert.Equal(t, "moderate", summary.HealthRiskLevel)
	assert.Equal(t, 100.0, summary.MonthlyMedicalExpenses)
	assert.Equal(t, 300.0, summary.MonthlyInsurancePremiums)
	assert.Equal(t, 15000.0, summary.RecommendedEmergencyFund)
	assert.Equal(t, "moderate", summary.FinancialVulnerability)
	assert.Greater(t, summary.PriorityAdjustment, 1.0, "Priority adjustment should be > 1.0 for moderate risk")

	mockProfileRepo.AssertExpectations(t)
	mockConditionRepo.AssertExpectations(t)
	mockExpenseRepo.AssertExpectations(t)
	mockPolicyRepo.AssertExpectations(t)
	mockRiskCalc.AssertExpectations(t)
	mockCostAnalyzer.AssertExpectations(t)
}

func TestHealthService_AddCondition_Success(t *testing.T) {
	// Arrange
	mockProfileRepo := &MockHealthProfileRepository{}
	mockConditionRepo := &MockMedicalConditionRepository{}
	mockExpenseRepo := &MockMedicalExpenseRepository{}
	mockPolicyRepo := &MockInsurancePolicyRepository{}
	mockRiskCalc := &MockRiskCalculator{}
	mockCostAnalyzer := &MockMedicalCostAnalyzer{}

	service := NewHealthService(
		mockProfileRepo,
		mockConditionRepo,
		mockExpenseRepo,
		mockPolicyRepo,
		mockRiskCalc,
		mockCostAnalyzer,
	)

	condition := &domain.MedicalCondition{
		UserID:             "user123",
		ProfileID:          "profile123",
		Name:               "Diabetes",
		Category:           "chronic",
		Severity:           "moderate",
		DiagnosedDate:      time.Now().AddDate(-1, 0, 0),
		RequiresMedication: true,
		MonthlyMedCost:     150.0,
		IsActive:           true,
	}

	// Set expectations
	mockConditionRepo.On("Create", mock.Anything, mock.MatchedBy(func(c *domain.MedicalCondition) bool {
		return c.Name == "Diabetes" && c.RiskFactor > 0
	})).Return(condition, nil)

	// Act
	err := service.AddCondition(context.Background(), condition)

	// Assert
	assert.NoError(t, err)
	assert.Greater(t, condition.RiskFactor, 0.0, "Risk factor should be calculated based on severity")
	mockConditionRepo.AssertExpectations(t)
}

func TestHealthService_AddInsurancePolicy_OverlapValidation(t *testing.T) {
	// Arrange
	mockProfileRepo := &MockHealthProfileRepository{}
	mockConditionRepo := &MockMedicalConditionRepository{}
	mockExpenseRepo := &MockMedicalExpenseRepository{}
	mockPolicyRepo := &MockInsurancePolicyRepository{}
	mockRiskCalc := &MockRiskCalculator{}
	mockCostAnalyzer := &MockMedicalCostAnalyzer{}

	service := NewHealthService(
		mockProfileRepo,
		mockConditionRepo,
		mockExpenseRepo,
		mockPolicyRepo,
		mockRiskCalc,
		mockCostAnalyzer,
	)

	// Existing active policy
	existingPolicy := &domain.InsurancePolicy{
		ID:        "existing1",
		UserID:    "user123",
		Type:      "health",
		IsActive:  true,
		StartDate: time.Now().AddDate(0, -6, 0), // 6 months ago
		EndDate:   time.Now().AddDate(1, 0, 0),  // 1 year from now
	}

	// New policy that overlaps
	newPolicy := &domain.InsurancePolicy{
		UserID:    "user123",
		Type:      "health",
		IsActive:  true,
		StartDate: time.Now().AddDate(0, -3, 0), // 3 months ago (overlaps)
		EndDate:   time.Now().AddDate(1, 6, 0),  // 1.5 years from now
	}

	// Set expectations
	mockPolicyRepo.On("GetByType", mock.Anything, "user123", "health").Return([]*domain.InsurancePolicy{existingPolicy}, nil)

	// Act
	err := service.AddInsurancePolicy(context.Background(), newPolicy)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "policy overlaps")
	mockPolicyRepo.AssertExpectations(t)
}