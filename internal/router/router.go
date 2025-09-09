package router

import (
	"github.com/gin-gonic/gin"
	
	"github.com/DuckDHD/BuyOrBye/internal/handlers"
	"github.com/DuckDHD/BuyOrBye/internal/middleware"
	"github.com/DuckDHD/BuyOrBye/internal/services"
)

// Router handles all application routes
type Router struct {
	authRouter    *AuthRouter
	financeRouter *FinanceRouter
}

// NewRouter creates a new main router with all domain routers
func NewRouter(authHandler *handlers.AuthHandler, financeHandler *handlers.FinanceHandler, jwtService services.JWTService) *Router {
	return &Router{
		authRouter:    NewAuthRouter(authHandler, jwtService),
		financeRouter: NewFinanceRouter(financeHandler, jwtService),
	}
}

// SetupRoutes configures all application routes
func (r *Router) SetupRoutes() *gin.Engine {
	// Create Gin router with default middleware
	router := gin.Default()

	// Add global middleware
	router.Use(middleware.CORS())
	router.Use(middleware.Logger())
	router.Use(middleware.Recovery())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
			"message": "BuyOrBye API is running",
		})
	})

	// API version group
	apiV1 := router.Group("/api/v1")
	{
		// Register domain routers
		r.authRouter.RegisterRoutes(apiV1)
		r.financeRouter.RegisterRoutes(apiV1)
	}

	return router
}