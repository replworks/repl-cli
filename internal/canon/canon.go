// Package canon obtains the canonical AGENTS.md from the project's website.
package canon

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/replworks/repl-cli/internal/fact"
)

// Document identifies one of the published AGENTS.md documents.
type Document string

const (
	AgentsDocument     Document = "https://www.repl.net/documents/agents.md"
	AgentsAutoDocument Document = "https://www.repl.net/documents/agents-auto.md"
	// SourceURL is the default document used by commands without a selection.
	SourceURL = string(AgentsDocument)
)

const (
	requestTimeout = 30 * time.Second
	maxPageSize    = 8 << 20
)

// Provider retrieves and caches the canonical AGENTS.md.
type Provider struct {
	client *http.Client
	url    string

	once      sync.Once
	canonical fact.Fact[string]
}

// New returns a Provider reading the default document. A nil client uses a
// default client with a timeout.
func New(client *http.Client) *Provider {
	return NewForDocument(client, AgentsDocument)
}

// NewForDocument returns a Provider reading document. A nil client uses a
// default client with a timeout.
func NewForDocument(client *http.Client, document Document) *Provider {
	if client == nil {
		client = &http.Client{Timeout: requestTimeout}
	}

	return &Provider{client: client, url: string(document)}
}

// Canonical returns the canonical AGENTS.md. The source is contacted at most
// once; every later call returns the same fact.
func (p *Provider) Canonical() fact.Fact[string] {
	p.once.Do(func() {
		p.canonical = p.retrieve()
	})

	return p.canonical
}

func (p *Provider) retrieve() fact.Fact[string] {
	page, err := p.fetch()
	if err != nil {
		return fact.Failed[string](err)
	}

	if len(page) == 0 {
		return fact.Failed[string](fmt.Errorf("reading %s: response is empty", p.url))
	}

	return fact.Present(string(page))
}

func (p *Provider) fetch() ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, p.url, nil)
	if err != nil {
		return nil, fmt.Errorf("requesting %s: %w", p.url, err)
	}

	response, err := p.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("requesting %s: %w", p.url, err)
	}
	defer func() {
		_ = response.Body.Close()
	}()

	if response.StatusCode != http.StatusOK {
		_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, maxPageSize))
		return nil, fmt.Errorf("requesting %s: unexpected status %s", p.url, response.Status)
	}

	page, err := io.ReadAll(io.LimitReader(response.Body, maxPageSize+1))
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", p.url, err)
	}
	if len(page) > maxPageSize {
		return nil, fmt.Errorf("reading %s: page is larger than %d bytes", p.url, maxPageSize)
	}

	return page, nil
}
