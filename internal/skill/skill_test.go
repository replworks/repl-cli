package skill

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"
)

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

func TestSourcesAreFixedAndWellFormed(t *testing.T) {
	want := []struct {
		name string
		path string
	}{
		{"repl-product-spec", ".claude/skills/repl-product-spec/SKILL.md"},
		{"repl-tech-stack", ".claude/skills/repl-tech-stack/SKILL.md"},
		{"repl-architecture", ".claude/skills/repl-architecture/SKILL.md"},
		{"repl-tasks", ".claude/skills/repl-tasks/SKILL.md"},
		{"repl-auto-run", ".claude/skills/repl-auto-run/SKILL.md"},
	}

	if len(Sources) != len(want) {
		t.Fatalf("Sources has %d entries, want %d", len(Sources), len(want))
	}

	for i, want := range want {
		source := Sources[i]
		if source.Name != want.name {
			t.Errorf("Sources[%d].Name = %q, want %q", i, source.Name, want.name)
		}
		if source.Path != want.path {
			t.Errorf("Sources[%d].Path = %q, want %q", i, source.Path, want.path)
		}
		if !strings.HasPrefix(source.URL, "https://www.repl.net/skills/"+source.Name+"/SKILL.md") {
			t.Errorf("Sources[%d].URL = %q, want the published skill URL", i, source.URL)
		}
	}
}

func TestFetchReturnsTheExactResponseOnce(t *testing.T) {
	want := "# SKILL.md\n\nname: repl-product-spec\n"
	client, transport := serveText(t, want)

	provider := New(client)
	for range []int{1, 2} {
		got := provider.Fetch(Sources[0].URL)
		if !got.IsPresent() {
			t.Fatal("expected the skill file to be present")
		}
		if got.Value() != want {
			t.Fatalf("Fetch() = %q, want %q", got.Value(), want)
		}
	}

	if requests := transport.requested(); len(requests) != 1 || requests[0] != Sources[0].URL {
		t.Fatalf("requests = %q, want one request to %q", requests, Sources[0].URL)
	}
}

func TestFetchFailsOnEmptyResponse(t *testing.T) {
	client, _ := serveText(t, "")

	provider := New(client)

	got := provider.Fetch(Sources[0].URL)
	if !got.IsFailed() {
		t.Fatal("expected the fetch to fail on an empty response")
	}
}

func TestFetchFailsOnUnexpectedStatus(t *testing.T) {
	client, _ := serve(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	provider := New(client)

	got := provider.Fetch(Sources[0].URL)
	if !got.IsFailed() {
		t.Fatal("expected the fetch to fail on an empty response")
	}
}
