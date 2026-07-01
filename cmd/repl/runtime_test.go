package main

import (
	"os"
	"strings"
	"testing"

	rt "repl-cli/internal/runtime"
)

func TestRuntimeStartCommand(t *testing.T) {
	// Clean up any existing .repl directory
	_ = os.RemoveAll(".repl")
	defer func() { _ = os.RemoveAll(".repl") }()

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
	_ = os.RemoveAll(".repl")
	defer func() { _ = os.RemoveAll(".repl") }()

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
	_ = os.RemoveAll(".repl")
	defer func() { _ = os.RemoveAll(".repl") }()

	// Initialize first
	_ = os.MkdirAll(".repl/runtime", 0755)
	_ = os.WriteFile(".repl/runtime/execution-state.json", []byte(`{"session_active":false}`), 0644)
	_ = os.WriteFile(".repl/runtime/task-progress.json", []byte(`{"tasks":{}}`), 0644)
	_ = os.WriteFile(".repl/runtime/execution-log.json", []byte(`{"logs":[]}`), 0644)

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
	_ = os.RemoveAll(".repl")
	defer func() { _ = os.RemoveAll(".repl") }()

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
	_ = os.RemoveAll(".repl")
	defer func() { _ = os.RemoveAll(".repl") }()

	// Initialize first with active session
	_ = os.MkdirAll(".repl/runtime", 0755)
	_ = os.WriteFile(".repl/runtime/execution-state.json", []byte(`{"session_active":true}`), 0644)
	_ = os.WriteFile(".repl/runtime/task-progress.json", []byte(`{"tasks":{}}`), 0644)
	_ = os.WriteFile(".repl/runtime/execution-log.json", []byte(`{"logs":[]}`), 0644)

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
	_ = os.RemoveAll(".repl")
	defer func() { _ = os.RemoveAll(".repl") }()

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
	_ = os.RemoveAll(".repl")
	defer func() { _ = os.RemoveAll(".repl") }()

	// Initialize first
	_ = os.MkdirAll(".repl/runtime", 0755)
	_ = os.WriteFile(".repl/runtime/execution-state.json", []byte(`{"session_active":false}`), 0644)
	_ = os.WriteFile(".repl/runtime/task-progress.json", []byte(`{"tasks":{"TASK_1":{"status":"done"}}}`), 0644)
	_ = os.WriteFile(".repl/runtime/execution-log.json", []byte(`{"logs":["test log"]}`), 0644)

	// Execute the runtime status command
	cmd := newRuntimeStatusCmd()
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Runtime status command failed: %v", err)
	}
}

func TestRuntimeStopClearsCurrentTask(t *testing.T) {
	_ = os.RemoveAll(".repl")
	defer func() { _ = os.RemoveAll(".repl") }()

	_ = os.MkdirAll(".repl/runtime", 0755)
	_ = os.WriteFile(".repl/runtime/execution-state.json", []byte(`{"session_active":true,"current_task":"TASK_1"}`), 0644)
	_ = os.WriteFile(".repl/runtime/task-progress.json", []byte(`{"tasks":{}}`), 0644)
	_ = os.WriteFile(".repl/runtime/execution-log.json", []byte(`{"logs":[]}`), 0644)

	err := runRuntimeStop()
	if err != nil {
		t.Fatalf("runRuntimeStop() failed: %v", err)
	}

	state, err := rt.ReadState()
	if err != nil {
		t.Fatalf("failed to read state: %v", err)
	}

	if state.CurrentTask != "" {
		t.Fatalf("expected current task to be cleared after stop, got %q", state.CurrentTask)
	}
}

func TestRuntimeApplyCommandUpdatesProgressAndCurrentTask(t *testing.T) {
	_ = os.RemoveAll(".repl")
	defer func() { _ = os.RemoveAll(".repl") }()

	_ = os.MkdirAll(".repl/runtime", 0755)
	_ = os.WriteFile(".repl/runtime/execution-state.json", []byte(`{"session_active":false}`), 0644)
	_ = os.WriteFile(".repl/runtime/task-progress.json", []byte(`{"tasks":{}}`), 0644)
	_ = os.WriteFile(".repl/runtime/execution-log.json", []byte(`{"logs":[]}`), 0644)

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create stdin pipe: %v", err)
	}
	defer func() { _ = r.Close() }()

	_, _ = w.WriteString(`{"action":"update_runtime","taskId":"TASK_1","status":"done","events":["step-1"]}`)
	_ = w.Close()

	oldStdin := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = oldStdin }()

	err = runRuntimeApply()
	if err != nil {
		t.Fatalf("runRuntimeApply() failed: %v", err)
	}

	progress, err := rt.ReadProgress()
	if err != nil {
		t.Fatalf("failed to read progress: %v", err)
	}

	if progress.Tasks["TASK_1"].Status != "done" {
		t.Fatalf("expected TASK_1 status to be done, got %q", progress.Tasks["TASK_1"].Status)
	}

	state, err := rt.ReadState()
	if err != nil {
		t.Fatalf("failed to read state: %v", err)
	}

	if state.CurrentTask != "TASK_1" {
		t.Fatalf("expected current task to be TASK_1, got %q", state.CurrentTask)
	}
}

func TestRuntimeApplyCommandRequiresReasonForBlockedTasks(t *testing.T) {
	_ = os.RemoveAll(".repl")
	defer func() { _ = os.RemoveAll(".repl") }()

	_ = os.MkdirAll(".repl/runtime", 0755)
	_ = os.WriteFile(".repl/runtime/execution-state.json", []byte(`{"session_active":false}`), 0644)
	_ = os.WriteFile(".repl/runtime/task-progress.json", []byte(`{"tasks":{}}`), 0644)
	_ = os.WriteFile(".repl/runtime/execution-log.json", []byte(`{"logs":[]}`), 0644)

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create stdin pipe: %v", err)
	}
	defer func() { _ = r.Close() }()

	_, _ = w.WriteString(`{"action":"update_runtime","taskId":"TASK_2","status":"blocked"}`)
	_ = w.Close()

	oldStdin := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = oldStdin }()

	err = runRuntimeApply()
	if err == nil {
		t.Fatal("expected runRuntimeApply() to fail when blocked tasks omit reason")
	}

	if !strings.Contains(err.Error(), "reason is required") {
		t.Fatalf("expected validation error, got %v", err)
	}
}

func TestRuntimeApplyCommandRejectsInvalidJSON(t *testing.T) {
	_ = os.RemoveAll(".repl")
	defer func() { _ = os.RemoveAll(".repl") }()

	_ = os.MkdirAll(".repl/runtime", 0755)
	_ = os.WriteFile(".repl/runtime/execution-state.json", []byte(`{"session_active":false}`), 0644)
	_ = os.WriteFile(".repl/runtime/task-progress.json", []byte(`{"tasks":{}}`), 0644)
	_ = os.WriteFile(".repl/runtime/execution-log.json", []byte(`{"logs":[]}`), 0644)

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create stdin pipe: %v", err)
	}
	defer func() { _ = r.Close() }()

	_, _ = w.WriteString(`{"action":"update_runtime","taskId":"TASK_3","status":}`)
	_ = w.Close()

	oldStdin := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = oldStdin }()

	err = runRuntimeApply()
	if err == nil {
		t.Fatal("expected runRuntimeApply() to fail for invalid JSON")
	}

	if !strings.Contains(err.Error(), "invalid JSON input") {
		t.Fatalf("expected invalid JSON error, got %v", err)
	}
}

func TestRuntimeApplyCommandRejectsUnknownFields(t *testing.T) {
	_ = os.RemoveAll(".repl")
	defer func() { _ = os.RemoveAll(".repl") }()

	_ = os.MkdirAll(".repl/runtime", 0755)
	_ = os.WriteFile(".repl/runtime/execution-state.json", []byte(`{"session_active":false}`), 0644)
	_ = os.WriteFile(".repl/runtime/task-progress.json", []byte(`{"tasks":{}}`), 0644)
	_ = os.WriteFile(".repl/runtime/execution-log.json", []byte(`{"logs":[]}`), 0644)

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create stdin pipe: %v", err)
	}
	defer func() { _ = r.Close() }()

	_, _ = w.WriteString(`{"action":"update_runtime","taskId":"TASK_4","status":"done","unexpected":"value"}`)
	_ = w.Close()

	oldStdin := os.Stdin
	os.Stdin = r
	defer func() { os.Stdin = oldStdin }()

	err = runRuntimeApply()
	if err == nil {
		t.Fatal("expected runRuntimeApply() to fail for unknown fields")
	}

	if !strings.Contains(err.Error(), "invalid JSON input") {
		t.Fatalf("expected unknown field error, got %v", err)
	}
}
