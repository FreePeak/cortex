# Logging Package

This package provides a structured logging implementation for the MCP Server SDK, built on top of [Uber's zap](https://github.com/uber-go/zap) library.

## Features

- Multiple log levels (Debug, Info, Warn, Error, Fatal, Panic)
- Structured logging with fields
- Context-aware logging
- Printf-style logging
- High performance JSON logging
- Configurable output paths

## Usage

### Basic Logging

```go
package main

import (
    "github.com/FreePeak/golang-mcp-server-sdk/internal/infrastructure/logging"
)

func main() {
    // Create a development logger (with debug level enabled)
    logger, err := logging.NewDevelopment()
    if err != nil {
        panic(err)
    }
    defer logger.Sync() // Flushes buffer, if any

    // Simple logging
    logger.Debug("Debug message")
    logger.Info("Info message")
    logger.Warn("Warning message")
    logger.Error("Error message")
    // logger.Fatal("Fatal message") // This would exit the program
    // logger.Panic("Panic message") // This would panic
}
```

### Structured Logging with Fields

```go
logger.Info("User logged in", logging.Fields{
    "user_id": 123,
    "email":   "user@example.com",
})

// Create a logger with default fields
userLogger := logger.With(logging.Fields{
    "user_id": 123,
    "email":   "user@example.com",
})

userLogger.Info("User profile updated")
userLogger.Error("Failed to update password")
```

### Context-Aware Logging

```go
ctx := context.Background()
logger.InfoContext(ctx, "Starting operation")
logger.ErrorContext(ctx, "Operation failed", logging.Fields{
    "error": "connection timeout",
})
```

### Formatted Logging

```go
logger.Infof("User %d logged in from %s", 123, "192.168.1.1")
logger.Errorf("Failed to process payment: %v", fmt.Errorf("insufficient funds"))
```

### Custom Configuration

```go
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
```

### Default Logger

The package provides a default logger that can be used throughout your application:

```go
// Get the default logger (production level)
logger := logging.Default()
logger.Info("Using default logger")

// You can also replace the default logger
customLogger, _ := logging.NewDevelopment()
logging.SetDefault(customLogger)
```

## Available Log Levels

- `DebugLevel`: Debug information, most verbose
- `InfoLevel`: General operational information
- `WarnLevel`: Warning conditions, not critical but should be checked
- `ErrorLevel`: Error conditions, likely requiring attention
- `FatalLevel`: Fatal conditions, will call `os.Exit(1)`
- `PanicLevel`: Panic conditions, will call `panic()` 