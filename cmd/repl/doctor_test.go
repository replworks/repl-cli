package main

import (
	"os"
	"strings"
	"testing"
)

func TestDoctorCommand(t *testing.T) {
	// Clean up any existing .repl directory
	os.RemoveAll(".repl")
	defer os.RemoveAll(".repl")

	// Create the doctor command
	cmd := newDoctorCmd()

	// Test that command is properly configured
	if cmd.Use != "doctor" {
		t.Errorf("Expected command use to be 'doctor', got '%s'", cmd.Use)
	}

	if cmd.Short == "" {
		t.Error("Expected command to have a short description")
	}

	if cmd.Long == "" {
		t.Error("Expected command to have a long description")
	}

	// Test that Run or RunE is set
	if cmd.Run == nil && cmd.RunE == nil {
		t.Error("Expected command to have Run or RunE function")
	}
}

func TestDoctorCommandNotInitialized(t *testing.T) {
	// Clean up any existing .repl directory
	os.RemoveAll(".repl")
	defer os.RemoveAll(".repl")

	// Execute the doctor command without initialization
	cmd := newDoctorCmd()
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected doctor command to fail when not initialized")
	}

	// Verify error message contains expected text
	if err != nil && !strings.Contains(err.Error(), ".repl/ directory does not exist") {
		t.Errorf("Expected error about .repl/ directory, got: %v", err)
	}
}

func TestDoctorCommandSuccess(t *testing.T) {
	// Clean up any existing .repl directory
	os.RemoveAll(".repl")
	defer os.RemoveAll(".repl")

	// Initialize first
	os.MkdirAll(".repl/runtime", 0755)

	// Create minimal required files
	os.WriteFile(".repl/runtime/execution-state.json", []byte(`{"session_active":false}`), 0644)
	os.WriteFile(".repl/runtime/task-progress.json", []byte(`{"tasks":{}}`), 0644)
	os.WriteFile(".repl/runtime/execution-log.json", []byte(`{"logs":[]}`), 0644)

	// Execute the doctor command
	cmd := newDoctorCmd()
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Doctor command failed: %v", err)
	}
}
