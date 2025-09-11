package main

import (
	"context"
	"fmt"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
	
	"github.com/DuckDHD/BuyOrBye/internal/config"
	"github.com/DuckDHD/BuyOrBye/internal/logging"
	"github.com/DuckDHD/BuyOrBye/internal/server"
)

func gracefulShutdown(apiServer *http.Server, done chan bool) {
	// Create context that listens for the interrupt signal from the OS.
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Listen for the interrupt signal.
	<-ctx.Done()

	logger := logging.GetLogger()
	logger.Info("Shutting down gracefully, press Ctrl+C again to force")
	stop() // Allow Ctrl+C to force shutdown

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := apiServer.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", logging.WithError(err))
	}

	logger.Info("Server exiting")

	// Notify the main goroutine that the shutdown is complete
	done <- true
}

func main() {
	// Load configuration first
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(fmt.Sprintf("Failed to load configuration: %v", err))
	}

	// Initialize logger with config
	if err := logging.InitLogger(logging.LogConfig{
		Environment: cfg.Logging.Environment,
		Level:       cfg.Logging.Level,
	}); err != nil {
		panic(fmt.Sprintf("Failed to initialize logger: %v", err))
	}
	logger := logging.GetLogger()
	logger.Info("Configuration loaded successfully", 
		logging.WithComponent("main"),
		zap.String("environment", cfg.Server.Environment),
		zap.String("config_file", config.GetConfigPath(cfg.Server.Environment)))

	// Initialize the server with config
	apiServer, err := server.NewServerWithConfig(cfg)
	if err != nil {
		logger.Fatal("Failed to initialize server", logging.WithError(err))
	}
	
	serverService := config.NewServerService(&cfg.Server)
	logger.Info("Starting BuyOrBye API server", 
		logging.WithComponent("main"), 
		zap.String("address", serverService.GetAddress()),
		zap.String("environment", cfg.Server.Environment))

	// Create a done channel to signal when the shutdown is complete
	done := make(chan bool, 1)

	// Run graceful shutdown in a separate goroutine
	go gracefulShutdown(apiServer, done)

	// Start the server
	err = apiServer.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		panic(fmt.Sprintf("http server error: %s", err))
	}

	// Wait for the graceful shutdown to complete
	<-done
	logger.Info("Graceful shutdown complete")
}
