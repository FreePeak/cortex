// Package logging provides a wrapper around zap for structured logging
package logging

import (
	"context"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger is a wrapper around zap.Logger providing a simplified API
type Logger struct {
	logger *zap.Logger
	sugar  *zap.SugaredLogger
}

// Fields is a type alias for key-value pairs
type Fields map[string]interface{}

// LogLevel represents the log severity level
type LogLevel string

// Available log levels
const (
	DebugLevel LogLevel = "debug"
	InfoLevel  LogLevel = "info"
	WarnLevel  LogLevel = "warn"
	ErrorLevel LogLevel = "error"
	FatalLevel LogLevel = "fatal"
	PanicLevel LogLevel = "panic"
)

// Config represents the logging configuration
type Config struct {
	Level         LogLevel
	Development   bool
	OutputPaths   []string
	InitialFields Fields
}

// DefaultConfig returns a default configuration for the logger
func DefaultConfig() Config {
	return Config{
		Level:       InfoLevel,
		Development: false,
		OutputPaths: []string{"stdout"},
	}
}

// DevelopmentConfig returns a development configuration for the logger
func DevelopmentConfig() Config {
	return Config{
		Level:       DebugLevel,
		Development: true,
		OutputPaths: []string{"stdout"},
	}
}

// ProductionConfig returns a production configuration for the logger
func ProductionConfig() Config {
	return Config{
		Level:       InfoLevel,
		Development: false,
		OutputPaths: []string{"stdout"},
	}
}

// New creates a new logger with the given configuration
func New(config Config) (*Logger, error) {
	// Convert log level to zapcore level
	var level zapcore.Level
	switch config.Level {
	case DebugLevel:
		level = zapcore.DebugLevel
	case InfoLevel:
		level = zapcore.InfoLevel
	case WarnLevel:
		level = zapcore.WarnLevel
	case ErrorLevel:
		level = zapcore.ErrorLevel
	case FatalLevel:
		level = zapcore.FatalLevel
	case PanicLevel:
		level = zapcore.PanicLevel
	default:
		level = zapcore.InfoLevel
	}

	// Create zap configuration
	zapConfig := zap.Config{
		Level:             zap.NewAtomicLevelAt(level),
		Development:       config.Development,
		DisableCaller:     !config.Development,
		DisableStacktrace: !config.Development,
		Encoding:          "json",
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "timestamp",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			FunctionKey:    zapcore.OmitKey,
			MessageKey:     "message",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		OutputPaths:      config.OutputPaths,
		ErrorOutputPaths: []string{"stderr"},
	}

	// Apply initial fields if any
	if config.InitialFields != nil {
		zapConfig.InitialFields = make(map[string]interface{})
		for k, v := range config.InitialFields {
			zapConfig.InitialFields[k] = v
		}
	}

	// Build the logger
	zapLogger, err := zapConfig.Build()
	if err != nil {
		return nil, err
	}

	// Create our logger wrapper
	return &Logger{
		logger: zapLogger,
		sugar:  zapLogger.Sugar(),
	}, nil
}

// NewDevelopment creates a new development logger
func NewDevelopment() (*Logger, error) {
	return New(DevelopmentConfig())
}

// NewProduction creates a new production logger
func NewProduction() (*Logger, error) {
	return New(ProductionConfig())
}

// With returns a logger with the given fields
func (l *Logger) With(fields Fields) *Logger {
	if len(fields) == 0 {
		return l
	}

	zapFields := make([]zap.Field, 0, len(fields))
	for k, v := range fields {
		zapFields = append(zapFields, zap.Any(k, v))
	}

	newLogger := l.logger.With(zapFields...)
	return &Logger{
		logger: newLogger,
		sugar:  newLogger.Sugar(),
	}
}

// Debug logs a message at debug level with optional fields
func (l *Logger) Debug(msg string, fields ...Fields) {
	if len(fields) > 0 {
		l.With(fields[0]).logger.Debug(msg)
	} else {
		l.logger.Debug(msg)
	}
}

// Info logs a message at info level with optional fields
func (l *Logger) Info(msg string, fields ...Fields) {
	if len(fields) > 0 {
		l.With(fields[0]).logger.Info(msg)
	} else {
		l.logger.Info(msg)
	}
}

// Warn logs a message at warn level with optional fields
func (l *Logger) Warn(msg string, fields ...Fields) {
	if len(fields) > 0 {
		l.With(fields[0]).logger.Warn(msg)
	} else {
		l.logger.Warn(msg)
	}
}

// Error logs a message at error level with optional fields
func (l *Logger) Error(msg string, fields ...Fields) {
	if len(fields) > 0 {
		l.With(fields[0]).logger.Error(msg)
	} else {
		l.logger.Error(msg)
	}
}

// Fatal logs a message at fatal level with optional fields and then calls os.Exit(1)
func (l *Logger) Fatal(msg string, fields ...Fields) {
	if len(fields) > 0 {
		l.With(fields[0]).logger.Fatal(msg)
	} else {
		l.logger.Fatal(msg)
	}
}

// Panic logs a message at panic level with optional fields and then panics
func (l *Logger) Panic(msg string, fields ...Fields) {
	if len(fields) > 0 {
		l.With(fields[0]).logger.Panic(msg)
	} else {
		l.logger.Panic(msg)
	}
}

// DebugContext logs a message at debug level with context and optional fields
func (l *Logger) DebugContext(ctx context.Context, msg string, fields ...Fields) {
	if len(fields) > 0 {
		l.With(fields[0]).logger.Debug(msg, zap.Any("context", ctx))
	} else {
		l.logger.Debug(msg, zap.Any("context", ctx))
	}
}

// InfoContext logs a message at info level with context and optional fields
func (l *Logger) InfoContext(ctx context.Context, msg string, fields ...Fields) {
	if len(fields) > 0 {
		l.With(fields[0]).logger.Info(msg, zap.Any("context", ctx))
	} else {
		l.logger.Info(msg, zap.Any("context", ctx))
	}
}

// WarnContext logs a message at warn level with context and optional fields
func (l *Logger) WarnContext(ctx context.Context, msg string, fields ...Fields) {
	if len(fields) > 0 {
		l.With(fields[0]).logger.Warn(msg, zap.Any("context", ctx))
	} else {
		l.logger.Warn(msg, zap.Any("context", ctx))
	}
}

// ErrorContext logs a message at error level with context and optional fields
func (l *Logger) ErrorContext(ctx context.Context, msg string, fields ...Fields) {
	if len(fields) > 0 {
		l.With(fields[0]).logger.Error(msg, zap.Any("context", ctx))
	} else {
		l.logger.Error(msg, zap.Any("context", ctx))
	}
}

// FatalContext logs a message at fatal level with context and optional fields and then calls os.Exit(1)
func (l *Logger) FatalContext(ctx context.Context, msg string, fields ...Fields) {
	if len(fields) > 0 {
		l.With(fields[0]).logger.Fatal(msg, zap.Any("context", ctx))
	} else {
		l.logger.Fatal(msg, zap.Any("context", ctx))
	}
}

// PanicContext logs a message at panic level with context and optional fields and then panics
func (l *Logger) PanicContext(ctx context.Context, msg string, fields ...Fields) {
	if len(fields) > 0 {
		l.With(fields[0]).logger.Panic(msg, zap.Any("context", ctx))
	} else {
		l.logger.Panic(msg, zap.Any("context", ctx))
	}
}

// Debugf logs a formatted message at debug level
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.sugar.Debugf(format, args...)
}

// Infof logs a formatted message at info level
func (l *Logger) Infof(format string, args ...interface{}) {
	l.sugar.Infof(format, args...)
}

// Warnf logs a formatted message at warn level
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.sugar.Warnf(format, args...)
}

// Errorf logs a formatted message at error level
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.sugar.Errorf(format, args...)
}

// Fatalf logs a formatted message at fatal level and then calls os.Exit(1)
func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.sugar.Fatalf(format, args...)
}

// Panicf logs a formatted message at panic level and then panics
func (l *Logger) Panicf(format string, args ...interface{}) {
	l.sugar.Panicf(format, args...)
}

// Sync flushes any buffered log entries
func (l *Logger) Sync() error {
	return l.logger.Sync()
}

// DefaultLogger returns a new default production logger
var defaultLogger, _ = NewProduction()

// Default returns the default logger
func Default() *Logger {
	return defaultLogger
}

// SetDefault sets the default logger
func SetDefault(logger *Logger) {
	defaultLogger = logger
}
