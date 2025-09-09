package router

import (
	"github.com/gin-gonic/gin"
	
	"github.com/DuckDHD/BuyOrBye/internal/handlers"
	"github.com/DuckDHD/BuyOrBye/internal/middleware"
	"github.com/DuckDHD/BuyOrBye/internal/services"
)

// AuthRouter handles all authentication-related routes
type AuthRouter struct {
	authHandler *handlers.AuthHandler
	jwtService  services.JWTService
}

// NewAuthRouter creates a new authentication router
func NewAuthRouter(authHandler *handlers.AuthHandler, jwtService services.JWTService) *AuthRouter {
	return &AuthRouter{
		authHandler: authHandler,
		jwtService:  jwtService,
	}
}

// RegisterRoutes registers all authentication routes to the given router group
func (ar *AuthRouter) RegisterRoutes(rg *gin.RouterGroup) {
	// Public authentication routes (no middleware required)
	authGroup := rg.Group("/auth")
	{
		// POST /api/auth/login - User login
		authGroup.POST("/login", ar.authHandler.Login)
		
		// POST /api/auth/register - User registration  
		authGroup.POST("/register", ar.authHandler.Register)
		
		// POST /api/auth/refresh - Refresh access token using refresh token
		authGroup.POST("/refresh", ar.authHandler.RefreshToken)
	}

	// Protected authentication routes (require valid token)
	protectedAuthGroup := rg.Group("/auth")
	protectedAuthGroup.Use(middleware.JWTAuth(ar.jwtService))
	{
		// POST /api/auth/logout - User logout (requires auth to invalidate token)
		protectedAuthGroup.POST("/logout", ar.authHandler.Logout)
	}
}