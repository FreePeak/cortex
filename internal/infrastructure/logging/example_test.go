package logging_test

import (
	"context"
	"fmt"

	"github.com/FreePeak/golang-mcp-server-sdk/internal/infrastructure/logging"
)

func Example() {
	// Create a development logger (with debug level enabled)
	logger, err := logging.NewDevelopment()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	// Simple logging
	logger.Debug("Debug message")
	logger.Info("Info message")
	logger.Warn("Warning message")
	logger.Error("Error message")
	// logger.Fatal("Fatal message") // This would exit the program
	// logger.Panic("Panic message") // This would panic

	// Logging with fields
	logger.Info("User logged in", logging.Fields{
		"user_id": 123,
		"email":   "user@example.com",
	})

	// Using With to create a logger with default fields
	userLogger := logger.With(logging.Fields{
		"user_id": 123,
		"email":   "user@example.com",
	})

	userLogger.Info("User profile updated")
	userLogger.Error("Failed to update password")

	// Formatted logging (using sugar)
	logger.Infof("User %d logged in from %s", 123, "192.168.1.1")
	logger.Errorf("Failed to process payment: %v", fmt.Errorf("insufficient funds"))

	// Context-aware logging
	ctx := context.Background()
	logger.InfoContext(ctx, "Starting operation")
	logger.ErrorContext(ctx, "Operation failed", logging.Fields{
		"error": "connection timeout",
	})

	// Using the default logger (production level)
	defaultLogger := logging.Default()
	defaultLogger.Info("Using default logger")
}

func Example_customConfig() {
	// Create a custom logger configuration
	config := logging.Config{
		Level:       logging.DebugLevel,
		Development: true,
		OutputPaths: []string{"stdout", "logs/app.log"},
		InitialFields: logging.Fields{
			"app": "example-app",
			"env": "testing",
		},
	}

	// Create a logger with custom config
	logger, err := logging.New(config)
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	logger.Info("Application started")
}

func Example_productionLogger() {
	// Create a production logger
	logger, err := logging.NewProduction()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	// Production loggers typically use JSON format
	// and have DEBUG level disabled
	logger.Debug("This won't be logged in production") // Not shown
	logger.Info("System is running")
	logger.Error("Failed to connect to database", logging.Fields{
		"error":     "connection refused",
		"database":  "postgres",
		"reconnect": true,
	})
}
