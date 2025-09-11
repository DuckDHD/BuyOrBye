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

// NewServerService creates a new server service from configuration
func NewServerService(config *ServerConfig) ServerService {
	return &serverService{
		config: config,
	}
}

type serverService struct {
	config *ServerConfig
}

func (s *serverService) CreateServer(handler http.Handler) *http.Server {
	return &http.Server{
		Addr:         s.GetAddress(),
		Handler:      handler,
		ReadTimeout:  s.config.ReadTimeout,
		WriteTimeout: s.config.WriteTimeout,
		IdleTimeout:  s.config.IdleTimeout,
	}
}

func (s *serverService) GetAddress() string {
	return fmt.Sprintf(":%d", s.config.Port)
}

func (s *serverService) IsProduction() bool {
	return s.config.Environment == "production"
}

func (s *serverService) IsDevelopment() bool {
	return s.config.Environment == "development"
}

func (s *serverService) IsTest() bool {
	return s.config.Environment == "test"
}

// GetPort returns the configured server port, with fallback to environment
func GetPort(config *ServerConfig) int {
	if config.Port > 0 {
		return config.Port
	}
	
	// Fallback to environment variable
	if portStr := getEnvWithDefault("PORT", "8080"); portStr != "" {
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
	
	if config.Environment == "" {
		return fmt.Errorf("environment cannot be empty")
	}
	
	validEnvironments := map[string]bool{
		"development": true,
		"production":  true,
		"test":       true,
	}
	
	if !validEnvironments[config.Environment] {
		return fmt.Errorf("invalid environment: %s (must be one of: development, production, test)", config.Environment)
	}
	
	if config.ReadTimeout <= 0 {
		return fmt.Errorf("read timeout must be positive")
	}
	
	if config.WriteTimeout <= 0 {
		return fmt.Errorf("write timeout must be positive")
	}
	
	if config.IdleTimeout <= 0 {
		return fmt.Errorf("idle timeout must be positive")
	}
	
	return nil
}