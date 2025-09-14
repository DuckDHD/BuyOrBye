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
	"github.com/DuckDHD/BuyOrBye/internal/types"
)

// MockDecisionService for testing
type MockDecisionService struct {
	mock.Mock
}

func (m *MockDecisionService) MakeDecision(ctx context.Context, intent domain.PurchaseIntent) (*domain.DecisionOutcome, error) {
	args := m.Called(ctx, intent)
	return args.Get(0).(*domain.DecisionOutcome), args.Error(1)
}

func (m *MockDecisionService) GetDecisionHistory(ctx context.Context, userID string, days int) ([]domain.PastDecision, error) {
	args := m.Called(ctx, userID, days)
	return args.Get(0).([]domain.PastDecision), args.Error(1)
}

func (m *MockDecisionService) GetDecisionByIntent(ctx context.Context, intentID string) (*domain.DecisionOutcome, error) {
	args := m.Called(ctx, intentID)
	return args.Get(0).(*domain.DecisionOutcome), args.Error(1)
}

// Helper function to create JWT token for testing
func createDecisionTestJWTToken(userID string) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(time.Hour).Unix(),
	})
	
	tokenString, _ := token.SignedString([]byte("test-secret"))
	return tokenString
}

// Helper function to setup test router with auth middleware
func setupDecisionTestRouter(handler *DecisionHandler) *gin.Engine {
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
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization_required", "message": "Authorization header is required"})
			c.Abort()
			return
		}
		c.Next()
	})
	
	// Register decision routes
	decision := router.Group("/decision")
	{
		decision.POST("/evaluate", handler.MakeDecision)
		decision.GET("/history", handler.GetDecisionHistory)
		decision.GET("/stats", handler.GetDecisionStats)
	}
	
	return router
}

func TestMakeDecision_Success(t *testing.T) {
	// Arrange
	mockService := new(MockDecisionService)
	handler := NewDecisionHandler(mockService)
	router := setupDecisionTestRouter(handler)
	
	intentDTO := types.PurchaseIntentDTO{
		ItemName:    "New Laptop",
		ItemCost:    1200.0,
		Category:    "electronics",
		Urgency:     "medium",
		Frequency:   "one_time",
		Purpose:     "Work productivity",
		Alternative: "Refurbished laptop",
	}
	
	expectedDecision := &domain.DecisionOutcome{
		ID:            "decision123",
		UserID:        "user123",
		IntentID:      "intent123",
		Decision:      "BUY",
		Confidence:    0.8,
		PrimaryReason: "Good value for productivity investment",
		Factors: []domain.DecisionFactor{
			{
				Category:    "financial",
				Impact:      "positive",
				Weight:      0.7,
				Description: "Purchase fits within budget",
			},
		},
		Recommendations: []string{"Shop around for best price", "Consider warranty options"},
		WaitPeriod:      0,
		MaxBudget:       1260.0,
		CreatedAt:       time.Now(),
		ProcessingTime:  150,
	}
	
	mockService.On("MakeDecision", mock.Anything, mock.MatchedBy(func(intent domain.PurchaseIntent) bool {
		return intent.ItemName == "New Laptop" && intent.ItemCost == 1200.0 && intent.Category == "electronics"
	})).Return(expectedDecision, nil)
	
	reqBody, _ := json.Marshal(intentDTO)
	req := httptest.NewRequest("POST", "/decision/evaluate", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+createDecisionTestJWTToken("user123"))
	
	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	
	var responseDTO types.DecisionResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &responseDTO)
	assert.NoError(t, err)
	assert.Equal(t, "BUY", responseDTO.Decision)
	assert.Equal(t, 0.8, responseDTO.Confidence)
	assert.Equal(t, "user123", responseDTO.UserID)
	assert.Len(t, responseDTO.Factors, 1)
	assert.Len(t, responseDTO.Recommendations, 2)
	
	mockService.AssertExpectations(t)
}

func TestMakeDecision_ValidationError(t *testing.T) {
	// Arrange
	mockService := new(MockDecisionService)
	handler := NewDecisionHandler(mockService)
	router := setupDecisionTestRouter(handler)
	
	// Invalid category in request body
	intentDTO := types.PurchaseIntentDTO{
		ItemName:    "Test Item",
		ItemCost:    100.0,
		Category:    "invalid_category", // Invalid category
		Urgency:     "medium",
		Frequency:   "one_time",
	}
	
	reqBody, _ := json.Marshal(intentDTO)
	req := httptest.NewRequest("POST", "/decision/evaluate", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+createDecisionTestJWTToken("user123"))
	
	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
	
	var errorResponse map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)
	assert.Contains(t, errorResponse, "error")
	assert.Equal(t, "validation_error", errorResponse["error"])
	assert.Contains(t, errorResponse, "message")
	assert.Contains(t, errorResponse, "timestamp")
	
	// Service should not be called due to validation failure
	mockService.AssertNotCalled(t, "MakeDecision")
}

func TestMakeDecision_RequiresAuth(t *testing.T) {
	// Arrange
	mockService := new(MockDecisionService)
	handler := NewDecisionHandler(mockService)
	router := setupDecisionTestRouter(handler)
	
	intentDTO := types.PurchaseIntentDTO{
		ItemName:    "Test Item",
		ItemCost:    100.0,
		Category:    "electronics",
		Urgency:     "medium",
		Frequency:   "one_time",
	}
	
	reqBody, _ := json.Marshal(intentDTO)
	req := httptest.NewRequest("POST", "/decision/evaluate", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	// No Authorization header - should trigger 401
	
	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	
	var errorResponse map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)
	assert.Equal(t, "authorization_required", errorResponse["error"])
	
	// Service should not be called due to auth failure
	mockService.AssertNotCalled(t, "MakeDecision")
}

func TestMakeDecision_ServiceFailure(t *testing.T) {
	// Arrange
	mockService := new(MockDecisionService)
	handler := NewDecisionHandler(mockService)
	router := setupDecisionTestRouter(handler)
	
	intentDTO := types.PurchaseIntentDTO{
		ItemName:    "Test Item",
		ItemCost:    100.0,
		Category:    "electronics",
		Urgency:     "medium",
		Frequency:   "one_time",
	}
	
	mockService.On("MakeDecision", mock.Anything, mock.AnythingOfType("domain.PurchaseIntent")).
		Return((*domain.DecisionOutcome)(nil), fmt.Errorf("service failure"))
	
	reqBody, _ := json.Marshal(intentDTO)
	req := httptest.NewRequest("POST", "/decision/evaluate", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+createDecisionTestJWTToken("user123"))
	
	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(t, http.StatusInternalServerError, w.Code)
	
	var errorResponse map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)
	assert.Equal(t, "service_failure", errorResponse["error"])
	assert.Contains(t, errorResponse["message"], "Unable to process decision request")
	
	mockService.AssertExpectations(t)
}

func TestMakeDecision_MissingUserID(t *testing.T) {
	// Arrange
	mockService := new(MockDecisionService)
	handler := NewDecisionHandler(mockService)
	router := gin.New()
	
	// Setup router without proper auth middleware (no userID in context)
	router.POST("/decision/evaluate", handler.MakeDecision)
	
	intentDTO := types.PurchaseIntentDTO{
		ItemName:    "Test Item",
		ItemCost:    100.0,
		Category:    "electronics",
		Urgency:     "medium",
		Frequency:   "one_time",
	}
	
	reqBody, _ := json.Marshal(intentDTO)
	req := httptest.NewRequest("POST", "/decision/evaluate", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	
	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	
	var errorResponse map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)
	assert.Equal(t, "missing_user_context", errorResponse["error"])
	
	mockService.AssertNotCalled(t, "MakeDecision")
}

func TestGetHistory_Success(t *testing.T) {
	// Arrange
	mockService := new(MockDecisionService)
	handler := NewDecisionHandler(mockService)
	router := setupDecisionTestRouter(handler)
	
	pastDecisions := []domain.PastDecision{
		{
			ItemName: "Smartphone",
			ItemCost: 800.0,
			Decision: "BUY",
			DaysAgo:  5,
			Category: "electronics",
		},
		{
			ItemName: "Gaming Console",
			ItemCost: 500.0,
			Decision: "WAIT",
			DaysAgo:  10,
			Category: "entertainment",
		},
	}
	
	mockService.On("GetDecisionHistory", mock.Anything, "user123", 30).Return(pastDecisions, nil)
	
	req := httptest.NewRequest("GET", "/decision/history?limit=10&offset=0", nil)
	req.Header.Set("Authorization", "Bearer "+createDecisionTestJWTToken("user123"))
	
	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	
	var responseDTO types.DecisionHistoryDTO
	err := json.Unmarshal(w.Body.Bytes(), &responseDTO)
	assert.NoError(t, err)
	assert.Equal(t, "user123", responseDTO.UserID)
	assert.Equal(t, 2, responseDTO.TotalDecisions)
	assert.Len(t, responseDTO.RecentDecisions, 2)
	assert.Equal(t, "Smartphone", responseDTO.RecentDecisions[0].ItemName)
	assert.Equal(t, "BUY", responseDTO.RecentDecisions[0].Decision)
	
	mockService.AssertExpectations(t)
}

func TestGetHistory_RequiresAuth(t *testing.T) {
	// Arrange
	mockService := new(MockDecisionService)
	handler := NewDecisionHandler(mockService)
	router := setupDecisionTestRouter(handler)
	
	req := httptest.NewRequest("GET", "/decision/history", nil)
	// No Authorization header
	
	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	
	mockService.AssertNotCalled(t, "GetDecisionHistory")
}

func TestGetHistory_Pagination(t *testing.T) {
	// Arrange
	mockService := new(MockDecisionService)
	handler := NewDecisionHandler(mockService)
	router := setupDecisionTestRouter(handler)
	
	// Mock many decisions to test pagination
	pastDecisions := make([]domain.PastDecision, 25) // More than default limit
	for i := 0; i < 25; i++ {
		pastDecisions[i] = domain.PastDecision{
			ItemName: fmt.Sprintf("Item %d", i+1),
			ItemCost: float64(100 * (i + 1)),
			Decision: "BUY",
			DaysAgo:  i + 1,
			Category: "electronics",
		}
	}
	
	mockService.On("GetDecisionHistory", mock.Anything, "user123", 30).Return(pastDecisions, nil)
	
	req := httptest.NewRequest("GET", "/decision/history?limit=20&offset=5", nil)
	req.Header.Set("Authorization", "Bearer "+createDecisionTestJWTToken("user123"))
	
	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	
	var responseDTO types.DecisionHistoryDTO
	err := json.Unmarshal(w.Body.Bytes(), &responseDTO)
	assert.NoError(t, err)
	assert.Equal(t, 25, responseDTO.TotalDecisions)
	assert.Len(t, responseDTO.RecentDecisions, 20) // Should be limited to 20 (limit - offset)
	
	mockService.AssertExpectations(t)
}

func TestGetStats_Success(t *testing.T) {
	// Arrange
	mockService := new(MockDecisionService)
	handler := NewDecisionHandler(mockService)
	router := setupDecisionTestRouter(handler)
	
	pastDecisions := []domain.PastDecision{
		{ItemName: "Item1", ItemCost: 100.0, Decision: "BUY", DaysAgo: 5, Category: "electronics"},
		{ItemName: "Item2", ItemCost: 200.0, Decision: "BUY", DaysAgo: 10, Category: "clothing"},
		{ItemName: "Item3", ItemCost: 300.0, Decision: "WAIT", DaysAgo: 15, Category: "electronics"},
		{ItemName: "Item4", ItemCost: 150.0, Decision: "BYE", DaysAgo: 20, Category: "entertainment"},
	}
	
	mockService.On("GetDecisionHistory", mock.Anything, "user123", 30).Return(pastDecisions, nil)
	
	req := httptest.NewRequest("GET", "/decision/stats?days=30", nil)
	req.Header.Set("Authorization", "Bearer "+createDecisionTestJWTToken("user123"))
	
	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	
	var stats map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &stats)
	assert.NoError(t, err)
	
	assert.Equal(t, "user123", stats["user_id"])
	assert.Equal(t, float64(4), stats["total_decisions"])
	
	// Check decision counts
	decisionPattern := stats["decision_pattern"].(map[string]interface{})
	assert.Equal(t, float64(2), decisionPattern["BUY"])
	assert.Equal(t, float64(1), decisionPattern["WAIT"])
	assert.Equal(t, float64(1), decisionPattern["BYE"])
	
	// Check spending (only BUY decisions count towards spending)
	assert.Equal(t, float64(300.0), stats["total_spending"]) // 100 + 200
	
	mockService.AssertExpectations(t)
}

func TestGetStats_DateRangeFilter(t *testing.T) {
	// Arrange
	mockService := new(MockDecisionService)
	handler := NewDecisionHandler(mockService)
	router := setupDecisionTestRouter(handler)
	
	pastDecisions := []domain.PastDecision{
		{ItemName: "Recent", ItemCost: 100.0, Decision: "BUY", DaysAgo: 5, Category: "electronics"},
		{ItemName: "Old", ItemCost: 200.0, Decision: "BUY", DaysAgo: 45, Category: "clothing"}, // Outside 30 day range
	}
	
	// Service should be called with days=7 filter
	mockService.On("GetDecisionHistory", mock.Anything, "user123", 7).Return(pastDecisions[:1], nil) // Only recent decision
	
	req := httptest.NewRequest("GET", "/decision/stats?days=7", nil)
	req.Header.Set("Authorization", "Bearer "+createDecisionTestJWTToken("user123"))
	
	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	
	var stats map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &stats)
	assert.NoError(t, err)
	
	assert.Equal(t, float64(1), stats["total_decisions"]) // Only 1 decision in 7 days
	assert.Equal(t, float64(100.0), stats["total_spending"]) // Only recent purchase
	
	mockService.AssertExpectations(t)
}

func TestGetStats_RequiresAuth(t *testing.T) {
	// Arrange
	mockService := new(MockDecisionService)
	handler := NewDecisionHandler(mockService)
	router := setupDecisionTestRouter(handler)
	
	req := httptest.NewRequest("GET", "/decision/stats", nil)
	// No Authorization header
	
	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	
	mockService.AssertNotCalled(t, "GetDecisionHistory")
}

func TestMakeDecision_InvalidJSON(t *testing.T) {
	// Arrange
	mockService := new(MockDecisionService)
	handler := NewDecisionHandler(mockService)
	router := setupDecisionTestRouter(handler)
	
	// Invalid JSON
	req := httptest.NewRequest("POST", "/decision/evaluate", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+createDecisionTestJWTToken("user123"))
	
	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
	
	var errorResponse map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)
	assert.Equal(t, "invalid_request", errorResponse["error"])
	
	mockService.AssertNotCalled(t, "MakeDecision")
}

func TestMakeDecision_MissingRequiredFields(t *testing.T) {
	// Arrange
	mockService := new(MockDecisionService)
	handler := NewDecisionHandler(mockService)
	router := setupDecisionTestRouter(handler)
	
	// Missing required fields
	intentDTO := types.PurchaseIntentDTO{
		ItemName: "", // Empty required field
		ItemCost: 0,  // Invalid cost
		// Missing other required fields
	}
	
	reqBody, _ := json.Marshal(intentDTO)
	req := httptest.NewRequest("POST", "/decision/evaluate", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+createDecisionTestJWTToken("user123"))
	
	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
	
	var errorResponse map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
	assert.NoError(t, err)
	assert.Equal(t, "validation_error", errorResponse["error"])
	assert.Contains(t, errorResponse, "details") // Should have validation details
	
	mockService.AssertNotCalled(t, "MakeDecision")
}

func TestGetStats_DefaultDaysFilter(t *testing.T) {
	// Arrange
	mockService := new(MockDecisionService)
	handler := NewDecisionHandler(mockService)
	router := setupDecisionTestRouter(handler)
	
	pastDecisions := []domain.PastDecision{
		{ItemName: "Item1", ItemCost: 100.0, Decision: "BUY", DaysAgo: 5, Category: "electronics"},
	}
	
	// Should default to 30 days if no days parameter provided
	mockService.On("GetDecisionHistory", mock.Anything, "user123", 30).Return(pastDecisions, nil)
	
	req := httptest.NewRequest("GET", "/decision/stats", nil) // No days parameter
	req.Header.Set("Authorization", "Bearer "+createDecisionTestJWTToken("user123"))
	
	// Act
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	
	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	mockService.AssertExpectations(t)
}