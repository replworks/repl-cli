package definition

import (
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/replworks/repl-cli/internal/docline"
)

var (
	identifierRun = regexp.MustCompile(`[A-Za-z0-9-]+`)
	formPattern   = regexp.MustCompile(`^(T-|FR-|EL-|ERR-|AC-|C|AR-)([0-9]+)$`)
)

// References returns the distinct identifiers referenced in excerpt, outside
// fenced code blocks, in first-reference order, excluding self. docs holds the
// texts of the defining documents that are present, by path; a range
// `<A> through <B>` expands to the identifiers of that form defined there
// whose number lies strictly between A and B, in ascending order.
func References(excerpt, self string, docs map[string]string) []string {
	var refs []string
	add := func(id string) {
		if id != self && !slices.Contains(refs, id) {
			refs = append(refs, id)
		}
	}

	for _, line := range docline.Classify(excerpt) {
		if line.InsideFence {
			continue
		}
		runs := identifierRun.FindAllStringIndex(line.Text, -1)
		for i, run := range runs {
			id := line.Text[run[0]:run[1]]
			if !Valid(id) {
				continue
			}
			add(id)

			if i+2 < len(runs) {
				word, next := runs[i+1], runs[i+2]
				if line.Text[run[1]:next[0]] == " through " && line.Text[word[0]:word[1]] == "through" {
					for _, between := range between(id, line.Text[next[0]:next[1]], docs) {
						add(between)
					}
				}
			}
		}
	}

	return refs
}

func between(from, to string, docs map[string]string) []string {
	fromForm, fromNumber, ok := split(from)
	if !ok {
		return nil
	}
	toForm, toNumber, ok := split(to)
	if !ok || !Valid(to) || toForm != fromForm || fromNumber >= toNumber {
		return nil
	}

	text, ok := docs[DocumentOf(from)]
	if !ok {
		return nil
	}

	var ids []string
	for _, id := range defined(fromForm, DocumentOf(from), text) {
		_, number, _ := split(id)
		if number > fromNumber && number < toNumber && !slices.Contains(ids, id) {
			ids = append(ids, id)
		}
	}
	slices.SortFunc(ids, func(a, b string) int {
		_, x, _ := split(a)
		_, y, _ := split(b)
		return x - y
	})

	return ids
}

// defined lists the identifiers of form that have a definition line in text.
func defined(form, path, text string) []string {
	var ids []string
	for _, line := range docline.Classify(text) {
		if line.InsideFence {
			continue
		}
		for _, run := range identifierRun.FindAllString(line.Text, 3) {
			if strings.HasPrefix(run, form) && Valid(run) && isDefinition(run, path, line) {
				if f, _, _ := split(run); f == form {
					ids = append(ids, run)
				}
			}
		}
	}

	return ids
}

func split(id string) (string, int, bool) {
	match := formPattern.FindStringSubmatch(id)
	if match == nil {
		return "", 0, false
	}
	number, err := strconv.Atoi(match[2])
	if err != nil {
		return "", 0, false
	}

	return match[1], number, true
}
