package cmd

import (
	"strings"
	"testing"
)

func TestLogsCmdMetadata(t *testing.T) {
	// Test command structure
	if logsCmd.Use != "logs [SERVICE]" {
		t.Errorf("Expected Use to be 'logs [SERVICE]', got '%s'", logsCmd.Use)
	}

	if logsCmd.Short != "View service logs" {
		t.Errorf("Expected Short to be 'View service logs', got '%s'", logsCmd.Short)
	}

	// Test that the important note is included in help text
	if !strings.Contains(logsCmd.Long, "Execute this command on the same machine") {
		t.Error("Expected Long description to contain location guidance")
	}

	// Test follow flag exists
	followFlagExists := logsCmd.Flags().Lookup("follow") != nil
	if !followFlagExists {
		t.Error("Expected --follow flag to exist")
	}

	// Test follow flag short form
	followFlag := logsCmd.Flags().Lookup("follow")
	if followFlag != nil && followFlag.Shorthand != "f" {
		t.Errorf("Expected --follow flag shorthand to be 'f', got '%s'", followFlag.Shorthand)
	}

	// Test aliases
	expectedAliases := []string{"log"}
	if len(logsCmd.Aliases) != len(expectedAliases) {
		t.Errorf("Expected %d aliases, got %d", len(expectedAliases), len(logsCmd.Aliases))
	}
	for i, alias := range expectedAliases {
		if i >= len(logsCmd.Aliases) || logsCmd.Aliases[i] != alias {
			t.Errorf("Expected alias '%s', got '%s'", alias, logsCmd.Aliases[i])
		}
	}
}

func TestLogsCmdRequiresServiceArg(t *testing.T) {
	// Test with no arguments - should fail
	cmd := logsCmd
	cmd.SetArgs([]string{})
	
	err := cmd.Args(cmd, []string{})
	if err == nil {
		t.Error("Expected error when no service argument provided")
	}

	// Test with exactly one argument - should pass
	err = cmd.Args(cmd, []string{"nginx"})
	if err != nil {
		t.Errorf("Expected no error with one service argument, got: %v", err)
	}

	// Test with too many arguments - should fail
	err = cmd.Args(cmd, []string{"nginx", "extra"})
	if err == nil {
		t.Error("Expected error when too many arguments provided")
	}
}

func TestGetLogStyle(t *testing.T) {
	tests := []struct {
		message       string
		expectedEmoji string
		expectedColor string
	}{
		{"[error] something failed", "🔴", "red"},
		{"[warn] be careful", "⚠️", "yellow"},
		{"[notice] 1#1: nginx/1.27.5", "🔷", "blue"},
		{"[info] getting checksum", "🔷", "blue"},
		{"Configuration complete; ready for start up", "🔷", "blue"},
		{"[debug] verbose output", "🔸", "cyan"},
		{"some general message", "🔶", "cyan"},
	}

	for _, test := range tests {
		emoji, colorFunc := getLogStyle(test.message)
		if emoji != test.expectedEmoji {
			t.Errorf("For message '%s', expected emoji '%s', got '%s'", 
				test.message, test.expectedEmoji, emoji)
		}
		
		// Test that colorFunc is not nil
		if colorFunc == nil {
			t.Errorf("For message '%s', expected color function, got nil", test.message)
		}
	}
}
