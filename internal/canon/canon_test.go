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

const (
	recordedPagePath = "testdata/agents.html"
	recordedTextPath = "testdata/canonical.txt"
)

func recordedPage(t *testing.T) []byte {
	t.Helper()

	page, err := os.ReadFile(recordedPagePath)
	if err != nil {
		t.Fatalf("reading the recorded page: %v", err)
	}

	return page
}

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

func servePage(t *testing.T, page string) (*http.Client, *recordingTransport) {
	t.Helper()

	return serve(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = io.WriteString(w, page)
	})
}

func TestCanonicalFromTheRecordedPage(t *testing.T) {
	client, _ := servePage(t, string(recordedPage(t)))

	got := New(client).Canonical()

	if !got.IsPresent() {
		t.Fatalf("Canonical() = %v, want present", got.Kind())
	}

	want := recordedText(t)
	if got.Value() != want {
		t.Fatalf("Canonical() = %q, want the recorded canonical text", got.Value())
	}
	if strings.Contains(got.Value(), separator) {
		t.Fatal("Canonical() still holds the line separator used by the source")
	}
	if strings.HasPrefix(got.Value(), "\n") || strings.HasSuffix(got.Value(), "\n") {
		t.Fatalf("Canonical() = %q, want no leading or trailing newline", got.Value())
	}
}

func TestCanonicalRequestsOnlyTheCanonicalSource(t *testing.T) {
	client, transport := servePage(t, string(recordedPage(t)))

	if got := New(client).Canonical(); !got.IsPresent() {
		t.Fatalf("Canonical() = %v, want present", got.Kind())
	}

	got := transport.requested()
	if len(got) != 1 {
		t.Fatalf("requests = %q, want 1 request", got)
	}
	if got[0] != SourceURL {
		t.Fatalf("requested %q, want %q", got[0], SourceURL)
	}
}

func TestCanonicalRetrievesTheSourceOnce(t *testing.T) {
	client, transport := servePage(t, string(recordedPage(t)))
	provider := New(client)

	first := provider.Canonical()
	second := provider.Canonical()

	if transport.count() != 1 {
		t.Fatalf("requests = %d, want 1", transport.count())
	}
	if !first.IsPresent() || !second.IsPresent() {
		t.Fatalf("Canonical() = %v and %v, want present", first.Kind(), second.Kind())
	}
	if first.Value() != second.Value() {
		t.Fatal("Canonical() returned two different texts")
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
	client, transport := servePage(t, string(recordedPage(t)))
	transport.target = "http://127.0.0.1:1"

	got := New(client).Canonical()

	if !got.IsFailed() {
		t.Fatalf("Canonical() = %v, want failed", got.Kind())
	}
	if got.Reason() == nil {
		t.Fatal("Canonical() reason = nil, want an error")
	}
}

func TestCanonicalFailsWhenTheResponseIsUnsuccessful(t *testing.T) {
	client, _ := serve(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	got := New(client).Canonical()

	if !got.IsFailed() {
		t.Fatalf("Canonical() = %v, want failed", got.Kind())
	}
	if got.Reason() == nil || !strings.Contains(got.Reason().Error(), "404") {
		t.Fatalf("Canonical() reason = %v, want it to mention the status", got.Reason())
	}
}

func TestCanonicalFailsWithoutACanonicalText(t *testing.T) {
	client, _ := servePage(t, "<html><body><pre>AGENTS.md</pre></body></html>")

	got := New(client).Canonical()

	if !got.IsFailed() {
		t.Fatalf("Canonical() = %v, want failed", got.Kind())
	}
}

func TestCanonicalFailsWithSeveralCanonicalTexts(t *testing.T) {
	page := `<html><body><button data-code="one">a</button><button data-code="two">b</button></body></html>`
	client, _ := servePage(t, page)

	got := New(client).Canonical()

	if !got.IsFailed() {
		t.Fatalf("Canonical() = %v, want failed", got.Kind())
	}
}

func TestExtractDecodesEntitiesAndSplitsLines(t *testing.T) {
	page := `<html><body><button data-code="&#x60;code&#x60;` + separator + `next"></button></body></html>`

	got, err := extract([]byte(page))
	if err != nil {
		t.Fatalf("extract() error = %v, want nil", err)
	}
	if want := "`code`\nnext"; got != want {
		t.Fatalf("extract() = %q, want %q", got, want)
	}
}
