package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/joho/godotenv/autoload"

	"github.com/DuckDHD/BuyOrBye/internal/database"
	"github.com/DuckDHD/BuyOrBye/internal/handlers"
	"github.com/DuckDHD/BuyOrBye/internal/middleware"
	"github.com/DuckDHD/BuyOrBye/internal/repositories"
	"github.com/DuckDHD/BuyOrBye/internal/services"
)

// AuthServer represents the main application server with authentication
type AuthServer struct {
	port             int
	gormService      *database.GormService
	
	// Services
	passwordService  services.PasswordService
	jwtService       services.JWTService
	authService      services.AuthService
	
	// Handlers
	authHandler      *handlers.AuthHandler
	
	// Middleware
	authMiddleware   *middleware.JWTAuthMiddleware
	loginRateLimit   *middleware.InMemoryRateLimiter
	apiRateLimit     *middleware.InMemoryRateLimiter
}

// NewAuthServer creates a new server instance with full authentication setup
func NewAuthServer() (*http.Server, error) {
	port, _ := strconv.Atoi(os.Getenv("PORT"))
	if port == 0 {
		port = 8080 // Default port
	}

	// Initialize database
	gormService, err := database.NewGormService()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	// Initialize password service
	passwordService := services.NewPasswordService()

	// Initialize JWT service
	jwtService, err := services.NewJWTService()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize JWT service: %w", err)
	}

	// Initialize repositories
	userRepo := repositories.NewUserRepository(gormService.GetDB())
	tokenRepo := repositories.NewTokenRepository(gormService.GetDB())

	// Initialize auth service
	authService := services.NewAuthService(userRepo, tokenRepo, passwordService, jwtService)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)

	// Initialize middleware
	authMiddleware := middleware.NewJWTAuthMiddleware(jwtService)
	loginRateLimit := middleware.NewLoginRateLimiter()
	apiRateLimit := middleware.NewAPIRateLimiter()

	server := &AuthServer{
		port:            port,
		gormService:     gormService,
		passwordService: passwordService,
		jwtService:      jwtService,
		authService:     authService,
		authHandler:     authHandler,
		authMiddleware:  authMiddleware,
		loginRateLimit:  loginRateLimit,
		apiRateLimit:    apiRateLimit,
	}

	// Create HTTP server
	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", server.port),
		Handler:      server.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return httpServer, nil
}

// RegisterRoutes sets up all routes with proper middleware
func (s *AuthServer) RegisterRoutes() http.Handler {
	// Set Gin mode based on environment
	if os.Getenv("APP_ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	// CORS middleware - allow credentials for authentication
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://localhost:3000"}, 
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
	}))

	// Apply general API rate limiting to all routes
	r.Use(s.apiRateLimit.RateLimit())
	r.Use(s.apiRateLimit.WithRateLimitHeaders())

	// Health check endpoints (no auth required)
	r.GET("/health", s.healthHandler)
	r.GET("/ping", s.pingHandler)

	// CSRF configuration for protected routes
	csrfConfig := middleware.DefaultCSRFConfig()
	csrfConfig.Secure = os.Getenv("APP_ENV") == "production" // Only secure in production
	csrfMiddleware := middleware.NewCSRFMiddleware(csrfConfig)

	// Authentication routes group
	authRoutes := r.Group("/auth")
	{
		// Public authentication endpoints (no auth middleware)
		public := authRoutes.Group("")
		{
			// Apply CSRF protection to state-changing operations
			public.Use(csrfMiddleware)
			
			// Apply rate limiting specifically to login endpoint
			public.POST("/login", s.loginRateLimit.RateLimit(), s.authHandler.Login)
			public.POST("/register", s.authHandler.Register)
			public.POST("/refresh", s.authHandler.RefreshToken)
		}

		// CSRF token endpoint (for SPAs to get CSRF tokens)
		authRoutes.GET("/csrf", csrfMiddleware, middleware.GetCSRFTokenHandler())

		// Protected authentication endpoints (require auth middleware)
		protected := authRoutes.Group("")
		protected.Use(s.authMiddleware.RequireAuth())
		{
			protected.Use(csrfMiddleware) // CSRF protection on authenticated routes too
			protected.POST("/logout", s.authHandler.Logout)
			protected.GET("/me", s.getUserProfileHandler)
		}
	}

	// API routes group (for future expansion)
	apiRoutes := r.Group("/api")
	{
		// Apply authentication middleware to all API routes
		apiRoutes.Use(s.authMiddleware.RequireAuth())
		apiRoutes.Use(csrfMiddleware)
		
		// Example protected API endpoint
		apiRoutes.GET("/protected", s.protectedHandler)
	}

	// Optional auth routes (authentication if token present, but not required)
	optionalRoutes := r.Group("/public")
	{
		optionalRoutes.Use(s.authMiddleware.OptionalAuth())
		optionalRoutes.GET("/info", s.publicInfoHandler)
	}

	return r
}

// Health check handler
func (s *AuthServer) healthHandler(c *gin.Context) {
	health := s.gormService.Health()
	
	if health["status"] == "up" {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"database": health,
			"timestamp": time.Now().UTC(),
		})
	} else {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status": "unhealthy",
			"database": health,
			"timestamp": time.Now().UTC(),
		})
	}
}

// Simple ping handler for load balancer checks
func (s *AuthServer) pingHandler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "pong"})
}

// Get current user profile handler
func (s *AuthServer) getUserProfileHandler(c *gin.Context) {
	// Get user information from JWT middleware context
	userID := middleware.GetUserID(c)
	email := middleware.GetUserEmail(c)
	claims := middleware.GetUserClaims(c)
	
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "user information not found in token",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user": gin.H{
			"id":         userID,
			"email":      email,
			"expires_at": claims.ExpiresAt,
		},
		"message": "user profile retrieved successfully",
	})
}

// Protected API handler example
func (s *AuthServer) protectedHandler(c *gin.Context) {
	userID := middleware.GetUserID(c)
	
	c.JSON(http.StatusOK, gin.H{
		"message":     "This is a protected resource",
		"user_id":     userID,
		"accessed_at": time.Now().UTC(),
	})
}

// Public info handler with optional authentication
func (s *AuthServer) publicInfoHandler(c *gin.Context) {
	response := gin.H{
		"message":     "This is public information",
		"timestamp":   time.Now().UTC(),
		"authenticated": middleware.IsAuthenticated(c),
	}
	
	// Add user info if authenticated
	if middleware.IsAuthenticated(c) {
		response["user_id"] = middleware.GetUserID(c)
	}
	
	c.JSON(http.StatusOK, response)
}

// Close gracefully shuts down the server and cleans up resources
func (s *AuthServer) Close() error {
	log.Println("Shutting down authentication server...")
	
	// Close rate limiters
	if s.loginRateLimit != nil {
		s.loginRateLimit.Close()
	}
	if s.apiRateLimit != nil {
		s.apiRateLimit.Close()
	}
	
	// Close database connection
	if s.gormService != nil {
		return s.gormService.Close()
	}
	
	return nil
}