package runtime

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	ReplDir      = ".repl"
	RuntimeDir   = ".repl/runtime"
	StateFile    = ".repl/runtime/execution-state.json"
	ProgressFile = ".repl/runtime/task-progress.json"
	LogFile      = ".repl/runtime/execution-log.json"
)

type ExecutionState struct {
	SessionActive bool   `json:"session_active"`
	CurrentTask   string `json:"current_task,omitempty"`
}

type TaskProgress struct {
	Tasks map[string]TaskStatus `json:"tasks"`
}

type TaskStatus struct {
	Status string `json:"status"`
}

type ExecutionLog struct {
	Logs []string `json:"logs"`
}

func Init() error {
	// Create .repl directory
	if err := os.MkdirAll(ReplDir, 0755); err != nil {
		return fmt.Errorf("failed to create .repl directory: %w", err)
	}

	// Create .repl/runtime directory
	if err := os.MkdirAll(RuntimeDir, 0755); err != nil {
		return fmt.Errorf("failed to create .repl/runtime directory: %w", err)
	}

	// Initialize execution-state.json
	state := ExecutionState{
		SessionActive: false,
	}
	if err := writeJSON(StateFile, state); err != nil {
		return fmt.Errorf("failed to initialize execution-state.json: %w", err)
	}

	// Initialize task-progress.json with tasks from tasks.md
	progress, err := loadTasksFromTasksMD()
	if err != nil {
		return fmt.Errorf("failed to load tasks from tasks.md: %w", err)
	}
	if err := writeJSON(ProgressFile, progress); err != nil {
		return fmt.Errorf("failed to initialize task-progress.json: %w", err)
	}

	// Initialize execution-log.json
	log := ExecutionLog{
		Logs: []string{},
	}
	if err := writeJSON(LogFile, log); err != nil {
		return fmt.Errorf("failed to initialize execution-log.json: %w", err)
	}

	return nil
}

func copyTasksMD() error {
	// Check if tasks.md already exists in .repl directory
	dstPath := filepath.Join(ReplDir, "tasks.md")
	if _, err := os.Stat(dstPath); err == nil {
		// tasks.md already exists in .repl/, no need to copy
		return nil
	}

	// Check if tasks.md exists in current directory
	srcPath := "tasks.md"
	if _, err := os.Stat(srcPath); os.IsNotExist(err) {
		// If tasks.md doesn't exist in current directory, nothing to copy
		return nil
	}

	// Copy tasks.md to .repl/tasks.md
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("failed to open source tasks.md: %w", err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dstPath)
	if err != nil {
		return fmt.Errorf("failed to create destination tasks.md: %w", err)
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return fmt.Errorf("failed to copy tasks.md: %w", err)
	}

	return nil
}

func loadTasksFromTasksMD() (TaskProgress, error) {
	progress := TaskProgress{
		Tasks: make(map[string]TaskStatus),
	}

	// Open tasks.md file from .repl directory
	tasksPath := filepath.Join(ReplDir, "tasks.md")
	file, err := os.Open(tasksPath)
	if err != nil {
		// If tasks.md doesn't exist, return empty progress
		return progress, nil
	}
	defer file.Close()

	// Parse tasks.md to extract task IDs
	scanner := bufio.NewScanner(file)
	lineCount := 0
	taskCount := 0
	for scanner.Scan() {
		lineCount++
		line := strings.TrimSpace(scanner.Text())
		// Look for lines like "# TASK_1 — repl init" or "# TASK_1 - repl init"
		if strings.HasPrefix(line, "# TASK_") {
			taskCount++
			// Extract task ID
			// Line format: "# TASK_1 — repl init" or "# TASK_1 - repl init"
			// Remove "# " prefix
			lineWithoutHash := strings.TrimPrefix(line, "# ")
			// Split by space
			parts := strings.Split(lineWithoutHash, " ")
			if len(parts) >= 1 {
				taskID := parts[0]
				// Remove any trailing em-dash or dash
				taskID = strings.TrimRight(taskID, "——-")
				taskID = strings.TrimSpace(taskID)
				if taskID != "" {
					progress.Tasks[taskID] = TaskStatus{Status: "pending"}
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return progress, fmt.Errorf("error reading tasks.md: %w", err)
	}

	// Only print in non-test environment
	if os.Getenv("REPL_DEBUG") != "" {
		fmt.Printf("DEBUG: Scanned %d lines, found %d tasks\n", lineCount, taskCount)
	}

	return progress, nil
}

func Exists() bool {
	_, err := os.Stat(ReplDir)
	return err == nil
}

func RuntimeExists() bool {
	_, err := os.Stat(RuntimeDir)
	return err == nil
}

func writeJSON(path string, data interface{}) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func ReadState() (*ExecutionState, error) {
	file, err := os.Open(StateFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var state ExecutionState
	if err := json.NewDecoder(file).Decode(&state); err != nil {
		return nil, err
	}

	return &state, nil
}

func WriteState(state *ExecutionState) error {
	return writeJSON(StateFile, state)
}

func ReadProgress() (*TaskProgress, error) {
	file, err := os.Open(ProgressFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var progress TaskProgress
	if err := json.NewDecoder(file).Decode(&progress); err != nil {
		return nil, err
	}

	return &progress, nil
}

func WriteProgress(progress *TaskProgress) error {
	return writeJSON(ProgressFile, progress)
}

func ReadLog() (*ExecutionLog, error) {
	file, err := os.Open(LogFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var log ExecutionLog
	if err := json.NewDecoder(file).Decode(&log); err != nil {
		return nil, err
	}

	return &log, nil
}

func WriteLog(log *ExecutionLog) error {
	return writeJSON(LogFile, log)
}

func AddLog(message string) error {
	log, err := ReadLog()
	if err != nil {
		return err
	}

	log.Logs = append(log.Logs, message)
	return WriteLog(log)
}

func Reset() error {
	// Remove runtime directory
	if err := os.RemoveAll(RuntimeDir); err != nil {
		return fmt.Errorf("failed to remove runtime directory: %w", err)
	}

	// Recreate runtime directory
	if err := os.MkdirAll(RuntimeDir, 0755); err != nil {
		return fmt.Errorf("failed to recreate runtime directory: %w", err)
	}

	// Reinitialize state files
	return Init()
}

func Validate() error {
	// Check if .repl directory exists
	if !Exists() {
		return fmt.Errorf(".repl directory does not exist")
	}

	// Check if .repl/runtime directory exists
	if !RuntimeExists() {
		return fmt.Errorf(".repl/runtime directory does not exist")
	}

	// Check if execution-state.json exists and is valid
	if _, err := ReadState(); err != nil {
		return fmt.Errorf("execution-state.json is missing or corrupted: %w", err)
	}

	// Check if task-progress.json exists and is valid
	if _, err := ReadProgress(); err != nil {
		return fmt.Errorf("task-progress.json is missing or corrupted: %w", err)
	}

	// Check if execution-log.json exists and is valid
	if _, err := ReadLog(); err != nil {
		return fmt.Errorf("execution-log.json is missing or corrupted: %w", err)
	}

	return nil
}
