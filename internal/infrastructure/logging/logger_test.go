package logging

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type testingWriter struct {
	tb   testing.TB
	logs *bytes.Buffer
}

func (w *testingWriter) Write(p []byte) (int, error) {
	n, err := w.logs.Write(p)
	return n, err
}

func (w *testingWriter) Sync() error {
	return nil
}

func newTestLogger(t *testing.T) (*Logger, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	writer := &testingWriter{
		tb:   t,
		logs: buf,
	}

	encoderConfig := zapcore.EncoderConfig{
		MessageKey:     "msg",
		LevelKey:       "level",
		TimeKey:        "time",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.AddSync(writer),
		zap.NewAtomicLevelAt(zapcore.DebugLevel),
	)

	zapLogger := zap.New(core)
	return &Logger{
		logger: zapLogger,
		sugar:  zapLogger.Sugar(),
	}, buf
}

func TestLoggerLevels(t *testing.T) {
	testLogger, buf := newTestLogger(t)
	defer testLogger.Sync()

	// Log messages at different levels
	testLogger.Debug("debug message")
	testLogger.Info("info message")
	testLogger.Warn("warning message")
	testLogger.Error("error message")

	// Check output contains all levels of messages
	output := buf.String()
	if !strings.Contains(output, "debug message") {
		t.Error("Debug message not found in logs")
	}
	if !strings.Contains(output, "info message") {
		t.Error("Info message not found in logs")
	}
	if !strings.Contains(output, "warning message") {
		t.Error("Warning message not found in logs")
	}
	if !strings.Contains(output, "error message") {
		t.Error("Error message not found in logs")
	}

	// Check log levels
	if !strings.Contains(output, `"level":"debug"`) {
		t.Error("Debug level not found in logs")
	}
	if !strings.Contains(output, `"level":"info"`) {
		t.Error("Info level not found in logs")
	}
	if !strings.Contains(output, `"level":"warn"`) {
		t.Error("Warn level not found in logs")
	}
	if !strings.Contains(output, `"level":"error"`) {
		t.Error("Error level not found in logs")
	}
}

func TestLoggerWithFields(t *testing.T) {
	testLogger, buf := newTestLogger(t)
	defer testLogger.Sync()

	// Log with fields
	testLogger.Info("user login", Fields{
		"user_id": 123,
		"action":  "login",
	})

	// Check output contains the fields
	output := buf.String()
	if !strings.Contains(output, `"user_id":123`) {
		t.Error("user_id field not found in logs")
	}
	if !strings.Contains(output, `"action":"login"`) {
		t.Error("action field not found in logs")
	}
}

func TestLoggerWithContext(t *testing.T) {
	testLogger, buf := newTestLogger(t)
	defer testLogger.Sync()

	// Create context
	ctx := context.Background()

	// Log with context
	testLogger.InfoContext(ctx, "operation started")

	// Check output contains context
	output := buf.String()
	if !strings.Contains(output, "context") {
		t.Error("context field not found in logs")
	}
	if !strings.Contains(output, "operation started") {
		t.Error("Message not found in logs")
	}
}

func TestLoggerWithFormattedMessages(t *testing.T) {
	testLogger, buf := newTestLogger(t)
	defer testLogger.Sync()

	// Log with formatting
	testLogger.Infof("User %d logged in from %s", 123, "192.168.1.1")

	// Check formatted message is in output
	output := buf.String()
	if !strings.Contains(output, "User 123 logged in from 192.168.1.1") {
		t.Error("Formatted message not found in logs")
	}
}

// Additional tests to improve coverage

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "Default config",
			config: Config{
				Level:       InfoLevel,
				Development: false,
				OutputPaths: []string{"stdout"},
			},
			wantErr: false,
		},
		{
			name: "Debug level",
			config: Config{
				Level:       DebugLevel,
				Development: true,
				OutputPaths: []string{"stdout"},
			},
			wantErr: false,
		},
		{
			name: "Warn level",
			config: Config{
				Level:       WarnLevel,
				Development: false,
				OutputPaths: []string{"stdout"},
			},
			wantErr: false,
		},
		{
			name: "Error level",
			config: Config{
				Level:       ErrorLevel,
				Development: false,
				OutputPaths: []string{"stdout"},
			},
			wantErr: false,
		},
		{
			name: "Fatal level",
			config: Config{
				Level:       FatalLevel,
				Development: false,
				OutputPaths: []string{"stdout"},
			},
			wantErr: false,
		},
		{
			name: "Panic level",
			config: Config{
				Level:       PanicLevel,
				Development: false,
				OutputPaths: []string{"stdout"},
			},
			wantErr: false,
		},
		{
			name: "Unknown level",
			config: Config{
				Level:       LogLevel("unknown"),
				Development: false,
				OutputPaths: []string{"stdout"},
			},
			wantErr: false, // Should default to InfoLevel
		},
		{
			name: "With initial fields",
			config: Config{
				Level:       InfoLevel,
				Development: false,
				OutputPaths: []string{"stdout"},
				InitialFields: Fields{
					"service": "test-service",
					"version": "1.0.0",
				},
			},
			wantErr: false,
		},
		{
			name: "Invalid output path",
			config: Config{
				Level:       InfoLevel,
				Development: false,
				OutputPaths: []string{"/invalid/path/that/doesnt/exist"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := New(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("New() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && logger == nil {
				t.Error("Expected non-nil logger")
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	if config.Level != InfoLevel {
		t.Errorf("Expected InfoLevel, got %v", config.Level)
	}
	if config.Development != false {
		t.Errorf("Expected Development to be false")
	}
	if len(config.OutputPaths) != 1 || config.OutputPaths[0] != "stdout" {
		t.Errorf("Expected OutputPaths to be [stdout], got %v", config.OutputPaths)
	}
}

func TestDevelopmentConfig(t *testing.T) {
	config := DevelopmentConfig()
	if config.Level != DebugLevel {
		t.Errorf("Expected DebugLevel, got %v", config.Level)
	}
	if config.Development != true {
		t.Errorf("Expected Development to be true")
	}
	if len(config.OutputPaths) != 1 || config.OutputPaths[0] != "stdout" {
		t.Errorf("Expected OutputPaths to be [stdout], got %v", config.OutputPaths)
	}
}

func TestProductionConfig(t *testing.T) {
	config := ProductionConfig()
	if config.Level != InfoLevel {
		t.Errorf("Expected InfoLevel, got %v", config.Level)
	}
	if config.Development != false {
		t.Errorf("Expected Development to be false")
	}
	if len(config.OutputPaths) != 1 || config.OutputPaths[0] != "stdout" {
		t.Errorf("Expected OutputPaths to be [stdout], got %v", config.OutputPaths)
	}
}

func TestNewDevelopment(t *testing.T) {
	logger, err := NewDevelopment()
	if err != nil {
		t.Errorf("NewDevelopment() error = %v", err)
		return
	}
	if logger == nil {
		t.Error("Expected non-nil logger")
	}
}

func TestNewProduction(t *testing.T) {
	logger, err := NewProduction()
	if err != nil {
		t.Errorf("NewProduction() error = %v", err)
		return
	}
	if logger == nil {
		t.Error("Expected non-nil logger")
	}
}

func TestWithEmptyFields(t *testing.T) {
	testLogger, _ := newTestLogger(t)
	defer testLogger.Sync()

	newLogger := testLogger.With(Fields{})
	if newLogger != testLogger {
		t.Error("Expected same logger instance when With is called with empty fields")
	}
}

func TestLoggerAllMethods(t *testing.T) {
	testLogger, buf := newTestLogger(t)
	defer testLogger.Sync()

	ctx := context.Background()

	// Test all methods
	tests := []struct {
		name     string
		logFunc  func()
		contains string
	}{
		{
			name: "Debug",
			logFunc: func() {
				testLogger.Debug("debug test")
			},
			contains: "debug test",
		},
		{
			name: "Debug with fields",
			logFunc: func() {
				testLogger.Debug("debug with fields", Fields{"key": "value"})
			},
			contains: `"key":"value"`,
		},
		{
			name: "Info",
			logFunc: func() {
				testLogger.Info("info test")
			},
			contains: "info test",
		},
		{
			name: "Info with fields",
			logFunc: func() {
				testLogger.Info("info with fields", Fields{"key": "value"})
			},
			contains: `"key":"value"`,
		},
		{
			name: "Warn",
			logFunc: func() {
				testLogger.Warn("warn test")
			},
			contains: "warn test",
		},
		{
			name: "Warn with fields",
			logFunc: func() {
				testLogger.Warn("warn with fields", Fields{"key": "value"})
			},
			contains: `"key":"value"`,
		},
		{
			name: "Error",
			logFunc: func() {
				testLogger.Error("error test")
			},
			contains: "error test",
		},
		{
			name: "Error with fields",
			logFunc: func() {
				testLogger.Error("error with fields", Fields{"key": "value"})
			},
			contains: `"key":"value"`,
		},
		// Skip Fatal and Panic as they would terminate the test
		{
			name: "DebugContext",
			logFunc: func() {
				testLogger.DebugContext(ctx, "debug context test")
			},
			contains: "debug context test",
		},
		{
			name: "DebugContext with fields",
			logFunc: func() {
				testLogger.DebugContext(ctx, "debug context with fields", Fields{"key": "value"})
			},
			contains: `"key":"value"`,
		},
		{
			name: "InfoContext",
			logFunc: func() {
				testLogger.InfoContext(ctx, "info context test")
			},
			contains: "info context test",
		},
		{
			name: "InfoContext with fields",
			logFunc: func() {
				testLogger.InfoContext(ctx, "info context with fields", Fields{"key": "value"})
			},
			contains: `"key":"value"`,
		},
		{
			name: "WarnContext",
			logFunc: func() {
				testLogger.WarnContext(ctx, "warn context test")
			},
			contains: "warn context test",
		},
		{
			name: "WarnContext with fields",
			logFunc: func() {
				testLogger.WarnContext(ctx, "warn context with fields", Fields{"key": "value"})
			},
			contains: `"key":"value"`,
		},
		{
			name: "ErrorContext",
			logFunc: func() {
				testLogger.ErrorContext(ctx, "error context test")
			},
			contains: "error context test",
		},
		{
			name: "ErrorContext with fields",
			logFunc: func() {
				testLogger.ErrorContext(ctx, "error context with fields", Fields{"key": "value"})
			},
			contains: `"key":"value"`,
		},
		// Skip FatalContext and PanicContext as they would terminate the test
		{
			name: "Debugf",
			logFunc: func() {
				testLogger.Debugf("debug %s test", "format")
			},
			contains: "debug format test",
		},
		{
			name: "Infof",
			logFunc: func() {
				testLogger.Infof("info %s test", "format")
			},
			contains: "info format test",
		},
		{
			name: "Warnf",
			logFunc: func() {
				testLogger.Warnf("warn %s test", "format")
			},
			contains: "warn format test",
		},
		{
			name: "Errorf",
			logFunc: func() {
				testLogger.Errorf("error %s test", "format")
			},
			contains: "error format test",
		},
		// Skip Fatalf and Panicf as they would terminate the test
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			tt.logFunc()
			output := buf.String()
			if !strings.Contains(output, tt.contains) {
				t.Errorf("Expected output to contain %q, got %q", tt.contains, output)
			}
		})
	}
}

func TestDefaultLogger(t *testing.T) {
	// Should not panic
	logger := Default()
	if logger == nil {
		t.Error("Expected non-nil default logger")
	}

	// Test setting default
	testLogger, _ := newTestLogger(t)
	SetDefault(testLogger)

	// Verify the default was set
	if Default() != testLogger {
		t.Error("Expected SetDefault to set the default logger")
	}
}
