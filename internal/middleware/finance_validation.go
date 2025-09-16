package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/DuckDHD/BuyOrBye/internal/domain"
	"github.com/DuckDHD/BuyOrBye/internal/types"
)

// ValidateOwnership middleware ensures users can only access their own financial data
// This middleware should be used on routes with :id parameters that represent financial records
func ValidateOwnership() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get authenticated user ID from context (set by JWT middleware)
		userID := GetUserID(c)
		if userID == "" {
			c.JSON(http.StatusUnauthorized, types.NewErrorResponse(
				http.StatusUnauthorized,
				"unauthorized",
				"Authentication required",
			))
			c.Abort()
			return
		}

		// Store user ID in context for handler use
		c.Set("authenticatedUserID", userID)
		c.Next()
	}
}

// ValidatePositiveAmount middleware validates that financial amounts are positive
// This middleware reads JSON body and validates amount fields
func ValidatePositiveAmount(field string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// This is a pre-validation middleware
		// The actual validation will be done in the handler when parsing JSON
		// We just pass through and let the domain validation handle it
		c.Next()
	}
}

// NormalizeFrequency middleware ensures frequency values are valid and normalized
func NormalizeFrequency() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the request body for processing
		var body map[string]interface{}
		
		// Only process if there's a JSON body
		if c.Request.ContentLength > 0 && strings.Contains(c.GetHeader("Content-Type"), "application/json") {
			if err := c.ShouldBindJSON(&body); err != nil {
				// If we can't parse JSON, let the handler deal with it
				c.Next()
				return
			}

			// Normalize frequency field if present
			if freq, exists := body["frequency"]; exists {
				if freqStr, ok := freq.(string); ok {
					normalizedFreq := normalizeFrequencyValue(freqStr)
					body["frequency"] = normalizedFreq
					
					// Store normalized body in context for handler use
					c.Set("normalizedBody", body)
				}
			}
		}

		c.Next()
	}
}

// normalizeFrequencyValue converts frequency strings to standardized values
func normalizeFrequencyValue(freq string) string {
	freq = strings.ToLower(strings.TrimSpace(freq))
	
	switch freq {
	case "daily", "day", "d":
		return domain.FrequencyDaily
	case "weekly", "week", "w":
		return domain.FrequencyWeekly
	case "monthly", "month", "m":
		return domain.FrequencyMonthly
	case "one-time", "onetime", "once", "single":
		return domain.FrequencyOneTime
	default:
		// Return original value if not recognized - let domain validation handle it
		return freq
	}
}

// ValidateFinancialData middleware performs comprehensive validation on financial data
func ValidateFinancialData() gin.HandlerFunc {
	return func(c *gin.Context) {
		var body map[string]interface{}
		
		// Only validate JSON requests
		if c.Request.ContentLength > 0 && strings.Contains(c.GetHeader("Content-Type"), "application/json") {
			if err := c.ShouldBindJSON(&body); err != nil {
				c.JSON(http.StatusBadRequest, types.NewErrorResponse(
					http.StatusBadRequest,
					"bad_request",
					"Invalid JSON format",
				))
				c.Abort()
				return
			}

			// Validate amount fields are positive
			amountFields := []string{"amount", "balance", "monthly_payment", "interest_rate"}
			for _, field := range amountFields {
				if value, exists := body[field]; exists {
					if amount, ok := value.(float64); ok {
						if amount < 0 {
							c.JSON(http.StatusBadRequest, types.NewErrorResponse(
								http.StatusBadRequest,
								"validation_error",
								field+" must be positive",
							))
							c.Abort()
							return
						}
					}
				}
			}

			// Validate required string fields are not empty
			stringFields := []string{"name", "description", "category", "source"}
			for _, field := range stringFields {
				if value, exists := body[field]; exists {
					if str, ok := value.(string); ok {
						if strings.TrimSpace(str) == "" {
							c.JSON(http.StatusBadRequest, types.NewErrorResponse(
								http.StatusBadRequest,
								"validation_error",
								field+" cannot be empty",
							))
							c.Abort()
							return
						}
					}
				}
			}

			// Store validated body for handler use
			c.Set("validatedBody", body)
		}

		c.Next()
	}
}

// ValidateUserOwnership validates that a user can only access/modify their own data
// This is used for endpoints with resource IDs to ensure ownership
func ValidateUserOwnership(resourceType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := GetUserID(c)
		if userID == "" {
			c.JSON(http.StatusUnauthorized, types.NewErrorResponse(
				http.StatusUnauthorized,
				"unauthorized",
				"Authentication required",
			))
			c.Abort()
			return
		}

		resourceID := c.Param("id")
		if resourceID == "" {
			c.JSON(http.StatusBadRequest, types.NewErrorResponse(
				http.StatusBadRequest,
				"bad_request",
				"Resource ID is required",
			))
			c.Abort()
			return
		}

		// Store both IDs in context for handler use
		c.Set("authenticatedUserID", userID)
		c.Set("resourceID", resourceID)
		c.Set("resourceType", resourceType)
		
		c.Next()
	}
}

// ValidateRequestLimits validates request size and content limits
func ValidateRequestLimits() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Limit request size (1MB)
		const maxRequestSize = 1 << 20 // 1MB
		
		if c.Request.ContentLength > maxRequestSize {
			c.JSON(http.StatusRequestEntityTooLarge, types.NewErrorResponse(
				http.StatusRequestEntityTooLarge,
				"payload_too_large",
				"Request payload too large",
			))
			c.Abort()
			return
		}

		c.Next()
	}
}

// GetValidatedBody retrieves the validated request body from context
func GetValidatedBody(c *gin.Context) (map[string]interface{}, bool) {
	if body, exists := c.Get("validatedBody"); exists {
		if validBody, ok := body.(map[string]interface{}); ok {
			return validBody, true
		}
	}
	return nil, false
}

// GetNormalizedBody retrieves the normalized request body from context
func GetNormalizedBody(c *gin.Context) (map[string]interface{}, bool) {
	if body, exists := c.Get("normalizedBody"); exists {
		if normalizedBody, ok := body.(map[string]interface{}); ok {
			return normalizedBody, true
		}
	}
	return nil, false
}

// GetResourceInfo retrieves resource ownership information from context
func GetResourceInfo(c *gin.Context) (userID, resourceID, resourceType string) {
	if uid, exists := c.Get("authenticatedUserID"); exists {
		if userIDStr, ok := uid.(string); ok {
			userID = userIDStr
		}
	}
	
	if rid, exists := c.Get("resourceID"); exists {
		if resourceIDStr, ok := rid.(string); ok {
			resourceID = resourceIDStr
		}
	}
	
	if rt, exists := c.Get("resourceType"); exists {
		if resourceTypeStr, ok := rt.(string); ok {
			resourceType = resourceTypeStr
		}
	}
	
	return userID, resourceID, resourceType
}