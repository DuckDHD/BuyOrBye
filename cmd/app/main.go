// cmd/app/main.go
package main

import (
	"context"
	"fmt"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/DuckDHD/BuyOrBye/internal/config"
	"github.com/DuckDHD/BuyOrBye/internal/database"
	"github.com/DuckDHD/BuyOrBye/internal/domain"
	"github.com/DuckDHD/BuyOrBye/internal/handlers"
	"github.com/DuckDHD/BuyOrBye/internal/logging"
	"github.com/DuckDHD/BuyOrBye/internal/middleware"
	"github.com/DuckDHD/BuyOrBye/internal/repositories"
	"github.com/DuckDHD/BuyOrBye/internal/services"
	"github.com/DuckDHD/BuyOrBye/internal/types"
)

func main() {
	// Load configuration first
	cfg, err := config.LoadConfig()
	if err != nil {
		panic("Failed to load configuration: " + err.Error())
	}

	// Initialize logger with config
	if err := logging.InitLogger(logging.LogConfig{
		Environment: cfg.App.Environment,
		Level:       "info", // Default to info level
	}); err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
	logger := logging.GetLogger()
	logger.Info("Configuration loaded successfully",
		logging.WithComponent("main"),
		zap.String("environment", cfg.App.Environment),
		zap.String("config_source", "environment-variables"))

	// Initialize database service (restore original working setup)
	dbService, err := database.NewGormService()
	if err != nil {
		logger.Fatal("Failed to initialize database", logging.WithError(err))
	}
	db := dbService.GetDB()

	// Initialize core services with config
	passwordService := services.NewPasswordService()
	jwtService, err := services.NewJWTServiceFromConfig(&cfg.Auth)
	if err != nil {
		logger.Fatal("Failed to initialize JWT service", logging.WithError(err))
	}

	// Initialize repositories
	userRepo := repositories.NewUserRepository(db)
	tokenRepo := repositories.NewTokenRepository(db)

	// Finance repositories
	incomeRepo := repositories.NewIncomeRepository(db)
	expenseRepo := repositories.NewExpenseRepository(db)
	loanRepo := repositories.NewLoanRepository(db)
	financeSummaryRepo := repositories.NewFinanceSummaryRepository()
	financeRepos := services.NewFinanceRepositories(incomeRepo, expenseRepo, loanRepo, financeSummaryRepo)

	// Health repositories
	healthProfileRepo := repositories.NewHealthProfileRepository(db)
	conditionRepo := repositories.NewMedicalConditionRepository(db)
	medicalExpenseRepo := repositories.NewMedicalExpenseRepository(db)
	policyRepo := repositories.NewInsurancePolicyRepository(db)

	// Health analysis services
	riskCalculator := services.NewRiskCalculator()
	costAnalyzer := services.NewMedicalCostAnalyzer()

	// Domain services
	authService := services.NewAuthService(userRepo, tokenRepo, passwordService, jwtService)
	financeService := services.NewFinanceService(financeRepos)
	_ = services.NewBudgetAnalyzer(financeService) // kept for future use

	healthService := services.NewHealthService(
		healthProfileRepo,
		conditionRepo,
		medicalExpenseRepo,
		policyRepo,
		riskCalculator,
		costAnalyzer,
	)

	// Mock decision service (placeholder)
	decisionServiceInterface := &mockDecisionServiceDirect{}
	var financeServiceInterface handlers.FinanceServiceInterface = financeService

	// Handlers
	authHandler := handlers.NewAuthHandler(authService)
	financeHandler := handlers.NewFinanceHandler(financeServiceInterface)
	healthHandler := handlers.NewHealthHandler(healthService)
	decisionHandler := handlers.NewDecisionHandler(decisionServiceInterface)

	// UI handlers
	uiHandlers := handlers.NewUIHandlers(authService, financeServiceInterface, healthService, decisionServiceInterface)

	// Structured chat handler
	structuredChatHandler := handlers.NewStructuredChatHandler()

	// Middlewares
	jwtAuthMiddleware := middleware.NewJWTAuthMiddleware(jwtService)

	// Gin router
	router := gin.Default()

	// Global middleware
	router.Use(middleware.CORS())
	middlewareConfig := config.GetMiddlewareConfig(cfg.App.Environment)
	loggingConfig := logging.HTTPLoggingConfig{
		SkipPaths:       middlewareConfig.SkipPaths,
		LogRequestBody:  middlewareConfig.LogRequestBody,
		LogResponseBody: middlewareConfig.LogResponseBody,
		MaxBodySize:     middlewareConfig.MaxBodySize,
	}
	router.Use(logging.HTTPLoggingMiddleware(loggingConfig))
	router.Use(logging.ErrorLoggingMiddleware())
	router.Use(logging.RequestIDMiddleware())
	router.Use(middleware.CSRFMiddleware())
	router.Use(middleware.ValidateRequestLimits())

	// Health check
	router.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "BuyOrBye API is running",
		})
	})

	// UI Routes (Frontend/Web Interface)
	setupUIRoutes(router, uiHandlers, jwtAuthMiddleware, structuredChatHandler)

	// API routes
	api := router.Group("/api/v1")

	// ===== Auth (public + protected) =====
	auth := api.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh", authHandler.RefreshToken)

		protected := auth.Group("")
		protected.Use(jwtAuthMiddleware.RequireAuth())
		{
			protected.POST("/logout", authHandler.Logout)
		}
	}

	// ===== Finance (protected) =====
	finance := api.Group("/finance")
	finance.Use(jwtAuthMiddleware.RequireAuth())
	finance.Use(middleware.ValidateOwnership())
	{
		// Income
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

		// Expense
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

		// Loan
		finance.POST("/loan",
			middleware.ValidateFinancialData(),
			financeHandler.AddLoan)
		finance.GET("/loans", financeHandler.GetLoans)
		finance.PUT("/loan/:id",
			middleware.ValidateUserOwnership("loan"),
			middleware.ValidateFinancialData(),
			financeHandler.UpdateLoan)

		// Analysis
		finance.GET("/summary", financeHandler.GetFinanceSummary)
		finance.GET("/affordability", financeHandler.GetAffordability)
	}

	// ===== Health (protected) =====
	health := api.Group("/health")
	health.Use(jwtAuthMiddleware.RequireAuth())
	health.Use(middleware.SanitizeSensitiveData())
	{
		// Profiles
		health.POST("/profiles",
			middleware.ValidateHealthProfileData(),
			healthHandler.CreateProfile)
		health.GET("/profiles/:id",
			middleware.ValidateHealthOwnership(),
			healthHandler.GetProfile)
		health.PUT("/profiles/:id",
			middleware.ValidateHealthOwnership(),
			middleware.ValidateHealthProfileData(),
			healthHandler.UpdateProfile)
		health.DELETE("/profiles/:id",
			middleware.ValidateHealthOwnership(),
			healthHandler.DeleteProfile)
		health.GET("/profiles/:id/summary",
			middleware.ValidateHealthOwnership(),
			healthHandler.GetHealthSummary)
		health.GET("/profiles/:id/risk",
			middleware.ValidateHealthOwnership(),
			healthHandler.CalculateRisk)

		// Conditions
		health.POST("/conditions",
			middleware.ValidateHealthOwnership(),
			healthHandler.CreateCondition)
		health.GET("/conditions/:id",
			middleware.ValidateHealthOwnership(),
			healthHandler.GetCondition)
		health.PUT("/conditions/:id",
			middleware.ValidateHealthOwnership(),
			healthHandler.UpdateCondition)
		health.DELETE("/conditions/:id",
			middleware.ValidateHealthOwnership(),
			healthHandler.RemoveCondition)
		health.GET("/profiles/:id/conditions",
			middleware.ValidateHealthOwnership(),
			healthHandler.GetConditionsByProfile)

		// Policies
		health.POST("/policies",
			middleware.ValidateInsuranceDates(),
			middleware.ValidateHealthOwnership(),
			healthHandler.CreatePolicy)
		health.GET("/policies/:id",
			middleware.ValidateHealthOwnership(),
			healthHandler.GetPolicy)
		health.PUT("/policies/:id",
			middleware.ValidateHealthOwnership(),
			middleware.ValidateInsuranceDates(),
			healthHandler.UpdatePolicy)
		health.DELETE("/policies/:id",
			middleware.ValidateHealthOwnership(),
			healthHandler.DeletePolicy)
		health.GET("/profiles/:id/policies",
			middleware.ValidateHealthOwnership(),
			healthHandler.GetPoliciesByProfile)

		// Expenses
		health.POST("/expenses",
			middleware.ValidateExpenseData(),
			middleware.ValidateHealthOwnership(),
			healthHandler.CreateExpense)
		health.GET("/expenses/:id",
			middleware.ValidateHealthOwnership(),
			healthHandler.GetExpense)
		health.PUT("/expenses/:id",
			middleware.ValidateHealthOwnership(),
			middleware.ValidateExpenseData(),
			healthHandler.UpdateExpense)
		health.DELETE("/expenses/:id",
			middleware.ValidateHealthOwnership(),
			healthHandler.DeleteExpense)
		health.GET("/profiles/:id/expenses",
			middleware.ValidateHealthOwnership(),
			healthHandler.GetExpensesByProfile)
	}

	// ===== Decision (protected - available) =====
	if decisionHandler != nil {
		decision := api.Group("/decision")
		decision.Use(jwtAuthMiddleware.RequireAuth())
		{
			decision.POST("/evaluate",
				middleware.ValidateRequestLimits(),
				decisionHandler.MakeDecision)
			decision.GET("/history", decisionHandler.GetDecisionHistory)
			decision.GET("/stats", decisionHandler.GetDecisionStats)
		}
	}

	// HTTP server
	serverService := config.NewServerService(cfg)
	server := serverService.CreateServer(router)

	logger.Info("Starting BuyOrBye server",
		logging.WithComponent("main"),
		zap.String("address", serverService.GetAddress()),
		zap.String("environment", cfg.App.Environment))

	// Graceful shutdown
	done := make(chan bool, 1)
	go gracefulShutdown(server, done)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatal("HTTP server error", logging.WithError(err))
	}

	<-done
	logger.Info("Graceful shutdown complete")
}

func gracefulShutdown(server *http.Server, done chan bool) {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()

	logger := logging.GetLogger()
	logger.Info("Shutting down gracefully, press Ctrl+C again to force")
	stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", logging.WithError(err))
	}
	logger.Info("Server exiting")

	done <- true
}

// setupUIRoutes organizes frontend UI routes with proper separation
func setupUIRoutes(
	router *gin.Engine,
	uiHandlers *handlers.UIHandlers,
	jwtAuthMiddleware *middleware.JWTAuthMiddleware,
	structuredChatHandler *handlers.StructuredChatHandler,
) {
	// Static assets with caching
	router.Static("/static", "cmd/web/static")
	router.StaticFile("/favicon.ico", "cmd/web/static/favicon.svg")
	router.StaticFile("/manifest.json", "cmd/web/static/manifest.json")
	router.StaticFile("/robots.txt", "cmd/web/static/robots.txt")
	router.StaticFile("/sw.js", "cmd/web/static/sw.js")

	// Public routes
	router.GET("/", uiHandlers.ChatPage)
	router.GET("/auth/login", uiHandlers.LoginPage)
	router.GET("/auth/register", uiHandlers.RegisterPage)

	// Public form actions
	router.POST("/auth/login", uiHandlers.LoginAction)
	router.POST("/auth/register", uiHandlers.RegisterAction)

	// Test POST route for CSRF protection
	router.POST("/test-csrf", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "CSRF protection passed!",
			"data":    c.PostForm("test_data"),
		})
	})

	// Protected routes
	protected := router.Group("")
	protected.Use(jwtAuthMiddleware.RequireAuth())
	{
		// Pages
		protected.GET("/dashboard", uiHandlers.DashboardPage)
		protected.GET("/finance", uiHandlers.FinanceOverviewPage)
		protected.GET("/health", uiHandlers.HealthProfilePage)
		protected.GET("/decisions/new", uiHandlers.DecisionNewPage)
		protected.GET("/decisions/history", uiHandlers.DecisionHistoryPage)

		// Partials
		ui := protected.Group("/ui/partials")
		{
			ui.GET("/dashboard/content", uiHandlers.DashboardContentPartial)

			ui.GET("/finance/overview", uiHandlers.FinanceOverviewPartial)
			ui.GET("/finance/summary", uiHandlers.FinanceSummaryPartial)
			ui.GET("/finance/income/list", uiHandlers.IncomeListPartial)
			ui.GET("/finance/expense/form", uiHandlers.ExpenseFormPartial)

			// NOTE: aligned param names for consistency
			ui.GET("/health/risk-gauge/:id", uiHandlers.HealthRiskGaugePartial)
			ui.GET("/health/condition/add", uiHandlers.ConditionAddPartial)
			ui.GET("/health/insurance/card/:id", uiHandlers.InsuranceCardPartial)

			ui.GET("/decision/result", uiHandlers.DecisionResultPartial)
			ui.GET("/decision/filter", uiHandlers.DecisionFilterPartial)
		}

		// Form actions
		protected.POST("/auth/logout", uiHandlers.LogoutAction)

		protected.POST("/finance/income", uiHandlers.FinanceCreateIncomeAction)
		protected.PUT("/finance/income/:id", uiHandlers.FinanceUpdateIncomeAction)
		protected.DELETE("/finance/income/:id", uiHandlers.FinanceDeleteIncomeAction)

		protected.POST("/finance/expense", uiHandlers.FinanceCreateExpenseAction)
		protected.PUT("/finance/expense/:id", uiHandlers.FinanceUpdateExpenseAction)
		protected.DELETE("/finance/expense/:id", uiHandlers.FinanceDeleteExpenseAction)

		protected.POST("/health/profile", uiHandlers.HealthCreateProfileAction)
		protected.PUT("/health/profile/:id", uiHandlers.HealthUpdateProfileAction)
		protected.DELETE("/health/profile/:id", uiHandlers.HealthDeleteProfileAction)

		protected.POST("/decisions", uiHandlers.DecisionCreateAction)
	}

	// Structured Chat API Routes (for enhanced purchase decision flow)
	structuredAPI := router.Group("/api/structured-chat")
	{
		structuredAPI.POST("/step", structuredChatHandler.ProcessStep)
		structuredAPI.POST("/recommendation", structuredChatHandler.GenerateRecommendation)
	}

	// Additional API routes for decision saving
	decisionAPI := router.Group("/api/decisions")
	decisionAPI.Use(jwtAuthMiddleware.RequireAuth())
	{
		decisionAPI.POST("/save", func(c *gin.Context) {
			c.JSON(200, gin.H{"message": "Decision saved successfully"})
		})
	}
}


// ================================
// MOCK DECISION SERVICE IMPLEMENTATION
// ================================

type mockDecisionServiceDirect struct{}

// MakeDecision implements a simple decision logic for testing
func (m *mockDecisionServiceDirect) MakeDecision(ctx context.Context, intentDTO types.PurchaseIntentDTO) (*types.DecisionResponseDTO, error) {
	var decision string
	var confidence float64
	var reason string
	var recommendations []string

	if intentDTO.ItemCost < 50 {
		decision = "BUY"
		confidence = 0.9
		reason = "Low cost item - safe to purchase"
		recommendations = []string{"Great choice!", "This is an affordable purchase"}
	} else if intentDTO.ItemCost < 200 {
		decision = "WAIT"
		confidence = 0.7
		reason = "Moderate cost - consider if this is necessary"
		recommendations = []string{"Wait for a sale", "Consider alternatives", "Check your budget"}
	} else {
		decision = "BYE"
		confidence = 0.8
		reason = "High cost item - recommend avoiding"
		recommendations = []string{"This exceeds recommended spending", "Look for cheaper alternatives", "Save money instead"}
	}

	if intentDTO.Category == "health" {
		decision = "BUY"
		confidence = 0.95
		reason = "Health purchases are always important"
		recommendations = []string{"Health is a priority", "Good investment in your wellbeing"}
	}

	response := &types.DecisionResponseDTO{
		ID:            generateMockDecisionID(),
		UserID:        intentDTO.UserID,
		IntentID:      intentDTO.ID,
		Decision:      decision,
		Confidence:    confidence,
		PrimaryReason: reason,
		Factors: []types.DecisionFactorDTO{
			{
				Category:    "cost",
				Impact:      "primary",
				Weight:      0.8,
				Description: fmt.Sprintf("Item cost: $%.2f", intentDTO.ItemCost),
			},
			{
				Category:    "category",
				Impact:      "secondary",
				Weight:      0.6,
				Description: fmt.Sprintf("Category: %s", intentDTO.Category),
			},
		},
		Recommendations: recommendations,
		WaitPeriod:      getWaitPeriod(decision),
		MaxBudget:       intentDTO.ItemCost * 0.8,
		CreatedAt:       time.Now(),
		ProcessingTime:  50,
	}

	return response, nil
}

func (m *mockDecisionServiceDirect) GetDecisionHistory(ctx context.Context, userID string, days int) (*types.DecisionHistoryDTO, error) {
	pastDecisions := []domain.PastDecision{
		{ItemName: "Coffee Machine", ItemCost: 299.99, Category: "electronics", Decision: "BUY", DaysAgo: 5},
		{ItemName: "Designer Shoes", ItemCost: 450.00, Category: "clothing", Decision: "BYE", DaysAgo: 12},
		{ItemName: "Gym Membership", ItemCost: 89.99, Category: "health", Decision: "WAIT", DaysAgo: 8},
		{ItemName: "Gaming Headset", ItemCost: 129.99, Category: "electronics", Decision: "BUY", DaysAgo: 15},
	}

	historyDTO := &types.DecisionHistoryDTO{}
	periodDescription := fmt.Sprintf("last_%d_days", days)
	historyDTO.FromDomainHistory(userID, pastDecisions, periodDescription)

	return historyDTO, nil
}

func generateMockDecisionID() string {
	return fmt.Sprintf("mock_decision_%d", time.Now().UnixNano())
}

func getWaitPeriod(decision string) int {
	switch decision {
	case "WAIT":
		return 30
	case "BYE":
		return 90
	default:
		return 0
	}
}
