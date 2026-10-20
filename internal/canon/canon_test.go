package canon

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
)

const recordedTextPath = "testdata/canonical.txt"

func recordedText(t *testing.T) string {
	t.Helper()

	text, err := os.ReadFile(recordedTextPath)
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

func (t *recordingTransport) count() int {
	t.mu.Lock()
	defer t.mu.Unlock()

	return len(t.requests)
}

func (t *recordingTransport) requested() []string {
	t.mu.Lock()
	defer t.mu.Unlock()

	return append([]string(nil), t.requests...)
}

func serve(t *testing.T, handler http.HandlerFunc) (*http.Client, *recordingTransport) {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	transport := &recordingTransport{target: server.URL}

	return &http.Client{Transport: transport}, transport
}

func serveText(t *testing.T, text string) (*http.Client, *recordingTransport) {
	t.Helper()

	return serve(t, func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, text)
	})
}

func TestCanonicalReturnsTheExactDocumentResponse(t *testing.T) {
	want := recordedText(t)
	client, _ := serveText(t, want)

	got := New(client).Canonical()

	if !got.IsPresent() {
		t.Fatalf("Canonical() = %v, want present", got.Kind())
	}
	if got.Value() != want {
		t.Fatalf("Canonical() = %q, want the exact response", got.Value())
	}
}

func TestCanonicalCanSelectTheAutomaticDocument(t *testing.T) {
	want := "# AGENTS.md\n\n## AUTO_RUN\n"
	client, transport := serveText(t, want)

	got := NewForDocument(client, AgentsAutoDocument).Canonical()

	if !got.IsPresent() || got.Value() != want {
		t.Fatalf("Canonical() = %v, want %q", got, want)
	}
	requests := transport.requested()
	if len(requests) != 1 || requests[0] != string(AgentsAutoDocument) {
		t.Fatalf("requests = %q, want one request to %q", requests, AgentsAutoDocument)
	}
}

func TestCanonicalCanRetrieveTheAskHumanScript(t *testing.T) {
	want := "#!/usr/bin/env bash\necho ask-human\n"
	client, transport := serveText(t, want)

	got := NewForURL(client, string(AskHumanScript)).Canonical()

	if !got.IsPresent() || got.Value() != want {
		t.Fatalf("Canonical() = %v, want %q", got, want)
	}
	requests := transport.requested()
	if len(requests) != 1 || requests[0] != string(AskHumanScript) {
		t.Fatalf("requests = %q, want one request to %q", requests, AskHumanScript)
	}
}

func TestCanonicalRequestsOnlyTheSelectedSource(t *testing.T) {
	client, transport := serveText(t, recordedText(t))

	if got := NewForDocument(client, AgentsAutoDocument).Canonical(); !got.IsPresent() {
		t.Fatalf("Canonical() = %v, want present", got.Kind())
	}

	got := transport.requested()
	if len(got) != 1 || got[0] != string(AgentsAutoDocument) {
		t.Fatalf("requests = %q, want one request to %q", got, AgentsAutoDocument)
	}
}

func TestCanonicalRetrievesTheSourceOnce(t *testing.T) {
	client, transport := serveText(t, recordedText(t))
	provider := New(client)

	first := provider.Canonical()
	second := provider.Canonical()

	if transport.count() != 1 {
		t.Fatalf("requests = %d, want 1", transport.count())
	}
	if !first.IsPresent() || !second.IsPresent() || first.Value() != second.Value() {
		t.Fatalf("Canonical() = %v and %v, want the same present fact", first.Kind(), second.Kind())
	}
}

func TestCanonicalRetrievesTheSourceOnceWhenItFails(t *testing.T) {
	client, transport := serve(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	provider := New(client)

	first := provider.Canonical()
	second := provider.Canonical()

	if transport.count() != 1 {
		t.Fatalf("requests = %d, want 1", transport.count())
	}
	if !first.IsFailed() || !second.IsFailed() {
		t.Fatalf("Canonical() = %v and %v, want failed", first.Kind(), second.Kind())
	}
}

func TestCanonicalFailsWhenTheSourceIsUnreachable(t *testing.T) {
	client, transport := serveText(t, recordedText(t))
	transport.target = "http://127.0.0.1:1"

	got := New(client).Canonical()

	if !got.IsFailed() || got.Reason() == nil {
		t.Fatalf("Canonical() = %v, want failed with a reason", got.Kind())
	}
	requests := transport.requested()
	if len(requests) != 1 || requests[0] != SourceURL {
		t.Fatalf("requests = %q, want one request to %q and no fallback", requests, SourceURL)
	}
}

func TestCanonicalFailsWhenTheResponseIsUnsuccessful(t *testing.T) {
	client, _ := serve(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	got := New(client).Canonical()

	if !got.IsFailed() || got.Reason() == nil || !strings.Contains(got.Reason().Error(), "404") {
		t.Fatalf("Canonical() = %v, want a failed fact mentioning 404", got)
	}
}

func TestCanonicalFailsWhenTheResponseIsEmpty(t *testing.T) {
	client, _ := serveText(t, "")

	got := New(client).Canonical()

	if !got.IsFailed() {
		t.Fatalf("Canonical() = %v, want failed", got.Kind())
	}
}
