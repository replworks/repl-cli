package runtime

import (
	"os"
	"testing"
)

func TestInit(t *testing.T) {
	// Clean up any existing .repl directory
	os.RemoveAll(ReplDir)
	defer os.RemoveAll(ReplDir)

	// Test initialization
	err := Init()
	if err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	// Verify directories created
	if !Exists() {
		t.Error("Init() did not create .repl directory")
	}

	if !RuntimeExists() {
		t.Error("Init() did not create .repl/runtime directory")
	}

	// Verify state files exist
	if _, err := ReadState(); err != nil {
		t.Errorf("Init() did not create execution-state.json: %v", err)
	}

	if _, err := ReadProgress(); err != nil {
		t.Errorf("Init() did not create task-progress.json: %v", err)
	}

	if _, err := ReadLog(); err != nil {
		t.Errorf("Init() did not create execution-log.json: %v", err)
	}
}

func TestExists(t *testing.T) {
	// Clean up
	os.RemoveAll(ReplDir)
	defer os.RemoveAll(ReplDir)

	// Should not exist initially
	if Exists() {
		t.Error("Exists() returned true when .repl does not exist")
	}

	// Create directory
	os.MkdirAll(ReplDir, 0755)

	// Should exist now
	if !Exists() {
		t.Error("Exists() returned false when .repl exists")
	}
}

func TestRuntimeExists(t *testing.T) {
	// Clean up
	os.RemoveAll(ReplDir)
	defer os.RemoveAll(ReplDir)

	// Should not exist initially
	if RuntimeExists() {
		t.Error("RuntimeExists() returned true when .repl/runtime does not exist")
	}

	// Create directories
	os.MkdirAll(RuntimeDir, 0755)

	// Should exist now
	if !RuntimeExists() {
		t.Error("RuntimeExists() returned false when .repl/runtime exists")
	}
}

func TestReadWriteState(t *testing.T) {
	// Clean up
	os.RemoveAll(ReplDir)
	defer os.RemoveAll(ReplDir)

	// Initialize
	if err := Init(); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	// Test reading initial state
	state, err := ReadState()
	if err != nil {
		t.Fatalf("ReadState() failed: %v", err)
	}

	if state.SessionActive {
		t.Error("Initial state should have SessionActive=false")
	}

	// Test writing state
	state.SessionActive = true
	state.CurrentTask = "TASK_1"
	if err := WriteState(state); err != nil {
		t.Fatalf("WriteState() failed: %v", err)
	}

	// Verify written state
	state, err = ReadState()
	if err != nil {
		t.Fatalf("ReadState() failed after write: %v", err)
	}

	if !state.SessionActive {
		t.Error("State.SessionActive should be true after write")
	}

	if state.CurrentTask != "TASK_1" {
		t.Errorf("State.CurrentTask should be 'TASK_1', got '%s'", state.CurrentTask)
	}
}

func TestReadWriteProgress(t *testing.T) {
	// Clean up
	os.RemoveAll(ReplDir)
	defer os.RemoveAll(ReplDir)

	// Initialize
	if err := Init(); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	// Test reading initial progress
	progress, err := ReadProgress()
	if err != nil {
		t.Fatalf("ReadProgress() failed: %v", err)
	}

	if len(progress.Tasks) != 0 {
		t.Error("Initial progress should have no tasks")
	}

	// Test writing progress
	progress.Tasks["TASK_1"] = TaskStatus{Status: "done"}
	progress.Tasks["TASK_2"] = TaskStatus{Status: "blocked"}
	if err := WriteProgress(progress); err != nil {
		t.Fatalf("WriteProgress() failed: %v", err)
	}

	// Verify written progress
	progress, err = ReadProgress()
	if err != nil {
		t.Fatalf("ReadProgress() failed after write: %v", err)
	}

	if len(progress.Tasks) != 2 {
		t.Errorf("Progress should have 2 tasks, got %d", len(progress.Tasks))
	}

	if progress.Tasks["TASK_1"].Status != "done" {
		t.Errorf("TASK_1 status should be 'done', got '%s'", progress.Tasks["TASK_1"].Status)
	}

	if progress.Tasks["TASK_2"].Status != "blocked" {
		t.Errorf("TASK_2 status should be 'blocked', got '%s'", progress.Tasks["TASK_2"].Status)
	}
}

func TestReadWriteLog(t *testing.T) {
	// Clean up
	os.RemoveAll(ReplDir)
	defer os.RemoveAll(ReplDir)

	// Initialize
	if err := Init(); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	// Test reading initial log
	log, err := ReadLog()
	if err != nil {
		t.Fatalf("ReadLog() failed: %v", err)
	}

	if len(log.Logs) != 0 {
		t.Error("Initial log should be empty")
	}

	// Test writing log
	log.Logs = []string{"entry1", "entry2", "entry3"}
	if err := WriteLog(log); err != nil {
		t.Fatalf("WriteLog() failed: %v", err)
	}

	// Verify written log
	log, err = ReadLog()
	if err != nil {
		t.Fatalf("ReadLog() failed after write: %v", err)
	}

	if len(log.Logs) != 3 {
		t.Errorf("Log should have 3 entries, got %d", len(log.Logs))
	}

	if log.Logs[0] != "entry1" {
		t.Errorf("First log entry should be 'entry1', got '%s'", log.Logs[0])
	}
}

func TestAddLog(t *testing.T) {
	// Clean up
	os.RemoveAll(ReplDir)
	defer os.RemoveAll(ReplDir)

	// Initialize
	if err := Init(); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	// Add log entries
	if err := AddLog("first entry"); err != nil {
		t.Fatalf("AddLog() failed: %v", err)
	}

	if err := AddLog("second entry"); err != nil {
		t.Fatalf("AddLog() failed: %v", err)
	}

	// Verify log
	log, err := ReadLog()
	if err != nil {
		t.Fatalf("ReadLog() failed: %v", err)
	}

	if len(log.Logs) != 2 {
		t.Errorf("Log should have 2 entries, got %d", len(log.Logs))
	}

	if log.Logs[0] != "first entry" {
		t.Errorf("First log entry should be 'first entry', got '%s'", log.Logs[0])
	}

	if log.Logs[1] != "second entry" {
		t.Errorf("Second log entry should be 'second entry', got '%s'", log.Logs[1])
	}
}

func TestReset(t *testing.T) {
	// Clean up
	os.RemoveAll(ReplDir)
	defer os.RemoveAll(ReplDir)

	// Initialize
	if err := Init(); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	// Modify state
	state, _ := ReadState()
	state.SessionActive = true
	WriteState(state)

	progress, _ := ReadProgress()
	progress.Tasks["TASK_1"] = TaskStatus{Status: "done"}
	WriteProgress(progress)

	log, _ := ReadLog()
	log.Logs = []string{"test entry"}
	WriteLog(log)

	// Reset
	if err := Reset(); err != nil {
		t.Fatalf("Reset() failed: %v", err)
	}

	// Verify state is reset
	state, err := ReadState()
	if err != nil {
		t.Fatalf("ReadState() failed after reset: %v", err)
	}

	if state.SessionActive {
		t.Error("State.SessionActive should be false after reset")
	}

	progress, err = ReadProgress()
	if err != nil {
		t.Fatalf("ReadProgress() failed after reset: %v", err)
	}

	if len(progress.Tasks) != 0 {
		t.Errorf("Progress should have no tasks after reset, got %d", len(progress.Tasks))
	}

	log, err = ReadLog()
	if err != nil {
		t.Fatalf("ReadLog() failed after reset: %v", err)
	}

	if len(log.Logs) != 0 {
		t.Errorf("Log should be empty after reset, got %d entries", len(log.Logs))
	}
}

func TestValidate(t *testing.T) {
	// Clean up
	os.RemoveAll(ReplDir)
	defer os.RemoveAll(ReplDir)

	// Should fail when not initialized
	if err := Validate(); err == nil {
		t.Error("Validate() should fail when .repl does not exist")
	}

	// Initialize
	if err := Init(); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	// Should pass when initialized
	if err := Validate(); err != nil {
		t.Errorf("Validate() should pass when initialized, got error: %v", err)
	}

	// Corrupt state file
	os.WriteFile(StateFile, []byte("invalid json"), 0644)

	// Should fail with corrupted state
	if err := Validate(); err == nil {
		t.Error("Validate() should fail when state file is corrupted")
	}
}
