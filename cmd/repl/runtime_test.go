package main

import (
	"os"
	"testing"
)

func TestRuntimeStartCommand(t *testing.T) {
	// Clean up any existing .repl directory
	os.RemoveAll(".repl")
	defer os.RemoveAll(".repl")

	// Create the runtime start command
	cmd := newRuntimeStartCmd()

	// Test that command is properly configured
	if cmd.Use != "start" {
		t.Errorf("Expected command use to be 'start', got '%s'", cmd.Use)
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

func TestRuntimeStartCommandNotInitialized(t *testing.T) {
	// Clean up any existing .repl directory
	os.RemoveAll(".repl")
	defer os.RemoveAll(".repl")

	// Execute the runtime start command without initialization
	cmd := newRuntimeStartCmd()
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected runtime start command to fail when not initialized")
	}
}

func TestRuntimeStartCommandSuccess(t *testing.T) {
	// Clean up any existing .repl directory
	os.RemoveAll(".repl")
	defer os.RemoveAll(".repl")

	// Initialize first
	os.MkdirAll(".repl/runtime", 0755)
	os.WriteFile(".repl/runtime/execution-state.json", []byte(`{"session_active":false}`), 0644)
	os.WriteFile(".repl/runtime/task-progress.json", []byte(`{"tasks":{}}`), 0644)
	os.WriteFile(".repl/runtime/execution-log.json", []byte(`{"logs":[]}`), 0644)

	// Execute the runtime start command
	cmd := newRuntimeStartCmd()
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Runtime start command failed: %v", err)
	}

	// Verify session is active
	content, _ := os.ReadFile(".repl/runtime/execution-state.json")
	if string(content) != "{\n  \"session_active\": true\n}\n" {
		t.Errorf("Expected session_active to be true, got: %s", string(content))
	}
}

func TestRuntimeStopCommand(t *testing.T) {
	// Clean up any existing .repl directory
	os.RemoveAll(".repl")
	defer os.RemoveAll(".repl")

	// Create the runtime stop command
	cmd := newRuntimeStopCmd()

	// Test that command is properly configured
	if cmd.Use != "stop" {
		t.Errorf("Expected command use to be 'stop', got '%s'", cmd.Use)
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

func TestRuntimeStopCommandSuccess(t *testing.T) {
	// Clean up any existing .repl directory
	os.RemoveAll(".repl")
	defer os.RemoveAll(".repl")

	// Initialize first with active session
	os.MkdirAll(".repl/runtime", 0755)
	os.WriteFile(".repl/runtime/execution-state.json", []byte(`{"session_active":true}`), 0644)
	os.WriteFile(".repl/runtime/task-progress.json", []byte(`{"tasks":{}}`), 0644)
	os.WriteFile(".repl/runtime/execution-log.json", []byte(`{"logs":[]}`), 0644)

	// Execute the runtime stop command
	cmd := newRuntimeStopCmd()
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Runtime stop command failed: %v", err)
	}

	// Verify session is inactive
	content, _ := os.ReadFile(".repl/runtime/execution-state.json")
	if string(content) != "{\n  \"session_active\": false\n}\n" {
		t.Errorf("Expected session_active to be false, got: %s", string(content))
	}
}

func TestRuntimeStatusCommand(t *testing.T) {
	// Clean up any existing .repl directory
	os.RemoveAll(".repl")
	defer os.RemoveAll(".repl")

	// Create the runtime status command
	cmd := newRuntimeStatusCmd()

	// Test that command is properly configured
	if cmd.Use != "status" {
		t.Errorf("Expected command use to be 'status', got '%s'", cmd.Use)
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

func TestRuntimeStatusCommandSuccess(t *testing.T) {
	// Clean up any existing .repl directory
	os.RemoveAll(".repl")
	defer os.RemoveAll(".repl")

	// Initialize first
	os.MkdirAll(".repl/runtime", 0755)
	os.WriteFile(".repl/runtime/execution-state.json", []byte(`{"session_active":false}`), 0644)
	os.WriteFile(".repl/runtime/task-progress.json", []byte(`{"tasks":{"TASK_1":{"status":"done"}}}`), 0644)
	os.WriteFile(".repl/runtime/execution-log.json", []byte(`{"logs":["test log"]}`), 0644)

	// Execute the runtime status command
	cmd := newRuntimeStatusCmd()
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Runtime status command failed: %v", err)
	}
}
