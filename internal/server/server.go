package server

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"

	"github.com/DuckDHD/BuyOrBye/internal/database"
	"github.com/DuckDHD/BuyOrBye/internal/handlers"
	"github.com/DuckDHD/BuyOrBye/internal/repositories"
	"github.com/DuckDHD/BuyOrBye/internal/router"
	"github.com/DuckDHD/BuyOrBye/internal/services"
)

type Server struct {
	port int
}

func NewServer() (*http.Server, error) {
	port, _ := strconv.Atoi(os.Getenv("PORT"))
	if port == 0 {
		port = 8080 // Default port
	}

	// Initialize GORM database service
	gormService, err := database.NewGormService()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Initialize services
	passwordService := services.NewPasswordService()
	jwtService, err := services.NewJWTService()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize JWT service: %w", err)
	}

	// Initialize repositories
	userRepo := repositories.NewUserRepository(gormService.GetDB())
	tokenRepo := repositories.NewTokenRepository(gormService.GetDB())
	incomeRepo := repositories.NewIncomeRepository(gormService.GetDB())
	expenseRepo := repositories.NewExpenseRepository(gormService.GetDB())
	loanRepo := repositories.NewLoanRepository(gormService.GetDB())
	financeSummaryRepo := repositories.NewFinanceSummaryRepository()

	// Create finance repositories aggregate
	financeRepos := services.NewFinanceRepositories(incomeRepo, expenseRepo, loanRepo, financeSummaryRepo)

	// Initialize services with proper dependencies
	authService := services.NewAuthService(userRepo, tokenRepo, passwordService, jwtService)
	financeService := services.NewFinanceService(financeRepos)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)
	financeHandler := handlers.NewFinanceHandler(financeService)

	// Initialize main router
	appRouter := router.NewRouter(authHandler, financeHandler, jwtService)

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      appRouter.SetupRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server, nil
}
