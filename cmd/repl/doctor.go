package main

import (
	"fmt"

	rt "repl-cli/internal/runtime"

	"github.com/spf13/cobra"
)

func newDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Validate REPL system integrity",
		Long:  "Check runtime environment and configuration integrity",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runDoctor()
		},
	}
}

func runDoctor() error {
	// TASK_2 - repl doctor
	fmt.Println("Checking REPL system integrity...")

	// Check if .repl directory exists
	if !rt.Exists() {
		return fmt.Errorf("FAIL: .repl/ directory does not exist")
	}
	fmt.Println("✓ .repl/ directory exists")

	// Check if .repl/runtime directory exists
	if !rt.RuntimeExists() {
		return fmt.Errorf("FAIL: .repl/runtime/ directory does not exist")
	}
	fmt.Println("✓ .repl/runtime/ directory exists")

	// Validate runtime state files
	if err := rt.Validate(); err != nil {
		return fmt.Errorf("FAIL: Runtime state validation failed: %w", err)
	}
	fmt.Println("✓ execution-state.json is valid")
	fmt.Println("✓ task-progress.json is valid")
	fmt.Println("✓ execution-log.json is valid")

	fmt.Println("\nREPL system is healthy")
	return nil
}
