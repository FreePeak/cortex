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
