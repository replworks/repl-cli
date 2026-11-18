// Package showproject prints the definition excerpts of given identifiers.
package showproject

import (
	"github.com/replworks/repl-cli/internal/definition"
	"github.com/replworks/repl-cli/internal/fact"
	"github.com/replworks/repl-cli/internal/workspace"
)

type Outcome struct {
	Definitions []definition.Excerpt
	Missing     []string
	Errors      []error
}

type Coordinator struct {
	Workspace workspace.Workspace
}

// Run looks up ids, which are valid, distinct, and in given order.
func (c Coordinator) Run(ids []string) Outcome {
	needed := map[string]bool{}
	for _, id := range ids {
		needed[definition.DocumentOf(id)] = true
	}

	docs, errs := ReadDocuments(c.Workspace, needed)
	outcome := Outcome{Errors: errs}
	for _, id := range ids {
		excerpt, found, known := Lookup(id, docs)
		switch {
		case found:
			outcome.Definitions = append(outcome.Definitions, excerpt)
		case known:
			outcome.Missing = append(outcome.Missing, id)
		}
	}

	return outcome
}

// ReadDocuments reads each needed defining document once, in the fixed
// document order, and returns one execution error per failed read.
func ReadDocuments(ws workspace.Workspace, needed map[string]bool) (map[string]fact.Fact[string], []error) {
	docs := map[string]fact.Fact[string]{}
	var errs []error
	for _, path := range definition.DocumentOrder {
		if !needed[path] {
			continue
		}
		text := ws.ReadText(path)
		docs[path] = text
		if text.IsFailed() {
			errs = append(errs, text.Reason())
		}
	}

	return docs, errs
}

// Lookup finds id in its defining document. known is false when that
// document could not be read, so the identifier is neither found nor missing.
func Lookup(id string, docs map[string]fact.Fact[string]) (excerpt definition.Excerpt, found bool, known bool) {
	text, ok := docs[definition.DocumentOf(id)]
	if !ok || text.IsFailed() {
		return definition.Excerpt{}, false, false
	}
	if text.IsAbsent() {
		return definition.Excerpt{}, false, true
	}

	excerpt, found = definition.Find(id, text.Value())
	return excerpt, found, true
}
