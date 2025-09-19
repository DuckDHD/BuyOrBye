package config

import (
	"os"
	"strconv"
)

// getEnvWithDefault gets environment variable with a default value (for backward compatibility)
func getEnvWithDefault(key, defaultValue string) string {
	return getEnv(key, defaultValue)
}

// getEnvAsInt gets environment variable as integer with default (for backward compatibility)
func getEnvAsInt(key string, defaultValue int) int {
	return getEnvInt(key, defaultValue)
}

// getEnvAsBool gets environment variable as boolean with default (for backward compatibility)
func getEnvAsBool(key string, defaultValue bool) bool {
	return getEnvBool(key, defaultValue)
}

// setEnvIfEmpty sets environment variable only if it's empty
func setEnvIfEmpty(key, value string) {
	if os.Getenv(key) == "" {
		os.Setenv(key, value)
	}
}

// validateRequiredEnv validates that required environment variables are set
func validateRequiredEnv(keys []string) []string {
	var missing []string
	for _, key := range keys {
		if os.Getenv(key) == "" {
			missing = append(missing, key)
		}
	}
	return missing
}

// ConfigSummary provides a summary of configuration for logging/debugging
type ConfigSummary struct {
	Environment      string            `json:"environment"`
	ServerPort       int               `json:"server_port"`
	DatabaseType     string            `json:"database_type"`
	DatabaseHost     string            `json:"database_host"`
	ConfigSource     string            `json:"config_source"`
	ValidationIssues []string          `json:"validation_issues,omitempty"`
	SecurityChecks   map[string]string `json:"security_checks,omitempty"`
}

// GetConfigSummary returns a summary of the current configuration
func GetConfigSummary(config *Config) ConfigSummary {
	summary := ConfigSummary{
		Environment:      config.App.Environment,
		ServerPort:       config.Server.Port,
		DatabaseType:     config.Database.Driver,
		DatabaseHost:     config.Database.Host,
		ConfigSource:     "environment-variables",
		ValidationIssues: []string{},
		SecurityChecks:   make(map[string]string),
	}

	// Run security checks
	summary.SecurityChecks["jwt_secret_length"] = strconv.Itoa(len(config.Auth.JWTSecret))
	summary.SecurityChecks["csrf_secret_length"] = strconv.Itoa(len(config.Security.CSRFSecret))
	summary.SecurityChecks["bcrypt_cost"] = strconv.Itoa(config.Auth.BCryptCost)

	if len(config.Auth.JWTSecret) >= 32 {
		summary.SecurityChecks["jwt_secret_secure"] = "true"
	} else {
		summary.SecurityChecks["jwt_secret_secure"] = "false"
		summary.ValidationIssues = append(summary.ValidationIssues, "JWT secret should be at least 32 characters")
	}

	if len(config.Security.CSRFSecret) >= 32 {
		summary.SecurityChecks["csrf_secret_secure"] = "true"
	} else {
		summary.SecurityChecks["csrf_secret_secure"] = "false"
		summary.ValidationIssues = append(summary.ValidationIssues, "CSRF secret should be at least 32 characters")
	}

	// Production-specific checks
	if config.App.Environment == "production" {
		if config.Auth.BCryptCost < 12 {
			summary.ValidationIssues = append(summary.ValidationIssues, "BCrypt cost should be at least 12 for production")
		}

		if len(config.Auth.JWTSecret) < 64 {
			summary.ValidationIssues = append(summary.ValidationIssues, "JWT secret should be at least 64 characters for production")
		}

		if len(config.Security.CSRFSecret) < 64 {
			summary.ValidationIssues = append(summary.ValidationIssues, "CSRF secret should be at least 64 characters for production")
		}
	}

	return summary
}

// PrintConfigSummary prints a human-readable configuration summary
func PrintConfigSummary(config *Config) {
	summary := GetConfigSummary(config)

	// This would be implemented to print a nice summary
	// For now, we'll keep it simple since we're focused on the core functionality
	_ = summary
}

// MergeConfigs merges configuration from multiple sources (useful for testing)
func MergeConfigs(base, override *Config) *Config {
	result := *base // Copy base config

	if override == nil {
		return &result
	}

	// Merge server config
	if override.Server.Port != 0 {
		result.Server.Port = override.Server.Port
	}

	// Merge app config
	if override.App.Environment != "" {
		result.App.Environment = override.App.Environment
	}

	// Merge database config
	if override.Database.Host != "" {
		result.Database.Host = override.Database.Host
	}
	if override.Database.Port != 0 {
		result.Database.Port = override.Database.Port
	}
	if override.Database.Database != "" {
		result.Database.Database = override.Database.Database
	}
	if override.Database.Username != "" {
		result.Database.Username = override.Database.Username
	}
	if override.Database.Password != "" {
		result.Database.Password = override.Database.Password
	}

	// Merge auth config
	if override.Auth.JWTSecret != "" {
		result.Auth.JWTSecret = override.Auth.JWTSecret
	}
	if override.Auth.BCryptCost != 0 {
		result.Auth.BCryptCost = override.Auth.BCryptCost
	}

	// Merge security config
	if override.Security.CSRFSecret != "" {
		result.Security.CSRFSecret = override.Security.CSRFSecret
	}

	// Merge finance config
	if override.Finance.HealthyDTIRatio != 0 {
		result.Finance.HealthyDTIRatio = override.Finance.HealthyDTIRatio
	}
	if override.Finance.MinSavingsRate != 0 {
		result.Finance.MinSavingsRate = override.Finance.MinSavingsRate
	}
	if override.Finance.EmergencyFundMonths != 0 {
		result.Finance.EmergencyFundMonths = override.Finance.EmergencyFundMonths
	}

	// Merge health config
	if override.Health.RiskThresholdLow != 0 {
		result.Health.RiskThresholdLow = override.Health.RiskThresholdLow
	}
	if override.Health.RiskThresholdModerate != 0 {
		result.Health.RiskThresholdModerate = override.Health.RiskThresholdModerate
	}
	if override.Health.RiskThresholdHigh != 0 {
		result.Health.RiskThresholdHigh = override.Health.RiskThresholdHigh
	}

	// Merge OpenAI config
	if override.OpenAI.APIKey != "" {
		result.OpenAI.APIKey = override.OpenAI.APIKey
	}

	return &result
}