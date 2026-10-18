// Package canon obtains the canonical AGENTS.md from the project's website.
package canon

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/html"

	"github.com/replworks/repl-cli/internal/fact"
)

// SourceURL is the only network resource the program may access.
const SourceURL = "https://www.repl.net/documents/agents/"

const (
	requestTimeout = 30 * time.Second
	maxPageSize    = 8 << 20
	separator      = "\x7f"
)

// Provider retrieves and caches the canonical AGENTS.md.
type Provider struct {
	client *http.Client
	url    string

	once      sync.Once
	canonical fact.Fact[string]
}

// New returns a Provider reading from SourceURL. A nil client uses a default
// client with a timeout.
func New(client *http.Client) *Provider {
	if client == nil {
		client = &http.Client{Timeout: requestTimeout}
	}

	return &Provider{client: client, url: SourceURL}
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

	text, err := extract(page)
	if err != nil {
		return fact.Failed[string](err)
	}

	return fact.Present(text)
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

func extract(page []byte) (string, error) {
	document, err := html.Parse(bytes.NewReader(page))
	if err != nil {
		return "", fmt.Errorf("parsing %s: %w", SourceURL, err)
	}

	var texts []string

	var visit func(node *html.Node)
	visit = func(node *html.Node) {
		if node.Type == html.ElementNode {
			for _, attribute := range node.Attr {
				if attribute.Key == "data-code" {
					texts = append(texts, attribute.Val)
				}
			}
		}
		for child := node.FirstChild; child != nil; child = child.NextSibling {
			visit(child)
		}
	}
	visit(document)

	if len(texts) != 1 {
		return "", fmt.Errorf("%s holds %d canonical texts, want 1", SourceURL, len(texts))
	}

	return strings.ReplaceAll(texts[0], separator, "\n"), nil
}
