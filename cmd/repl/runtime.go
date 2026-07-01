package main

import (
	"encoding/json"
	"fmt"
	"os"

	rt "repl-cli/internal/runtime"

	"github.com/spf13/cobra"
)

func newRuntimeCmd() *cobra.Command {
	runtimeCmd := &cobra.Command{
		Use:   "runtime",
		Short: "Runtime execution management",
		Long:  "Manage runtime execution sessions and state",
	}

	runtimeCmd.AddCommand(newRuntimeStartCmd())
	runtimeCmd.AddCommand(newRuntimeStopCmd())
	runtimeCmd.AddCommand(newRuntimeApplyCmd())
	runtimeCmd.AddCommand(newRuntimeStatusCmd())

	return runtimeCmd
}

func newRuntimeStartCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "start",
		Short: "Start execution session",
		Long:  "Activate execution session and prepare context for AI",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRuntimeStart()
		},
	}
}

func newRuntimeStopCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "Stop execution session",
		Long:  "Terminate execution session without state mutation",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRuntimeStop()
		},
	}
}

func newRuntimeApplyCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "apply",
		Short: "Apply AI execution result",
		Long:  "Apply external AI-generated execution result to runtime state",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRuntimeApply()
		},
	}
}

func newRuntimeStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Display runtime state",
		Long:  "Show current runtime state and task progress",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRuntimeStatus()
		},
	}
}

func runRuntimeStart() error {
	// TASK_4 - repl runtime start
	// Check if .repl directory exists
	if !rt.Exists() {
		return fmt.Errorf("REPL project not initialized. Run 'repl init' first")
	}

	fmt.Println("Starting REPL runtime execution session...")

	// Read current state
	state, err := rt.ReadState()
	if err != nil {
		return fmt.Errorf("failed to read execution state: %w", err)
	}

	// Check if session is already active
	if state.SessionActive {
		return fmt.Errorf("runtime session is already active")
	}

	// Activate session
	state.SessionActive = true
	if err := rt.WriteState(state); err != nil {
		return fmt.Errorf("failed to update execution state: %w", err)
	}

	// Add log entry
	if err := rt.AddLog("Runtime session started"); err != nil {
		return fmt.Errorf("failed to add log entry: %w", err)
	}

	fmt.Println("Runtime execution session activated")
	fmt.Println("System is ready for AI execution input")
	fmt.Println("\nNext step: Use 'repl runtime apply' to submit AI execution results")

	return nil
}

func runRuntimeStop() error {
	// TASK_5 - repl runtime stop
	fmt.Println("Stopping REPL runtime execution session...")

	// Check if .repl directory exists
	if !rt.Exists() {
		return fmt.Errorf("REPL project not initialized. Run 'repl init' first")
	}

	// Read current state
	state, err := rt.ReadState()
	if err != nil {
		return fmt.Errorf("failed to read execution state: %w", err)
	}

	// Check if session is active
	if !state.SessionActive {
		return fmt.Errorf("no active runtime session to stop")
	}

	// Deactivate session
	state.SessionActive = false
	if err := rt.WriteState(state); err != nil {
		return fmt.Errorf("failed to update execution state: %w", err)
	}

	// Add log entry
	if err := rt.AddLog("Runtime session stopped"); err != nil {
		return fmt.Errorf("failed to add log entry: %w", err)
	}

	fmt.Println("Runtime execution session terminated")
	fmt.Println("No active task remains in execution mode")

	return nil
}

func runRuntimeApply() error {
	// TASK_6 - repl runtime apply
	fmt.Println("Applying AI execution result...")

	// Check if .repl directory exists
	if !rt.Exists() {
		return fmt.Errorf("REPL project not initialized. Run 'repl init' first")
	}

	// Read JSON from stdin
	var input struct {
		Action string   `json:"action"`
		TaskID string   `json:"taskId"`
		Status string   `json:"status"`
		Reason string   `json:"reason,omitempty"`
		Events []string `json:"events,omitempty"`
	}

	decoder := json.NewDecoder(os.Stdin)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		return fmt.Errorf("invalid JSON input: %w", err)
	}

	// Validate action
	if input.Action != "update_runtime" {
		return fmt.Errorf("invalid action: must be 'update_runtime'")
	}

	// Validate taskId
	if input.TaskID == "" {
		return fmt.Errorf("taskId is required")
	}

	// Validate status
	if input.Status != "done" && input.Status != "blocked" {
		return fmt.Errorf("invalid status: must be 'done' or 'blocked'")
	}

	// If status is blocked, reason is required
	if input.Status == "blocked" && input.Reason == "" {
		return fmt.Errorf("reason is required when status is 'blocked'")
	}

	// Read current progress
	progress, err := rt.ReadProgress()
	if err != nil {
		return fmt.Errorf("failed to read task progress: %w", err)
	}

	// Update task status
	progress.Tasks[input.TaskID] = rt.TaskStatus{
		Status: input.Status,
	}

	// Write updated progress
	if err := rt.WriteProgress(progress); err != nil {
		return fmt.Errorf("failed to update task progress: %w", err)
	}

	// Add log entries
	logMessages := []string{
		fmt.Sprintf("Runtime apply executed for task: %s", input.TaskID),
		fmt.Sprintf("Task status updated to: %s", input.Status),
	}
	if input.Reason != "" {
		logMessages = append(logMessages, fmt.Sprintf("Reason: %s", input.Reason))
	}
	if len(input.Events) > 0 {
		logMessages = append(logMessages, fmt.Sprintf("Events: %v", input.Events))
	}

	for _, msg := range logMessages {
		if err := rt.AddLog(msg); err != nil {
			return fmt.Errorf("failed to add log entry: %w", err)
		}
	}

	// Output result
	fmt.Printf("Task %s marked as: %s\n", input.TaskID, input.Status)
	if input.Status == "done" {
		fmt.Println("Task completed successfully")
	} else {
		fmt.Printf("Task blocked: %s\n", input.Reason)
	}

	return nil
}

func runRuntimeStatus() error {
	// TASK_7 - repl runtime status
	fmt.Println("REPL Runtime Status")
	fmt.Println("===================")

	// Check if .repl directory exists
	if !rt.Exists() {
		return fmt.Errorf("REPL project not initialized. Run 'repl init' first")
	}

	// Read execution state
	state, err := rt.ReadState()
	if err != nil {
		return fmt.Errorf("failed to read execution state: %w", err)
	}

	// Display session status
	fmt.Println("\nSession Status:")
	if state.SessionActive {
		fmt.Println("  State: active")
	} else {
		fmt.Println("  State: inactive")
	}

	// Display current task if any
	if state.CurrentTask != "" {
		fmt.Printf("  Current Task: %s\n", state.CurrentTask)
	} else {
		fmt.Println("  Current Task: none")
	}

	// Read and display task progress
	progress, err := rt.ReadProgress()
	if err != nil {
		return fmt.Errorf("failed to read task progress: %w", err)
	}

	fmt.Println("\nTask Progress:")
	if len(progress.Tasks) == 0 {
		fmt.Println("  No tasks recorded")
	} else {
		for taskID, taskStatus := range progress.Tasks {
			fmt.Printf("  %s: %s\n", taskID, taskStatus.Status)
		}
	}

	// Read and display recent logs
	log, err := rt.ReadLog()
	if err != nil {
		return fmt.Errorf("failed to read execution log: %w", err)
	}

	fmt.Println("\nExecution Log (last 5 entries):")
	if len(log.Logs) == 0 {
		fmt.Println("  No log entries")
	} else {
		start := len(log.Logs) - 5
		if start < 0 {
			start = 0
		}
		for i := start; i < len(log.Logs); i++ {
			fmt.Printf("  - %s\n", log.Logs[i])
		}
	}

	fmt.Println("\nRuntime state is readable and consistent")
	return nil
}
