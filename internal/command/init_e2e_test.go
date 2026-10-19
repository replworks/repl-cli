//go:build e2e

package command

import (
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/replworks/repl-cli/internal/canon"
)

func liveCanonical(t *testing.T) string {
	return liveDocument(t, canon.AgentsDocument)
}

func liveDocument(t *testing.T, document canon.Document) string {
	t.Helper()

	canonical := canon.NewForDocument(nil, document).Canonical()
	if !canonical.IsPresent() {
		t.Fatalf("the live canonical source is unreachable: %v", canonical.Reason())
	}
	if !strings.HasPrefix(canonical.Value(), "# AGENTS.md") {
		t.Fatalf("the live canonical text starts with %q, want the AGENTS.md heading", firstLine(canonical.Value()))
	}

	return canonical.Value()
}

func TestInitAgainstTheLiveAutomaticSource(t *testing.T) {
	want := liveDocument(t, canon.AgentsAutoDocument)

	dir := t.TempDir()
	t.Chdir(dir)

	code, stdout, stderr := runCommandWithInput(t, nil, "2\n", "init")

	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (stdout = %q, stderr = %q)", code, stdout, stderr)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}
	content, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	if err != nil {
		t.Fatalf("reading AGENTS.md: %v", err)
	}
	if string(content) != want {
		t.Fatal("AGENTS.md does not hold the selected live AGENTS.md")
	}
	if info, err := os.Stat(filepath.Join(dir, ".replworks")); err != nil || !info.IsDir() {
		t.Fatalf("stat .replworks: %v, want a directory", err)
	}
}

func TestInitAgainstTheLiveSource(t *testing.T) {
	want := liveCanonical(t)

	dir := t.TempDir()
	t.Chdir(dir)

	code, stdout, stderr := runCommand(t, nil, "init")

	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr = %q)", code, stderr)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}

	content, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	if err != nil {
		t.Fatalf("reading AGENTS.md: %v", err)
	}
	if string(content) != want {
		t.Fatal("AGENTS.md does not hold the live canonical AGENTS.md")
	}

	info, err := os.Stat(filepath.Join(dir, ".replworks"))
	if err != nil {
		t.Fatalf("stat .replworks: %v", err)
	}
	if !info.IsDir() {
		t.Fatal(".replworks is not a directory")
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("reading the project directory: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("project directory holds %d entries, want 2", len(entries))
	}

	documents := []string{
		".replworks/PRODUCT_SPEC.md",
		".replworks/TECH_STACK.md",
		".replworks/ARCHITECTURE.md",
		".replworks/TASKS.md",
	}
	for _, document := range documents {
		content, err := os.ReadFile(filepath.Join(dir, document))
		if err != nil {
			t.Fatalf("reading %s: %v", document, err)
		}
		if len(content) != 0 {
			t.Fatalf("%s = %q, want an empty file", document, content)
		}
	}

	for _, want := range []string{
		"created: AGENTS.md",
		"created: .replworks",
		"created: .replworks/PRODUCT_SPEC.md",
		"created: .replworks/TECH_STACK.md",
		"created: .replworks/ARCHITECTURE.md",
		"created: .replworks/TASKS.md",
	} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("stdout = %q, want it to contain %q", stdout, want)
		}
	}
}

func TestInitAgainstTheLiveSourceKeepsAnExistingAgentsMD(t *testing.T) {
	liveCanonical(t)

	dir := t.TempDir()
	existing := "# My own AGENTS.md\r\n\r\nKeep me byte for byte.\n"
	if err := os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte(existing), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	t.Chdir(dir)

	code, stdout, stderr := runCommand(t, nil, "init")

	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr = %q)", code, stderr)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}

	content, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	if err != nil {
		t.Fatalf("reading AGENTS.md: %v", err)
	}
	if string(content) != existing {
		t.Fatalf("AGENTS.md = %q, want the existing content byte for byte", content)
	}

	if info, err := os.Stat(filepath.Join(dir, ".replworks")); err != nil || !info.IsDir() {
		t.Fatalf("stat .replworks: %v, want an existing directory", err)
	}

	for _, want := range []string{
		"skipped: AGENTS.md",
		"created: .replworks",
		"created: .replworks/PRODUCT_SPEC.md",
		"created: .replworks/TECH_STACK.md",
		"created: .replworks/ARCHITECTURE.md",
		"created: .replworks/TASKS.md",
	} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("stdout = %q, want it to contain %q", stdout, want)
		}
	}
}

func TestInitAgainstTheLiveSourceCreatesAGENTSMDWhenReplworksExists(t *testing.T) {
	want := liveCanonical(t)

	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, ".replworks"), 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	t.Chdir(dir)

	code, stdout, stderr := runCommand(t, nil, "init")

	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr = %q)", code, stderr)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}

	content, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	if err != nil {
		t.Fatalf("reading AGENTS.md: %v", err)
	}
	if string(content) != want {
		t.Fatal("AGENTS.md does not hold the live canonical AGENTS.md")
	}

	for _, text := range []string{
		"created: AGENTS.md",
		"skipped: .replworks",
		"created: .replworks/PRODUCT_SPEC.md",
		"created: .replworks/TECH_STACK.md",
		"created: .replworks/ARCHITECTURE.md",
		"created: .replworks/TASKS.md",
	} {
		if !strings.Contains(stdout, text) {
			t.Fatalf("stdout = %q, want it to contain %q", stdout, text)
		}
	}
}

func TestInitAgainstTheLiveSourceSkipsEveryExistingItem(t *testing.T) {
	liveCanonical(t)

	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, ".replworks"), 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	for _, document := range []string{"PRODUCT_SPEC.md", "TECH_STACK.md", "ARCHITECTURE.md", "TASKS.md"} {
		if err := os.WriteFile(filepath.Join(dir, ".replworks", document), []byte(document+"\n"), 0o644); err != nil {
			t.Fatalf("setup: %v", err)
		}
	}
	if err := os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte("mine\n"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	t.Chdir(dir)

	before := directoryState(t, dir)

	code, stdout, stderr := runCommand(t, nil, "init")

	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr = %q)", code, stderr)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}
	if after := directoryState(t, dir); after != before {
		t.Fatalf("project directory changed:\nbefore: %s\nafter:  %s", before, after)
	}

	for _, want := range []string{
		"skipped: AGENTS.md",
		"skipped: .replworks",
		"skipped: .replworks/PRODUCT_SPEC.md",
		"skipped: .replworks/TECH_STACK.md",
		"skipped: .replworks/ARCHITECTURE.md",
		"skipped: .replworks/TASKS.md",
	} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("stdout = %q, want it to contain %q", stdout, want)
		}
	}
}

func TestInitAgainstTheLiveSourceCreatesOnlyTheMissingDocuments(t *testing.T) {
	liveCanonical(t)

	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, ".replworks"), 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	existing := "# My own TECH_STACK.md\n"
	if err := os.WriteFile(filepath.Join(dir, ".replworks", "TECH_STACK.md"), []byte(existing), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	t.Chdir(dir)

	code, stdout, stderr := runCommand(t, nil, "init")

	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr = %q)", code, stderr)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}

	content, err := os.ReadFile(filepath.Join(dir, ".replworks", "TECH_STACK.md"))
	if err != nil {
		t.Fatalf("reading TECH_STACK.md: %v", err)
	}
	if string(content) != existing {
		t.Fatalf("TECH_STACK.md = %q, want the existing content byte for byte", content)
	}

	for _, document := range []string{"PRODUCT_SPEC.md", "ARCHITECTURE.md", "TASKS.md"} {
		info, err := os.Stat(filepath.Join(dir, ".replworks", document))
		if err != nil {
			t.Fatalf("stat %s: %v", document, err)
		}
		if info.Size() != 0 {
			t.Fatalf("%s size = %d, want 0", document, info.Size())
		}
	}

	for _, want := range []string{
		"skipped: .replworks/TECH_STACK.md",
		"created: .replworks/PRODUCT_SPEC.md",
		"created: .replworks/ARCHITECTURE.md",
		"created: .replworks/TASKS.md",
	} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("stdout = %q, want it to contain %q", stdout, want)
		}
	}
}

func TestInitAgainstTheLiveSourceWritesNothingWhenTheSourceIsUnobtainable(t *testing.T) {
	liveCanonical(t)

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte("mine\n"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	t.Chdir(dir)

	before := directoryState(t, dir)

	client := &http.Client{Transport: &recordingTransport{target: "http://127.0.0.1:1"}}

	code, stdout, stderr := runCommand(t, client, "init")

	if code != 2 {
		t.Fatalf("exit code = %d, want 2", code)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}
	if count := strings.Count(stdout, "repl:"); count != 1 {
		t.Fatalf("stdout has %d error lines (%q), want 1", count, stdout)
	}
	if after := directoryState(t, dir); after != before {
		t.Fatalf("project directory changed:\nbefore: %s\nafter:  %s", before, after)
	}
}

func TestInitAgainstTheLiveSourceReportsAWriteFailure(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("a read-only directory does not prevent writes on Windows")
	}

	liveCanonical(t)

	dir := t.TempDir()
	t.Chdir(dir)
	if err := os.Chmod(dir, 0o555); err != nil {
		t.Fatalf("setup: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chmod(dir, 0o755)
	})

	code, stdout, stderr := runCommand(t, nil, "init")

	if code != 2 {
		t.Fatalf("exit code = %d, want 2", code)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}
	if count := strings.Count(stdout, "repl:"); count != 1 {
		t.Fatalf("stdout has %d error lines (%q), want 1", count, stdout)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("reading the project directory: %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("project directory holds %d entries, want none", len(entries))
	}
}

func firstLine(text string) string {
	if index := strings.Index(text, "\n"); index >= 0 {
		return text[:index]
	}

	return text
}
