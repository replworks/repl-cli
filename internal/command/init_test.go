package command

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

// recordedTestdata holds the page and the canonical text observed live from
// the canonical source. The path is resolved once so that tests may change the
// working directory.
var recordedTestdata = func() string {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	return filepath.Join(dir, "..", "canon", "testdata")
}()

func recordedPage(t *testing.T) []byte {
	t.Helper()

	page, err := os.ReadFile(filepath.Join(recordedTestdata, "agents.html"))
	if err != nil {
		t.Fatalf("reading the recorded page: %v", err)
	}

	return page
}

func recordedText(t *testing.T) string {
	t.Helper()

	text, err := os.ReadFile(filepath.Join(recordedTestdata, "canonical.txt"))
	if err != nil {
		t.Fatalf("reading the recorded canonical text: %v", err)
	}

	return string(text)
}

type recordingTransport struct {
	target string

	mu       sync.Mutex
	requests []string
}

func (t *recordingTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	t.mu.Lock()
	t.requests = append(t.requests, request.URL.String())
	t.mu.Unlock()

	target, err := url.Parse(t.target)
	if err != nil {
		return nil, err
	}

	clone := request.Clone(request.Context())
	clone.URL.Scheme = target.Scheme
	clone.URL.Host = target.Host

	return http.DefaultTransport.RoundTrip(clone)
}

func (t *recordingTransport) requested() []string {
	t.mu.Lock()
	defer t.mu.Unlock()

	return append([]string(nil), t.requests...)
}

func servePage(t *testing.T, page string) (*http.Client, *recordingTransport) {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = io.WriteString(w, page)
	}))
	t.Cleanup(server.Close)

	transport := &recordingTransport{target: server.URL}

	return &http.Client{Transport: transport}, transport
}

func runInit(t *testing.T, client *http.Client, args ...string) (int, string, string) {
	t.Helper()

	var stdout, stderr bytes.Buffer
	app := App{
		Version:    "dev",
		Commit:     "none",
		Date:       "unknown",
		Stdout:     &stdout,
		Stderr:     &stderr,
		HTTPClient: client,
	}

	return app.Execute(args), stdout.String(), stderr.String()
}

func TestInitCreatesAgentsAndReplworks(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	client, transport := servePage(t, string(recordedPage(t)))

	code, stdout, stderr := runInit(t, client, "init")

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
	if string(content) != recordedText(t) {
		t.Fatal("AGENTS.md does not hold the canonical AGENTS.md")
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

	inside, err := os.ReadDir(filepath.Join(dir, ".replworks"))
	if err != nil {
		t.Fatalf("reading .replworks: %v", err)
	}
	if len(inside) != 0 {
		t.Fatalf(".replworks holds %d entries, want none", len(inside))
	}

	if got := transport.requested(); len(got) != 1 {
		t.Fatalf("requests = %q, want 1", got)
	}

	for _, want := range []string{"created: AGENTS.md", "created: .replworks"} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("stdout = %q, want it to contain %q", stdout, want)
		}
	}
}

func TestInitLeavesExistingEntriesUnchanged(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte("mine\n"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := os.Mkdir(filepath.Join(dir, ".replworks"), 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	t.Chdir(dir)

	client, _ := servePage(t, string(recordedPage(t)))
	before := directoryState(t, dir)

	code, stdout, stderr := runInit(t, client, "init")

	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr = %q)", code, stderr)
	}
	if after := directoryState(t, dir); after != before {
		t.Fatalf("project directory changed:\nbefore: %s\nafter:  %s", before, after)
	}

	for _, want := range []string{"skipped: AGENTS.md", "skipped: .replworks"} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("stdout = %q, want it to contain %q", stdout, want)
		}
	}
}

func TestInitCreatesTheMissingEntryOnly(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte("mine\n"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	t.Chdir(dir)

	client, _ := servePage(t, string(recordedPage(t)))

	code, stdout, stderr := runInit(t, client, "init")

	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr = %q)", code, stderr)
	}

	content, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	if err != nil {
		t.Fatalf("reading AGENTS.md: %v", err)
	}
	if string(content) != "mine\n" {
		t.Fatalf("AGENTS.md = %q, want the existing content to be unchanged", content)
	}
	if info, err := os.Stat(filepath.Join(dir, ".replworks")); err != nil || !info.IsDir() {
		t.Fatalf("stat .replworks: %v, want an existing directory", err)
	}
	if !strings.Contains(stdout, "created: .replworks") {
		t.Fatalf("stdout = %q, want it to report the created directory", stdout)
	}
}

func TestInitFailsWhenTheCanonicalSourceIsUnavailable(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "AGENTS.md"), []byte("mine\n"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	t.Chdir(dir)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(server.Close)
	client := &http.Client{Transport: &recordingTransport{target: server.URL}}

	before := directoryState(t, dir)

	code, stdout, stderr := runInit(t, client, "init")

	if code != 2 {
		t.Fatalf("exit code = %d, want 2", code)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}

	lines := errorLines(t, stdout)
	if len(lines) != 1 {
		t.Fatalf("stdout has %d lines (%q), want 1", len(lines), stdout)
	}
	if !strings.Contains(lines[0], "repl:") {
		t.Fatalf("error line = %q, want it to start with the program name", lines[0])
	}

	if after := directoryState(t, dir); after != before {
		t.Fatalf("project directory changed:\nbefore: %s\nafter:  %s", before, after)
	}
}

func TestInitFailsWithoutAnExtractableCanonicalText(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	client, _ := servePage(t, "<html><body><pre>AGENTS.md</pre></body></html>")

	before := directoryState(t, dir)

	code, stdout, stderr := runInit(t, client, "init")

	if code != 2 {
		t.Fatalf("exit code = %d, want 2", code)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}
	if lines := errorLines(t, stdout); len(lines) != 1 {
		t.Fatalf("stdout has %d lines (%q), want 1", len(lines), stdout)
	}

	if after := directoryState(t, dir); after != before {
		t.Fatalf("project directory changed:\nbefore: %s\nafter:  %s", before, after)
	}
}

func TestInitSkipsEntriesOfAnyKind(t *testing.T) {
	dir := t.TempDir()
	if err := os.Mkdir(filepath.Join(dir, "AGENTS.md"), 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".replworks"), []byte("not a directory\n"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	t.Chdir(dir)

	client, _ := servePage(t, string(recordedPage(t)))
	before := directoryState(t, dir)

	code, stdout, stderr := runInit(t, client, "init")

	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr = %q)", code, stderr)
	}
	if stderr != "" {
		t.Fatalf("stderr = %q, want empty", stderr)
	}
	if after := directoryState(t, dir); after != before {
		t.Fatalf("project directory changed:\nbefore: %s\nafter:  %s", before, after)
	}

	for _, want := range []string{"skipped: AGENTS.md", "skipped: .replworks"} {
		if !strings.Contains(stdout, want) {
			t.Fatalf("stdout = %q, want it to contain %q", stdout, want)
		}
	}
}

func TestInitRejectsArguments(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	code, stdout, stderr := runInit(t, nil, "init", "AGENTS.md")

	if code != 2 {
		t.Fatalf("exit code = %d, want 2", code)
	}
	if stdout != "" {
		t.Fatalf("stdout = %q, want empty", stdout)
	}
	if !strings.Contains(stderr, "unknown command") && !strings.Contains(stderr, "accepts 0 arg") {
		t.Fatalf("stderr = %q, want it to reject the extra argument", stderr)
	}
}
