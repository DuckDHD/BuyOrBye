package logging

import (
	"os"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	logger *zap.Logger
	once   sync.Once
)

// LogConfig holds the configuration for the logger
type LogConfig struct {
	Environment string // "production", "development", or "test"
	Level       string // "debug", "info", "warn", "error"
}

// InitLogger initializes the global logger instance
// This should be called once at application startup
func InitLogger(config LogConfig) error {
	var err error
	once.Do(func() {
		logger, err = createLogger(config)
	})
	return err
}

// GetLogger returns the global logger instance
// Returns nil if InitLogger has not been called (for graceful degradation)
func GetLogger() *zap.Logger {
	return logger
}

// MustGetLogger returns the global logger instance
// Panics if InitLogger has not been called (use for critical paths)
func MustGetLogger() *zap.Logger {
	if logger == nil {
		panic("logger not initialized - call InitLogger first")
	}
	return logger
}

// SetTestLogger sets the global logger for testing purposes
// This bypasses the sync.Once initialization and should only be used in tests
func SetTestLogger(testLogger *zap.Logger) {
	logger = testLogger
}

// Sync flushes any buffered log entries
// Should be called before application shutdown
func Sync() error {
	if logger != nil {
		return logger.Sync()
	}
	return nil
}

// createLogger creates a new zap logger based on the configuration
func createLogger(config LogConfig) (*zap.Logger, error) {
	// Determine log level
	logLevel, err := zapcore.ParseLevel(config.Level)
	if err != nil {
		logLevel = zapcore.InfoLevel // Default to info level
	}

	// Configure encoder based on environment
	var encoderConfig zapcore.EncoderConfig
	var encoder zapcore.Encoder

	switch config.Environment {
	case "production":
		// JSON encoder for production
		encoderConfig = zap.NewProductionEncoderConfig()
		encoderConfig.TimeKey = "timestamp"
		encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		encoderConfig.CallerKey = "caller"
		encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
		encoder = zapcore.NewJSONEncoder(encoderConfig)

	case "test":
		// Minimal encoder for testing
		encoderConfig = zap.NewDevelopmentEncoderConfig()
		encoderConfig.TimeKey = ""      // Disable timestamps in tests
		encoderConfig.CallerKey = ""    // Disable caller in tests
		encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
		encoder = zapcore.NewConsoleEncoder(encoderConfig)

	default: // development
		// Console encoder for development
		encoderConfig = zap.NewDevelopmentEncoderConfig()
		encoderConfig.TimeKey = "time"
		encoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoderConfig.CallerKey = "caller"
		encoderConfig.EncodeCaller = zapcore.ShortCallerEncoder
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// Configure writer (always stdout for now, could be extended to support files)
	writer := zapcore.AddSync(os.Stdout)

	// Create core
	core := zapcore.NewCore(encoder, writer, logLevel)

	// Create logger with caller information
	zapLogger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

	// Add stack traces for error level and above in development
	if config.Environment == "development" {
		zapLogger = zapLogger.WithOptions(zap.AddStacktrace(zapcore.ErrorLevel))
	}

	return zapLogger, nil
}

// LoggerFromEnvironment creates a logger configuration from environment variables
func LoggerFromEnvironment() LogConfig {
	env := os.Getenv("GO_ENV")
	if env == "" {
		env = os.Getenv("GIN_MODE")
	}
	if env == "" {
		env = "development"
	}

	// Map gin modes to our environment names
	switch env {
	case "release":
		env = "production"
	case "test":
		env = "test" 
	default:
		env = "development"
	}

	level := os.Getenv("LOG_LEVEL")
	if level == "" {
		if env == "production" {
			level = "info"
		} else {
			level = "debug"
		}
	}

	return LogConfig{
		Environment: env,
		Level:       level,
	}
}

// Helper functions for common logging patterns

// WithUserID adds user ID as a structured field
func WithUserID(userID string) zap.Field {
	return zap.String("user_id", userID)
}

// WithRequestID adds request ID as a structured field
func WithRequestID(requestID string) zap.Field {
	return zap.String("request_id", requestID)
}

// WithError adds error as a structured field
func WithError(err error) zap.Field {
	return zap.Error(err)
}

// WithDuration adds duration as a structured field
func WithDuration(field string, duration interface{}) zap.Field {
	return zap.Any(field, duration)
}

// WithHTTPStatus adds HTTP status code as a structured field
func WithHTTPStatus(status int) zap.Field {
	return zap.Int("http_status", status)
}

// WithMethod adds HTTP method as a structured field
func WithMethod(method string) zap.Field {
	return zap.String("method", method)
}

// WithPath adds HTTP path as a structured field
func WithPath(path string) zap.Field {
	return zap.String("path", path)
}

// WithIP adds client IP as a structured field
func WithIP(ip string) zap.Field {
	return zap.String("client_ip", ip)
}

// WithLatency adds request latency as a structured field
func WithLatency(latency interface{}) zap.Field {
	return zap.Any("latency", latency)
}

// WithComponent adds component name as a structured field
func WithComponent(component string) zap.Field {
	return zap.String("component", component)
}

// WithOperation adds operation name as a structured field
func WithOperation(operation string) zap.Field {
	return zap.String("operation", operation)
}

// WithEntityID adds entity ID as a structured field
func WithEntityID(entityType, entityID string) zap.Field {
	return zap.String(entityType+"_id", entityID)
}

// WithTable adds database table name as a structured field
func WithTable(table string) zap.Field {
	return zap.String("table", table)
}

// WithQuery adds database query information as a structured field
func WithQuery(query string) zap.Field {
	return zap.String("query", query)
}

// WithRowsAffected adds rows affected count as a structured field
func WithRowsAffected(count int64) zap.Field {
	return zap.Int64("rows_affected", count)
}

// Component-specific loggers with pre-configured fields

// HandlerLogger returns a logger pre-configured for handler layer
func HandlerLogger() *zap.Logger {
	if base := GetLogger(); base != nil {
		return base.With(WithComponent("handler"))
	}
	return nil
}

// ServiceLogger returns a logger pre-configured for service layer
func ServiceLogger() *zap.Logger {
	if base := GetLogger(); base != nil {
		return base.With(WithComponent("service"))
	}
	return nil
}

// RepositoryLogger returns a logger pre-configured for repository layer
func RepositoryLogger() *zap.Logger {
	if base := GetLogger(); base != nil {
		return base.With(WithComponent("repository"))
	}
	return nil
}

// MiddlewareLogger returns a logger pre-configured for middleware
func MiddlewareLogger() *zap.Logger {
	if base := GetLogger(); base != nil {
		return base.With(WithComponent("middleware"))
	}
	return nil
}

// DatabaseLogger returns a logger pre-configured for database operations
func DatabaseLogger() *zap.Logger {
	if base := GetLogger(); base != nil {
		return base.With(WithComponent("database"))
	}
	return nil
}