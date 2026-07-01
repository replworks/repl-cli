package main

import (
	"fmt"

	rt "repl-cli/internal/runtime"

	"github.com/spf13/cobra"
)

func newResetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "reset",
		Short: "Reset runtime environment to clean state",
		Long:  "Clear .repl/runtime/ state and return to initial state",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runReset()
		},
	}
}

func runReset() error {
	// TASK_3 - repl reset
	// Check if .repl directory exists
	if !rt.Exists() {
		return fmt.Errorf("REPL project not initialized. Run 'repl init' first")
	}

	fmt.Println("Resetting REPL runtime environment...")

	// Reset runtime environment
	if err := rt.Reset(); err != nil {
		return fmt.Errorf("failed to reset runtime: %w", err)
	}

	fmt.Println("Runtime environment reset successfully")
	fmt.Printf("Cleared: %s/\n", rt.RuntimeDir)
	fmt.Println("Restored: execution-state.json")
	fmt.Println("Restored: task-progress.json")
	fmt.Println("Restored: execution-log.json")

	return nil
}
