package config

import (
	"fmt"
	"net/http"
	"strconv"
	"time"
)

// ServerService provides HTTP server configuration and setup
type ServerService interface {
	// CreateServer creates an HTTP server with the configured settings
	CreateServer(handler http.Handler) *http.Server

	// GetAddress returns the server listen address
	GetAddress() string

	// IsProduction returns true if running in production environment
	IsProduction() bool

	// IsDevelopment returns true if running in development environment
	IsDevelopment() bool

	// IsTest returns true if running in test environment
	IsTest() bool
}

// NewServerService creates a new server service from full configuration
func NewServerService(config *Config) ServerService {
	return &serverService{
		serverConfig: &config.Server,
		appConfig:    &config.App,
	}
}

// NewServerServiceFromConfig creates server service from server config only (for compatibility)
func NewServerServiceFromConfig(serverConfig *ServerConfig) ServerService {
	cfg := &Config{
		Server: *serverConfig,
		App: AppConfig{
			Environment: "development", // Default fallback
		},
	}
	return NewServerService(cfg)
}

type serverService struct {
	serverConfig *ServerConfig
	appConfig    *AppConfig
}

func (s *serverService) CreateServer(handler http.Handler) *http.Server {
	readTimeout, writeTimeout, idleTimeout := GetTimeoutConfig(s.appConfig.Environment)
	return &http.Server{
		Addr:         s.GetAddress(),
		Handler:      handler,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}
}

func (s *serverService) GetAddress() string {
	return fmt.Sprintf(":%d", s.serverConfig.Port)
}

func (s *serverService) IsProduction() bool {
	return s.appConfig.Environment == "production"
}

func (s *serverService) IsDevelopment() bool {
	return s.appConfig.Environment == "development"
}

func (s *serverService) IsTest() bool {
	return s.appConfig.Environment == "test"
}

// GetPort returns the configured server port, with fallback to environment
func GetPort(config *ServerConfig) int {
	if config.Port > 0 {
		return config.Port
	}

	// Fallback to environment variable
	if portStr := getEnv("PORT", "8080"); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			return port
		}
	}

	return 8080
}

// GetTimeoutConfig returns timeout configuration with sensible defaults
func GetTimeoutConfig(environment string) (read, write, idle time.Duration) {
	switch environment {
	case "production":
		return 10 * time.Second, 30 * time.Second, 60 * time.Second
	case "test":
		return 5 * time.Second, 10 * time.Second, 30 * time.Second
	default: // development
		return 10 * time.Second, 30 * time.Second, 60 * time.Second
	}
}

// ValidateServerConfig validates server configuration
func ValidateServerConfig(config *ServerConfig) error {
	if config.Port <= 0 || config.Port > 65535 {
		return fmt.Errorf("invalid port: %d (must be between 1 and 65535)", config.Port)
	}

	return nil
}

// ValidateAppConfig validates app configuration
func ValidateAppConfig(config *AppConfig) error {
	if config.Environment == "" {
		return fmt.Errorf("environment cannot be empty")
	}

	validEnvironments := map[string]bool{
		"development": true,
		"production":  true,
		"test":        true,
	}

	if !validEnvironments[config.Environment] {
		return fmt.Errorf("invalid environment: %s (must be one of: development, production, test)", config.Environment)
	}

	return nil
}