package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

// Config represents the complete application configuration
type Config struct {
	Server   ServerConfig   `mapstructure:"server" validate:"required"`
	Database DatabaseConfig `mapstructure:"database" validate:"required"`
	Auth     AuthConfig     `mapstructure:"auth" validate:"required"`
	Logging  LoggingConfig  `mapstructure:"logging" validate:"required"`
	Finance  FinanceConfig  `mapstructure:"finance" validate:"required"`
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port         int           `mapstructure:"port" validate:"min=1,max=65535"`
	Environment  string        `mapstructure:"environment" validate:"required,oneof=development production test"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout" validate:"required"`
	WriteTimeout time.Duration `mapstructure:"write_timeout" validate:"required"`
	IdleTimeout  time.Duration `mapstructure:"idle_timeout" validate:"required"`
}

// DatabaseConfig holds database-related configuration
type DatabaseConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port" validate:"min=0,max=65535"`
	Database        string        `mapstructure:"database" validate:"required"`
	Username        string        `mapstructure:"username"`
	Password        string        `mapstructure:"password"`
	MaxIdleConns    int           `mapstructure:"max_idle_conns" validate:"min=1"`
	MaxOpenConns    int           `mapstructure:"max_open_conns" validate:"min=1"`
	ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
}

// AuthConfig holds authentication-related configuration
type AuthConfig struct {
	JWTSecret       string        `mapstructure:"jwt_secret" validate:"required,min=32"`
	BCryptCost      int           `mapstructure:"bcrypt_cost" validate:"min=4,max=20"`
	AccessTokenTTL  time.Duration `mapstructure:"access_token_ttl" validate:"required"`
	RefreshTokenTTL time.Duration `mapstructure:"refresh_token_ttl" validate:"required"`
	CSRFSecret      string        `mapstructure:"csrf_secret" validate:"required,min=32"`
}

// LoggingConfig holds logging-related configuration
type LoggingConfig struct {
	Level       string `mapstructure:"level" validate:"required,oneof=debug info warn error"`
	Environment string `mapstructure:"environment" validate:"required,oneof=development production test"`
}

// FinanceConfig holds finance-related configuration
type FinanceConfig struct {
	HealthyDTIRatio     float64 `mapstructure:"healthy_dti_ratio" validate:"min=0,max=1"`
	MinSavingsRate      float64 `mapstructure:"min_savings_rate" validate:"min=0,max=1"`
	EmergencyFundMonths int     `mapstructure:"emergency_fund_months" validate:"min=1"`
}

// LoadConfig loads configuration from files and environment variables
func LoadConfig() (*Config, error) {
	// Determine environment
	env := getEnvironment()

	// Setup Viper
	v := viper.New()

	// Set config name and paths
	v.SetConfigName(env)
	v.SetConfigType("yaml")
	v.AddConfigPath("./configs")
	v.AddConfigPath("../configs")
	v.AddConfigPath("../../configs")

	// Enable environment variable substitution
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Expand environment variables in config values with defaults
	v.Set("database.host", expandEnvWithDefault(v.GetString("database.host"), "localhost"))
	v.Set("database.port", expandEnvIntWithDefault(v.GetString("database.port"), 3306))
	v.Set("database.database", expandEnvWithDefault(v.GetString("database.database"), "blueprint"))
	v.Set("database.username", expandEnvWithDefault(v.GetString("database.username"), ""))
	v.Set("database.password", expandEnvWithDefault(v.GetString("database.password"), ""))
	v.Set("auth.jwt_secret", expandEnvWithDefault(v.GetString("auth.jwt_secret"), ""))
	v.Set("auth.csrf_secret", expandEnvWithDefault(v.GetString("auth.csrf_secret"), ""))

	// Unmarshal into config struct
	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate configuration
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}

// getEnvironment determines the current environment
func getEnvironment() string {
	// Check various environment variables
	if env := os.Getenv("GO_ENV"); env != "" {
		return normalizeEnvironment(env)
	}
	if env := os.Getenv("GIN_MODE"); env != "" {
		return normalizeEnvironment(env)
	}
	if env := os.Getenv("APP_ENV"); env != "" {
		return normalizeEnvironment(env)
	}

	// Default to development
	return "development"
}

// normalizeEnvironment maps various environment names to our standard names
func normalizeEnvironment(env string) string {
	env = strings.ToLower(env)
	switch env {
	case "prod", "production", "release":
		return "production"
	case "test", "testing":
		return "test"
	case "dev", "development", "local":
		return "development"
	default:
		return "development"
	}
}

// expandEnv expands environment variables in strings
func expandEnv(value string) string {
	if strings.HasPrefix(value, "${") && strings.HasSuffix(value, "}") {
		envVar := value[2 : len(value)-1]
		if envValue := os.Getenv(envVar); envValue != "" {
			return envValue
		}
		// Return original value if env var not found
		return value
	}
	return value
}

// expandEnvWithDefault expands environment variables with a default value
func expandEnvWithDefault(value, defaultValue string) string {
	if strings.HasPrefix(value, "${") && strings.HasSuffix(value, "}") {
		envVar := value[2 : len(value)-1]
		if envValue := os.Getenv(envVar); envValue != "" {
			return envValue
		}
		// Return default value if env var not found
		return defaultValue
	}
	return value
}

// expandEnvInt expands environment variables for integer values
func expandEnvInt(value string) interface{} {
	expanded := expandEnv(value)
	// If it's still a string after expansion, return as-is for Viper to handle
	return expanded
}

// expandEnvIntWithDefault expands environment variables for integer values with a default
func expandEnvIntWithDefault(value string, defaultValue int) int {
	if strings.HasPrefix(value, "${") && strings.HasSuffix(value, "}") {
		envVar := value[2 : len(value)-1]
		if envValue := os.Getenv(envVar); envValue != "" {
			if intValue, err := strconv.Atoi(envValue); err == nil {
				return intValue
			}
		}
		// Return default value if env var not found or can't be parsed
		return defaultValue
	}
	// Try to parse the original value as int, return default if it fails
	if intValue, err := strconv.Atoi(value); err == nil {
		return intValue
	}
	return defaultValue
}

// validateConfig validates the configuration using struct tags
func validateConfig(config *Config) error {
	validator := validator.New()
	if err := validator.Struct(config); err != nil {
		return fmt.Errorf("validation errors: %w", err)
	}
	return nil
}

// GetConfigPath returns the path to the config file being used
func GetConfigPath(env string) string {
	configPaths := []string{
		"./configs",
		"../configs",
		"../../configs",
	}

	for _, path := range configPaths {
		configFile := filepath.Join(path, env+".yaml")
		if _, err := os.Stat(configFile); err == nil {
			return configFile
		}
	}

	return filepath.Join("configs", env+".yaml")
}

// MustLoadConfig loads configuration and panics on error
func MustLoadConfig() *Config {
	config, err := LoadConfig()
	if err != nil {
		panic(fmt.Sprintf("failed to load configuration: %v", err))
	}
	return config
}

// GetDSN returns the database connection string
func (d *DatabaseConfig) GetDSN() string {
	if d.Database == ":memory:" {
		// SQLite in-memory database
		return ":memory:"
	}
	if d.Host == "" {
		// SQLite file database
		return d.Database
	}
	// MySQL connection string
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
		d.Username, d.Password, d.Host, d.Port, d.Database)
}

// IsMySQL returns true if this is a MySQL database configuration
func (d *DatabaseConfig) IsMySQL() bool {
	return d.Host != "" && d.Port > 0
}

// IsSQLite returns true if this is a SQLite database configuration
func (d *DatabaseConfig) IsSQLite() bool {
	return d.Host == "" || d.Database == ":memory:"
}
