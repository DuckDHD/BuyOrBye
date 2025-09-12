package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/DuckDHD/BuyOrBye/internal/domain"
	"github.com/DuckDHD/BuyOrBye/internal/dtos"
)

// MockHealthService for testing
type MockHealthService struct {
	mock.Mock
}

func (m *MockHealthService) CreateProfile(ctx context.Context, profile *domain.HealthProfile) error {
	args := m.Called(ctx, profile)
	return args.Error(0)
}

func (m *MockHealthService) GetProfile(ctx context.Context, userID string) (*domain.HealthProfile, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(*domain.HealthProfile), args.Error(1)
}

func (m *MockHealthService) UpdateProfile(ctx context.Context, profile *domain.HealthProfile) error {
	args := m.Called(ctx, profile)
	return args.Error(0)
}

func (m *MockHealthService) AddCondition(ctx context.Context, condition *domain.MedicalCondition) error {
	args := m.Called(ctx, condition)
	return args.Error(0)
}

func (m *MockHealthService) GetConditions(ctx context.Context, userID string) ([]domain.MedicalCondition, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]domain.MedicalCondition), args.Error(1)
}

func (m *MockHealthService) UpdateCondition(ctx context.Context, condition *domain.MedicalCondition) error {
	args := m.Called(ctx, condition)
	return args.Error(0)
}

func (m *MockHealthService) RemoveCondition(ctx context.Context, userID, conditionID string) error {
	args := m.Called(ctx, userID, conditionID)
	return args.Error(0)
}

func (m *MockHealthService) AddExpense(ctx context.Context, expense *domain.MedicalExpense) error {
	args := m.Called(ctx, expense)
	return args.Error(0)
}

func (m *MockHealthService) GetExpenses(ctx context.Context, userID string) ([]domain.MedicalExpense, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]domain.MedicalExpense), args.Error(1)
}

func (m *MockHealthService) GetRecurringExpenses(ctx context.Context, userID string) ([]domain.MedicalExpense, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]domain.MedicalExpense), args.Error(1)
}

func (m *MockHealthService) AddInsurancePolicy(ctx context.Context, policy *domain.InsurancePolicy) error {
	args := m.Called(ctx, policy)
	return args.Error(0)
}

func (m *MockHealthService) GetActivePolicies(ctx context.Context, userID string) ([]domain.InsurancePolicy, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]domain.InsurancePolicy), args.Error(1)
}

func (m *MockHealthService) UpdateDeductibleProgress(ctx context.Context, policyID string, amount float64) error {
	args := m.Called(ctx, policyID, amount)
	return args.Error(0)
}

func (m *MockHealthService) CalculateHealthSummary(ctx context.Context, userID string) (*domain.HealthSummary, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(*domain.HealthSummary), args.Error(1)
}

// Helper function to create JWT token for testing
func createTestJWTToken(userID string) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour).Unix(),
	})
	
	tokenString, _ := token.SignedString([]byte("test-secret"))
	return tokenString
}

// Helper function to setup test router with auth middleware
func setupHealthTestRouter(handler *HealthHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	
	// Add auth middleware that sets user context
	router.Use(func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			token, _ := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				return []byte("test-secret"), nil
			})
			
			if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
				c.Set("user_id", claims["user_id"])
			}
		}
		c.Next()
	})
	
	// Register health routes
	health := router.Group("/health")
	{
		health.POST("/profile", handler.CreateProfile)
		health.GET("/profile", handler.GetProfile)
		health.PUT("/profile", handler.UpdateProfile)
		health.POST("/conditions", handler.AddCondition)
		health.GET("/conditions", handler.GetConditions)
		health.PUT("/conditions/:id", handler.UpdateCondition)
		health.DELETE("/conditions/:id", handler.RemoveCondition)
		health.POST("/expenses", handler.AddExpense)
		health.GET("/expenses", handler.GetExpenses)
		health.GET("/expenses/recurring", handler.GetRecurringExpenses)
		health.POST("/policies", handler.AddInsurancePolicy)
		health.GET("/policies", handler.GetActivePolicies)
		health.PUT("/policies/:id/deductible", handler.UpdateDeductibleProgress)
		health.GET("/summary", handler.GetHealthSummary)
	}
	
	return router
}

func TestCreateProfile_Success(t *testing.T) {
	// Arrange
	mockService := new(MockHealthService)
	handler := NewHealthHandler(mockService)
	router := setupHealthTestRouter(handler)
	
	profileDTO := dtos.CreateHealthProfileRequestDTO{
		UserID:               "user123",
		Age:                  30,
		Gender:               "male",
		Height:               175.0,
		Weight:               70.0,
		FamilySize:           2,
		HasChronicConditions: false,
		EmergencyFundHealth:  1000.0,
	}
	
	mockService.On("CreateProfile", mock.Anything, mock.MatchedBy(func(profile *domain.HealthProfile) bool {
		return profile.UserID == "user123" && profile.Age == 30
	})).Return(nil)
	
	reqBody, _ := json.Marshal(profileDTO)
	req := httptest.NewRequest("POST", "/health/profile", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+createTestJWTToken("user123"))
	
	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(t, http.StatusCreated, w.Code)
	mockService.AssertExpectations(t)
}

func TestCreateProfile_DuplicateUser(t *testing.T) {
	// Arrange
	mockService := new(MockHealthService)
	handler := NewHealthHandler(mockService)
	router := setupHealthTestRouter(handler)
	
	profileDTO := dtos.CreateHealthProfileRequestDTO{
		UserID:               "user123",
		Age:                  30,
		Gender:               "male",
		Height:               175.0,
		Weight:               70.0,
		FamilySize:           2,
		HasChronicConditions: false,
		EmergencyFundHealth:  1000.0,
	}
	
	mockService.On("CreateProfile", mock.Anything, mock.AnythingOfType("*domain.HealthProfile")).
		Return(fmt.Errorf("user already has a health profile"))
	
	reqBody, _ := json.Marshal(profileDTO)
	req := httptest.NewRequest("POST", "/health/profile", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+createTestJWTToken("user123"))
	
	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(t, http.StatusConflict, w.Code)
	mockService.AssertExpectations(t)
}

func TestGetProfile_Success(t *testing.T) {
	// Arrange
	mockService := new(MockHealthService)
	handler := NewHealthHandler(mockService)
	router := setupHealthTestRouter(handler)
	
	expectedProfile := &domain.HealthProfile{
		ID:                   "profile123",
		UserID:               "user123",
		Age:                  30,
		Gender:               "male",
		Height:               175.0,
		Weight:               70.0,
		BMI:                  22.86,
		FamilySize:           2,
		HasChronicConditions: false,
		EmergencyFundHealth:  1000.0,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}
	
	mockService.On("GetProfile", mock.Anything, "user123").Return(expectedProfile, nil)
	
	req := httptest.NewRequest("GET", "/health/profile", nil)
	req.Header.Set("Authorization", "Bearer "+createTestJWTToken("user123"))
	
	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	
	var responseDTO dtos.HealthProfileResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &responseDTO)
	assert.NoError(t, err)
	assert.Equal(t, "user123", responseDTO.UserID)
	assert.Equal(t, 30, responseDTO.Age)
	
	mockService.AssertExpectations(t)
}

func TestAddCondition_ValidationError(t *testing.T) {
	// Arrange
	mockService := new(MockHealthService)
	handler := NewHealthHandler(mockService)
	router := setupHealthTestRouter(handler)
	
	// Invalid condition DTO - missing required fields
	conditionDTO := dtos.CreateMedicalConditionRequestDTO{
		Name: "", // Empty name should trigger validation error
		// Missing other required fields
	}
	
	reqBody, _ := json.Marshal(conditionDTO)
	req := httptest.NewRequest("POST", "/health/conditions", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+createTestJWTToken("user123"))
	
	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
	
	var errorResponse map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)
	assert.Contains(t, errorResponse, "error")
	
	// Service should not be called due to validation failure
	mockService.AssertNotCalled(t, "AddCondition")
}

func TestAddExpense_RequiresAuth(t *testing.T) {
	// Arrange
	mockService := new(MockHealthService)
	handler := NewHealthHandler(mockService)
	router := setupHealthTestRouter(handler)
	
	expenseDTO := dtos.CreateMedicalExpenseRequestDTO{
		UserID:           "user123",
		ProfileID:        "profile123",
		Amount:           100.0,
		Category:         "doctor_visit",
		Description:      "Annual checkup",
		Date:             time.Now(),
		IsCovered:        true,
		InsurancePayment: 80.0,
		OutOfPocket:      20.0,
		IsRecurring:      false,
		Frequency:        "one_time",
	}
	
	reqBody, _ := json.Marshal(expenseDTO)
	req := httptest.NewRequest("POST", "/health/expenses", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	// No Authorization header - should trigger 401
	
	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	
	// Service should not be called due to auth failure
	mockService.AssertNotCalled(t, "AddExpense")
}

func TestGetHealthSummary_Success(t *testing.T) {
	// Arrange
	mockService := new(MockHealthService)
	handler := NewHealthHandler(mockService)
	router := setupHealthTestRouter(handler)
	
	expectedSummary := &domain.HealthSummary{
		UserID:                    "user123",
		HealthRiskScore:           45,
		HealthRiskLevel:           "moderate",
		MonthlyMedicalExpenses:    250.0,
		MonthlyInsurancePremiums:  150.0,
		AnnualDeductibleRemaining: 1000.0,
		OutOfPocketRemaining:      500.0,
		TotalHealthCosts:          400.0,
		CoverageGapRisk:           200.0,
		RecommendedEmergencyFund:  2000.0,
		FinancialVulnerability:    "moderate",
		PriorityAdjustment:        1.1,
		UpdatedAt:                 time.Now(),
	}
	
	mockService.On("CalculateHealthSummary", mock.Anything, "user123").Return(expectedSummary, nil)
	
	req := httptest.NewRequest("GET", "/health/summary", nil)
	req.Header.Set("Authorization", "Bearer "+createTestJWTToken("user123"))
	
	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	
	var responseDTO dtos.HealthSummaryResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &responseDTO)
	assert.NoError(t, err)
	assert.Equal(t, "user123", responseDTO.UserID)
	assert.Equal(t, 45, responseDTO.HealthRiskScore)
	assert.Equal(t, "moderate", responseDTO.HealthRiskLevel)
	
	mockService.AssertExpectations(t)
}

func TestAddInsurancePolicy_UniqueNumber(t *testing.T) {
	// Arrange
	mockService := new(MockHealthService)
	handler := NewHealthHandler(mockService)
	router := setupHealthTestRouter(handler)
	
	policyDTO := dtos.CreateInsurancePolicyRequestDTO{
		UserID:             "user123",
		PolicyNumber:       "POL123456",
		Provider:           "HealthCorp",
		Type:               "health",
		CoveragePercentage: 80.0,
		Deductible:         1000.0,
		OutOfPocketMax:     5000.0,
		MonthlyPremium:     200.0,
		StartDate:          time.Now(),
		EndDate:            time.Now().AddDate(1, 0, 0),
		IsActive:           true,
	}
	
	// Mock service to return duplicate policy error
	mockService.On("AddInsurancePolicy", mock.Anything, mock.AnythingOfType("*domain.InsurancePolicy")).
		Return(fmt.Errorf("policy with number POL123456 already exists"))
	
	reqBody, _ := json.Marshal(policyDTO)
	req := httptest.NewRequest("POST", "/health/policies", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+createTestJWTToken("user123"))
	
	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(t, http.StatusConflict, w.Code)
	
	var errorResponse map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)
	assert.Contains(t, errorResponse, "error")
	
	mockService.AssertExpectations(t)
}

func TestUpdateDeductible_OnlyOwner(t *testing.T) {
	// Arrange
	mockService := new(MockHealthService)
	handler := NewHealthHandler(mockService)
	router := setupHealthTestRouter(handler)
	
	deductibleDTO := dtos.UpdateDeductibleRequestDTO{
		Amount: 500.0,
	}
	
	// Mock service to return permission denied error for different user
	mockService.On("UpdateDeductibleProgress", mock.Anything, "policy123", 500.0).
		Return(fmt.Errorf("user not authorized to update this policy"))
	
	reqBody, _ := json.Marshal(deductibleDTO)
	req := httptest.NewRequest("PUT", "/health/policies/policy123/deductible", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+createTestJWTToken("user456")) // Different user
	
	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(t, http.StatusForbidden, w.Code)
	
	var errorResponse map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)
	assert.Contains(t, errorResponse, "error")
	
	mockService.AssertExpectations(t)
}

// Additional comprehensive tests for better coverage

func TestCreateProfile_Unauthorized(t *testing.T) {
	mockService := new(MockHealthService)
	handler := NewHealthHandler(mockService)
	router := setupHealthTestRouter(handler)
	
	profileDTO := dtos.CreateHealthProfileRequestDTO{
		UserID:               "user123",
		Age:                  30,
		Gender:               "male",
		Height:               175.0,
		Weight:               70.0,
		FamilySize:           2,
		HasChronicConditions: false,
		EmergencyFundHealth:  1000.0,
	}
	
	reqBody, _ := json.Marshal(profileDTO)
	req := httptest.NewRequest("POST", "/health/profile", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	mockService.AssertNotCalled(t, "CreateProfile")
}

func TestGetProfile_NotFound(t *testing.T) {
	mockService := new(MockHealthService)
	handler := NewHealthHandler(mockService)
	router := setupHealthTestRouter(handler)
	
	mockService.On("GetProfile", mock.Anything, "user123").
		Return((*domain.HealthProfile)(nil), fmt.Errorf("profile not found"))
	
	req := httptest.NewRequest("GET", "/health/profile", nil)
	req.Header.Set("Authorization", "Bearer "+createTestJWTToken("user123"))
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusNotFound, w.Code)
	mockService.AssertExpectations(t)
}

func TestUpdateProfile_Success(t *testing.T) {
	mockService := new(MockHealthService)
	handler := NewHealthHandler(mockService)
	router := setupHealthTestRouter(handler)
	
	existingProfile := &domain.HealthProfile{
		ID:                   "profile123",
		UserID:               "user123",
		Age:                  30,
		Gender:               "male",
		Height:               175.0,
		Weight:               70.0,
		BMI:                  22.86,
		FamilySize:           2,
		HasChronicConditions: false,
		EmergencyFundHealth:  1000.0,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}
	
	updateDTO := dtos.UpdateHealthProfileRequestDTO{
		Age:                  35,
		Weight:               75.0,
		FamilySize:           3, // Add required field
		HasChronicConditions: true,
		EmergencyFundHealth:  1500.0, // Add required field
	}
	
	mockService.On("GetProfile", mock.Anything, "user123").Return(existingProfile, nil)
	mockService.On("UpdateProfile", mock.Anything, mock.MatchedBy(func(profile *domain.HealthProfile) bool {
		return profile.Age == 35 && profile.Weight == 75.0 && profile.FamilySize == 3 && 
		       profile.HasChronicConditions == true && profile.EmergencyFundHealth == 1500.0
	})).Return(nil)
	
	reqBody, _ := json.Marshal(updateDTO)
	req := httptest.NewRequest("PUT", "/health/profile", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+createTestJWTToken("user123"))
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}

func TestGetConditions_Success(t *testing.T) {
	mockService := new(MockHealthService)
	handler := NewHealthHandler(mockService)
	router := setupHealthTestRouter(handler)
	
	conditions := []domain.MedicalCondition{
		{
			ID:       "condition1",
			UserID:   "user123",
			Name:     "Diabetes",
			Category: "chronic",
			Severity: "moderate",
		},
	}
	
	mockService.On("GetConditions", mock.Anything, "user123").Return(conditions, nil)
	
	req := httptest.NewRequest("GET", "/health/conditions", nil)
	req.Header.Set("Authorization", "Bearer "+createTestJWTToken("user123"))
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response dtos.MedicalConditionListResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 1, response.Total)
	assert.Len(t, response.Conditions, 1)
	
	mockService.AssertExpectations(t)
}

func TestGetExpenses_Success(t *testing.T) {
	mockService := new(MockHealthService)
	handler := NewHealthHandler(mockService)
	router := setupHealthTestRouter(handler)
	
	expenses := []domain.MedicalExpense{
		{
			ID:          "expense1",
			UserID:      "user123",
			Amount:      100.0,
			Category:    "doctor_visit",
			Description: "Checkup",
		},
	}
	
	mockService.On("GetExpenses", mock.Anything, "user123").Return(expenses, nil)
	
	req := httptest.NewRequest("GET", "/health/expenses", nil)
	req.Header.Set("Authorization", "Bearer "+createTestJWTToken("user123"))
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response dtos.MedicalExpenseListResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 1, response.Total)
	
	mockService.AssertExpectations(t)
}

func TestGetActivePolicies_Success(t *testing.T) {
	mockService := new(MockHealthService)
	handler := NewHealthHandler(mockService)
	router := setupHealthTestRouter(handler)
	
	policies := []domain.InsurancePolicy{
		{
			ID:           "policy1",
			UserID:       "user123",
			PolicyNumber: "POL123",
			Provider:     "TestInsurance",
			Type:         "health",
			IsActive:     true,
		},
	}
	
	mockService.On("GetActivePolicies", mock.Anything, "user123").Return(policies, nil)
	
	req := httptest.NewRequest("GET", "/health/policies", nil)
	req.Header.Set("Authorization", "Bearer "+createTestJWTToken("user123"))
	
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response dtos.InsurancePolicyListResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, 1, response.Total)
	
	mockService.AssertExpectations(t)
}