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
	insuranceEvaluator := services.NewInsuranceEvaluator()

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

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)
	financeHandler := handlers.NewFinanceHandler(financeService)
	healthHandler := handlers.NewHealthHandler(healthService)

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
	health.Use(middleware.ValidateHealthOwnership())
	health.Use(middleware.SanitizeSensitiveData())
	{
		// Profile endpoints
		health.POST("/profile",
			middleware.ValidateHealthProfileData(),
			healthHandler.CreateProfile)
		health.GET("/profile", healthHandler.GetProfile)
		health.PUT("/profile",
			middleware.ValidateHealthProfileData(),
			healthHandler.UpdateProfile)

		// Condition endpoints
		health.POST("/conditions",
			middleware.ValidateHealthOwnership(),
			healthHandler.AddCondition)
		health.GET("/conditions", healthHandler.GetConditions)
		health.PUT("/conditions/:id",
			middleware.ValidateHealthOwnership(),
			healthHandler.UpdateCondition)
		health.DELETE("/conditions/:id",
			middleware.ValidateHealthOwnership(),
			healthHandler.RemoveCondition)

		// Expense endpoints
		health.POST("/expenses",
			middleware.ValidateExpenseData(),
			middleware.ValidateHealthOwnership(),
			healthHandler.AddExpense)
		health.GET("/expenses", healthHandler.GetExpenses)
		health.GET("/expenses/recurring", healthHandler.GetRecurringExpenses)

		// Insurance endpoints
		health.POST("/insurance",
			middleware.ValidateInsuranceDates(),
			middleware.ValidateHealthOwnership(),
			healthHandler.AddInsurancePolicy)
		health.GET("/insurance", healthHandler.GetActivePolicies)
		health.PUT("/insurance/:id/deductible",
			middleware.ValidateHealthOwnership(),
			healthHandler.UpdateDeductibleProgress)

		// Analysis endpoints
		health.GET("/summary", healthHandler.GetHealthSummary)

		// Future endpoints for health context integration
		// health.GET("/risk-score", healthHandler.GetRiskScore)
		// health.GET("/context", healthHandler.GetHealthContext)
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
