package command

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func nextProject(t *testing.T, tasks string) string {
	t.Helper()

	dir := t.TempDir()
	writeReplworksFile(t, dir, "PRODUCT_SPEC.md", "- FR-001: one\n- AC-010: a\n- AC-011: b\n- AC-012: c\n")
	writeReplworksFile(t, dir, "ARCHITECTURE.md", "### C3. Third\n- C3 body\n### C4. Fourth\n")
	writeTasksFile(t, dir, tasks)
	t.Chdir(dir)

	return dir
}

func TestNextPrintsTaskThenDefinitions(t *testing.T) {
	nextProject(t, "- [x] T-001 One\n\n- [ ] T-002 Two\n  - Satisfies: FR-001; C3\n")

	code, stdout, _ := runApp(t, "next")

	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (stdout = %q)", code, stdout)
	}
	order := []string{"- [ ] T-002 Two", "  - Satisfies: FR-001; C3", "- FR-001: one", "### C3. Third", "- C3 body"}
	last := -1
	for _, want := range order {
		i := strings.Index(stdout, want)
		if i <= last {
			t.Fatalf("stdout = %q, want %q in order", stdout, order)
		}
		last = i
	}
	if strings.Contains(stdout, "C4") {
		t.Fatalf("stdout = %q, want the C3 section to end before C4", stdout)
	}
}

func TestNextReportsNoIncompleteTask(t *testing.T) {
	nextProject(t, "- [x] T-001 One\n")

	code, stdout, _ := runApp(t, "next")

	if code != 0 || !strings.Contains(stdout, "no incomplete task") {
		t.Fatalf("exit code = %d, stdout = %q", code, stdout)
	}
}

func TestNextReportsMissingWithExitZero(t *testing.T) {
	nextProject(t, "- [ ] T-001 One\n  - AC-999\n")

	code, stdout, _ := runApp(t, "next")

	if code != 0 || !strings.Contains(stdout, "AC-999: no definition") {
		t.Fatalf("exit code = %d, stdout = %q", code, stdout)
	}
}

func TestNextExpandsRangesAndIgnoresFences(t *testing.T) {
	nextProject(t, "- [ ] T-001 One\n  - AC-010 through AC-012\n  ```\n  FR-001\n  ```\n")

	_, stdout, _ := runApp(t, "next")

	for _, want := range []string{"- AC-010: a", "- AC-011: b", "- AC-012: c"} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("stdout = %q, want %q", stdout, want)
		}
	}
	if strings.Contains(stdout, "- FR-001: one") {
		t.Fatalf("stdout = %q, want no FR-001 excerpt", stdout)
	}
}

func TestNextWithoutTasksExitsTwo(t *testing.T) {
	t.Chdir(t.TempDir())

	code, stdout, _ := runApp(t, "next")

	if code != 2 || strings.Count(stdout, "repl: ") != 1 {
		t.Fatalf("exit code = %d, stdout = %q", code, stdout)
	}
}

func TestNextUnreadableArchitecture(t *testing.T) {
	dir := nextProject(t, "- [ ] T-001 One\n  - C3\n")
	path := filepath.Join(dir, ".replworks", "ARCHITECTURE.md")
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(path, 0o755); err != nil {
		t.Fatal(err)
	}

	code, stdout, _ := runApp(t, "next")

	if code != 2 || strings.Count(stdout, "repl: ") != 1 || strings.Contains(stdout, "C3: no definition") {
		t.Fatalf("exit code = %d, stdout = %q", code, stdout)
	}
}

func TestNextWritesNothing(t *testing.T) {
	dir := nextProject(t, "- [ ] T-001 One\n  - FR-001 AC-999\n")
	before := snapshotTree(t, dir)

	runApp(t, "next")

	if after := snapshotTree(t, dir); after != before {
		t.Fatalf("tree changed")
	}
}

func TestNextJSON(t *testing.T) {
	nextProject(t, "- [ ] T-001 One\n  - FR-001, C3\n")

	code, stdout, _ := runApp(t, "next", "--json")

	if code != 0 || strings.Count(stdout, "\n") != 1 || strings.Contains(stdout, "\x1b") {
		t.Fatalf("exit code = %d, stdout = %q", code, stdout)
	}
	var got struct {
		Task *struct {
			ID   string `json:"id"`
			Path string `json:"path"`
			Line int    `json:"line"`
			Text string `json:"text"`
		} `json:"task"`
		Definitions []struct {
			ID string `json:"id"`
		} `json:"definitions"`
		Missing []string `json:"missing"`
		Errors  []string `json:"errors"`
	}
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("stdout = %q: %v", stdout, err)
	}
	if got.Task == nil || got.Task.ID != "T-001" || got.Task.Path != ".replworks/TASKS.md" || got.Task.Line != 1 || got.Task.Text != "- [ ] T-001 One\n  - FR-001, C3" {
		t.Fatalf("task = %+v", got.Task)
	}
	if len(got.Definitions) != 2 || got.Definitions[0].ID != "FR-001" || got.Definitions[1].ID != "C3" || got.Missing == nil || got.Errors == nil {
		t.Fatalf("got = %+v", got)
	}
}

func TestNextJSONNoTask(t *testing.T) {
	nextProject(t, "- [x] T-001 One\n")

	code, stdout, _ := runApp(t, "next", "--json")

	if code != 0 || strings.TrimSpace(stdout) != `{"task":null,"definitions":[],"missing":[],"errors":[]}` {
		t.Fatalf("exit code = %d, stdout = %q", code, stdout)
	}
}
