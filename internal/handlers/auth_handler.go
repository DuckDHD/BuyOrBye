package handlers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/DuckDHD/BuyOrBye/internal/domain"
	"github.com/DuckDHD/BuyOrBye/internal/logging"
	"github.com/DuckDHD/BuyOrBye/internal/services"
	"github.com/DuckDHD/BuyOrBye/internal/types"
)

// AuthService interface from services package - handlers act as DTO-domain adapters

// AuthHandler handles HTTP requests for authentication endpoints
type AuthHandler struct {
	authService services.AuthService
	validator   *validator.Validate
}

// NewAuthHandler creates a new authentication handler with dependency injection
func NewAuthHandler(authService services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		validator:   validator.New(),
	}
}

// Login handles POST /api/auth/login requests
// Authenticates user with email and password
// Returns JWT token pair on successful authentication
func (h *AuthHandler) Login(c *gin.Context) {
	logger := logging.ContextLogger(c).With(logging.WithOperation("login"))
	logger.Info("Login request started")

	var request types.LoginRequestDTO

	// Parse and bind JSON request
	if err := c.ShouldBindJSON(&request); err != nil {
		logger.Warn("Login request failed - invalid JSON", logging.WithError(err))
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

	// Transform DTO to domain object
	credentials := domain.Credentials{
		Email:    request.Email,
		Password: request.Password,
	}

	// Call service layer with domain object
	logger.Info("Authenticating user credentials", logging.WithUserID(request.Email))
	tokenPair, err := h.authService.Login(c.Request.Context(), credentials)
	if err != nil {
		logger.Error("Login failed", logging.WithUserID(request.Email), logging.WithError(err))
		h.handleAuthError(c, err)
		return
	}

	// Transform domain object to DTO response
	response := types.TokenResponseDTO{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
	}

	logger.Info("Login successful", logging.WithUserID(request.Email))
	c.JSON(http.StatusOK, response)
}

// Register handles POST /api/auth/register requests
// Creates new user account and returns JWT token pair
func (h *AuthHandler) Register(c *gin.Context) {
	logger := logging.ContextLogger(c).With(logging.WithOperation("register"))
	logger.Info("Registration request started")

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

	// Transform DTO to domain object
	user := request.ToDomain()

	// Call service layer with domain object
	tokenPair, err := h.authService.Register(c.Request.Context(), user, request.Password)
	if err != nil {
		// Handle specific registration errors
		if strings.Contains(err.Error(), "already exists") {
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

	// Transform domain object to DTO response
	response := types.TokenResponseDTO{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
	}

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

	// Call service layer with token string
	tokenPair, err := h.authService.RefreshToken(c.Request.Context(), request.RefreshToken)
	if err != nil {
		h.handleAuthError(c, err)
		return
	}

	// Transform domain object to DTO response
	response := types.TokenResponseDTO{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresIn:    tokenPair.ExpiresIn,
	}

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

	// Call service layer with token string
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
	errMsg := err.Error()
	switch {
	case strings.Contains(errMsg, "invalid credentials"):
		c.JSON(http.StatusUnauthorized, types.NewErrorResponse(
			http.StatusUnauthorized,
			"unauthorized",
			"Invalid email or password",
		))
	case strings.Contains(errMsg, "account inactive"):
		c.JSON(http.StatusUnauthorized, types.NewErrorResponse(
			http.StatusUnauthorized,
			"unauthorized",
			"Your account is inactive. Please contact support",
		))
	case strings.Contains(errMsg, "invalid token") || strings.Contains(errMsg, "malformed token"):
		c.JSON(http.StatusUnauthorized, types.NewErrorResponse(
			http.StatusUnauthorized,
			"unauthorized",
			"Invalid or malformed token",
		))
	case strings.Contains(errMsg, "token expired") || strings.Contains(errMsg, "expired"):
		c.JSON(http.StatusUnauthorized, types.NewErrorResponse(
			http.StatusUnauthorized,
			"unauthorized",
			"Token has expired",
		))
	case strings.Contains(errMsg, "token revoked") || strings.Contains(errMsg, "revoked"):
		c.JSON(http.StatusUnauthorized, types.NewErrorResponse(
			http.StatusUnauthorized,
			"unauthorized",
			"Token has been revoked",
		))
	case strings.Contains(errMsg, "user not found") || strings.Contains(errMsg, "not found"):
		// Map user not found to invalid credentials for security
		c.JSON(http.StatusUnauthorized, types.NewErrorResponse(
			http.StatusUnauthorized,
			"unauthorized",
			"Invalid email or password",
		))
	case strings.Contains(errMsg, "already exists"):
		c.JSON(http.StatusConflict, types.NewErrorResponse(
			http.StatusConflict,
			"conflict",
			"User with this email already exists",
		))
	case strings.Contains(errMsg, "invalid user data"):
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
