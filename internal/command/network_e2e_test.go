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
		wantRequests []string
	}{
		{name: "init", wantRequests: []string{canon.SourceURL, string(canon.AskHumanScript)}},
		{name: "status", documents: true},
		{name: "check", documents: true, wantRequests: []string{canon.SourceURL}},
		{name: "sync", agents: "# stale\n", documents: true, wantRequests: []string{canon.SourceURL, string(canon.AskHumanScript)}},
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
			if len(requests) != len(tt.wantRequests) {
				t.Fatalf("requests = %q, want %q", requests, tt.wantRequests)
			}
			for i, request := range requests {
				if request != tt.wantRequests[i] {
					t.Fatalf("requests = %q, want %q", requests, tt.wantRequests)
				}
			}
		})
	}
}
