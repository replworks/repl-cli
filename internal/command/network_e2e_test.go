//go:build e2e

package command

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/replworks/repl-cli/internal/canon"
)

func TestCommandsRequestOnlyTheCanonicalSource(t *testing.T) {
	tests := []struct {
		name         string
		agents       string
		documents    bool
		wantRequests int
	}{
		{name: "init", wantRequests: 1},
		{name: "status", documents: true, wantRequests: 0},
		{name: "check", documents: true, wantRequests: 1},
		{name: "sync", agents: "# stale\n", documents: true, wantRequests: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			if tt.documents {
				agents := tt.agents
				if agents == "" {
					agents = liveCanonical(t)
				}
				writeDocuments(t, dir, agents)
				writeTasks(t, dir, "- [ ] T-001 first\n")
			}
			if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("readme\n"), 0o644); err != nil {
				t.Fatalf("setup: %v", err)
			}
			initRepository(t, dir)
			t.Chdir(dir)

			client, transport := liveRecordingClient()

			if code, _, stderr := runCommand(t, client, tt.name); code != 0 {
				t.Fatalf("exit code = %d, want 0 (stderr = %q)", code, stderr)
			}

			requests := transport.requested()
			if len(requests) != tt.wantRequests {
				t.Fatalf("requests = %q, want %d", requests, tt.wantRequests)
			}
			for _, request := range requests {
				if request != canon.SourceURL {
					t.Fatalf("requested %q, want %q", request, canon.SourceURL)
				}
			}
		})
	}
}
