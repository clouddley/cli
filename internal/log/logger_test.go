package log

import (
	"bytes"
	"strings"
	"testing"

	"github.com/charmbracelet/log"
)

func TestLoggerInit(t *testing.T) {
	// Test that the logger is properly initialized
	// We can't easily test the actual output, but we can verify the logger exists
	// and has the expected configuration
	
	// Create a buffer to capture log output
	var buf bytes.Buffer
	
	// Create a new logger for testing
	testLogger := log.NewWithOptions(&buf, log.Options{
		ReportCaller:    false,
		ReportTimestamp: true,
		TimeFormat:      "3:04PM",
		Prefix:          "Clouddley:",
	})
	
	// Test basic logging
	testLogger.Info("Test info message")
	testLogger.Error("Test error message")
	testLogger.Debug("Test debug message")
	
	// Verify output contains expected elements
	output := buf.String()
	
	if !strings.Contains(output, "Test info message") {
		t.Error("Expected info message not found in output")
	}
	
	if !strings.Contains(output, "Test error message") {
		t.Error("Expected error message not found in output")
	}
	
	if !strings.Contains(output, "Clouddley:") {
		t.Error("Expected prefix 'Clouddley:' not found in output")
	}
}

func TestLogLevel(t *testing.T) {
	var buf bytes.Buffer
	
	// Create logger with different levels
	infoLogger := log.NewWithOptions(&buf, log.Options{
		Level:           log.InfoLevel,
		ReportCaller:    false,
		ReportTimestamp: false,
		Prefix:          "Test:",
	})
	
	// Test that debug messages are filtered out at INFO level
	infoLogger.Debug("This debug message should not appear")
	infoLogger.Info("This info message should appear")
	
	output := buf.String()
	
	if strings.Contains(output, "debug message should not appear") {
		t.Error("Debug message appeared when logger level is INFO")
	}
	
	if !strings.Contains(output, "info message should appear") {
		t.Error("Info message did not appear when logger level is INFO")
	}
}

func TestLoggerPrefix(t *testing.T) {
	tests := []struct {
		name   string
		prefix string
	}{
		{
			name:   "Standard prefix",
			prefix: "Clouddley:",
		},
		{
			name:   "Custom prefix",
			prefix: "TestApp:",
		},
		{
			name:   "Empty prefix",
			prefix: "",
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			
			testLogger := log.NewWithOptions(&buf, log.Options{
				ReportCaller:    false,
				ReportTimestamp: false,
				Prefix:          tt.prefix,
			})
			
			testLogger.Info("Test message")
			output := buf.String()
			
			if tt.prefix != "" {
				if !strings.Contains(output, tt.prefix) {
					t.Errorf("Expected prefix '%s' not found in output: %s", tt.prefix, output)
				}
			}
			
			if !strings.Contains(output, "Test message") {
				t.Error("Expected message not found in output")
			}
		})
	}
}

func TestLoggerTimestampFormat(t *testing.T) {
	var buf bytes.Buffer
	
	testLogger := log.NewWithOptions(&buf, log.Options{
		ReportCaller:    false,
		ReportTimestamp: true,
		TimeFormat:      "3:04PM",
		Prefix:          "",
	})
	
	testLogger.Info("Test message with timestamp")
	output := buf.String()
	
	// Check that output contains a timestamp pattern (like "9:36PM")
	// We can't check exact time, but we can check the format pattern
	hasTimePattern := false
	for _, char := range output {
		if char >= '0' && char <= '9' {
			hasTimePattern = true
			break
		}
	}
	
	if !hasTimePattern {
		t.Error("Expected timestamp pattern not found in output")
	}
}

func TestLoggerFields(t *testing.T) {
	var buf bytes.Buffer
	
	testLogger := log.NewWithOptions(&buf, log.Options{
		ReportCaller:    false,
		ReportTimestamp: false,
		Prefix:          "",
	})
	
	// Test logging with fields (key-value pairs)
	testLogger.Info("Test message", "key1", "value1", "key2", "value2")
	output := buf.String()
	
	if !strings.Contains(output, "Test message") {
		t.Error("Expected message not found in output")
	}
	
	if !strings.Contains(output, "key1") || !strings.Contains(output, "value1") {
		t.Error("Expected key-value pair 'key1=value1' not found in output")
	}
	
	if !strings.Contains(output, "key2") || !strings.Contains(output, "value2") {
		t.Error("Expected key-value pair 'key2=value2' not found in output")
	}
}

func TestLoggerErrorHandling(t *testing.T) {
	var buf bytes.Buffer
	
	testLogger := log.NewWithOptions(&buf, log.Options{
		ReportCaller:    false,
		ReportTimestamp: false,
		Prefix:          "Error:",
	})
	
	// Test error logging
	testError := "something went wrong"
	testLogger.Error("Operation failed", "error", testError)
	
	output := buf.String()
	
	if !strings.Contains(output, "Operation failed") {
		t.Error("Expected error message not found in output")
	}
	
	if !strings.Contains(output, testError) {
		t.Error("Expected error details not found in output")
	}
	
	if !strings.Contains(output, "Error:") {
		t.Error("Expected error prefix not found in output")
	}
}

func TestLoggerMultipleMessages(t *testing.T) {
	var buf bytes.Buffer
	
	testLogger := log.NewWithOptions(&buf, log.Options{
		ReportCaller:    false,
		ReportTimestamp: false,
		Prefix:          "",
	})
	
	// Test multiple log messages
	messages := []string{
		"First message",
		"Second message", 
		"Third message",
	}
	
	for _, msg := range messages {
		testLogger.Info(msg)
	}
	
	output := buf.String()
	
	for _, msg := range messages {
		if !strings.Contains(output, msg) {
			t.Errorf("Expected message '%s' not found in output", msg)
		}
	}
}

func TestLoggerStructuredData(t *testing.T) {
	var buf bytes.Buffer
	
	testLogger := log.NewWithOptions(&buf, log.Options{
		ReportCaller:    false,
		ReportTimestamp: false,
		Prefix:          "",
	})
	
	// Test structured logging with various data types
	testLogger.Info("Structured log test",
		"string_field", "test_value",
		"int_field", 42,
		"bool_field", true,
		"instance_id", "i-1234567890abcdef0",
	)
	
	output := buf.String()
	
	expectedPairs := map[string]string{
		"string_field": "test_value",
		"int_field":    "42",
		"bool_field":   "true",
		"instance_id":  "i-1234567890abcdef0",
	}
	
	for key, value := range expectedPairs {
		if !strings.Contains(output, key) {
			t.Errorf("Expected key '%s' not found in output", key)
		}
		if !strings.Contains(output, value) {
			t.Errorf("Expected value '%s' not found in output", value)
		}
	}
}
