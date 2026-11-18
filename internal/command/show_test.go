package command

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeReplworksFile(t *testing.T, dir, name, text string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Join(dir, ".replworks"), 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".replworks", name), []byte(text), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
}

func showProject(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	writeReplworksFile(t, dir, "PRODUCT_SPEC.md", ""+
		"- FR-001: one\n"+
		"- FR-002: two\n"+
		"  - first continuation\n"+
		"  - second continuation\n"+
		"- FR-003: three, not part of FR-002\n")
	writeReplworksFile(t, dir, "ARCHITECTURE.md", ""+
		"### C2. Second\n"+
		"- C2 body\n"+
		"\n"+
		"### C3. Third\n"+
		"- C3 body\n"+
		"## 8. Architectural Rules\n"+
		"- AR-03: Pure components.\n")
	writeTasksFile(t, dir, "- [ ] T-001 First\n  - Satisfies: FR-001\n")
	t.Chdir(dir)

	return dir
}

func TestShowPrintsContinuationLines(t *testing.T) {
	showProject(t)

	code, stdout, _ := runApp(t, "show", "FR-002")

	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (stdout = %q)", code, stdout)
	}
	for _, want := range []string{"FR-002 .replworks/PRODUCT_SPEC.md:2", "- FR-002: two", "  - first continuation", "  - second continuation"} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("stdout = %q, want it to contain %q", stdout, want)
		}
	}
	if strings.Contains(stdout, "FR-003") {
		t.Fatalf("stdout = %q, want no FR-003 line", stdout)
	}
}

func TestShowPrintsInGivenOrderOnce(t *testing.T) {
	showProject(t)

	code, stdout, _ := runApp(t, "show", "C2", "AR-03", "C2")

	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (stdout = %q)", code, stdout)
	}
	c2 := strings.Index(stdout, "### C2. Second")
	ar := strings.Index(stdout, "- AR-03: Pure components.")
	if c2 < 0 || ar < c2 || strings.Count(stdout, "### C2. Second") != 1 {
		t.Fatalf("stdout = %q, want C2 once and then AR-03", stdout)
	}
	if strings.Contains(stdout, "### C3.") {
		t.Fatalf("stdout = %q, want the C2 section to end before C3", stdout)
	}
}

func TestShowReportsMissingWithExitOne(t *testing.T) {
	showProject(t)

	code, stdout, _ := runApp(t, "show", "FR-001", "FR-999")

	if code != 1 {
		t.Fatalf("exit code = %d, want 1 (stdout = %q)", code, stdout)
	}
	if !strings.Contains(stdout, "- FR-001: one") || !strings.Contains(stdout, "FR-999") {
		t.Fatalf("stdout = %q, want FR-001 and a FR-999 line", stdout)
	}
}

func TestShowRejectsMissingOrInvalidArguments(t *testing.T) {
	for _, args := range [][]string{{"show"}, {"show", "foo"}, {"show", "FR-001", "foo"}} {
		dir := t.TempDir()
		// An unreadable document proves that no file is read.
		if err := os.MkdirAll(filepath.Join(dir, ".replworks", "PRODUCT_SPEC.md"), 0o755); err != nil {
			t.Fatal(err)
		}
		t.Chdir(dir)

		code, stdout, _ := runApp(t, args...)

		if code != 2 {
			t.Fatalf("%q: exit code = %d, want 2", args, code)
		}
		lines := strings.Split(strings.TrimSpace(stdout), "\n")
		if len(lines) != 1 || !strings.HasPrefix(lines[0], "repl: ") {
			t.Fatalf("%q: stdout = %q, want one error line", args, stdout)
		}
		if len(args) > 1 && !strings.Contains(stdout, "foo") {
			t.Fatalf("%q: stdout = %q, want it to name foo", args, stdout)
		}
	}
}

func TestShowReadsOnlyDefiningDocuments(t *testing.T) {
	dir := showProject(t)
	if err := os.Remove(filepath.Join(dir, ".replworks", "TASKS.md")); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, ".replworks", "TASKS.md"), 0o755); err != nil {
		t.Fatal(err)
	}

	if code, stdout, _ := runApp(t, "show", "FR-001"); code != 0 {
		t.Fatalf("exit code = %d, want 0 (stdout = %q)", code, stdout)
	}
	if code, stdout, _ := runApp(t, "show", "T-001"); code != 2 || !strings.Contains(stdout, "repl: ") || strings.Contains(stdout, "no definition") {
		t.Fatalf("exit code = %d, stdout = %q, want 2 with one error and no missing line", code, stdout)
	}
}

func TestShowWritesNothing(t *testing.T) {
	dir := showProject(t)
	before := snapshotTree(t, dir)

	runApp(t, "show", "FR-001", "C2", "T-001", "FR-999")

	if after := snapshotTree(t, dir); after != before {
		t.Fatalf("tree changed:\nbefore %s\nafter  %s", before, after)
	}
}

func TestShowJSON(t *testing.T) {
	showProject(t)

	code, stdout, _ := runApp(t, "show", "FR-002", "FR-999", "--json")

	if code != 1 {
		t.Fatalf("exit code = %d, want 1", code)
	}
	if strings.Count(stdout, "\n") != 1 || strings.Contains(stdout, "\x1b") {
		t.Fatalf("stdout = %q, want one line without escape sequences", stdout)
	}
	var got struct {
		Definitions []struct {
			ID   string `json:"id"`
			Path string `json:"path"`
			Line int    `json:"line"`
			Text string `json:"text"`
		} `json:"definitions"`
		Missing []string `json:"missing"`
		Errors  []string `json:"errors"`
	}
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("stdout = %q: %v", stdout, err)
	}
	if len(got.Definitions) != 1 || got.Definitions[0].ID != "FR-002" || got.Definitions[0].Path != ".replworks/PRODUCT_SPEC.md" ||
		got.Definitions[0].Line != 2 || got.Definitions[0].Text != "- FR-002: two\n  - first continuation\n  - second continuation" {
		t.Fatalf("definitions = %+v", got.Definitions)
	}
	if len(got.Missing) != 1 || got.Missing[0] != "FR-999" || got.Errors == nil || len(got.Errors) != 0 {
		t.Fatalf("missing = %q, errors = %q", got.Missing, got.Errors)
	}
}

func TestShowJSONMissingOnly(t *testing.T) {
	showProject(t)

	code, stdout, _ := runApp(t, "show", "FR-999", "--json")

	if code != 1 || strings.TrimSpace(stdout) != `{"definitions":[],"missing":["FR-999"],"errors":[]}` {
		t.Fatalf("exit code = %d, stdout = %q", code, stdout)
	}
}

func snapshotTree(t *testing.T, dir string) string {
	t.Helper()

	var b strings.Builder
	err := filepath.WalkDir(dir, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		b.WriteString(path + " " + info.Mode().String() + " " + info.ModTime().String())
		if !entry.IsDir() {
			content, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			b.Write(content)
		}
		b.WriteString("\n")
		return nil
	})
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}

	return b.String()
}
