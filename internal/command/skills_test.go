package command

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/replworks/repl-cli/internal/skill"
)

func serveSkills(t *testing.T, fail map[string]bool) (*http.Client, *recordingTransport) {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, src := range skill.Sources {
			if strings.HasSuffix(r.URL.Path, "/"+src.Name+"/SKILL.md") {
				if fail[src.Name] {
					http.Error(w, "boom", http.StatusNotFound)
					return
				}
				_, _ = w.Write([]byte("# " + src.Name + "\nbody for " + src.Name + "\n"))
				return
			}
		}

		http.NotFound(w, r)
	}))
	t.Cleanup(server.Close)

	transport := &recordingTransport{target: server.URL}

	return &http.Client{Transport: transport}, transport
}

func skillPaths(dir string) map[string]string {
	files := map[string]string{}

	for _, src := range skill.Sources {
		content, err := os.ReadFile(filepath.Join(dir, src.Path))
		if err == nil {
			files[src.Path] = string(content)
		}
	}

	return files
}

func TestSkillsSyncInstallsTheFiveSkills(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	client, _ := serveSkills(t, nil)

	code, stdout, stderr := runCommand(t, client, "skills", "sync")

	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (stdout = %q, stderr = %q)", code, stdout, stderr)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}

	for _, src := range skill.Sources {
		content, err := os.ReadFile(filepath.Join(dir, src.Path))
		if err != nil {
			t.Fatalf("reading %s: %v", src.Path, err)
		}
		if string(content) != "# "+src.Name+"\nbody for "+src.Name+"\n" {
			t.Errorf("%s = %q, want the served body", src.Path, content)
		}
		if !strings.Contains(stdout, src.Name+": "+src.Path+": created\n") {
			t.Errorf("stdout = %q, want a created line for %s", stdout, src.Name)
		}
	}
}

func TestSkillsSyncRerunLeavesIdenticalFilesUntouched(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	client, _ := serveSkills(t, nil)

	if code, _, stderr := runCommand(t, client, "skills", "sync"); code != 0 {
		t.Fatalf("setup: exit code = %d (stderr = %q)", code, stderr)
	}
	before := skillPaths(dir)

	code, stdout, stderr := runCommand(t, client, "skills", "sync")

	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr = %q)", code, stderr)
	}
	if after := skillPaths(dir); !mapsEqual(before, after) {
		t.Fatal("a rerun modified a skill file")
	}
	for _, src := range skill.Sources {
		if !strings.Contains(stdout, src.Name+": "+src.Path+": up to date\n") {
			t.Errorf("stdout = %q, want an up to date line for %s", stdout, src.Name)
		}
	}
}

func TestSkillsSyncReplacesOnlyDifferingFiles(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	client, _ := serveSkills(t, nil)

	if code, _, stderr := runCommand(t, client, "skills", "sync"); code != 0 {
		t.Fatalf("setup: exit code = %d (stderr = %q)", code, stderr)
	}

	path := filepath.Join(dir, ".claude", "skills", "repl-tasks", "SKILL.md")
	if err := os.WriteFile(path, []byte("custom"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	code, stdout, stderr := runCommand(t, client, "skills", "sync")

	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr = %q)", code, stderr)
	}
	if content, err := os.ReadFile(path); err != nil || string(content) != "# repl-tasks\nbody for repl-tasks\n" {
		t.Fatalf("repl-tasks skill = %q, err = %v, want the served body", content, err)
	}
	for _, src := range skill.Sources {
		state := "up to date"
		if src.Name == "repl-tasks" {
			state = "replaced"
		}
		if !strings.Contains(stdout, src.Name+": "+src.Path+": "+state+"\n") {
			t.Errorf("stdout = %q, want a %s line for %s", stdout, state, src.Name)
		}
	}
}

func TestSkillsSyncFetchFailureWritesNothing(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	client, _ := serveSkills(t, map[string]bool{"repl-auto-run": true})

	code, stdout, _ := runCommand(t, client, "skills", "sync")

	if code != 2 {
		t.Fatalf("exit code = %d, want 2 (stdout = %q)", code, stdout)
	}
	if entries, err := os.ReadDir(dir); err != nil || len(entries) != 0 {
		t.Fatalf("project directory is not empty after a failed run: %v, err = %v", entries, err)
	}
}

func TestSkillsSyncRequestsOnlyTheFiveSkillURLs(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	client, transport := serveSkills(t, nil)

	code, _, stderr := runCommand(t, client, "skills", "sync")

	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr = %q)", code, stderr)
	}

	requests := transport.requested()
	if len(requests) != len(skill.Sources) {
		t.Fatalf("requests = %q, want the five skill URLs", requests)
	}
	for i, src := range skill.Sources {
		if requests[i] != src.URL {
			t.Fatalf("requests = %q, want %q at index %d", requests, src.URL, i)
		}
	}
}

func mapsEqual(a, b map[string]string) bool {
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if b[k] != v {
			return false
		}
	}

	return true
}

func TestSkillsSyncJSONReportsCreatedSkills(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	client, _ := serveSkills(t, nil)

	code, stdout, stderr := runCommand(t, client, "--json", "skills", "sync")

	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (stdout = %q, stderr = %q)", code, stdout, stderr)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}

	value := parseJSON(t, stdout)
	items, ok := value["items"].([]any)
	if !ok || len(items) != len(skill.Sources) {
		t.Fatalf("items = %v, want the five skills", value["items"])
	}
	for i, item := range items {
		info, ok := item.(map[string]any)
		if !ok {
			t.Fatalf("item %d = %v, want an object", i, item)
		}
		if info["name"] != skill.Sources[i].Name || info["path"] != skill.Sources[i].Path || info["state"] != "created" {
			t.Fatalf("item %d = %v, want the created skill", i, info)
		}
	}
	if errors, ok := value["errors"].([]any); !ok || len(errors) != 0 {
		t.Fatalf("errors = %v, want empty array", value["errors"])
	}
}

func TestSkillsSyncJSONReportsUpToDateSkills(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	client, _ := serveSkills(t, nil)

	if code, _, stderr := runCommand(t, client, "skills", "sync"); code != 0 {
		t.Fatalf("setup: exit code = %d (stderr = %q)", code, stderr)
	}

	code, stdout, _ := runCommand(t, client, "--json", "skills", "sync")

	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (stdout = %q)", code, stdout)
	}

	value := parseJSON(t, stdout)
	for _, item := range value["items"].([]any) {
		state := item.(map[string]any)["state"]
		if state != "up to date" {
			t.Fatalf("state = %v, want up to date", state)
		}
	}
}

func TestSkillsSyncJSONExitsTwoOnFetchFailure(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	client, _ := serveSkills(t, map[string]bool{"repl-auto-run": true})

	code, stdout, _ := runCommand(t, client, "--json", "skills", "sync")

	if code != 2 {
		t.Fatalf("exit code = %d, want 2 (stdout = %q)", code, stdout)
	}

	value := parseJSON(t, stdout)
	if errors, ok := value["errors"].([]any); !ok || len(errors) != 1 {
		t.Fatalf("errors = %v, want one error", value["errors"])
	}
}
