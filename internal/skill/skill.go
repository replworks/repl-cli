// Package skill obtains the five published Agent Skills for installation.
package skill

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/replworks/repl-cli/internal/fact"
)

// Source identifies one installable Agent Skill.
type Source struct {
	Name string
	URL  string
	Path string
}

// Sources is the fixed installation order of the five Agent Skills.
var Sources = []Source{
	{Name: "repl-product-spec", URL: "https://www.repl.net/skills/repl-product-spec/SKILL.md", Path: ".claude/skills/repl-product-spec/SKILL.md"},
	{Name: "repl-tech-stack", URL: "https://www.repl.net/skills/repl-tech-stack/SKILL.md", Path: ".claude/skills/repl-tech-stack/SKILL.md"},
	{Name: "repl-architecture", URL: "https://www.repl.net/skills/repl-architecture/SKILL.md", Path: ".claude/skills/repl-architecture/SKILL.md"},
	{Name: "repl-tasks", URL: "https://www.repl.net/skills/repl-tasks/SKILL.md", Path: ".claude/skills/repl-tasks/SKILL.md"},
	{Name: "repl-auto-run", URL: "https://www.repl.net/skills/repl-auto-run/SKILL.md", Path: ".claude/skills/repl-auto-run/SKILL.md"},
}

const (
	requestTimeout = 30 * time.Second
	maxPageSize    = 8 << 20
)

// Provider retrieves the published skill files. Each URL is contacted at most
// once per Provider.
type Provider struct {
	client *http.Client

	mu    sync.Mutex
	cache map[string]fact.Fact[string]
}

// New returns a Provider using client. A nil client uses a default client
// with a timeout.
func New(client *http.Client) *Provider {
	if client == nil {
		client = &http.Client{Timeout: requestTimeout}
	}

	return &Provider{client: client, cache: map[string]fact.Fact[string]{}}
}

// Fetch returns the skill file published at url. The URL is contacted at most
// once; every later call returns the same fact.
func (p *Provider) Fetch(url string) fact.Fact[string] {
	p.mu.Lock()
	defer p.mu.Unlock()

	if cached, ok := p.cache[url]; ok {
		return cached
	}

	cached := p.retrieve(url)
	p.cache[url] = cached

	return cached
}

func (p *Provider) retrieve(url string) fact.Fact[string] {
	page, err := p.fetch(url)
	if err != nil {
		return fact.Failed[string](err)
	}

	if len(page) == 0 {
		return fact.Failed[string](fmt.Errorf("reading %s: response is empty", url))
	}

	return fact.Present(string(page))
}

func (p *Provider) fetch(url string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("requesting %s: %w", url, err)
	}

	response, err := p.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("requesting %s: %w", url, err)
	}
	defer func() {
		_ = response.Body.Close()
	}()

	if response.StatusCode != http.StatusOK {
		_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, maxPageSize))
		return nil, fmt.Errorf("requesting %s: unexpected status %s", url, response.Status)
	}

	page, err := io.ReadAll(io.LimitReader(response.Body, maxPageSize+1))
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", url, err)
	}
	if len(page) > maxPageSize {
		return nil, fmt.Errorf("reading %s: page is larger than %d bytes", url, maxPageSize)
	}

	return page, nil
}
