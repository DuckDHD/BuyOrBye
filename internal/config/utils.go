package config

import (
	"os"
	"strconv"
	"strings"
)

// getEnvWithDefault gets environment variable with a default value
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvAsInt gets environment variable as integer with default
func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}

// getEnvAsBool gets environment variable as boolean with default
func getEnvAsBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return strings.ToLower(value) == "true"
	}
	return defaultValue
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
	Environment     string            `json:"environment"`
	ServerPort      int               `json:"server_port"`
	DatabaseType    string            `json:"database_type"`
	LogLevel        string            `json:"log_level"`
	ConfigFile      string            `json:"config_file"`
	ValidationIssues []string         `json:"validation_issues,omitempty"`
	SecurityChecks  map[string]string `json:"security_checks,omitempty"`
}

// GetConfigSummary returns a summary of the current configuration
func GetConfigSummary(config *Config) ConfigSummary {
	summary := ConfigSummary{
		Environment:     config.Server.Environment,
		ServerPort:      config.Server.Port,
		LogLevel:        config.Logging.Level,
		ConfigFile:      GetConfigPath(config.Server.Environment),
		ValidationIssues: []string{},
		SecurityChecks:  make(map[string]string),
	}
	
	// Determine database type
	if config.Database.IsMySQL() {
		summary.DatabaseType = "MySQL"
	} else {
		summary.DatabaseType = "SQLite"
	}
	
	// Run security checks for production
	if config.Server.Environment == "production" {
		authIssues := IsProductionReady(&config.Auth)
		if len(authIssues) > 0 {
			summary.ValidationIssues = append(summary.ValidationIssues, authIssues...)
		}
		
		// Add security status checks
		summary.SecurityChecks["jwt_secret_length"] = strconv.Itoa(len(config.Auth.JWTSecret))
		summary.SecurityChecks["csrf_secret_length"] = strconv.Itoa(len(config.Auth.CSRFSecret))
		summary.SecurityChecks["bcrypt_cost"] = strconv.Itoa(config.Auth.BCryptCost)
		
		if len(config.Auth.JWTSecret) >= 64 {
			summary.SecurityChecks["jwt_secret_secure"] = "true"
		} else {
			summary.SecurityChecks["jwt_secret_secure"] = "false"
		}
		
		if len(config.Auth.CSRFSecret) >= 64 {
			summary.SecurityChecks["csrf_secret_secure"] = "true" 
		} else {
			summary.SecurityChecks["csrf_secret_secure"] = "false"
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
	if override.Server.Environment != "" {
		result.Server.Environment = override.Server.Environment
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
	
	// Merge auth config
	if override.Auth.JWTSecret != "" {
		result.Auth.JWTSecret = override.Auth.JWTSecret
	}
	if override.Auth.BCryptCost != 0 {
		result.Auth.BCryptCost = override.Auth.BCryptCost
	}
	
	// Merge logging config  
	if override.Logging.Level != "" {
		result.Logging.Level = override.Logging.Level
	}
	if override.Logging.Environment != "" {
		result.Logging.Environment = override.Logging.Environment
	}
	
	return &result
}