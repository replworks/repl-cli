package main

import (
	"fmt"

	rt "repl-cli/internal/runtime"

	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize REPL project runtime environment",
		Long:  "Create .repl/ directory structure and initialize default runtime state",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit()
		},
	}
}

func runInit() error {
	// TASK_1 - repl init
	// Check if already initialized
	if rt.Exists() {
		return fmt.Errorf("REPL project already initialized")
	}

	// Initialize runtime environment
	if err := rt.Init(); err != nil {
		return fmt.Errorf("failed to initialize REPL runtime: %w", err)
	}

	fmt.Println("REPL project initialized successfully")
	fmt.Printf("Created: %s/\n", rt.ReplDir)
	fmt.Printf("Created: %s/\n", rt.RuntimeDir)
	fmt.Println("Initialized: execution-state.json")
	fmt.Println("Initialized: task-progress.json")
	fmt.Println("Initialized: execution-log.json")

	// Copy templates/.repl/agent.md → .repl/agent.md
	if err := rt.CopyAgentMD(); err != nil {
		return fmt.Errorf("failed to copy agent.md: %w", err)
	}
	fmt.Println("Copied: .repl/agent.md")

	// Copy templates/AGENTS.md → AGENTS.md (prepend if already exists)
	prepended, err := rt.CopyOrPrependAgentsMD()
	if err != nil {
		return fmt.Errorf("failed to copy AGENTS.md: %w", err)
	}
	if prepended {
		fmt.Println("Prepended: AGENTS.md")
	} else {
		fmt.Println("Copied: AGENTS.md")
	}

	return nil
}
