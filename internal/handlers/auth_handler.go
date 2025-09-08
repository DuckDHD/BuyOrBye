package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/DuckDHD/BuyOrBye/internal/domain"
	"github.com/DuckDHD/BuyOrBye/internal/services"
	"github.com/DuckDHD/BuyOrBye/internal/types"
)

// AuthService interface defines the contract for authentication operations
// This matches the interface from the services package to avoid circular imports
type AuthService = services.AuthService

// AuthHandler handles HTTP requests for authentication endpoints
type AuthHandler struct {
	authService AuthService
	validator   *validator.Validate
}

// NewAuthHandler creates a new authentication handler with dependency injection
func NewAuthHandler(authService AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		validator:   validator.New(),
	}
}

// Login handles POST /api/auth/login requests
// Authenticates user with email and password
// Returns JWT token pair on successful authentication
func (h *AuthHandler) Login(c *gin.Context) {
	var request types.LoginRequestDTO

	// Parse and bind JSON request
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(
			http.StatusBadRequest,
			"bad_request",
			"Invalid JSON format",
		))
		return
	}

	// Validate request fields
	if err := h.validator.Struct(&request); err != nil {
		validationErrors := make(map[string]interface{})
		for _, err := range err.(validator.ValidationErrors) {
			field := err.Field()
			switch err.Tag() {
			case "required":
				validationErrors[field] = field + " is required"
			case "email":
				validationErrors[field] = field + " must be a valid email address"
			case "min":
				validationErrors[field] = field + " must be at least " + err.Param() + " characters"
			default:
				validationErrors[field] = field + " is invalid"
			}
		}

		c.JSON(http.StatusBadRequest, types.NewValidationErrorResponse(
			"Validation failed",
			validationErrors,
		))
		return
	}

	// Convert DTO to domain struct
	credentials := request.ToDomain()

	// Call service layer
	tokenPair, err := h.authService.Login(c.Request.Context(), credentials)
	if err != nil {
		h.handleAuthError(c, err)
		return
	}

	// Convert domain response to DTO
	var response types.TokenResponseDTO
	response.FromDomain(tokenPair)

	c.JSON(http.StatusOK, response)
}

// Register handles POST /api/auth/register requests
// Creates new user account and returns JWT token pair
func (h *AuthHandler) Register(c *gin.Context) {
	var request types.RegisterRequestDTO

	// Parse and bind JSON request
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(
			http.StatusBadRequest,
			"bad_request",
			"Invalid JSON format",
		))
		return
	}

	// Validate request fields
	if err := h.validator.Struct(&request); err != nil {
		validationErrors := make(map[string]interface{})
		for _, err := range err.(validator.ValidationErrors) {
			field := err.Field()
			switch err.Tag() {
			case "required":
				validationErrors[field] = field + " is required"
			case "email":
				validationErrors[field] = field + " must be a valid email address"
			case "min":
				validationErrors[field] = field + " must be at least " + err.Param() + " characters"
			default:
				validationErrors[field] = field + " is invalid"
			}
		}

		c.JSON(http.StatusBadRequest, types.NewValidationErrorResponse(
			"Validation failed",
			validationErrors,
		))
		return
	}

	// Convert DTO to domain struct
	user := request.ToDomain()

	// Call service layer
	tokenPair, err := h.authService.Register(c.Request.Context(), user, request.Password)
	if err != nil {
		// Handle specific registration errors
		if errors.Is(err, domain.ErrUserAlreadyExists) {
			c.JSON(http.StatusConflict, types.NewErrorResponse(
				http.StatusConflict,
				"conflict",
				"User with this email already exists",
			))
			return
		}
		h.handleAuthError(c, err)
		return
	}

	// Convert domain response to DTO
	var response types.TokenResponseDTO
	response.FromDomain(tokenPair)

	c.JSON(http.StatusCreated, response)
}

// RefreshToken handles POST /api/auth/refresh requests
// Generates new JWT token pair using valid refresh token
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var request types.RefreshTokenRequestDTO

	// Parse and bind JSON request
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(
			http.StatusBadRequest,
			"bad_request",
			"Invalid JSON format",
		))
		return
	}

	// Validate request fields
	if err := h.validator.Struct(&request); err != nil {
		validationErrors := make(map[string]interface{})
		for _, err := range err.(validator.ValidationErrors) {
			field := err.Field()
			if err.Tag() == "required" {
				validationErrors[field] = field + " is required"
			} else {
				validationErrors[field] = field + " is invalid"
			}
		}

		c.JSON(http.StatusBadRequest, types.NewValidationErrorResponse(
			"Validation failed",
			validationErrors,
		))
		return
	}

	// Call service layer
	tokenPair, err := h.authService.RefreshToken(c.Request.Context(), request.RefreshToken)
	if err != nil {
		h.handleAuthError(c, err)
		return
	}

	// Convert domain response to DTO
	var response types.TokenResponseDTO
	response.FromDomain(tokenPair)

	c.JSON(http.StatusOK, response)
}

// Logout handles POST /api/auth/logout requests
// Revokes the provided refresh token
func (h *AuthHandler) Logout(c *gin.Context) {
	var request types.RefreshTokenRequestDTO

	// Parse and bind JSON request
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(
			http.StatusBadRequest,
			"bad_request",
			"Invalid JSON format",
		))
		return
	}

	// Validate request fields
	if err := h.validator.Struct(&request); err != nil {
		validationErrors := make(map[string]interface{})
		for _, err := range err.(validator.ValidationErrors) {
			field := err.Field()
			if err.Tag() == "required" {
				validationErrors[field] = field + " is required"
			} else {
				validationErrors[field] = field + " is invalid"
			}
		}

		c.JSON(http.StatusBadRequest, types.NewValidationErrorResponse(
			"Validation failed",
			validationErrors,
		))
		return
	}

	// Call service layer
	if err := h.authService.Logout(c.Request.Context(), request.RefreshToken); err != nil {
		h.handleAuthError(c, err)
		return
	}

	// Return success response
	c.JSON(http.StatusOK, gin.H{
		"message": "logged out successfully",
	})
}

// handleAuthError handles authentication-specific errors and maps them to appropriate HTTP responses
func (h *AuthHandler) handleAuthError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidCredentials):
		c.JSON(http.StatusUnauthorized, types.NewErrorResponse(
			http.StatusUnauthorized,
			"unauthorized",
			"Invalid email or password",
		))
	case errors.Is(err, domain.ErrAccountInactive):
		c.JSON(http.StatusUnauthorized, types.NewErrorResponse(
			http.StatusUnauthorized,
			"unauthorized",
			"Your account is inactive. Please contact support",
		))
	case errors.Is(err, domain.ErrInvalidToken):
		c.JSON(http.StatusUnauthorized, types.NewErrorResponse(
			http.StatusUnauthorized,
			"unauthorized",
			"Invalid or malformed token",
		))
	case errors.Is(err, domain.ErrTokenExpired):
		c.JSON(http.StatusUnauthorized, types.NewErrorResponse(
			http.StatusUnauthorized,
			"unauthorized",
			"token has expired",
		))
	case errors.Is(err, domain.ErrTokenRevoked):
		c.JSON(http.StatusUnauthorized, types.NewErrorResponse(
			http.StatusUnauthorized,
			"unauthorized",
			"Token has been revoked",
		))
	case errors.Is(err, domain.ErrUserNotFound):
		// Map user not found to invalid credentials for security
		c.JSON(http.StatusUnauthorized, types.NewErrorResponse(
			http.StatusUnauthorized,
			"unauthorized",
			"Invalid email or password",
		))
	case errors.Is(err, domain.ErrUserAlreadyExists):
		c.JSON(http.StatusConflict, types.NewErrorResponse(
			http.StatusConflict,
			"conflict",
			"User with this email already exists",
		))
	case errors.Is(err, domain.ErrInvalidUserData):
		c.JSON(http.StatusBadRequest, types.NewErrorResponse(
			http.StatusBadRequest,
			"bad_request",
			"Invalid user data provided",
		))
	default:
		// Internal server error for unexpected errors
		c.JSON(http.StatusInternalServerError, types.NewErrorResponse(
			http.StatusInternalServerError,
			"internal_error",
			"An internal error occurred. Please try again later",
		))
	}
}