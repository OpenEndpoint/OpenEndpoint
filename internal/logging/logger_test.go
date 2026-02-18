package logging

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestLoggerLevels(t *testing.T) {
	tests := []struct {
		level   Level
		wantStr string
	}{
		{DebugLevel, "DEBUG"},
		{InfoLevel, "INFO"},
		{WarnLevel, "WARN"},
		{ErrorLevel, "ERROR"},
		{PanicLevel, "PANIC"},
		{FatalLevel, "FATAL"},
	}

	for _, tt := range tests {
		t.Run(tt.wantStr, func(t *testing.T) {
			if tt.level.String() != tt.wantStr {
				t.Errorf("Level.String() = %v, want %v", tt.level.String(), tt.wantStr)
			}
		})
	}
}

func TestLoggerDefaultLevel(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New(buf, InfoLevel)

	// Debug should not be logged
	logger.Debug("debug message")
	if buf.Len() != 0 {
		t.Error("Debug should not be logged when level is InfoLevel")
	}

	// Info should be logged
	logger.Info("info message")
	if !strings.Contains(buf.String(), "info message") {
		t.Error("Info should be logged")
	}
}

func TestLoggerLevelFiltering(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New(buf, ErrorLevel)

	logger.Debug("debug")
	logger.Info("info")
	logger.Warn("warn")
	logger.Error("error")

	output := buf.String()

	if strings.Contains(output, "debug") {
		t.Error("Debug should be filtered")
	}
	if strings.Contains(output, "info") {
		t.Error("Info should be filtered")
	}
	if strings.Contains(output, "warn") {
		t.Error("Warn should be filtered")
	}
	if !strings.Contains(output, "error") {
		t.Error("Error should be logged")
	}
}

func TestLoggerSetLevel(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New(buf, ErrorLevel)

	// Initially, Info should be filtered
	logger.Info("info message")
	if buf.Len() != 0 {
		t.Error("Info should be filtered at ErrorLevel")
	}

	// Change level to Info
	logger.SetLevel(InfoLevel)
	logger.Info("info message")

	if !strings.Contains(buf.String(), "info message") {
		t.Error("Info should be logged after level change")
	}
}

func TestLoggerFormat(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New(buf, DebugLevel)
	logger.format = "[%s] %s: %s\n"

	logger.Info("test message")

	output := buf.String()
	if !strings.Contains(output, "INFO") {
		t.Errorf("Output should contain INFO, got: %s", output)
	}
	if !strings.Contains(output, "test message") {
		t.Errorf("Output should contain test message, got: %s", output)
	}
}

func TestLoggerMethods(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New(buf, DebugLevel)

	logger.Debug("debug")
	logger.Debugf("debug %s", "formatted")
	logger.Info("info")
	logger.Infof("info %s", "formatted")
	logger.Warn("warn")
	logger.Warnf("warn %s", "formatted")
	logger.Error("error")
	logger.Errorf("error %s", "formatted")

	output := buf.String()

	if !strings.Contains(output, "DEBUG") {
		t.Error("Debug should be logged")
	}
	if !strings.Contains(output, "INFO") {
		t.Error("Info should be logged")
	}
	if !strings.Contains(output, "WARN") {
		t.Error("Warn should be logged")
	}
	if !strings.Contains(output, "ERROR") {
		t.Error("Error should be logged")
	}
	if !strings.Contains(output, "formatted") {
		t.Error("Formatted messages should be logged")
	}
}

func TestLoggerConcurrent(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New(buf, DebugLevel)

	done := make(chan bool)

	go func() {
		for i := 0; i < 100; i++ {
			logger.Info("message from goroutine 1")
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 100; i++ {
			logger.Info("message from goroutine 2")
		}
		done <- true
	}()

	<-done
	<-done

	output := buf.String()
	count := strings.Count(output, "message from goroutine")
	if count != 200 {
		t.Errorf("Expected 200 messages, got %d", count)
	}
}

func TestLoggerWithTimeLocation(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New(buf, DebugLevel)
	logger.WithTimeLocation(nil) // Should not panic

	_ = buf.String() // Use the buffer
}

func TestLoggerWithFormat(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New(buf, DebugLevel)
	logger.WithFormat("%s - %s\n")

	logger.Info("test")

	_ = buf.String() // Use the buffer
}

func TestGlobalLogger(t *testing.T) {
	// Save original global logger
	orig := Global

	// Replace with test logger
	buf := &bytes.Buffer{}
	Global = New(buf, DebugLevel)

	Info("test info message")
	if !strings.Contains(buf.String(), "test info message") {
		t.Error("Global Info should log message")
	}

	// Restore
	Global = orig
}

func TestNewFile(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := tmpDir + "/test.log"

	logger, err := NewFile(logFile, DebugLevel)
	if err != nil {
		t.Fatalf("NewFile() error = %v", err)
	}
	defer logger.Close()

	logger.Info("test message")

	// Read the file to verify
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	if !strings.Contains(string(content), "test message") {
		t.Error("Log file should contain the message")
	}
}

func TestNewStd(t *testing.T) {
	logger := NewStd(InfoLevel)
	if logger == nil {
		t.Error("NewStd should return a logger")
	}
}

func TestLoggerClose(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New(buf, DebugLevel)

	// Closing a non-file logger should not panic
	err := logger.Close()
	if err != nil {
		t.Errorf("Close() error = %v", err)
	}
}
