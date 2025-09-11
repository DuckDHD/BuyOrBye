package config

import (
	"fmt"
	"strings"
)

// LoggingService provides logging configuration management
type LoggingService interface {
	// GetLevel returns the configured log level
	GetLevel() string
	
	// GetEnvironment returns the logging environment (for formatter selection)
	GetEnvironment() string
	
	// ShouldLogLevel returns true if the given level should be logged
	ShouldLogLevel(level string) bool
	
	// IsStructuredLogging returns true if structured logging should be used
	IsStructuredLogging() bool
	
	// IsJSONLogging returns true if JSON format should be used
	IsJSONLogging() bool
}

// NewLoggingService creates a new logging service from configuration
func NewLoggingService(config *LoggingConfig) LoggingService {
	return &loggingService{
		config: config,
	}
}

type loggingService struct {
	config *LoggingConfig
}

func (l *loggingService) GetLevel() string {
	return l.config.Level
}

func (l *loggingService) GetEnvironment() string {
	return l.config.Environment
}

func (l *loggingService) ShouldLogLevel(level string) bool {
	levelPriority := getLogLevelPriority(level)
	configPriority := getLogLevelPriority(l.config.Level)
	return levelPriority >= configPriority
}

func (l *loggingService) IsStructuredLogging() bool {
	// Always use structured logging
	return true
}

func (l *loggingService) IsJSONLogging() bool {
	// Use JSON logging in production
	return l.config.Environment == "production"
}

// getLogLevelPriority returns numeric priority for log levels
func getLogLevelPriority(level string) int {
	switch strings.ToLower(level) {
	case "debug":
		return 0
	case "info":
		return 1
	case "warn", "warning":
		return 2
	case "error":
		return 3
	case "fatal":
		return 4
	default:
		return 1 // Default to info level
	}
}

// ValidateLoggingConfig validates logging configuration
func ValidateLoggingConfig(config *LoggingConfig) error {
	if config.Level == "" {
		return fmt.Errorf("log level cannot be empty")
	}
	
	validLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
		"fatal": true,
	}
	
	if !validLevels[strings.ToLower(config.Level)] {
		return fmt.Errorf("invalid log level: %s (must be one of: debug, info, warn, error, fatal)", config.Level)
	}
	
	if config.Environment == "" {
		return fmt.Errorf("logging environment cannot be empty")
	}
	
	validEnvironments := map[string]bool{
		"development": true,
		"production":  true,
		"test":       true,
	}
	
	if !validEnvironments[config.Environment] {
		return fmt.Errorf("invalid logging environment: %s (must be one of: development, production, test)", config.Environment)
	}
	
	return nil
}

// GetLogLevelForEnvironment returns appropriate log level for environment
func GetLogLevelForEnvironment(environment string) string {
	switch environment {
	case "production":
		return "info"
	case "development":
		return "debug"
	case "test":
		return "warn"
	default:
		return "info"
	}
}

// GetLoggingFormatForEnvironment returns appropriate logging format for environment
func GetLoggingFormatForEnvironment(environment string) string {
	switch environment {
	case "production":
		return "json"
	case "development":
		return "console"
	case "test":
		return "minimal"
	default:
		return "console"
	}
}

// LoggingMiddlewareConfig returns middleware configuration for logging
type LoggingMiddlewareConfig struct {
	SkipPaths       []string
	LogRequestBody  bool
	LogResponseBody bool
	MaxBodySize     int64
}

// GetMiddlewareConfig returns logging middleware configuration for environment
func GetMiddlewareConfig(environment string) LoggingMiddlewareConfig {
	switch environment {
	case "production":
		return LoggingMiddlewareConfig{
			SkipPaths: []string{
				"/health",
				"/healthz", 
				"/ping",
				"/metrics",
				"/favicon.ico",
			},
			LogRequestBody:  false, // Disabled for security
			LogResponseBody: false, // Disabled for performance
			MaxBodySize:     512,   // 512 bytes limit
		}
	case "development":
		return LoggingMiddlewareConfig{
			SkipPaths: []string{
				"/health",
				"/healthz",
				"/ping",
			},
			LogRequestBody:  true,  // Enabled for debugging
			LogResponseBody: true,  // Enabled for debugging
			MaxBodySize:     2048,  // 2KB limit
		}
	case "test":
		return LoggingMiddlewareConfig{
			SkipPaths: []string{
				"/health",
				"/healthz",
				"/ping", 
				"/metrics",
			},
			LogRequestBody:  false, // Disabled for test speed
			LogResponseBody: false, // Disabled for test speed
			MaxBodySize:     256,   // 256 bytes limit
		}
	default:
		return GetMiddlewareConfig("development")
	}
}

// IsDebugMode returns true if debug logging is enabled
func IsDebugMode(config *LoggingConfig) bool {
	return strings.ToLower(config.Level) == "debug"
}

// IsProductionLogging returns true if production logging is configured
func IsProductionLogging(config *LoggingConfig) bool {
	return config.Environment == "production"
}

// GetComponentLogger returns logger configuration for a specific component
type ComponentLoggerConfig struct {
	Component string
	Level     string
	Enabled   bool
}

// GetComponentConfigs returns logging configuration for different components
func GetComponentConfigs(environment string, baseLevel string) map[string]ComponentLoggerConfig {
	configs := map[string]ComponentLoggerConfig{
		"handler": {
			Component: "handler",
			Level:     baseLevel,
			Enabled:   true,
		},
		"service": {
			Component: "service", 
			Level:     baseLevel,
			Enabled:   true,
		},
		"repository": {
			Component: "repository",
			Level:     baseLevel,
			Enabled:   true,
		},
		"middleware": {
			Component: "middleware",
			Level:     baseLevel,
			Enabled:   true,
		},
		"database": {
			Component: "database",
			Level:     baseLevel,
			Enabled:   true,
		},
	}
	
	// Adjust for different environments
	switch environment {
	case "test":
		// Reduce database logging verbosity in tests
		if dbConfig, exists := configs["database"]; exists {
			if baseLevel == "debug" {
				dbConfig.Level = "info"
			}
			configs["database"] = dbConfig
		}
	case "production":
		// All components use the base level in production
		// No special adjustments needed
	}
	
	return configs
}