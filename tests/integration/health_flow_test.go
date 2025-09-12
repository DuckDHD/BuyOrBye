//go:build integration
// +build integration

package integration

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DuckDHD/BuyOrBye/internal/database"
	"github.com/DuckDHD/BuyOrBye/internal/domain"
	"github.com/DuckDHD/BuyOrBye/internal/dtos"
	"github.com/DuckDHD/BuyOrBye/internal/handlers"
	"github.com/DuckDHD/BuyOrBye/internal/models"
	"github.com/DuckDHD/BuyOrBye/internal/repositories"
	"github.com/DuckDHD/BuyOrBye/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mysql"
	"gorm.io/gorm"

	"github.com/DuckDHD/BuyOrBye/tests/testutils"
)

// HealthFlowTestSuite represents the comprehensive health flow test suite
type HealthFlowTestSuite struct {
	suite.Suite
	server *testutils.TestServer
	client *testutils.HTTPClient
}

// SetupSuite runs before all tests in the suite
func (s *HealthFlowTestSuite) SetupSuite() {
	testutils.SetupIntegrationTest()
	s.server = testutils.NewTestServer(s.T())
	s.client = testutils.NewHTTPClient(s.server.BaseURL)
}

// TearDownSuite runs after all tests in the suite
func (s *HealthFlowTestSuite) TearDownSuite() {
	if s.server != nil {
		s.server.Close()
	}
	testutils.TeardownIntegrationTest()
}

// SetupTest runs before each test
func (s *HealthFlowTestSuite) SetupTest() {
	// Reset database for clean state
	s.server.ResetDatabase(s.T())
	// Clear any existing token
	s.client.SetAccessToken("")
}

// TestCompleteHealthFlow tests the complete health flow from registration to risk calculation
func (s *HealthFlowTestSuite) TestCompleteHealthFlow() {
	t := s.T()
	
	// Step 1: User Registration
	user := testutils.NewTestUser("health.user@example.com", "Health User", "password123")
	user.Register(t, s.client)
	
	// Step 2: Create Health Profile
	profileData := dtos.CreateHealthProfileRequestDTO{
		Age:        35,
		Gender:     "male",
		Height:     175.0, // 175cm
		Weight:     80.0,  // 80kg
		FamilySize: 2,
	}
	
	resp, body := s.client.POST(t, "/api/v1/health/profile", profileData)
	assert.Equal(t, http.StatusCreated, resp.StatusCode, "Failed to create health profile: %s", string(body))
	
	var profileResp dtos.HealthProfileResponseDTO
	err := json.Unmarshal(body, &profileResp)
	require.NoError(t, err, "Failed to parse profile response")
	
	// Verify BMI calculation (80 / (1.75^2) = 26.12)
	assert.InDelta(t, 26.12, profileResp.BMI, 0.1, "BMI should be calculated correctly")
	
	// Step 3: Add Medical Conditions
	conditions := []dtos.CreateMedicalConditionRequestDTO{
		{
			Name:               "Hypertension",
			Category:           "chronic",
			Severity:           "moderate",
			DiagnosedDate:      time.Now().AddDate(-2, 0, 0), // 2 years ago
			RequiresMedication: true,
			MonthlyMedCost:     45.00,
		},
		{
			Name:               "Type 2 Diabetes",
			Category:           "chronic", 
			Severity:           "severe",
			DiagnosedDate:      time.Now().AddDate(-1, 0, 0), // 1 year ago
			RequiresMedication: true,
			MonthlyMedCost:     120.00,
		},
	}
	
	// Track initial risk score
	initialRiskScore := s.getHealthRiskScore(t, "/api/v1/health/summary")
	
	for _, condition := range conditions {
		resp, body := s.client.POST(t, "/api/v1/health/conditions", condition)
		assert.Equal(t, http.StatusCreated, resp.StatusCode, "Failed to add condition %s: %s", condition.Name, string(body))
	}
	
	// Step 4: Verify Risk Score Increased
	updatedRiskScore := s.getHealthRiskScore(t, "/api/v1/health/summary")
	assert.Greater(t, updatedRiskScore, initialRiskScore, "Risk score should increase after adding conditions")
	
	// Expected risk calculation: Age (35) = 5 points, BMI (26.12) = 8 points, Hypertension (moderate) = 5 points, Diabetes (severe) = 10 points
	// Total expected: ~28 points
	assert.InDelta(t, 28, updatedRiskScore, 5, "Risk score should be around 28 points")
	
	// Step 5: Add Medical Expenses
	expenses := []dtos.CreateMedicalExpenseRequestDTO{
		{
			Amount:      85.00,
			Category:    "doctor_visit",
			Description: "Quarterly diabetes checkup",
			IsRecurring: true,
			Frequency:   "quarterly",
			IsCovered:   false,
			Date:        time.Now().AddDate(0, 0, -30), // 30 days ago
		},
		{
			Amount:      250.00,
			Category:    "lab_test",
			Description: "Comprehensive blood panel",
			IsRecurring: false,
			IsCovered:   false,
			Date:        time.Now().AddDate(0, 0, -15), // 15 days ago
		},
		{
			Amount:      165.00, // Monthly medications (45+120)
			Category:    "medication",
			Description: "Monthly medication costs",
			IsRecurring: true,
			Frequency:   "monthly",
			IsCovered:   false,
			Date:        time.Now().AddDate(0, 0, -10), // 10 days ago
		},
	}
	
	for _, expense := range expenses {
		resp, body := s.client.POST(t, "/api/v1/health/expenses", expense)
		assert.Equal(t, http.StatusCreated, resp.StatusCode, "Failed to add expense: %s", string(body))
	}
	
	// Step 6: Verify Health Summary Calculations
	resp, body = s.client.GET(t, "/api/v1/health/summary")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Failed to get health summary")
	
	var summary dtos.HealthSummaryResponseDTO
	err = json.Unmarshal(body, &summary)
	require.NoError(t, err, "Failed to parse health summary")
	
	// Verify monthly medical expenses calculation
	// Expected: 165 (monthly meds) + 85/3 (quarterly visit) = ~193
	assert.Greater(t, summary.MonthlyMedicalExpenses, 190.0, "Monthly expenses should be calculated correctly")
	assert.Less(t, summary.MonthlyMedicalExpenses, 200.0, "Monthly expenses should be reasonable")
	
	// Verify risk level determination
	expectedRiskLevel := "moderate" // Risk score ~28 should be moderate (26-50)
	assert.Equal(t, expectedRiskLevel, summary.HealthRiskLevel, "Risk level should be moderate")
	
	// Verify recommended emergency fund is calculated
	assert.Greater(t, summary.RecommendedEmergencyFund, 1000.0, "Emergency fund should be recommended based on health risks")
	
	// Step 7: Test Financial Vulnerability Assessment
	assert.NotEmpty(t, summary.FinancialVulnerability, "Financial vulnerability should be assessed")
}

// TestInsuranceCoverageFlow tests complete insurance coverage workflow
func (s *HealthFlowTestSuite) TestInsuranceCoverageFlow() {
	t := s.T()
	
	// Setup: Create user and health profile
	user := testutils.NewTestUser("insurance.user@example.com", "Insurance User", "password123")
	user.Register(t, s.client)
	
	profileData := dtos.CreateHealthProfileRequestDTO{
		Age:        42,
		Gender:     "female",
		Height:     165.0,
		Weight:     70.0,
		FamilySize: 3,
	}
	
	resp, body := s.client.POST(t, "/api/v1/health/profile", profileData)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "Failed to create profile")
	
	// Step 1: Add Insurance Policy
	policyData := dtos.CreateInsurancePolicyRequestDTO{
		Provider:           "Blue Cross Blue Shield",
		PolicyNumber:       "BCBS-12345678",
		Type:               "health",
		MonthlyPremium:     450.00,
		Deductible:         2000.00,
		OutOfPocketMax:     6000.00,
		CoveragePercentage: 80.0, // 80% after deductible
		StartDate:          time.Now().AddDate(0, -6, 0), // 6 months ago
		EndDate:            time.Now().AddDate(1, 0, 0),  // 1 year from now
	}
	
	resp, body = s.client.POST(t, "/api/v1/health/policies", policyData)
	assert.Equal(t, http.StatusCreated, resp.StatusCode, "Failed to add insurance policy: %s", string(body))
	
	// Step 2: Add Medical Expense (covered by insurance)
	expenseData := dtos.CreateMedicalExpenseRequestDTO{
		Amount:           1200.00,
		Category:         "hospital",
		Description:      "Emergency room visit",
		IsRecurring:      false,
		IsCovered:        true,
		InsurancePayment: 0.00, // Will be calculated by service
		Date:            time.Now().AddDate(0, 0, -5), // 5 days ago
	}
	
	resp, body = s.client.POST(t, "/api/v1/health/expenses", expenseData)
	assert.Equal(t, http.StatusCreated, resp.StatusCode, "Failed to add covered expense: %s", string(body))
	
	// Step 3: Verify Insurance Coverage Applied
	resp, body = s.client.GET(t, "/api/v1/health/summary")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	
	var summary dtos.HealthSummaryResponseDTO
	err := json.Unmarshal(body, &summary)
	require.NoError(t, err)
	
	// Verify insurance premiums are included
	assert.Equal(t, 450.00, summary.MonthlyInsurancePremiums, "Monthly premiums should match policy")
	
	// Verify deductible tracking
	assert.Greater(t, summary.AnnualDeductibleRemaining, 0.0, "Deductible remaining should be tracked")
	assert.Less(t, summary.AnnualDeductibleRemaining, 2000.00, "Some deductible should have been applied")
	
	// Step 4: Test High-Cost Medical Expense with Insurance
	highCostExpense := dtos.CreateMedicalExpenseRequestDTO{
		Amount:      5000.00,
		Category:    "hospital",
		Description: "Surgical procedure",
		IsRecurring: false,
		IsCovered:   true,
		Date:        time.Now().AddDate(0, 0, -1), // Yesterday
	}
	
	resp, body = s.client.POST(t, "/api/v1/health/expenses", highCostExpense)
	assert.Equal(t, http.StatusCreated, resp.StatusCode, "Failed to add high-cost expense: %s", string(body))
	
	// Step 5: Verify Insurance Calculations Update
	resp, body = s.client.GET(t, "/api/v1/health/summary")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	
	err = json.Unmarshal(body, &summary)
	require.NoError(t, err)
	
	// After high-cost expense, deductible should be met and coverage should apply
	expectedCoverageGapRisk := summary.CoverageGapRisk
	assert.Greater(t, expectedCoverageGapRisk, 0.0, "Coverage gap should be calculated")
}

// TestProfileUniquenessConstraint tests that only one health profile can exist per user
func (s *HealthFlowTestSuite) TestProfileUniquenessConstraint() {
	t := s.T()
	
	// Setup: Create user
	user := testutils.NewTestUser("unique.user@example.com", "Unique User", "password123")
	user.Register(t, s.client)
	
	// Step 1: Create first health profile
	profileData := dtos.CreateHealthProfileRequestDTO{
		Age:        30,
		Gender:     "other",
		Height:     170.0,
		Weight:     65.0,
		FamilySize: 1,
	}
	
	resp, body := s.client.POST(t, "/api/v1/health/profile", profileData)
	assert.Equal(t, http.StatusCreated, resp.StatusCode, "First profile should be created successfully")
	
	// Step 2: Attempt to create second health profile
	secondProfileData := dtos.CreateHealthProfileRequestDTO{
		Age:        31,
		Gender:     "male",
		Height:     180.0,
		Weight:     75.0,
		FamilySize: 2,
	}
	
	resp, body = s.client.POST(t, "/api/v1/health/profile", secondProfileData)
	assert.Equal(t, http.StatusConflict, resp.StatusCode, "Second profile creation should fail: %s", string(body))
	
	// Verify error message indicates uniqueness constraint
	assert.Contains(t, string(body), "already has a health profile", "Error message should indicate constraint violation")
	
	// Step 3: Verify profile update works (replacing existing profile)
	updateData := dtos.UpdateHealthProfileRequestDTO{
		Age:        32,
		Gender:     "female",
		Height:     168.0,
		Weight:     62.0,
		FamilySize: 1,
	}
	
	resp, body = s.client.PUT(t, "/api/v1/health/profile", updateData)
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Profile update should succeed: %s", string(body))
	
	var updatedProfile dtos.HealthProfileResponseDTO
	err := json.Unmarshal(body, &updatedProfile)
	require.NoError(t, err)
	
	assert.Equal(t, 32, updatedProfile.Age, "Profile should be updated with new age")
	assert.Equal(t, "female", updatedProfile.Gender, "Profile should be updated with new gender")
}

// TestCascadeDeleteFunctionality tests that deleting a profile removes all related data
func (s *HealthFlowTestSuite) TestCascadeDeleteFunctionality() {
	t := s.T()
	
	// Setup: Create user with complete health data
	user := testutils.NewTestUser("cascade.user@example.com", "Cascade User", "password123")
	user.Register(t, s.client)
	
	// Create health profile
	profileData := dtos.CreateHealthProfileRequestDTO{
		Age:        45,
		Gender:     "male",
		Height:     178.0,
		Weight:     85.0,
		FamilySize: 4,
	}
	
	resp, body := s.client.POST(t, "/api/v1/health/profile", profileData)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "Profile creation should succeed")
	
	// Add medical conditions
	conditionData := dtos.CreateMedicalConditionRequestDTO{
		Name:               "Asthma",
		Category:           "chronic",
		Severity:           "mild",
		DiagnosedDate:      time.Now().AddDate(-5, 0, 0), // 5 years ago
		RequiresMedication: true,
		MonthlyMedCost:     25.00,
	}
	
	resp, body = s.client.POST(t, "/api/v1/health/conditions", conditionData)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "Condition creation should succeed")
	
	// Add medical expenses
	expenseData := dtos.CreateMedicalExpenseRequestDTO{
		Amount:      75.00,
		Category:    "medication",
		Description: "Asthma inhaler",
		IsRecurring: true,
		Frequency:   "monthly",
		IsCovered:   false,
		Date:        time.Now().AddDate(0, 0, -7), // 7 days ago
	}
	
	resp, body = s.client.POST(t, "/api/v1/health/expenses", expenseData)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "Expense creation should succeed")
	
	// Add insurance policy
	policyData := dtos.CreateInsurancePolicyRequestDTO{
		Provider:           "Aetna Health",
		PolicyNumber:       "AETNA-87654321",
		Type:               "health",
		MonthlyPremium:     380.00,
		Deductible:         1500.00,
		OutOfPocketMax:     5000.00,
		CoveragePercentage: 85.0,
		StartDate:          time.Now().AddDate(0, -3, 0), // 3 months ago
		EndDate:            time.Now().AddDate(1, 0, 0),  // 1 year from now
	}
	
	resp, body = s.client.POST(t, "/api/v1/health/policies", policyData)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "Policy creation should succeed")
	
	// Verify all data exists
	resp, _ = s.client.GET(t, "/api/v1/health/profile")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Profile should exist")
	
	resp, _ = s.client.GET(t, "/api/v1/health/conditions")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Conditions should exist")
	
	resp, _ = s.client.GET(t, "/api/v1/health/expenses")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Expenses should exist")
	
	resp, _ = s.client.GET(t, "/api/v1/health/policies")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Policies should exist")
	
	// Delete health profile
	resp, body = s.client.DELETE(t, "/api/v1/health/profile")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Profile deletion should succeed: %s", string(body))
	
	// Verify all related data is removed
	resp, _ = s.client.GET(t, "/api/v1/health/profile")
	assert.Equal(t, http.StatusNotFound, resp.StatusCode, "Profile should not exist after deletion")
	
	resp, body = s.client.GET(t, "/api/v1/health/conditions")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Conditions endpoint should be accessible")
	// Parse response to check empty array
	var conditions []dtos.MedicalConditionResponseDTO
	err := json.Unmarshal(body, &conditions)
	require.NoError(t, err)
	assert.Empty(t, conditions, "Conditions should be empty after profile deletion")
	
	resp, body = s.client.GET(t, "/api/v1/health/expenses")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Expenses endpoint should be accessible")
	var expenses []dtos.MedicalExpenseResponseDTO
	err = json.Unmarshal(body, &expenses)
	require.NoError(t, err)
	assert.Empty(t, expenses, "Expenses should be empty after profile deletion")
	
	resp, body = s.client.GET(t, "/api/v1/health/policies")
	assert.Equal(t, http.StatusOK, resp.StatusCode, "Policies endpoint should be accessible")
	var policies []dtos.InsurancePolicyResponseDTO
	err = json.Unmarshal(body, &policies)
	require.NoError(t, err)
	assert.Empty(t, policies, "Policies should be empty after profile deletion")
}

// TestRiskScoreTransitions tests how risk score changes as conditions are added/modified
func (s *HealthFlowTestSuite) TestRiskScoreTransitions() {
	t := s.T()
	
	// Setup: Create user with baseline health profile
	user := testutils.NewTestUser("risk.user@example.com", "Risk User", "password123")
	user.Register(t, s.client)
	
	// Create healthy young profile (low risk baseline)
	profileData := dtos.CreateHealthProfileRequestDTO{
		Age:        25,
		Gender:     "female",
		Height:     165.0,
		Weight:     60.0, // BMI ~22 (normal)
		FamilySize: 1,
	}
	
	resp, body := s.client.POST(t, "/api/v1/health/profile", profileData)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "Profile creation should succeed")
	
	// Baseline risk score (should be low)
	baselineRisk := s.getHealthRiskScore(t, "/api/v1/health/summary")
	assert.LessOrEqual(t, baselineRisk, 10, "Young healthy person should have low risk score")
	
	// Step 1: Add mild condition
	mildCondition := dtos.CreateMedicalConditionRequestDTO{
		Name:               "Seasonal Allergies",
		Category:           "preventive",
		Severity:           "mild",
		DiagnosedDate:      time.Now().AddDate(-1, 0, 0),
		RequiresMedication: false,
		MonthlyMedCost:     0.00,
	}
	
	resp, body = s.client.POST(t, "/api/v1/health/conditions", mildCondition)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "Mild condition should be added")
	
	mildRisk := s.getHealthRiskScore(t, "/api/v1/health/summary")
	// Preventive conditions shouldn't increase risk much
	assert.LessOrEqual(t, mildRisk, baselineRisk+2, "Preventive mild condition should barely affect risk")
	
	// Step 2: Add moderate chronic condition
	moderateCondition := dtos.CreateMedicalConditionRequestDTO{
		Name:               "Hypothyroidism",
		Category:           "chronic",
		Severity:           "moderate",
		DiagnosedDate:      time.Now().AddDate(-2, 0, 0),
		RequiresMedication: true,
		MonthlyMedCost:     30.00,
	}
	
	resp, body = s.client.POST(t, "/api/v1/health/conditions", moderateCondition)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "Moderate condition should be added")
	
	moderateRisk := s.getHealthRiskScore(t, "/api/v1/health/summary")
	assert.Greater(t, moderateRisk, mildRisk, "Moderate chronic condition should increase risk")
	assert.InDelta(t, mildRisk+5, moderateRisk, 3, "Moderate condition should add ~5 risk points")
	
	// Step 3: Add severe chronic condition
	severeCondition := dtos.CreateMedicalConditionRequestDTO{
		Name:               "Chronic Kidney Disease",
		Category:           "chronic",
		Severity:           "severe",
		DiagnosedDate:      time.Now().AddDate(-3, 0, 0),
		RequiresMedication: true,
		MonthlyMedCost:     200.00,
	}
	
	resp, body = s.client.POST(t, "/api/v1/health/conditions", severeCondition)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "Severe condition should be added")
	
	severeRisk := s.getHealthRiskScore(t, "/api/v1/health/summary")
	assert.Greater(t, severeRisk, moderateRisk, "Severe chronic condition should significantly increase risk")
	assert.InDelta(t, moderateRisk+10, severeRisk, 3, "Severe condition should add ~10 risk points")
	
	// Step 4: Add critical condition (should push to high risk)
	criticalCondition := dtos.CreateMedicalConditionRequestDTO{
		Name:               "Advanced Heart Disease",
		Category:           "chronic",
		Severity:           "critical",
		DiagnosedDate:      time.Now().AddDate(-1, -6, 0),
		RequiresMedication: true,
		MonthlyMedCost:     350.00,
	}
	
	resp, body = s.client.POST(t, "/api/v1/health/conditions", criticalCondition)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "Critical condition should be added")
	
	criticalRisk := s.getHealthRiskScore(t, "/api/v1/health/summary")
	assert.Greater(t, criticalRisk, severeRisk, "Critical condition should greatly increase risk")
	assert.Greater(t, criticalRisk, 30, "Critical condition should push into high risk territory")
	
	// Verify risk level progression
	resp, body = s.client.GET(t, "/api/v1/health/summary")
	require.Equal(t, http.StatusOK, resp.StatusCode)
	
	var summary dtos.HealthSummaryResponseDTO
	err := json.Unmarshal(body, &summary)
	require.NoError(t, err)
	
	// With critical condition, should be "high" or "critical" risk level
	assert.Contains(t, []string{"high", "critical"}, summary.HealthRiskLevel, "Should be in high risk category")
	
	// Test deactivating condition reduces risk
	// Note: This would require an update condition endpoint, which is assumed to exist
	// This test demonstrates the expected behavior pattern
}

// TestFinancialVulnerabilityAssessment tests vulnerability calculations with different expense levels
func (s *HealthFlowTestSuite) TestFinancialVulnerabilityAssessment() {
	t := s.T()
	
	// Setup: Create user
	user := testutils.NewTestUser("vulnerability.user@example.com", "Vulnerability User", "password123")
	user.Register(t, s.client)
	
	// Create profile
	profileData := dtos.CreateHealthProfileRequestDTO{
		Age:        55,
		Gender:     "male",
		Height:     175.0,
		Weight:     90.0, // BMI ~29 (overweight)
		FamilySize: 3,
	}
	
	resp, body := s.client.POST(t, "/api/v1/health/profile", profileData)
	require.Equal(t, http.StatusCreated, resp.StatusCode, "Profile creation should succeed")
	
	// Test Scenario 1: Low medical expenses (secure)
	lowExpenses := []dtos.CreateMedicalExpenseRequestDTO{
		{
			Amount:      50.00,
			Category:    "doctor_visit",
			Description: "Annual checkup",
			IsRecurring: true,
			Frequency:   "annually",
			IsCovered:   true,
			InsurancePayment: 40.00,
			Date:        time.Now().AddDate(0, 0, -30),
		},
	}
	
	for _, expense := range lowExpenses {
		resp, body := s.client.POST(t, "/api/v1/health/expenses", expense)
		require.Equal(t, http.StatusCreated, resp.StatusCode, "Low expense should be added")
	}
	
	resp, body = s.client.GET(t, "/api/v1/health/summary")
	require.Equal(t, http.StatusOK, resp.StatusCode)
	
	var lowSummary dtos.HealthSummaryResponseDTO
	err := json.Unmarshal(body, &lowSummary)
	require.NoError(t, err)
	
	// With low expenses, should be "secure" or "moderate" vulnerability
	assert.Contains(t, []string{"secure", "moderate"}, lowSummary.FinancialVulnerability, "Low expenses should result in lower vulnerability")
	
	// Clear expenses for next scenario
	s.server.ResetDatabase(t)
	user.Register(t, s.client)
	resp, body = s.client.POST(t, "/api/v1/health/profile", profileData)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	
	// Test Scenario 2: High medical expenses (vulnerable/critical)
	highExpenses := []dtos.CreateMedicalExpenseRequestDTO{
		{
			Amount:      800.00,
			Category:    "medication",
			Description: "Specialty medications",
			IsRecurring: true,
			Frequency:   "monthly",
			IsCovered:   true,
			InsurancePayment: 400.00, // 50% coverage
			Date:        time.Now().AddDate(0, 0, -15),
		},
		{
			Amount:      1500.00,
			Category:    "therapy",
			Description: "Physical therapy sessions",
			IsRecurring: true,
			Frequency:   "monthly",
			IsCovered:   true,
			InsurancePayment: 900.00, // 60% coverage
			Date:        time.Now().AddDate(0, 0, -20),
		},
		{
			Amount:      3000.00,
			Category:    "hospital",
			Description: "Emergency procedure",
			IsRecurring: false,
			IsCovered:   true,
			InsurancePayment: 2400.00, // 80% coverage
			Date:        time.Now().AddDate(0, 0, -5),
		},
	}
	
	for _, expense := range highExpenses {
		resp, body := s.client.POST(t, "/api/v1/health/expenses", expense)
		require.Equal(t, http.StatusCreated, resp.StatusCode, "High expense should be added")
	}
	
	resp, body = s.client.GET(t, "/api/v1/health/summary")
	require.Equal(t, http.StatusOK, resp.StatusCode)
	
	var highSummary dtos.HealthSummaryResponseDTO
	err = json.Unmarshal(body, &highSummary)
	require.NoError(t, err)
	
	// With high expenses, should be "vulnerable" or "critical"
	assert.Contains(t, []string{"vulnerable", "critical"}, highSummary.FinancialVulnerability, "High expenses should result in higher vulnerability")
	
	// Verify monthly costs are significantly higher
	assert.Greater(t, highSummary.MonthlyMedicalExpenses, lowSummary.MonthlyMedicalExpenses+1000, "High expense scenario should show much higher monthly costs")
	
	// Verify emergency fund recommendation increases with vulnerability
	assert.Greater(t, highSummary.RecommendedEmergencyFund, lowSummary.RecommendedEmergencyFund, "Higher vulnerability should recommend larger emergency fund")
}

// Helper function to extract health risk score from summary
func (s *HealthFlowTestSuite) getHealthRiskScore(t *testing.T, endpoint string) int {
	resp, body := s.client.GET(t, endpoint)
	require.Equal(t, http.StatusOK, resp.StatusCode, "Summary endpoint should be accessible")
	
	var summary dtos.HealthSummaryResponseDTO
	err := json.Unmarshal(body, &summary)
	require.NoError(t, err, "Should parse summary response")
	
	return summary.HealthRiskScore
}

func TestHealthFlow_CompleteUserJourney_Success(t *testing.T) {
	// Setup test container and database
	ctx := context.Background()
	db := setupTestDatabase(t, ctx)
	defer cleanupTestDatabase(t, ctx, db)

	// Setup services and handler
	router := setupHealthRouter(db)

	// Test data
	userID := uuid.New()
	
	// Step 1: Create Health Profile
	profileReq := dtos.HealthProfileRequestDTO{
		UserID: userID,
		Age:    35,
		Gender: "male",
		Height: 175.5,
		Weight: 80.0,
	}

	profileJSON, _ := json.Marshal(profileReq)
	w1 := httptest.NewRecorder()
	req1 := httptest.NewRequest("POST", "/api/health/profiles", bytes.NewBuffer(profileJSON))
	req1.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w1, req1)

	require.Equal(t, http.StatusCreated, w1.Code)
	
	var profileResp dtos.HealthProfileResponseDTO
	err := json.Unmarshal(w1.Body.Bytes(), &profileResp)
	require.NoError(t, err)
	profileID := profileResp.ID

	// Step 2: Add Medical Conditions
	conditions := []dtos.MedicalConditionRequestDTO{
		{
			ProfileID:   profileID,
			Name:        "Hypertension",
			Severity:    "moderate",
			DiagnosedAt: time.Now().AddDate(-2, 0, 0),
			Status:      "active",
		},
		{
			ProfileID:   profileID,
			Name:        "Type 2 Diabetes",
			Severity:    "severe",
			DiagnosedAt: time.Now().AddDate(-1, 0, 0),
			Status:      "active",
		},
	}

	var conditionIDs []uuid.UUID
	for _, condition := range conditions {
		condJSON, _ := json.Marshal(condition)
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/health/conditions", bytes.NewBuffer(condJSON))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusCreated, w.Code)
		var condResp dtos.MedicalConditionResponseDTO
		err := json.Unmarshal(w.Body.Bytes(), &condResp)
		require.NoError(t, err)
		conditionIDs = append(conditionIDs, condResp.ID)
	}

	// Step 3: Add Insurance Policy
	policyReq := dtos.InsurancePolicyRequestDTO{
		ProfileID:        profileID,
		Provider:         "HealthCorp Insurance",
		PolicyNumber:     "HC-2024-001",
		CoverageType:     "comprehensive",
		CoverageAmount:   500000.0,
		Deductible:       5000.0,
		MonthlyPremium:   450.0,
		StartDate:        time.Now().AddDate(0, -6, 0),
		EndDate:          time.Now().AddDate(1, 6, 0),
		Status:           "active",
	}

	policyJSON, _ := json.Marshal(policyReq)
	w3 := httptest.NewRecorder()
	req3 := httptest.NewRequest("POST", "/api/health/policies", bytes.NewBuffer(policyJSON))
	req3.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w3, req3)

	require.Equal(t, http.StatusCreated, w3.Code)
	var policyResp dtos.InsurancePolicyResponseDTO
	err = json.Unmarshal(w3.Body.Bytes(), &policyResp)
	require.NoError(t, err)
	policyID := policyResp.ID

	// Step 4: Add Medical Expenses
	expenses := []dtos.MedicalExpenseRequestDTO{
		{
			ProfileID:     profileID,
			ConditionID:   &conditionIDs[0],
			PolicyID:      &policyID,
			Description:   "Monthly hypertension medication",
			Amount:        150.0,
			ExpenseType:   "medication",
			ExpenseDate:   time.Now().AddDate(0, -1, 0),
			Provider:      "City Pharmacy",
			CoveredAmount: 120.0,
			Status:        "approved",
		},
		{
			ProfileID:     profileID,
			ConditionID:   &conditionIDs[1],
			PolicyID:      &policyID,
			Description:   "Diabetes specialist consultation",
			Amount:        300.0,
			ExpenseType:   "consultation",
			ExpenseDate:   time.Now().AddDate(0, -1, -15),
			Provider:      "Diabetes Care Center",
			CoveredAmount: 240.0,
			Status:        "approved",
		},
		{
			ProfileID:   profileID,
			Description: "Emergency room visit",
			Amount:      2500.0,
			ExpenseType: "emergency",
			ExpenseDate: time.Now().AddDate(0, -2, 0),
			Provider:    "City Hospital ER",
			Status:      "pending",
		},
	}

	var expenseIDs []uuid.UUID
	for _, expense := range expenses {
		expJSON, _ := json.Marshal(expense)
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/health/expenses", bytes.NewBuffer(expJSON))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		require.Equal(t, http.StatusCreated, w.Code)
		var expResp dtos.MedicalExpenseResponseDTO
		err := json.Unmarshal(w.Body.Bytes(), &expResp)
		require.NoError(t, err)
		expenseIDs = append(expenseIDs, expResp.ID)
	}

	// Step 5: Calculate Risk Score
	w5 := httptest.NewRecorder()
	req5 := httptest.NewRequest("GET", fmt.Sprintf("/api/health/profiles/%s/risk", profileID), nil)
	router.ServeHTTP(w5, req5)

	require.Equal(t, http.StatusOK, w5.Code)
	var riskResp map[string]interface{}
	err = json.Unmarshal(w5.Body.Bytes(), &riskResp)
	require.NoError(t, err)

	// Verify risk calculation
	assert.NotZero(t, riskResp["risk_score"])
	assert.Contains(t, riskResp, "risk_factors")
	riskScore := riskResp["risk_score"].(float64)
	assert.Greater(t, riskScore, 0.6) // Should be high risk due to diabetes + hypertension

	// Step 6: Get Health Summary
	w6 := httptest.NewRecorder()
	req6 := httptest.NewRequest("GET", fmt.Sprintf("/api/health/profiles/%s/summary", profileID), nil)
	router.ServeHTTP(w6, req6)

	require.Equal(t, http.StatusOK, w6.Code)
	var summaryResp dtos.HealthSummaryResponseDTO
	err = json.Unmarshal(w6.Body.Bytes(), &summaryResp)
	require.NoError(t, err)

	// Verify summary data
	assert.Equal(t, profileID, summaryResp.ProfileID)
	assert.Equal(t, 2, summaryResp.TotalConditions)
	assert.Equal(t, 1, summaryResp.TotalPolicies)
	assert.Equal(t, 3, summaryResp.TotalExpenses)
	assert.Equal(t, 2950.0, summaryResp.TotalExpenseAmount)
	assert.Equal(t, 360.0, summaryResp.TotalCoveredAmount)
	assert.NotZero(t, summaryResp.RiskScore)
}

func TestHealthFlow_InsuranceCoverageApplication_Success(t *testing.T) {
	ctx := context.Background()
	db := setupTestDatabase(t, ctx)
	defer cleanupTestDatabase(t, ctx, db)

	router := setupHealthRouter(db)
	userID := uuid.New()

	// Create profile
	profileReq := dtos.HealthProfileRequestDTO{
		UserID: userID,
		Age:    28,
		Gender: "female",
		Height: 165.0,
		Weight: 60.0,
	}

	profileJSON, _ := json.Marshal(profileReq)
	w1 := httptest.NewRecorder()
	req1 := httptest.NewRequest("POST", "/api/health/profiles", bytes.NewBuffer(profileJSON))
	req1.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w1, req1)

	require.Equal(t, http.StatusCreated, w1.Code)
	var profileResp dtos.HealthProfileResponseDTO
	json.Unmarshal(w1.Body.Bytes(), &profileResp)
	profileID := profileResp.ID

	// Add insurance policy with specific coverage
	policyReq := dtos.InsurancePolicyRequestDTO{
		ProfileID:        profileID,
		Provider:         "Premium Health",
		PolicyNumber:     "PH-2024-TEST",
		CoverageType:     "premium",
		CoverageAmount:   1000000.0,
		Deductible:       2000.0,
		MonthlyPremium:   800.0,
		StartDate:        time.Now().AddDate(0, -3, 0),
		EndDate:          time.Now().AddDate(1, 9, 0),
		Status:           "active",
	}

	policyJSON, _ := json.Marshal(policyReq)
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("POST", "/api/health/policies", bytes.NewBuffer(policyJSON))
	req2.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w2, req2)

	require.Equal(t, http.StatusCreated, w2.Code)
	var policyResp dtos.InsurancePolicyResponseDTO
	json.Unmarshal(w2.Body.Bytes(), &policyResp)
	policyID := policyResp.ID

	// Add expense that should be covered
	expenseReq := dtos.MedicalExpenseRequestDTO{
		ProfileID:     profileID,
		PolicyID:      &policyID,
		Description:   "Annual health checkup",
		Amount:        500.0,
		ExpenseType:   "preventive",
		ExpenseDate:   time.Now(),
		Provider:      "Family Health Clinic",
		CoveredAmount: 450.0, // 90% coverage after deductible considerations
		Status:        "approved",
	}

	expenseJSON, _ := json.Marshal(expenseReq)
	w3 := httptest.NewRecorder()
	req3 := httptest.NewRequest("POST", "/api/health/expenses", bytes.NewBuffer(expenseJSON))
	req3.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w3, req3)

	require.Equal(t, http.StatusCreated, w3.Code)
	var expenseResp dtos.MedicalExpenseResponseDTO
	json.Unmarshal(w3.Body.Bytes(), &expenseResp)

	// Verify coverage was applied correctly
	assert.Equal(t, 500.0, expenseResp.Amount)
	assert.Equal(t, 450.0, expenseResp.CoveredAmount)
	assert.Equal(t, "approved", expenseResp.Status)

	// Verify policy is linked
	assert.NotNil(t, expenseResp.PolicyID)
	assert.Equal(t, policyID, *expenseResp.PolicyID)
}

func TestHealthFlow_ProfileUniqueness_ShouldFail(t *testing.T) {
	ctx := context.Background()
	db := setupTestDatabase(t, ctx)
	defer cleanupTestDatabase(t, ctx, db)

	router := setupHealthRouter(db)
	userID := uuid.New()

	// Create first profile
	profileReq := dtos.HealthProfileRequestDTO{
		UserID: userID,
		Age:    30,
		Gender: "male",
		Height: 180.0,
		Weight: 75.0,
	}

	profileJSON, _ := json.Marshal(profileReq)
	
	// First request should succeed
	w1 := httptest.NewRecorder()
	req1 := httptest.NewRequest("POST", "/api/health/profiles", bytes.NewBuffer(profileJSON))
	req1.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w1, req1)

	require.Equal(t, http.StatusCreated, w1.Code)

	// Second request with same userID should fail
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("POST", "/api/health/profiles", bytes.NewBuffer(profileJSON))
	req2.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusConflict, w2.Code)
}

func TestHealthFlow_CascadeDelete_Success(t *testing.T) {
	ctx := context.Background()
	db := setupTestDatabase(t, ctx)
	defer cleanupTestDatabase(t, ctx, db)

	router := setupHealthRouter(db)
	userID := uuid.New()

	// Create complete health profile with all related data
	profileReq := dtos.HealthProfileRequestDTO{
		UserID: userID,
		Age:    45,
		Gender: "female",
		Height: 170.0,
		Weight: 65.0,
	}

	profileJSON, _ := json.Marshal(profileReq)
	w1 := httptest.NewRecorder()
	req1 := httptest.NewRequest("POST", "/api/health/profiles", bytes.NewBuffer(profileJSON))
	req1.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w1, req1)

	require.Equal(t, http.StatusCreated, w1.Code)
	var profileResp dtos.HealthProfileResponseDTO
	json.Unmarshal(w1.Body.Bytes(), &profileResp)
	profileID := profileResp.ID

	// Add condition
	conditionReq := dtos.MedicalConditionRequestDTO{
		ProfileID:   profileID,
		Name:        "Test Condition",
		Severity:    "mild",
		DiagnosedAt: time.Now(),
		Status:      "active",
	}

	condJSON, _ := json.Marshal(conditionReq)
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("POST", "/api/health/conditions", bytes.NewBuffer(condJSON))
	req2.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w2, req2)
	require.Equal(t, http.StatusCreated, w2.Code)

	// Add policy
	policyReq := dtos.InsurancePolicyRequestDTO{
		ProfileID:      profileID,
		Provider:       "Test Insurance",
		PolicyNumber:   "TEST-001",
		CoverageType:   "basic",
		CoverageAmount: 100000.0,
		Deductible:     1000.0,
		MonthlyPremium: 200.0,
		StartDate:      time.Now(),
		EndDate:        time.Now().AddDate(1, 0, 0),
		Status:         "active",
	}

	policyJSON, _ := json.Marshal(policyReq)
	w3 := httptest.NewRecorder()
	req3 := httptest.NewRequest("POST", "/api/health/policies", bytes.NewBuffer(policyJSON))
	req3.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w3, req3)
	require.Equal(t, http.StatusCreated, w3.Code)

	// Add expense
	expenseReq := dtos.MedicalExpenseRequestDTO{
		ProfileID:   profileID,
		Description: "Test Expense",
		Amount:      100.0,
		ExpenseType: "consultation",
		ExpenseDate: time.Now(),
		Provider:    "Test Provider",
		Status:      "pending",
	}

	expenseJSON, _ := json.Marshal(expenseReq)
	w4 := httptest.NewRecorder()
	req4 := httptest.NewRequest("POST", "/api/health/expenses", bytes.NewBuffer(expenseJSON))
	req4.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w4, req4)
	require.Equal(t, http.StatusCreated, w4.Code)

	// Verify data exists
	var count int64
	db.Model(&models.HealthProfileModel{}).Where("id = ?", profileID).Count(&count)
	assert.Equal(t, int64(1), count)

	db.Model(&models.MedicalConditionModel{}).Where("profile_id = ?", profileID).Count(&count)
	assert.Equal(t, int64(1), count)

	db.Model(&models.InsurancePolicyModel{}).Where("profile_id = ?", profileID).Count(&count)
	assert.Equal(t, int64(1), count)

	db.Model(&models.MedicalExpenseModel{}).Where("profile_id = ?", profileID).Count(&count)
	assert.Equal(t, int64(1), count)

	// Delete profile
	w5 := httptest.NewRecorder()
	req5 := httptest.NewRequest("DELETE", fmt.Sprintf("/api/health/profiles/%s", profileID), nil)
	router.ServeHTTP(w5, req5)

	require.Equal(t, http.StatusNoContent, w5.Code)

	// Verify cascade delete worked
	db.Model(&models.HealthProfileModel{}).Where("id = ?", profileID).Count(&count)
	assert.Equal(t, int64(0), count)

	db.Model(&models.MedicalConditionModel{}).Where("profile_id = ?", profileID).Count(&count)
	assert.Equal(t, int64(0), count)

	db.Model(&models.InsurancePolicyModel{}).Where("profile_id = ?", profileID).Count(&count)
	assert.Equal(t, int64(0), count)

	db.Model(&models.MedicalExpenseModel{}).Where("profile_id = ?", profileID).Count(&count)
	assert.Equal(t, int64(0), count)
}

func TestHealthFlow_RiskScoreTransitions_Success(t *testing.T) {
	ctx := context.Background()
	db := setupTestDatabase(t, ctx)
	defer cleanupTestDatabase(t, ctx, db)

	router := setupHealthRouter(db)
	userID := uuid.New()

	// Create healthy profile
	profileReq := dtos.HealthProfileRequestDTO{
		UserID: userID,
		Age:    25,
		Gender: "male",
		Height: 180.0,
		Weight: 75.0,
	}

	profileJSON, _ := json.Marshal(profileReq)
	w1 := httptest.NewRecorder()
	req1 := httptest.NewRequest("POST", "/api/health/profiles", bytes.NewBuffer(profileJSON))
	req1.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w1, req1)

	require.Equal(t, http.StatusCreated, w1.Code)
	var profileResp dtos.HealthProfileResponseDTO
	json.Unmarshal(w1.Body.Bytes(), &profileResp)
	profileID := profileResp.ID

	// Calculate initial risk (should be low)
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("GET", fmt.Sprintf("/api/health/profiles/%s/risk", profileID), nil)
	router.ServeHTTP(w2, req2)

	require.Equal(t, http.StatusOK, w2.Code)
	var initialRisk map[string]interface{}
	json.Unmarshal(w2.Body.Bytes(), &initialRisk)
	initialScore := initialRisk["risk_score"].(float64)
	assert.Less(t, initialScore, 0.3) // Should be low risk

	// Add mild condition
	conditionReq1 := dtos.MedicalConditionRequestDTO{
		ProfileID:   profileID,
		Name:        "Seasonal Allergies",
		Severity:    "mild",
		DiagnosedAt: time.Now(),
		Status:      "active",
	}

	condJSON1, _ := json.Marshal(conditionReq1)
	w3 := httptest.NewRecorder()
	req3 := httptest.NewRequest("POST", "/api/health/conditions", bytes.NewBuffer(condJSON1))
	req3.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w3, req3)
	require.Equal(t, http.StatusCreated, w3.Code)

	// Calculate risk after mild condition
	w4 := httptest.NewRecorder()
	req4 := httptest.NewRequest("GET", fmt.Sprintf("/api/health/profiles/%s/risk", profileID), nil)
	router.ServeHTTP(w4, req4)

	require.Equal(t, http.StatusOK, w4.Code)
	var mildRisk map[string]interface{}
	json.Unmarshal(w4.Body.Bytes(), &mildRisk)
	mildScore := mildRisk["risk_score"].(float64)
	assert.Greater(t, mildScore, initialScore) // Should increase slightly

	// Add severe condition
	conditionReq2 := dtos.MedicalConditionRequestDTO{
		ProfileID:   profileID,
		Name:        "Heart Disease",
		Severity:    "severe",
		DiagnosedAt: time.Now(),
		Status:      "active",
	}

	condJSON2, _ := json.Marshal(conditionReq2)
	w5 := httptest.NewRecorder()
	req5 := httptest.NewRequest("POST", "/api/health/conditions", bytes.NewBuffer(condJSON2))
	req5.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w5, req5)
	require.Equal(t, http.StatusCreated, w5.Code)

	// Calculate risk after severe condition
	w6 := httptest.NewRecorder()
	req6 := httptest.NewRequest("GET", fmt.Sprintf("/api/health/profiles/%s/risk", profileID), nil)
	router.ServeHTTP(w6, req6)

	require.Equal(t, http.StatusOK, w6.Code)
	var severeRisk map[string]interface{}
	json.Unmarshal(w6.Body.Bytes(), &severeRisk)
	severeScore := severeRisk["risk_score"].(float64)
	assert.Greater(t, severeScore, mildScore) // Should increase significantly
	assert.Greater(t, severeScore, 0.7) // Should be high risk
}

func TestHealthFlow_FinancialVulnerability_Success(t *testing.T) {
	ctx := context.Background()
	db := setupTestDatabase(t, ctx)
	defer cleanupTestDatabase(t, ctx, db)

	router := setupHealthRouter(db)
	userID := uuid.New()

	// Create profile
	profileReq := dtos.HealthProfileRequestDTO{
		UserID: userID,
		Age:    40,
		Gender: "female",
		Height: 165.0,
		Weight: 70.0,
	}

	profileJSON, _ := json.Marshal(profileReq)
	w1 := httptest.NewRecorder()
	req1 := httptest.NewRequest("POST", "/api/health/profiles", bytes.NewBuffer(profileJSON))
	req1.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w1, req1)

	require.Equal(t, http.StatusCreated, w1.Code)
	var profileResp dtos.HealthProfileResponseDTO
	json.Unmarshal(w1.Body.Bytes(), &profileResp)
	profileID := profileResp.ID

	// Add multiple high-cost expenses
	expenses := []dtos.MedicalExpenseRequestDTO{
		{
			ProfileID:   profileID,
			Description: "Emergency surgery",
			Amount:      50000.0,
			ExpenseType: "surgery",
			ExpenseDate: time.Now().AddDate(0, -1, 0),
			Provider:    "City Hospital",
			Status:      "approved",
		},
		{
			ProfileID:   profileID,
			Description: "Cancer treatment",
			Amount:      75000.0,
			ExpenseType: "treatment",
			ExpenseDate: time.Now().AddDate(0, -2, 0),
			Provider:    "Cancer Center",
			Status:      "approved",
		},
		{
			ProfileID:   profileID,
			Description: "Specialist consultations",
			Amount:      15000.0,
			ExpenseType: "consultation",
			ExpenseDate: time.Now().AddDate(0, -3, 0),
			Provider:    "Medical Group",
			Status:      "approved",
		},
	}

	for _, expense := range expenses {
		expJSON, _ := json.Marshal(expense)
		w := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/health/expenses", bytes.NewBuffer(expJSON))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		require.Equal(t, http.StatusCreated, w.Code)
	}

	// Get summary to check financial impact
	w2 := httptest.NewRecorder()
	req2 := httptest.NewRequest("GET", fmt.Sprintf("/api/health/profiles/%s/summary", profileID), nil)
	router.ServeHTTP(w2, req2)

	require.Equal(t, http.StatusOK, w2.Code)
	var summaryResp dtos.HealthSummaryResponseDTO
	json.Unmarshal(w2.Body.Bytes(), &summaryResp)

	// Verify high expense totals indicate financial vulnerability
	assert.Equal(t, 140000.0, summaryResp.TotalExpenseAmount)
	assert.Equal(t, 3, summaryResp.TotalExpenses)
	assert.Equal(t, 0.0, summaryResp.TotalCoveredAmount) // No insurance coverage
	
	// High expenses without coverage should indicate vulnerability
	outOfPocket := summaryResp.TotalExpenseAmount - summaryResp.TotalCoveredAmount
	assert.Equal(t, 140000.0, outOfPocket)
}

// Helper functions

func setupTestDatabase(t *testing.T, ctx context.Context) *gorm.DB {
	mysqlContainer, err := mysql.Run(ctx,
		"mysql:8.0",
		mysql.WithDatabase("testdb"),
		mysql.WithUsername("testuser"),
		mysql.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			testcontainers.ForSQL("3306/tcp", "mysql", func(host string, port int) string {
				return fmt.Sprintf("testuser:testpass@tcp(%s:%d)/testdb", host, port)
			}),
		),
	)
	require.NoError(t, err)

	host, err := mysqlContainer.Host(ctx)
	require.NoError(t, err)

	port, err := mysqlContainer.MappedPort(ctx, "3306/tcp")
	require.NoError(t, err)

	dsn := fmt.Sprintf("testuser:testpass@tcp(%s:%s)/testdb?charset=utf8mb4&parseTime=True&loc=Local",
		host, port.Port())

	db, err := database.ConnectMySQL(dsn)
	require.NoError(t, err)

	// Run migrations
	err = db.AutoMigrate(
		&models.HealthProfileModel{},
		&models.MedicalConditionModel{},
		&models.InsurancePolicyModel{},
		&models.MedicalExpenseModel{},
	)
	require.NoError(t, err)

	// Store container for cleanup
	t.Cleanup(func() {
		mysqlContainer.Terminate(ctx)
	})

	return db
}

func cleanupTestDatabase(t *testing.T, ctx context.Context, db *gorm.DB) {
	// Clean up test data
	db.Exec("DELETE FROM medical_expenses")
	db.Exec("DELETE FROM insurance_policies") 
	db.Exec("DELETE FROM medical_conditions")
	db.Exec("DELETE FROM health_profiles")
}

func setupHealthRouter(db *gorm.DB) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.Default()

	// Setup repositories
	healthProfileRepo := repositories.NewHealthProfileRepository(db)
	medicalConditionRepo := repositories.NewMedicalConditionRepository(db)
	insurancePolicyRepo := repositories.NewInsurancePolicyRepository(db)
	medicalExpenseRepo := repositories.NewMedicalExpenseRepository(db)

	// Setup services
	riskCalculator := services.NewRiskCalculator()
	costAnalyzer := services.NewMedicalCostAnalyzer()
	insuranceEvaluator := services.NewInsuranceEvaluator()

	healthService := services.NewHealthService(
		healthProfileRepo,
		medicalConditionRepo,
		insurancePolicyRepo,
		medicalExpenseRepo,
		riskCalculator,
		costAnalyzer,
		insuranceEvaluator,
	)

	// Setup handlers
	healthHandler := handlers.NewHealthHandler(healthService)

	// Setup routes
	api := router.Group("/api")
	health := api.Group("/health")
	{
		// Profile routes
		health.POST("/profiles", healthHandler.CreateProfile)
		health.GET("/profiles/:id", healthHandler.GetProfile)
		health.PUT("/profiles/:id", healthHandler.UpdateProfile)
		health.DELETE("/profiles/:id", healthHandler.DeleteProfile)
		health.GET("/profiles/:id/summary", healthHandler.GetHealthSummary)
		health.GET("/profiles/:id/risk", healthHandler.CalculateRisk)

		// Condition routes
		health.POST("/conditions", healthHandler.CreateCondition)
		health.GET("/conditions/:id", healthHandler.GetCondition)
		health.PUT("/conditions/:id", healthHandler.UpdateCondition)
		health.DELETE("/conditions/:id", healthHandler.DeleteCondition)
		health.GET("/profiles/:profileId/conditions", healthHandler.GetConditionsByProfile)

		// Policy routes
		health.POST("/policies", healthHandler.CreatePolicy)
		health.GET("/policies/:id", healthHandler.GetPolicy)
		health.PUT("/policies/:id", healthHandler.UpdatePolicy)
		health.DELETE("/policies/:id", healthHandler.DeletePolicy)
		health.GET("/profiles/:profileId/policies", healthHandler.GetPoliciesByProfile)

		// Expense routes
		health.POST("/expenses", healthHandler.CreateExpense)
		health.GET("/expenses/:id", healthHandler.GetExpense)
		health.PUT("/expenses/:id", healthHandler.UpdateExpense)
		health.DELETE("/expenses/:id", healthHandler.DeleteExpense)
		health.GET("/profiles/:profileId/expenses", healthHandler.GetExpensesByProfile)
	}

	return router
}

// Run the test suite
func TestHealthFlowTestSuite(t *testing.T) {
	suite.Run(t, new(HealthFlowTestSuite))
}