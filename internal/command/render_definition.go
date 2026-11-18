package command

import (
	"fmt"
	"io"

	"github.com/replworks/repl-cli/internal/definition"
)

type definitionJSON struct {
	ID   string `json:"id"`
	Path string `json:"path"`
	Line int    `json:"line"`
	Text string `json:"text"`
}

type showJSON struct {
	Definitions []definitionJSON `json:"definitions"`
	Missing     []string         `json:"missing"`
	Errors      []string         `json:"errors"`
}

func toDefinitionJSON(excerpt definition.Excerpt) definitionJSON {
	return definitionJSON{ID: excerpt.ID, Path: excerpt.Path, Line: excerpt.Line, Text: excerpt.Text()}
}

func definitionsJSON(excerpts []definition.Excerpt) []definitionJSON {
	out := make([]definitionJSON, 0, len(excerpts))
	for _, excerpt := range excerpts {
		out = append(out, toDefinitionJSON(excerpt))
	}

	return out
}

func nonNil(values []string) []string {
	if values == nil {
		return []string{}
	}

	return values
}

func renderExcerpt(out io.Writer, excerpt definition.Excerpt) {
	_, _ = fmt.Fprintf(out, "==> %s %s:%d\n", excerpt.ID, excerpt.Path, excerpt.Line)
	for _, line := range excerpt.Lines {
		_, _ = fmt.Fprintln(out, line)
	}
}

func renderExcerpts(out io.Writer, excerpts []definition.Excerpt) {
	for _, excerpt := range excerpts {
		renderExcerpt(out, excerpt)
	}
}

func renderMissing(out io.Writer, missing []string) {
	for _, id := range missing {
		_, _ = fmt.Fprintf(out, "%s: no definition\n", id)
	}
}
