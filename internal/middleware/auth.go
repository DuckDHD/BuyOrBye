package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/DuckDHD/BuyOrBye/internal/domain"
	"github.com/DuckDHD/BuyOrBye/internal/services"
	"github.com/DuckDHD/BuyOrBye/internal/types"
)

// JWTAuthMiddleware provides JWT authentication middleware for Gin
type JWTAuthMiddleware struct {
	jwtService services.JWTService
}

// NewJWTAuthMiddleware creates a new JWT authentication middleware
func NewJWTAuthMiddleware(jwtService services.JWTService) *JWTAuthMiddleware {
	return &JWTAuthMiddleware{
		jwtService: jwtService,
	}
}

// RequireAuth is a Gin middleware that validates JWT tokens
// Extracts JWT from Authorization header (Bearer token format)
// Validates token using JWTService and adds user claims to context
// Returns 401 for invalid, expired, or missing tokens
func (j *JWTAuthMiddleware) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, types.NewErrorResponse(
				http.StatusUnauthorized,
				"unauthorized",
				"Authorization header is required",
			))
			c.Abort()
			return
		}

		// Check Bearer token format
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, types.NewErrorResponse(
				http.StatusUnauthorized,
				"unauthorized",
				"Invalid authorization header format. Expected 'Bearer <token>'",
			))
			c.Abort()
			return
		}

		// Extract the actual token
		tokenString := tokenParts[1]
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, types.NewErrorResponse(
				http.StatusUnauthorized,
				"unauthorized",
				"Access token is required",
			))
			c.Abort()
			return
		}

		// Validate the access token using JWTService
		claims, err := j.jwtService.ValidateAccessToken(tokenString)
		if err != nil {
			// Determine specific error type for appropriate response
			statusCode := http.StatusUnauthorized
			errorCode := "unauthorized"
			message := "Invalid access token"

			if strings.Contains(err.Error(), "expired") {
				message = "Access token has expired"
			} else if strings.Contains(err.Error(), "signature") {
				message = "Invalid token signature"
			} else if strings.Contains(err.Error(), "malformed") || strings.Contains(err.Error(), "parse") {
				message = "Malformed access token"
			}

			c.JSON(statusCode, types.NewErrorResponse(
				statusCode,
				errorCode,
				message,
			))
			c.Abort()
			return
		}

		// Additional validation: Check if token has expired using domain logic
		if claims.IsExpired() {
			c.JSON(http.StatusUnauthorized, types.NewErrorResponse(
				http.StatusUnauthorized,
				"unauthorized",
				"Access token has expired",
			))
			c.Abort()
			return
		}

		// Store user claims in Gin context for use by handlers
		c.Set("userID", claims.UserID)
		c.Set("userEmail", claims.Email)
		c.Set("tokenClaims", claims)

		// Continue to the next middleware/handler
		c.Next()
	}
}

// OptionalAuth is a Gin middleware that validates JWT tokens when present
// Unlike RequireAuth, this middleware doesn't return 401 for missing tokens
// If a token is present and valid, it adds user claims to context
// If no token is present or token is invalid, it continues without claims
func (j *JWTAuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// No token provided, continue without authentication
			c.Next()
			return
		}

		// Check Bearer token format
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			// Invalid format, continue without authentication
			c.Next()
			return
		}

		// Extract the actual token
		tokenString := tokenParts[1]
		if tokenString == "" {
			// Empty token, continue without authentication
			c.Next()
			return
		}

		// Validate the access token using JWTService
		claims, err := j.jwtService.ValidateAccessToken(tokenString)
		if err != nil || claims.IsExpired() {
			// Invalid or expired token, continue without authentication
			c.Next()
			return
		}

		// Store user claims in Gin context for use by handlers
		c.Set("userID", claims.UserID)
		c.Set("userEmail", claims.Email)
		c.Set("tokenClaims", claims)

		// Continue to the next middleware/handler
		c.Next()
	}
}

// GetUserClaims extracts user claims from Gin context
// Returns nil if no authenticated user is found
func GetUserClaims(c *gin.Context) *domain.TokenClaims {
	claims, exists := c.Get("tokenClaims")
	if !exists {
		return nil
	}
	
	tokenClaims, ok := claims.(*domain.TokenClaims)
	if !ok {
		return nil
	}
	
	return tokenClaims
}

// GetUserID extracts user ID from Gin context
// Returns empty string if no authenticated user is found
func GetUserID(c *gin.Context) string {
	userID, exists := c.Get("userID")
	if !exists {
		return ""
	}
	
	id, ok := userID.(string)
	if !ok {
		return ""
	}
	
	return id
}

// GetUserEmail extracts user email from Gin context
// Returns empty string if no authenticated user is found
func GetUserEmail(c *gin.Context) string {
	email, exists := c.Get("userEmail")
	if !exists {
		return ""
	}
	
	userEmail, ok := email.(string)
	if !ok {
		return ""
	}
	
	return userEmail
}

// IsAuthenticated checks if the current request has valid authentication
// Returns true if user claims are present in context
func IsAuthenticated(c *gin.Context) bool {
	return GetUserClaims(c) != nil
}