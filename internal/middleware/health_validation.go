package middleware

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ValidateHealthOwnership ensures users can only access their own health data
func ValidateHealthOwnership() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from JWT context
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
			c.Abort()
			return
		}

		userIDStr, ok := userID.(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid user ID format"})
			c.Abort()
			return
		}

		// For JSON requests, validate user_id in body matches authenticated user
		if c.Request.Method == "POST" || c.Request.Method == "PUT" {
			var requestBody map[string]interface{}
			if err := c.ShouldBindJSON(&requestBody); err == nil {
				if bodyUserID, exists := requestBody["user_id"]; exists {
					if bodyUserIDStr, ok := bodyUserID.(string); ok {
						if bodyUserIDStr != userIDStr {
							c.JSON(http.StatusForbidden, gin.H{
								"error": "Cannot perform action on another user's health data",
							})
							c.Abort()
							return
						}
					}
				}
				
				// Re-bind the JSON for the handler to use
				c.Set("validated_request_body", requestBody)
			}
		}

		// Store validated user ID for handler use
		c.Set("validated_user_id", userIDStr)
		c.Next()
	}
}

// SanitizeSensitiveData removes sensitive health information from logs
func SanitizeSensitiveData() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Store original path for logging
		originalPath := c.Request.URL.Path
		
		// Process the request
		c.Next()

		// Get logger from context or create new one
		logger := GetLoggerFromContext(c)
		if logger == nil {
			logger = zap.L()
		}

		// Sanitize logged data based on health endpoints
		if strings.Contains(originalPath, "/health/") {
			// Remove sensitive fields from logging context
			sanitizedFields := []zap.Field{
				zap.String("path", originalPath),
				zap.Int("status", c.Writer.Status()),
				zap.String("method", c.Request.Method),
			}

			// Add user ID for audit purposes (but not sensitive health details)
			if userID, exists := c.Get("validated_user_id"); exists {
				sanitizedFields = append(sanitizedFields, zap.String("user_id", userID.(string)))
			}

			// Log with sanitized fields only
			if c.Writer.Status() >= 400 {
				logger.Warn("Health endpoint access", sanitizedFields...)
			} else {
				logger.Info("Health endpoint accessed", sanitizedFields...)
			}
		}
	}
}

// ValidateInsuranceDates ensures policy dates are logical
func ValidateInsuranceDates() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only validate for insurance policy endpoints
		if !strings.Contains(c.Request.URL.Path, "/insurance") {
			c.Next()
			return
		}

		if c.Request.Method == "POST" || c.Request.Method == "PUT" {
			var requestBody map[string]interface{}
			if err := c.ShouldBindJSON(&requestBody); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
				c.Abort()
				return
			}

			// Validate insurance policy dates
			if err := validatePolicyDates(requestBody); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				c.Abort()
				return
			}

			// Validate coverage percentage
			if err := validateCoveragePercentage(requestBody); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				c.Abort()
				return
			}

			// Validate deductible and out-of-pocket maximums
			if err := validateInsuranceLimits(requestBody); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				c.Abort()
				return
			}

			// Re-bind the validated JSON for the handler
			c.Set("validated_request_body", requestBody)
		}

		c.Next()
	}
}

// validatePolicyDates checks that start and end dates are logical
func validatePolicyDates(requestBody map[string]interface{}) error {
	startDateStr, hasStart := requestBody["start_date"]
	endDateStr, hasEnd := requestBody["end_date"]

	if !hasStart || !hasEnd {
		return nil // Optional for updates
	}

	startDateVal, startOk := startDateStr.(string)
	endDateVal, endOk := endDateStr.(string)

	if !startOk || !endOk {
		return nil // Will be caught by DTO validation
	}

	startDate, err := time.Parse(time.RFC3339, startDateVal)
	if err != nil {
		return nil // Will be caught by DTO validation
	}

	endDate, err := time.Parse(time.RFC3339, endDateVal)
	if err != nil {
		return nil // Will be caught by DTO validation
	}

	// Validate date logic
	if endDate.Before(startDate) {
		return errors.New("policy end date cannot be before start date")
	}

	// Validate reasonable policy duration (not more than 10 years)
	maxDuration := time.Hour * 24 * 365 * 10 // 10 years
	if endDate.Sub(startDate) > maxDuration {
		return errors.New("policy duration cannot exceed 10 years")
	}

	// Validate policy doesn't start too far in the past (more than 5 years)
	fiveYearsAgo := time.Now().AddDate(-5, 0, 0)
	if startDate.Before(fiveYearsAgo) {
		return errors.New("policy start date cannot be more than 5 years in the past")
	}

	return nil
}

// validateCoveragePercentage ensures coverage percentage is reasonable
func validateCoveragePercentage(requestBody map[string]interface{}) error {
	coverageVal, exists := requestBody["coverage_percentage"]
	if !exists {
		return nil // Optional for updates
	}

	coverage, ok := coverageVal.(float64)
	if !ok {
		return nil // Will be caught by DTO validation
	}

	if coverage < 0 || coverage > 100 {
		return errors.New("coverage percentage must be between 0 and 100")
	}

	// Warn about unusually low coverage (less than 50%)
	if coverage < 50 {
		// Log warning but don't block
		zap.L().Warn("Low insurance coverage percentage detected",
			zap.Float64("coverage", coverage))
	}

	return nil
}

// validateInsuranceLimits ensures deductible and out-of-pocket limits are reasonable
func validateInsuranceLimits(requestBody map[string]interface{}) error {
	deductibleVal, hasDeductible := requestBody["deductible"]
	outOfPocketVal, hasOutOfPocket := requestBody["out_of_pocket_max"]

	if hasDeductible && hasOutOfPocket {
		deductible, deductibleOk := deductibleVal.(float64)
		outOfPocketMax, outOfPocketOk := outOfPocketVal.(float64)

		if deductibleOk && outOfPocketOk {
			// Deductible should not exceed out-of-pocket maximum
			if deductible > outOfPocketMax {
				return errors.New("deductible cannot exceed out-of-pocket maximum")
			}

			// Validate reasonable limits (not more than $50,000 annually)
			maxReasonableAmount := 50000.0
			if deductible > maxReasonableAmount {
				return errors.New("deductible amount seems unreasonably high")
			}
			if outOfPocketMax > maxReasonableAmount {
				return errors.New("out-of-pocket maximum seems unreasonably high")
			}
		}
	}

	// Validate monthly premium is reasonable
	premiumVal, hasPremium := requestBody["monthly_premium"]
	if hasPremium {
		premium, ok := premiumVal.(float64)
		if ok {
			if premium < 0 {
				return errors.New("monthly premium cannot be negative")
			}
			if premium > 2000 { // Reasonable upper limit for individual coverage
				zap.L().Warn("High monthly premium detected",
					zap.Float64("premium", premium))
			}
		}
	}

	return nil
}

// ValidateHealthProfileData validates health profile specific constraints
func ValidateHealthProfileData() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only validate for profile endpoints
		if !strings.Contains(c.Request.URL.Path, "/profile") {
			c.Next()
			return
		}

		if c.Request.Method == "POST" || c.Request.Method == "PUT" {
			var requestBody map[string]interface{}
			if err := c.ShouldBindJSON(&requestBody); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
				c.Abort()
				return
			}

			// Validate family size is reasonable
			if familySizeVal, exists := requestBody["family_size"]; exists {
				if familySize, ok := familySizeVal.(float64); ok {
					if familySize < 1 || familySize > 20 { // MAX_FAMILY_SIZE from env
						c.JSON(http.StatusBadRequest, gin.H{
							"error": "Family size must be between 1 and 20",
						})
						c.Abort()
						return
					}
				}
			}

			// Validate BMI if height and weight are provided
			if err := validateBMIConsistency(requestBody); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				c.Abort()
				return
			}

			// Validate emergency fund amount is reasonable
			if fundVal, exists := requestBody["emergency_fund_health"]; exists {
				if fund, ok := fundVal.(float64); ok {
					if fund < 0 {
						c.JSON(http.StatusBadRequest, gin.H{
							"error": "Emergency fund cannot be negative",
						})
						c.Abort()
						return
					}
					if fund > 1000000 { // $1M reasonable upper limit
						zap.L().Warn("Very high emergency fund reported",
							zap.Float64("fund", fund))
					}
				}
			}

			c.Set("validated_request_body", requestBody)
		}

		c.Next()
	}
}

// validateBMIConsistency checks if height and weight result in reasonable BMI
func validateBMIConsistency(requestBody map[string]interface{}) error {
	heightVal, hasHeight := requestBody["height"]
	weightVal, hasWeight := requestBody["weight"]

	if hasHeight && hasWeight {
		height, heightOk := heightVal.(float64)
		weight, weightOk := weightVal.(float64)

		if heightOk && weightOk && height > 0 {
			// Calculate BMI (weight in kg, height in cm)
			heightInMeters := height / 100
			bmi := weight / (heightInMeters * heightInMeters)

			// Validate BMI is in reasonable range (10-100)
			if bmi < 10 || bmi > 100 {
				return errors.New("height and weight combination results in unrealistic BMI")
			}

			// Log extreme BMI values for review
			if bmi < 15 || bmi > 50 {
				zap.L().Warn("Extreme BMI value detected",
					zap.Float64("bmi", bmi),
					zap.Float64("height", height),
					zap.Float64("weight", weight))
			}
		}
	}

	return nil
}

// GetLoggerFromContext retrieves logger from gin context
func GetLoggerFromContext(c *gin.Context) *zap.Logger {
	if logger, exists := c.Get("logger"); exists {
		if zapLogger, ok := logger.(*zap.Logger); ok {
			return zapLogger
		}
	}
	return nil
}

// Helper function to format error messages
func formatValidationError(message string) error {
	return errors.New(message)
}

// ValidateExpenseData validates medical expense specific constraints
func ValidateExpenseData() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only validate for expense endpoints
		if !strings.Contains(c.Request.URL.Path, "/expense") {
			c.Next()
			return
		}

		if c.Request.Method == "POST" || c.Request.Method == "PUT" {
			var requestBody map[string]interface{}
			if err := c.ShouldBindJSON(&requestBody); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
				c.Abort()
				return
			}

			// Validate expense amounts are consistent
			if err := validateExpenseAmounts(requestBody); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				c.Abort()
				return
			}

			// Validate expense date is not in future
			if err := validateExpenseDate(requestBody); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				c.Abort()
				return
			}

			c.Set("validated_request_body", requestBody)
		}

		c.Next()
	}
}

// validateExpenseAmounts ensures insurance payment doesn't exceed total amount
func validateExpenseAmounts(requestBody map[string]interface{}) error {
	amountVal, hasAmount := requestBody["amount"]
	insuranceVal, hasInsurance := requestBody["insurance_payment"]
	outOfPocketVal, hasOutOfPocket := requestBody["out_of_pocket"]

	if hasAmount && hasInsurance {
		amount, amountOk := amountVal.(float64)
		insurance, insuranceOk := insuranceVal.(float64)

		if amountOk && insuranceOk {
			if insurance > amount {
				return errors.New("insurance payment cannot exceed total expense amount")
			}
			if insurance < 0 {
				return errors.New("insurance payment cannot be negative")
			}
		}
	}

	if hasAmount && hasOutOfPocket {
		amount, amountOk := amountVal.(float64)
		outOfPocket, outOfPocketOk := outOfPocketVal.(float64)

		if amountOk && outOfPocketOk {
			if outOfPocket > amount {
				return errors.New("out-of-pocket amount cannot exceed total expense amount")
			}
			if outOfPocket < 0 {
				return errors.New("out-of-pocket amount cannot be negative")
			}
		}
	}

	return nil
}

// validateExpenseDate ensures expense date is not in the future
func validateExpenseDate(requestBody map[string]interface{}) error {
	dateVal, hasDate := requestBody["date"]
	if !hasDate {
		return nil
	}

	dateStr, ok := dateVal.(string)
	if !ok {
		return nil // Will be caught by DTO validation
	}

	expenseDate, err := time.Parse(time.RFC3339, dateStr)
	if err != nil {
		return nil // Will be caught by DTO validation
	}

	if expenseDate.After(time.Now()) {
		return errors.New("expense date cannot be in the future")
	}

	// Warn about very old expenses (more than 5 years)
	fiveYearsAgo := time.Now().AddDate(-5, 0, 0)
	if expenseDate.Before(fiveYearsAgo) {
		zap.L().Warn("Very old expense date detected",
			zap.Time("expense_date", expenseDate))
	}

	return nil
}