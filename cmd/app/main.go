package main

import (
	"context"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/DuckDHD/BuyOrBye/internal/config"
	"github.com/DuckDHD/BuyOrBye/internal/database"
	"github.com/DuckDHD/BuyOrBye/internal/handlers"
	"github.com/DuckDHD/BuyOrBye/internal/logging"
	"github.com/DuckDHD/BuyOrBye/internal/middleware"
	"github.com/DuckDHD/BuyOrBye/internal/repositories"
	"github.com/DuckDHD/BuyOrBye/internal/services"
)

func main() {
	// Load configuration first
	cfg, err := config.LoadConfig()
	if err != nil {
		panic("Failed to load configuration: " + err.Error())
	}

	// Initialize logger with config
	if err := logging.InitLogger(logging.LogConfig{
		Environment: cfg.Logging.Environment,
		Level:       cfg.Logging.Level,
	}); err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
	logger := logging.GetLogger()
	logger.Info("Configuration loaded successfully",
		logging.WithComponent("main"),
		zap.String("environment", cfg.Server.Environment),
		zap.String("config_file", config.GetConfigPath(cfg.Server.Environment)))

	// Initialize database service with config
	dbService, err := config.NewDatabaseService(&cfg.Database, &cfg.Logging)
	if err != nil {
		logger.Fatal("Failed to initialize database", logging.WithError(err))
	}

	db := dbService.GetDB()

	// Run migrations
	if err := database.RunAllMigrations(db); err != nil {
		logger.Fatal("Migration failed", logging.WithError(err))
	}
	logger.Info("Database migrations completed successfully", logging.WithComponent("main"))

	// Initialize core services with config
	passwordService := services.NewPasswordService()
	jwtService, err := services.NewJWTServiceFromConfig(&cfg.Auth)
	if err != nil {
		logger.Fatal("Failed to initialize JWT service", logging.WithError(err))
	}

	// Initialize repositories
	userRepo := repositories.NewUserRepository(db)
	tokenRepo := repositories.NewTokenRepository(db)

	// Initialize finance repositories
	incomeRepo := repositories.NewIncomeRepository(db)
	expenseRepo := repositories.NewExpenseRepository(db)
	loanRepo := repositories.NewLoanRepository(db)
	financeSummaryRepo := repositories.NewFinanceSummaryRepository()

	// Create finance repositories aggregate
	financeRepos := services.NewFinanceRepositories(incomeRepo, expenseRepo, loanRepo, financeSummaryRepo)

	// Initialize health repositories
	healthProfileRepo := repositories.NewHealthProfileRepository(db)
	conditionRepo := repositories.NewMedicalConditionRepository(db)
	medicalExpenseRepo := repositories.NewMedicalExpenseRepository(db)
	policyRepo := repositories.NewInsurancePolicyRepository(db)

	// Initialize health analysis services
	riskCalculator := services.NewRiskCalculator()
	costAnalyzer := services.NewMedicalCostAnalyzer()

	// Initialize services
	authService := services.NewAuthService(userRepo, tokenRepo, passwordService, jwtService)
	financeService := services.NewFinanceService(financeRepos)
	// budgetAnalyzer will be used for future analysis endpoints
	_ = services.NewBudgetAnalyzer(financeService)

	// Initialize health service
	healthService := services.NewHealthService(
		healthProfileRepo,
		conditionRepo,
		medicalExpenseRepo,
		policyRepo,
		riskCalculator,
		costAnalyzer,
	)

	// TODO: Initialize decision domain components
	// Currently disabled due to interface compatibility issues that need to be resolved
	// The following components are available but need proper adapter interfaces:
	//
	// decisionRepo := repositories.NewDecisionRepository(db)
	// promptLogRepo := repositories.NewPromptLogRepository(db)
	// openaiClient := clients.NewOpenAIClient()
	// contextAggregator := services.NewContextAggregator(financeService, healthService, decisionRepo)
	// promptBuilder := services.NewPromptBuilder()
	// decisionInterpreter := services.NewDecisionInterpreter()
	// recommendationEngine := services.NewRecommendationEngine()
	// enhancedDecisionService := services.NewEnhancedDecisionService(...)

	// Decision service interface placeholder - will be implemented after interface resolution
	var decisionServiceInterface handlers.DecisionServiceInterface = nil
	var financeServiceInterface handlers.FinanceServiceInterface = nil

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)
	financeHandler := handlers.NewFinanceHandler(financeServiceInterface)
	healthHandler := handlers.NewHealthHandler(healthService)
	// Only initialize decision handler if service is available
	var decisionHandler *handlers.DecisionHandler
	if decisionServiceInterface != nil {
		decisionHandler = handlers.NewDecisionHandler(decisionServiceInterface)
	}

	// Initialize middlewares
	jwtAuthMiddleware := middleware.NewJWTAuthMiddleware(jwtService)

	// Setup Gin router
	router := gin.Default()

	// Global middleware with config
	router.Use(middleware.CORS())

	// Configure logging middleware based on environment
	middlewareConfig := config.GetMiddlewareConfig(cfg.Server.Environment)
	loggingConfig := logging.HTTPLoggingConfig{
		SkipPaths:       middlewareConfig.SkipPaths,
		LogRequestBody:  middlewareConfig.LogRequestBody,
		LogResponseBody: middlewareConfig.LogResponseBody,
		MaxBodySize:     middlewareConfig.MaxBodySize,
	}
	router.Use(logging.HTTPLoggingMiddleware(loggingConfig))
	router.Use(logging.ErrorLoggingMiddleware())
	router.Use(logging.RequestIDMiddleware())
	router.Use(middleware.ValidateRequestLimits())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "BuyOrBye API is running",
		})
	})

	// UI Routes (Frontend/Web Interface)
	setupUIRoutes(router, authHandler, financeHandler, healthHandler, decisionHandler, jwtAuthMiddleware)

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

		// Add spending insights endpoint when implemented
		// finance.GET("/insights", financeHandler.GetSpendingInsights)
	}

	// Health routes (all require auth)
	health := api.Group("/health")
	health.Use(jwtAuthMiddleware.RequireAuth())
	health.Use(middleware.SanitizeSensitiveData())
	{
		// Profile endpoints
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

		// Condition endpoints
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
		health.GET("/profiles/:profileId/conditions",
			middleware.ValidateHealthOwnership(),
			healthHandler.GetConditionsByProfile)

		// Policy endpoints
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
		health.GET("/profiles/:profileId/policies",
			middleware.ValidateHealthOwnership(),
			healthHandler.GetPoliciesByProfile)

		// Expense endpoints
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
		health.GET("/profiles/:profileId/expenses",
			middleware.ValidateHealthOwnership(),
			healthHandler.GetExpensesByProfile)
	}

	// Decision routes (all require auth) - Only enable if service is available
	if decisionHandler != nil {
		decision := api.Group("/decision")
		decision.Use(jwtAuthMiddleware.RequireAuth())
		{
			// Decision evaluation endpoint
			decision.POST("/evaluate",
				middleware.ValidateRequestLimits(),
				decisionHandler.MakeDecision)

			// Decision history endpoint
			decision.GET("/history", decisionHandler.GetDecisionHistory)

			// Decision statistics endpoint
			decision.GET("/stats", decisionHandler.GetDecisionStats)
		}
	}

	// Create HTTP server with config
	serverService := config.NewServerService(&cfg.Server)
	server := serverService.CreateServer(router)

	logger.Info("Starting BuyOrBye server",
		logging.WithComponent("main"),
		zap.String("address", serverService.GetAddress()),
		zap.String("environment", cfg.Server.Environment))

	// Create a done channel to signal when the shutdown is complete
	done := make(chan bool, 1)

	// Run graceful shutdown in a separate goroutine
	go gracefulShutdown(server, done)

	// Start the server
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatal("HTTP server error", logging.WithError(err))
	}

	// Wait for the graceful shutdown to complete
	<-done
	logger.Info("Graceful shutdown complete")
}

func gracefulShutdown(server *http.Server, done chan bool) {
	// Create context that listens for the interrupt signal from the OS.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Listen for the interrupt signal.
	<-ctx.Done()

	logger := logging.GetLogger()
	logger.Info("Shutting down gracefully, press Ctrl+C again to force")
	stop() // Allow Ctrl+C to force shutdown

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", logging.WithError(err))
	}

	logger.Info("Server exiting")

	// Notify the main goroutine that the shutdown is complete
	done <- true
}

// setupUIRoutes organizes frontend UI routes with proper separation
func setupUIRoutes(
	router *gin.Engine,
	authHandler *handlers.AuthHandler,
	financeHandler *handlers.FinanceHandler,
	healthHandler *handlers.HealthHandler,
	decisionHandler *handlers.DecisionHandler,
	jwtAuthMiddleware *middleware.JWTAuthMiddleware,
) {
	// Static assets with caching
	router.Static("/static", "cmd/web/static")
	router.StaticFile("/favicon.ico", "cmd/web/static/favicon.svg")
	router.StaticFile("/manifest.json", "cmd/web/static/manifest.json")
	router.StaticFile("/robots.txt", "cmd/web/static/robots.txt")

	// Public UI routes (no auth required)
	router.GET("/", redirectToDashboard)
	router.GET("/login", renderLoginPage)
	router.GET("/register", renderRegisterPage)
	router.GET("/forgot-password", renderForgotPasswordPage)

	// Protected UI routes group
	protected := router.Group("")
	protected.Use(jwtAuthMiddleware.RequireAuth())
	{
		// Main page routes - return full HTML pages
		protected.GET("/dashboard", renderDashboardPage)
		protected.GET("/finance", renderFinanceOverviewPage)
		protected.GET("/health", renderHealthProfilePage)
		protected.GET("/decisions/new", renderDecisionNewPage)
		protected.GET("/decisions/history", renderDecisionHistoryPage)

		// HTMX partial routes group - return HTML fragments only
		ui := protected.Group("/ui/partials")
		{
			// Dashboard partials
			ui.GET("/dashboard/stats", renderDashboardStatsPartial)
			ui.GET("/dashboard/quick-decision", renderQuickDecisionPartial)
			ui.GET("/dashboard/recent-decisions", renderRecentDecisionsPartial)
			ui.GET("/dashboard/content", renderDashboardContentPartial)
			ui.GET("/dashboard/below-fold", renderDashboardBelowFoldPartial)
			ui.GET("/dashboard/insights", renderDashboardInsightsPartial)

			// Finance partials
			ui.GET("/finance/overview", renderFinanceOverviewPartial)
			ui.GET("/finance/summary", renderFinanceSummaryPartial)
			ui.GET("/finance/income/list", renderIncomeListPartial)
			ui.GET("/finance/expense/form", renderExpenseFormPartial)

			// Health partials
			ui.GET("/health/profile", renderHealthProfilePartial)
			ui.GET("/health/risk-gauge/:profileId", renderHealthRiskGaugePartial)
			ui.GET("/health/condition/add", renderConditionAddPartial)
			ui.GET("/health/insurance/card/:policyId", renderInsuranceCardPartial)

			// Decision partials
			if decisionHandler != nil {
				ui.GET("/decision/result", renderDecisionResultPartial)
				ui.GET("/decision/filter", renderDecisionFilterPartial)
			}
		}
	}
}

// Page handlers - return full HTML pages

func redirectToDashboard(c *gin.Context) {
	c.Redirect(302, "/dashboard")
}

func renderLoginPage(c *gin.Context) {
	// TODO: Implement login page rendering using existing templates
	c.String(200, "Login Page - TODO: Implement with existing auth templates")
}

func renderRegisterPage(c *gin.Context) {
	// TODO: Implement register page rendering using existing templates
	c.String(200, "Register Page - TODO: Implement with existing auth templates")
}

func renderForgotPasswordPage(c *gin.Context) {
	// TODO: Implement forgot password page rendering using existing templates
	c.String(200, "Forgot Password Page - TODO: Implement with existing auth templates")
}

func renderDashboardPage(c *gin.Context) {
	// TODO: Implement dashboard page rendering using existing dashboard_page.templ
	c.String(200, "Dashboard Page - TODO: Implement with existing dashboard templates")
}

func renderFinanceOverviewPage(c *gin.Context) {
	// TODO: Implement finance overview page rendering using existing finance_overview_page.templ
	c.String(200, "Finance Overview Page - TODO: Implement with existing finance templates")
}

func renderHealthProfilePage(c *gin.Context) {
	// TODO: Implement health profile page rendering using existing health_profile_page.templ
	c.String(200, "Health Profile Page - TODO: Implement with existing health templates")
}

func renderDecisionNewPage(c *gin.Context) {
	// TODO: Implement new decision page rendering using existing decision_new_page.templ
	c.String(200, "New Decision Page - TODO: Implement with existing decision templates")
}

func renderDecisionHistoryPage(c *gin.Context) {
	// TODO: Implement decision history page rendering using existing decision_history_page.templ
	c.String(200, "Decision History Page - TODO: Implement with existing decision templates")
}

// Partial handlers - return HTML fragments for HTMX

func renderDashboardStatsPartial(c *gin.Context) {
	// TODO: Implement dashboard stats partial using existing dashboard_stats_partial.templ
	c.String(200, "Dashboard Stats Partial - TODO: Implement with existing partial templates")
}

func renderQuickDecisionPartial(c *gin.Context) {
	// TODO: Implement quick decision partial using existing quick_decision_partial.templ
	c.String(200, "Quick Decision Partial - TODO: Implement with existing partial templates")
}

func renderRecentDecisionsPartial(c *gin.Context) {
	// TODO: Implement recent decisions partial using existing recent_decisions_partial.templ
	c.String(200, "Recent Decisions Partial - TODO: Implement with existing partial templates")
}

func renderDashboardContentPartial(c *gin.Context) {
	// TODO: Implement dashboard content partial (main dashboard body without layout)
	c.String(200, "Dashboard Content Partial - TODO: Implement")
}

func renderDashboardBelowFoldPartial(c *gin.Context) {
	// TODO: Implement dashboard below fold content (lazy loaded sections)
	c.String(200, "Dashboard Below Fold Partial - TODO: Implement")
}

func renderDashboardInsightsPartial(c *gin.Context) {
	// TODO: Implement dashboard insights partial (performance metrics)
	c.String(200, "Dashboard Insights Partial - TODO: Implement")
}

func renderFinanceOverviewPartial(c *gin.Context) {
	// TODO: Implement finance overview partial using existing finance_summary_partial.templ
	c.String(200, "Finance Overview Partial - TODO: Implement with existing partial templates")
}

func renderFinanceSummaryPartial(c *gin.Context) {
	// TODO: Implement finance summary partial using existing finance_summary_partial.templ
	c.String(200, "Finance Summary Partial - TODO: Implement with existing partial templates")
}

func renderIncomeListPartial(c *gin.Context) {
	// TODO: Implement income list partial using existing income_list_partial.templ
	c.String(200, "Income List Partial - TODO: Implement with existing partial templates")
}

func renderExpenseFormPartial(c *gin.Context) {
	// TODO: Implement expense form partial using existing expense_form_partial.templ
	c.String(200, "Expense Form Partial - TODO: Implement with existing partial templates")
}

func renderHealthProfilePartial(c *gin.Context) {
	// TODO: Implement health profile partial content
	c.String(200, "Health Profile Partial - TODO: Implement")
}

func renderHealthRiskGaugePartial(c *gin.Context) {
	// TODO: Implement health risk gauge partial using existing health_risk_gauge_partial.templ
	c.String(200, "Health Risk Gauge Partial - TODO: Implement with existing partial templates")
}

func renderConditionAddPartial(c *gin.Context) {
	// TODO: Implement condition add partial using existing condition_add_partial.templ
	c.String(200, "Condition Add Partial - TODO: Implement with existing partial templates")
}

func renderInsuranceCardPartial(c *gin.Context) {
	// TODO: Implement insurance card partial using existing insurance_card_partial.templ
	c.String(200, "Insurance Card Partial - TODO: Implement with existing partial templates")
}

func renderDecisionResultPartial(c *gin.Context) {
	// TODO: Implement decision result partial using existing decision_result_partial.templ
	c.String(200, "Decision Result Partial - TODO: Implement with existing partial templates")
}

func renderDecisionFilterPartial(c *gin.Context) {
	// TODO: Implement decision filter partial using existing decision_filter_partial.templ
	c.String(200, "Decision Filter Partial - TODO: Implement with existing partial templates")
}
