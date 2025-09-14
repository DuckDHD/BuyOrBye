//go:build integration
// +build integration

package testutils

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/DuckDHD/BuyOrBye/internal/database"
	"github.com/DuckDHD/BuyOrBye/internal/domain"
	"github.com/DuckDHD/BuyOrBye/internal/handlers"
	"github.com/DuckDHD/BuyOrBye/internal/logging"
	"github.com/DuckDHD/BuyOrBye/internal/middleware"
	"github.com/DuckDHD/BuyOrBye/internal/repositories"
	"github.com/DuckDHD/BuyOrBye/internal/services"
)

// TestServer represents an integration test server instance
type TestServer struct {
	Server         *httptest.Server
	Router         *gin.Engine
	BaseURL        string
	GormService    *database.GormService
	AuthService    handlers.AuthService
	FinanceService services.FinanceService
	HealthService  services.HealthService
	DecisionService handlers.DecisionServiceInterface
}

// NewTestServer creates a new test server instance for integration tests
func NewTestServer(t *testing.T) *TestServer {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)
	
	// Initialize test database
	gormService, err := database.NewGormService()
	require.NoError(t, err, "Failed to initialize test database")
	
	db := gormService.GetDB()
	
	// Initialize core services
	passwordService := services.NewPasswordService()
	jwtService, err := services.NewJWTService()
	require.NoError(t, err, "Failed to initialize JWT service")
	
	// Initialize repositories
	userRepo := repositories.NewUserRepository(db)
	tokenRepo := repositories.NewTokenRepository(db)
	
	// Initialize finance repositories
	incomeRepo := repositories.NewIncomeRepository(db)
	expenseRepo := repositories.NewExpenseRepository(db)
	loanRepo := repositories.NewLoanRepository(db)
	financeSummaryRepo := repositories.NewFinanceSummaryRepository()
	
	// Initialize health repositories
	healthProfileRepo := repositories.NewHealthProfileRepository(db)
	insurancePolicyRepo := repositories.NewInsurancePolicyRepository(db)
	medicalExpenseRepo := repositories.NewMedicalExpenseRepository(db)
	
	// Initialize decision repositories
	decisionRepo := repositories.NewDecisionRepository(db)
	
	// Create repositories aggregates
	financeRepos := services.NewFinanceRepositories(incomeRepo, expenseRepo, loanRepo, financeSummaryRepo)
	healthRepos := services.NewHealthRepositories(healthProfileRepo, insurancePolicyRepo, medicalExpenseRepo)
	
	// Initialize services
	authService := services.NewAuthService(userRepo, tokenRepo, passwordService, jwtService)
	financeService := services.NewFinanceService(financeRepos)
	healthService := services.NewHealthService(healthRepos)
	
	// Initialize decision service components
	mockAIClient := NewMockAIClient() // Use a mock AI client for testing
	promptBuilder := services.NewPromptBuilder()
	decisionInterpreter := services.NewDecisionInterpreter()
	
	// For test purposes, use the legacy decision service
	decisionService := services.NewDecisionService(
		&MockLegacyDecisionRepository{repo: decisionRepo},
		financeService,
		healthService,
		mockAIClient,
		promptBuilder,
		decisionInterpreter,
	)
	
	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)
	financeHandler := handlers.NewFinanceHandler(financeService)
	healthHandler := handlers.NewHealthHandler(healthService)
	decisionHandler := handlers.NewDecisionHandler(decisionService)
	
	// Initialize middlewares
	jwtAuthMiddleware := middleware.NewJWTAuthMiddleware(jwtService)
	
	// Setup Gin router
	router := gin.New()
	
	// Global middleware
	router.Use(middleware.CORS())
	router.Use(middleware.Recovery())
	router.Use(middleware.ValidateRequestLimits())
	
	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"message": "BuyOrBye Test API is running",
		})
	})
	
	// API routes
	api := router.Group("/api/v1")
	
	// Auth routes (public)
	auth := api.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh", authHandler.RefreshToken)
		
		// Protected auth routes
		protected := auth.Group("")
		protected.Use(jwtAuthMiddleware.RequireAuth())
		{
			protected.POST("/logout", authHandler.Logout)
		}
	}
	
	// Finance routes (all require auth)
	finance := api.Group("/finance")
	finance.Use(jwtAuthMiddleware.RequireAuth())
	finance.Use(middleware.ValidateOwnership())
	{
		// Income endpoints
		finance.POST("/income", 
			middleware.ValidateFinancialData(),
			middleware.NormalizeFrequency(),
			financeHandler.AddIncome)
		finance.GET("/income", financeHandler.GetIncomes)
		finance.PUT("/income/:id", 
			middleware.ValidateUserOwnership("income"),
			middleware.ValidateFinancialData(),
			middleware.NormalizeFrequency(),
			financeHandler.UpdateIncome)
		finance.DELETE("/income/:id", 
			middleware.ValidateUserOwnership("income"),
			financeHandler.DeleteIncome)
		
		// Expense endpoints
		finance.POST("/expense", 
			middleware.ValidateFinancialData(),
			middleware.NormalizeFrequency(),
			financeHandler.AddExpense)
		finance.GET("/expenses", financeHandler.GetExpenses)
		finance.PUT("/expense/:id", 
			middleware.ValidateUserOwnership("expense"),
			middleware.ValidateFinancialData(),
			middleware.NormalizeFrequency(),
			financeHandler.UpdateExpense)
		finance.DELETE("/expense/:id", 
			middleware.ValidateUserOwnership("expense"),
			financeHandler.DeleteExpense)
		
		// Loan endpoints
		finance.POST("/loan", 
			middleware.ValidateFinancialData(),
			financeHandler.AddLoan)
		finance.GET("/loans", financeHandler.GetLoans)
		finance.PUT("/loan/:id", 
			middleware.ValidateUserOwnership("loan"),
			middleware.ValidateFinancialData(),
			financeHandler.UpdateLoan)
		
		// Analysis endpoints
		finance.GET("/summary", financeHandler.GetFinanceSummary)
		finance.GET("/affordability", financeHandler.GetAffordability)
	}
	
	// Health routes (all require auth)
	health := api.Group("/health")
	health.Use(jwtAuthMiddleware.RequireAuth())
	health.Use(middleware.ValidateOwnership())
	{
		// Health profile endpoints
		health.POST("/profile", healthHandler.AddHealthProfile)
		health.GET("/profile", healthHandler.GetHealthProfile)
		health.PUT("/profile/:id", 
			middleware.ValidateUserOwnership("health_profile"),
			healthHandler.UpdateHealthProfile)
		
		// Insurance endpoints
		health.POST("/insurance", healthHandler.AddInsurancePolicy)
		health.GET("/insurance", healthHandler.GetInsurancePolicies)
		health.PUT("/insurance/:id", 
			middleware.ValidateUserOwnership("insurance_policy"),
			healthHandler.UpdateInsurancePolicy)
		
		// Medical expense endpoints
		health.POST("/expense", healthHandler.AddMedicalExpense)
		health.GET("/expenses", healthHandler.GetMedicalExpenses)
		health.PUT("/expense/:id", 
			middleware.ValidateUserOwnership("medical_expense"),
			healthHandler.UpdateMedicalExpense)
		
		// Analysis endpoints
		health.GET("/summary", healthHandler.GetHealthSummary)
		health.GET("/risk-assessment", healthHandler.GetRiskAssessment)
	}
	
	// Decision routes (all require auth)
	decision := api.Group("/decision")
	decision.Use(jwtAuthMiddleware.RequireAuth())
	{
		decision.POST("/evaluate", decisionHandler.MakeDecision)
		decision.GET("/history", decisionHandler.GetDecisionHistory)
		decision.GET("/stats", decisionHandler.GetDecisionStats)
	}
	
	// Create test server
	server := httptest.NewServer(router)
	
	return &TestServer{
		Server:         server,
		Router:         router,
		BaseURL:        server.URL,
		GormService:    gormService,
		AuthService:    authService,
		FinanceService: financeService,
		HealthService:  healthService,
		DecisionService: decisionService,
	}
}

// Close closes the test server and cleans up resources
func (ts *TestServer) Close() {
	if ts.Server != nil {
		ts.Server.Close()
	}
	
	// Clean up database connections
	if ts.GormService != nil {
		db := ts.GormService.GetDB()
		if db != nil {
			sqlDB, err := db.DB()
			if err == nil {
				sqlDB.Close()
			}
		}
	}
}

// GetPort returns the port number the test server is running on
func (ts *TestServer) GetPort() string {
	_, port, _ := net.SplitHostPort(ts.Server.Listener.Addr().String())
	return port
}

// SetupIntegrationTest sets up environment for integration tests
func SetupIntegrationTest() {
	// Set test environment variables for MySQL (use in-memory or test database)
	os.Setenv("BLUEPRINT_DB_HOST", "localhost")
	os.Setenv("BLUEPRINT_DB_PORT", "3306")
	os.Setenv("BLUEPRINT_DB_DATABASE", "buyorbye_test")
	os.Setenv("BLUEPRINT_DB_USERNAME", "test")
	os.Setenv("BLUEPRINT_DB_PASSWORD", "test")
	os.Setenv("JWT_SECRET", "test-jwt-secret-key-for-integration-tests")
	os.Setenv("JWT_EXPIRY", "15m")
	os.Setenv("JWT_REFRESH_EXPIRY", "7d")
	os.Setenv("GIN_MODE", "test")
	os.Setenv("APP_ENV", "test")
}

// TeardownIntegrationTest cleans up after integration tests
func TeardownIntegrationTest() {
	// Clean up environment variables if needed
	os.Unsetenv("DATABASE_URL")
	os.Unsetenv("JWT_SECRET")
	os.Unsetenv("JWT_EXPIRY")
	os.Unsetenv("JWT_REFRESH_EXPIRY")
	os.Unsetenv("GIN_MODE")
}

// WaitForServer waits for the server to be ready
func WaitForServer(baseURL string, maxAttempts int) error {
	client := &http.Client{Timeout: 1 * time.Second}
	
	for i := 0; i < maxAttempts; i++ {
		resp, err := client.Get(baseURL + "/health")
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}
		time.Sleep(100 * time.Millisecond)
	}
	
	return fmt.Errorf("server not ready after %d attempts", maxAttempts)
}

// ResetDatabase truncates all tables to ensure clean state for tests
func (ts *TestServer) ResetDatabase(t *testing.T) {
	db := ts.GormService.GetDB()
	
	// List of tables to clear (in dependency order)
	tables := []string{
		"refresh_tokens",
		"decision_records",
		"medical_expenses",
		"insurance_policies",
		"health_profiles",
		"finance_summaries", 
		"loans",
		"expenses", 
		"incomes",
		"users",
	}
	
	// Disable foreign key checks for SQLite
	err := db.Exec("PRAGMA foreign_keys = OFF").Error
	require.NoError(t, err, "Failed to disable foreign key checks")
	
	// Clear all tables
	for _, table := range tables {
		err := db.Exec(fmt.Sprintf("DELETE FROM %s", table)).Error
		if err != nil {
			// Log warning but don't fail test if table doesn't exist
			logger := logging.GetLogger()
			logger.Warn("Failed to clear table during test cleanup",
				logging.WithTable(table),
				logging.WithError(err))
		}
	}
	
	// Re-enable foreign key checks
	err = db.Exec("PRAGMA foreign_keys = ON").Error
	require.NoError(t, err, "Failed to re-enable foreign key checks")
}

// Mock implementations for testing

// MockAIClient provides controlled AI responses for testing
type MockAIClient struct {
	responses      []domain.AIResponse
	currentIndex   int
	shouldTimeout  bool
	shouldError    bool
	errorMessage   string
}

// NewMockAIClient creates a new mock AI client
func NewMockAIClient() *MockAIClient {
	return &MockAIClient{
		responses: make([]domain.AIResponse, 0),
	}
}

// GenerateDecision implements the AIClient interface
func (m *MockAIClient) GenerateDecision(ctx context.Context, prompt domain.AIPrompt) (*domain.AIResponse, error) {
	if m.shouldTimeout {
		return nil, fmt.Errorf("context deadline exceeded")
	}

	if m.shouldError {
		return nil, fmt.Errorf("openai error: %s", m.errorMessage)
	}

	// Return default response if no specific response set
	if len(m.responses) == 0 {
		return &domain.AIResponse{
			Decision:   "BUY",
			Confidence: 0.75,
			Reasoning:  "Default mock AI decision",
		}, nil
	}

	if m.currentIndex >= len(m.responses) {
		m.currentIndex = 0 // Reset to first response
	}

	response := m.responses[m.currentIndex]
	m.currentIndex++

	return &response, nil
}

// MockLegacyDecisionRepository adapts the new DecisionRepository to the legacy interface
type MockLegacyDecisionRepository struct {
	repo services.DecisionRepository
}

// Save implements the legacy interface
func (m *MockLegacyDecisionRepository) Save(ctx context.Context, decision domain.DecisionOutcome) error {
	// Create a dummy intent since the legacy interface expects both
	intent := domain.PurchaseIntent{
		ID:       decision.IntentID,
		UserID:   decision.UserID,
		ItemName: "Mock Item",
		ItemCost: 100.0,
		Category: "other",
		Urgency:  "medium",
		Frequency: "one_time",
	}
	return m.repo.SaveDecision(ctx, decision, intent)
}

// GetByIntentID implements the legacy interface
func (m *MockLegacyDecisionRepository) GetByIntentID(ctx context.Context, intentID string) (*domain.DecisionOutcome, error) {
	// For testing, we can't easily get by intent ID with the new repository
	// So we'll return a mock decision for now
	return &domain.DecisionOutcome{
		ID:       "mock-decision-1",
		UserID:   "mock-user",
		IntentID: intentID,
		Decision: "BUY",
		Confidence: 0.8,
		PrimaryReason: "Mock decision for testing",
	}, nil
}

// GetRecentDecisions implements the legacy interface
func (m *MockLegacyDecisionRepository) GetRecentDecisions(ctx context.Context, userID string, days int) ([]domain.PastDecision, error) {
	// Return empty slice for testing - real implementation would convert from DecisionOutcome
	return []domain.PastDecision{}, nil
}