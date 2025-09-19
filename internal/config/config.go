package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config represents the complete application configuration
type Config struct {
	Server   ServerConfig   `json:"server"`
	Database DatabaseConfig `json:"database"`
	Auth     AuthConfig     `json:"auth"`
	Security SecurityConfig `json:"security"`
	Finance  FinanceConfig  `json:"finance"`
	Health   HealthConfig   `json:"health"`
	App      AppConfig      `json:"app"`
	OpenAI   OpenAIConfig   `json:"openai"`
}

// ServerConfig holds server-related configuration
type ServerConfig struct {
	Port int `json:"port"`
}

// DatabaseConfig holds database-related configuration
type DatabaseConfig struct {
	Host            string        `json:"host"`
	Port            int           `json:"port"`
	Database        string        `json:"database"`
	Username        string        `json:"username"`
	Password        string        `json:"password"`
	RootPassword    string        `json:"root_password"`
	Driver          string        `json:"driver"`
	MaxOpenConns    int           `json:"max_open_conns"`
	MaxIdleConns    int           `json:"max_idle_conns"`
	ConnMaxLifetime time.Duration `json:"conn_max_lifetime"`
}

// AuthConfig holds authentication-related configuration
type AuthConfig struct {
	JWTSecret       string        `json:"jwt_secret"`
	BCryptCost      int           `json:"bcrypt_cost"`
	AccessTokenTTL  time.Duration `json:"access_token_ttl"`
	RefreshTokenTTL time.Duration `json:"refresh_token_ttl"`
}

// SecurityConfig holds security-related configuration
type SecurityConfig struct {
	CSRFSecret string `json:"csrf_secret"`
}

// FinanceConfig holds finance-related configuration
type FinanceConfig struct {
	HealthyDTIRatio     float64 `json:"healthy_dti_ratio"`
	MinSavingsRate      float64 `json:"min_savings_rate"`
	EmergencyFundMonths int     `json:"emergency_fund_months"`
}

// HealthConfig holds health-related configuration
type HealthConfig struct {
	RiskThresholdLow              int           `json:"risk_threshold_low"`
	RiskThresholdModerate         int           `json:"risk_threshold_moderate"`
	RiskThresholdHigh             int           `json:"risk_threshold_high"`
	EmergencyFundBaseMonths       int           `json:"emergency_fund_base_months"`
	MaxFamilySize                 int           `json:"max_family_size"`
	DataEncryptionKey             string        `json:"data_encryption_key"`
	MedicalRecordRetentionDays    int           `json:"medical_record_retention_days"`
	InsuranceVerificationTimeout  time.Duration `json:"insurance_verification_timeout"`
	AuditLogEnabled               bool          `json:"audit_log_enabled"`
}

// AppConfig holds application-related configuration
type AppConfig struct {
	Environment string `json:"environment"`
}

// OpenAIConfig holds OpenAI-related configuration
type OpenAIConfig struct {
	APIKey string `json:"api_key"`
}

// LoggingConfig holds logging configuration (for compatibility)
type LoggingConfig struct {
	Level       string `json:"level"`
	Environment string `json:"environment"`
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	// Load .env file if it exists (ignore errors if file doesn't exist)
	godotenv.Load()

	config := &Config{
		Server: ServerConfig{
			Port: getEnvInt("PORT", 8080),
		},
		Database: DatabaseConfig{
			Host:            getEnv("BLUEPRINT_DB_HOST", "mysql_bp"),
			Port:            getEnvInt("BLUEPRINT_DB_PORT", 3306),
			Database:        getEnv("BLUEPRINT_DB_DATABASE", "blueprint"),
			Username:        getEnv("BLUEPRINT_DB_USERNAME", "melkey"),
			Password:        getEnv("BLUEPRINT_DB_PASSWORD", "password1234"),
			RootPassword:    getEnv("BLUEPRINT_DB_ROOT_PASSWORD", "password4321"),
			Driver:          "mysql", // Fixed as MySQL only
			MaxOpenConns:    getEnvInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    getEnvInt("DB_MAX_IDLE_CONNS", 5),
			ConnMaxLifetime: getEnvDuration("DB_CONN_MAX_LIFETIME", "300s"),
		},
		Auth: AuthConfig{
			JWTSecret:       getEnv("JWT_SECRET", "your-very-secure-32-character-jwt-secret-key-here-2024-buyorbye"),
			BCryptCost:      getEnvInt("BCRYPT_COST", 14),
			AccessTokenTTL:  getEnvDuration("ACCESS_TOKEN_TTL", "15m"),
			RefreshTokenTTL: getEnvDuration("REFRESH_TOKEN_TTL", "168h"),
		},
		Security: SecurityConfig{
			CSRFSecret: getEnv("CSRF_SECRET", "your-very-secure-32-character-csrf-secret-key-here-2024-secure"),
		},
		Finance: FinanceConfig{
			HealthyDTIRatio:     getEnvFloat("HEALTHY_DTI_RATIO", 0.36),
			MinSavingsRate:      getEnvFloat("MIN_SAVINGS_RATE", 0.20),
			EmergencyFundMonths: getEnvInt("EMERGENCY_FUND_MONTHS", 6),
		},
		Health: HealthConfig{
			RiskThresholdLow:              getEnvInt("HEALTH_RISK_THRESHOLD_LOW", 25),
			RiskThresholdModerate:         getEnvInt("HEALTH_RISK_THRESHOLD_MODERATE", 50),
			RiskThresholdHigh:             getEnvInt("HEALTH_RISK_THRESHOLD_HIGH", 75),
			EmergencyFundBaseMonths:       getEnvInt("EMERGENCY_FUND_BASE_MONTHS", 6),
			MaxFamilySize:                 getEnvInt("MAX_FAMILY_SIZE", 20),
			DataEncryptionKey:             getEnv("HEALTH_DATA_ENCRYPTION_KEY", ""),
			MedicalRecordRetentionDays:    getEnvInt("MEDICAL_RECORD_RETENTION_DAYS", 2555),
			InsuranceVerificationTimeout:  getEnvDuration("INSURANCE_VERIFICATION_TIMEOUT", "30s"),
			AuditLogEnabled:               getEnvBool("HEALTH_AUDIT_LOG_ENABLED", true),
		},
		App: AppConfig{
			Environment: normalizeEnvironment(getEnv("APP_ENV", "local")),
		},
		OpenAI: OpenAIConfig{
			APIKey: getEnv("OPENAI_API_KEY", ""),
		},
	}

	return config, nil
}

// GetDSN returns the MySQL database connection string
func (d *DatabaseConfig) GetDSN() string {
	return d.Username + ":" + d.Password + "@tcp(" + d.Host + ":" + strconv.Itoa(d.Port) + ")/" + d.Database + "?charset=utf8mb4&parseTime=True&loc=Local"
}

// Helper functions
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getEnvFloat(key string, defaultValue float64) float64 {
	if value := os.Getenv(key); value != "" {
		if floatValue, err := strconv.ParseFloat(value, 64); err == nil {
			return floatValue
		}
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue string) time.Duration {
	value := getEnv(key, defaultValue)
	duration, err := time.ParseDuration(value)
	if err != nil {
		duration, _ = time.ParseDuration(defaultValue)
	}
	return duration
}

// normalizeEnvironment maps various environment names to our standard names
func normalizeEnvironment(env string) string {
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

// MustLoadConfig loads configuration and panics on error (for compatibility)
func MustLoadConfig() *Config {
	config, err := LoadConfig()
	if err != nil {
		panic("failed to load configuration: " + err.Error())
	}
	return config
}

// GetConfigPath returns a dummy path for compatibility
func GetConfigPath(env string) string {
	return "env-based-config"
}