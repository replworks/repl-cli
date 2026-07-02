package main

import (
	"os"
	"testing"
)

func TestInitCommand(t *testing.T) {
	// Clean up any existing .repl directory
	_ = os.RemoveAll(".repl")
	defer func() { _ = os.RemoveAll(".repl") }()

	// Create the init command
	cmd := newInitCmd()

	// Test that command is properly configured
	if cmd.Use != "init" {
		t.Errorf("Expected command use to be 'init', got '%s'", cmd.Use)
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

func TestInitCommandExecution(t *testing.T) {
	// Clean up any existing .repl directory and generated files
	_ = os.RemoveAll(".repl")
	_ = os.Remove("AGENTS.md")
	defer func() {
		_ = os.RemoveAll(".repl")
		_ = os.Remove("AGENTS.md")
		_ = os.RemoveAll("templates")
	}()

	// Set up template fixtures required by init
	_ = os.MkdirAll("templates/.repl", 0755)
	_ = os.WriteFile("templates/.repl/agent.md", []byte("# agent.md\n"), 0644)
	_ = os.MkdirAll("templates", 0755)
	_ = os.WriteFile("templates/AGENTS.md", []byte("# AGENTS.md\n\n## Rule\n\nread agent.md\n"), 0644)

	// Execute the init command
	cmd := newInitCmd()
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("Init command failed: %v", err)
	}

	// Verify .repl directory was created
	if _, err := os.Stat(".repl"); os.IsNotExist(err) {
		t.Error("Expected .repl directory to be created")
	}

	// Verify .repl/runtime directory was created
	if _, err := os.Stat(".repl/runtime"); os.IsNotExist(err) {
		t.Error("Expected .repl/runtime directory to be created")
	}

	// Verify execution-state.json was created
	if _, err := os.Stat(".repl/runtime/execution-state.json"); os.IsNotExist(err) {
		t.Error("Expected execution-state.json to be created")
	}

	// Verify task-progress.json was created
	if _, err := os.Stat(".repl/runtime/task-progress.json"); os.IsNotExist(err) {
		t.Error("Expected task-progress.json to be created")
	}

	// Verify execution-log.json was created
	if _, err := os.Stat(".repl/runtime/execution-log.json"); os.IsNotExist(err) {
		t.Error("Expected execution-log.json to be created")
	}

	// Verify .repl/agent.md was copied
	if _, err := os.Stat(".repl/agent.md"); os.IsNotExist(err) {
		t.Error("Expected .repl/agent.md to be created")
	}

	// Verify AGENTS.md was copied
	if _, err := os.Stat("AGENTS.md"); os.IsNotExist(err) {
		t.Error("Expected AGENTS.md to be created")
	}
}


func TestInitCommandAlreadyInitialized(t *testing.T) {
	// Clean up any existing .repl directory
	_ = os.RemoveAll(".repl")
	defer func() { _ = os.RemoveAll(".repl") }()

	// Create .repl directory to simulate already initialized project
	_ = os.MkdirAll(".repl", 0755)

	// Execute the init command
	cmd := newInitCmd()
	cmd.SetArgs([]string{})

	err := cmd.Execute()
	if err == nil {
		t.Error("Expected init command to fail when already initialized")
	}
}
