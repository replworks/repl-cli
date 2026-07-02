package runtime

import (
	"bufio"
	"encoding/json"
	"fmt"
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
	defer func() {
		_ = file.Close()
	}()

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
				taskID = strings.TrimRight(taskID, "—")
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
	defer func() {
		_ = file.Close()
	}()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(data)
}

func ReadState() (*ExecutionState, error) {
	file, err := os.Open(StateFile)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = file.Close()
	}()

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
	defer func() {
		_ = file.Close()
	}()

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
	defer func() {
		_ = file.Close()
	}()

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

// CopyAgentMD copies templates/.repl/agent.md to .repl/agent.md.
func CopyAgentMD() error {
	const src = "templates/.repl/agent.md"
	const dst = ".repl/agent.md"

	data, err := os.ReadFile(src)
	if err != nil {
		return fmt.Errorf("failed to read template %s: %w", src, err)
	}

	if err := os.WriteFile(dst, data, 0644); err != nil {
		return fmt.Errorf("failed to write %s: %w", dst, err)
	}

	return nil
}

// CopyOrPrependAgentsMD copies templates/AGENTS.md to AGENTS.md.
// If AGENTS.md already exists, the template content (excluding the first heading line)
// is prepended to the existing file.
func CopyOrPrependAgentsMD() (bool, error) {
	const src = "templates/AGENTS.md"
	const dst = "AGENTS.md"

	templateData, err := os.ReadFile(src)
	if err != nil {
		return false, fmt.Errorf("failed to read template %s: %w", src, err)
	}

	// Check if AGENTS.md already exists
	existingData, err := os.ReadFile(dst)
	if err != nil && !os.IsNotExist(err) {
		return false, fmt.Errorf("failed to read %s: %w", dst, err)
	}

	if os.IsNotExist(err) {
		// File does not exist: plain copy
		if err := os.WriteFile(dst, templateData, 0644); err != nil {
			return false, fmt.Errorf("failed to write %s: %w", dst, err)
		}
		return false, nil
	}

	// File exists: strip the first heading line from the template, then prepend
	templateLines := strings.Split(string(templateData), "\n")
	var contentLines []string
	for _, line := range templateLines {
		if strings.HasPrefix(strings.TrimSpace(line), "# AGENTS") {
			continue
		}
		contentLines = append(contentLines, line)
	}

	// Remove leading blank lines after heading removal
	for len(contentLines) > 0 && strings.TrimSpace(contentLines[0]) == "" {
		contentLines = contentLines[1:]
	}

	prependContent := strings.Join(contentLines, "\n")
	merged := prependContent + "\n" + string(existingData)

	if err := os.WriteFile(dst, []byte(merged), 0644); err != nil {
		return false, fmt.Errorf("failed to write %s: %w", dst, err)
	}

	return true, nil
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
