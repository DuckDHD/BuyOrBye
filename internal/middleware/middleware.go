package middleware

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	
	"github.com/DuckDHD/BuyOrBye/internal/services"
)

// CORS returns a CORS middleware with default configuration
func CORS() gin.HandlerFunc {
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	config.AllowHeaders = []string{"*"}
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	return cors.New(config)
}

// Logger returns Gin's default logger middleware
func Logger() gin.HandlerFunc {
	return gin.Logger()
}

// Recovery returns Gin's default recovery middleware
func Recovery() gin.HandlerFunc {
	return gin.Recovery()
}

// JWTAuth returns a JWT authentication middleware instance
// This is a helper function that requires proper service injection
func JWTAuth(jwtService services.JWTService) gin.HandlerFunc {
	middleware := NewJWTAuthMiddleware(jwtService)
	return middleware.RequireAuth()
}