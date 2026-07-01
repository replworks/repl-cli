package main

import (
	"os"
	"testing"
)

func TestResetCommand(t *testing.T) {
	// Clean up any existing .repl directory
	os.RemoveAll(".repl")
	defer os.RemoveAll(".repl")

	// Create the reset command
	cmd := newResetCmd()

	// Test that command is properly configured
	if cmd.Use != "reset" {
		t.Errorf("Expected command use to be 'reset', got '%s'", cmd.Use)
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

func TestResetCommandNotInitialized(t *testing.T) {
	// Clean up any existing .repl directory
	os.RemoveAll(".repl")
	defer os.RemoveAll(".repl")

	// Execute the reset command without initialization
	cmd := newResetCmd()
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected reset command to fail when not initialized")
	}

	// Verify error message
	if err.Error() != "REPL project not initialized. Run 'repl init' first" {
		t.Errorf("Expected error message about initialization, got '%s'", err.Error())
	}
}

func TestResetCommandSuccess(t *testing.T) {
	// Clean up any existing .repl directory
	os.RemoveAll(".repl")
	defer os.RemoveAll(".repl")

	// Initialize first
	os.MkdirAll(".repl/runtime", 0755)
	os.WriteFile(".repl/runtime/execution-state.json", []byte(`{"session_active":true,"current_task":"TASK_1"}`), 0644)
	os.WriteFile(".repl/runtime/task-progress.json", []byte(`{"tasks":{"TASK_1":{"status":"done"}}}`), 0644)
	os.WriteFile(".repl/runtime/execution-log.json", []byte(`{"logs":["test log"]}`), 0644)

	// Execute the reset command
	cmd := newResetCmd()
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Reset command failed: %v", err)
	}

	// Verify state was reset - execution-state.json should have session_active=false
	content, _ := os.ReadFile(".repl/runtime/execution-state.json")
	if string(content) != "{\n  \"session_active\": false\n}\n" {
		t.Errorf("Expected execution-state.json to be reset, got: %s", string(content))
	}
}
