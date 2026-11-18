// Package rule evaluates the check rules over gathered facts and owns the rule
// identifiers, their fixed evaluation order, and their violation messages.
package rule

import (
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/replworks/repl-cli/internal/agentblock"
	"github.com/replworks/repl-cli/internal/docline"
	"github.com/replworks/repl-cli/internal/fact"
	"github.com/replworks/repl-cli/internal/ledger"
	"github.com/replworks/repl-cli/internal/mark"
)

// RuleDocumentsExist is reported for each authoritative document that does not
// exist.
const RuleDocumentsExist = "documents-exist"

// RuleTaskIDsUnchanged is reported for each task ID that exists on the base
// branch and does not exist in the working tree.
const RuleTaskIDsUnchanged = "task-ids-unchanged"

// RuleSingleTaskCompletion is reported when the working tree holds more than
// one newly complete task.
const RuleSingleTaskCompletion = "single-task-completion"

// RuleNoUnverified is reported for each UNVERIFIED mark while the working tree
// holds at least one newly complete task.
const RuleNoUnverified = "no-unverified"

// RuleAgentsCanonical is reported for an AGENTS.md that differs from the
// expected AGENTS.md or contains an unclosed block opening line.
const RuleAgentsCanonical = "agents-canonical"

// RuleDuplicateDefinition is reported for each identifier defined more than
// once in its defining document.
const RuleDuplicateDefinition = "duplicate-definition"

// RuleUnresolvedReference is reported for each identifier that appears in
// TASKS.md without a definition in its defining document.
const RuleUnresolvedReference = "unresolved-reference"

const (
	agentsPath       = "AGENTS.md"
	tasksPath        = ".replworks/TASKS.md"
	productSpecPath  = ".replworks/PRODUCT_SPEC.md"
	architecturePath = ".replworks/ARCHITECTURE.md"
)

// Violation is a failed rule.
type Violation struct {
	Rule    string
	Path    string
	Message string
}

// Document is the fact of one authoritative document.
type Document struct {
	Path string
	Text fact.Fact[string]
}

// MarkList is the fact of the marks of one document.
type MarkList struct {
	Path  string
	Marks fact.Fact[[]mark.Mark]
}

// Facts is the bundle of facts that rule evaluation reads.
type Facts struct {
	Documents     []Document
	Agents        fact.Fact[string]
	Canonical     fact.Fact[string]
	WorkingLedger fact.Fact[ledger.Ledger]
	BaseLedger    fact.Fact[ledger.Ledger]
	MarkLists     []MarkList
	ProductSpec   fact.Fact[string]
	Architecture  fact.Fact[string]
	TasksText     fact.Fact[string]
}

// Evaluator owns the rule identifiers, their evaluation order, and their
// violation messages.
type Evaluator struct{}

// Evaluate returns the violations of every rule whose preconditions hold, in
// the fixed rule order.
func (Evaluator) Evaluate(facts Facts) []Violation {
	var violations []Violation

	for _, evaluate := range order {
		violations = append(violations, evaluate(facts)...)
	}

	return violations
}

// order is the fixed rule evaluation order.
var order = []func(Facts) []Violation{
	documentsExist,
	taskIDsUnchanged,
	singleTaskCompletion,
	noUnverified,
	agentsCanonical,
	duplicateDefinition,
	unresolvedReference,
}

func documentsExist(facts Facts) []Violation {
	var violations []Violation

	for _, document := range facts.Documents {
		if !document.Text.IsAbsent() {
			continue
		}

		violations = append(violations, Violation{
			Rule:    RuleDocumentsExist,
			Path:    document.Path,
			Message: "document does not exist",
		})
	}

	return violations
}

func taskIDsUnchanged(facts Facts) []Violation {
	if !facts.WorkingLedger.IsPresent() || facts.BaseLedger.IsFailed() {
		return nil
	}

	base := ledger.Ledger{}
	if facts.BaseLedger.IsPresent() {
		base = facts.BaseLedger.Value()
	}

	var violations []Violation
	working := facts.WorkingLedger.Value().IDs

	for _, id := range base.IDs {
		if slices.Contains(working, id) {
			continue
		}

		violations = append(violations, Violation{
			Rule:    RuleTaskIDsUnchanged,
			Path:    tasksPath,
			Message: fmt.Sprintf("missing task ID %s", id),
		})
	}

	return violations
}

func singleTaskCompletion(facts Facts) []Violation {
	count, ok := newlyCompleteCount(facts)
	if !ok || count <= 1 {
		return nil
	}

	return []Violation{{
		Rule:    RuleSingleTaskCompletion,
		Path:    tasksPath,
		Message: fmt.Sprintf("%d tasks are newly complete", count),
	}}
}

// newlyCompleteCount counts the working entries that are complete and whose
// task ID is not complete in the base ledger. Entries are counted
// individually, so duplicate IDs are separate entries. The second result is
// false when the rule preconditions do not hold.
func newlyCompleteCount(facts Facts) (int, bool) {
	if !facts.WorkingLedger.IsPresent() || facts.BaseLedger.IsFailed() {
		return 0, false
	}

	complete := map[string]bool{}
	if facts.BaseLedger.IsPresent() {
		for _, entry := range facts.BaseLedger.Value().Entries {
			if entry.Complete {
				complete[entry.ID] = true
			}
		}
	}

	count := 0
	for _, entry := range facts.WorkingLedger.Value().Entries {
		if entry.Complete && !complete[entry.ID] {
			count++
		}
	}

	return count, true
}

func noUnverified(facts Facts) []Violation {
	count, ok := newlyCompleteCount(facts)
	if !ok || count == 0 {
		return nil
	}

	for _, list := range facts.MarkLists {
		if list.Marks.IsFailed() {
			return nil
		}
	}

	var violations []Violation

	for _, list := range facts.MarkLists {
		if !list.Marks.IsPresent() {
			continue
		}

		for _, m := range list.Marks.Value() {
			violations = append(violations, Violation{
				Rule:    RuleNoUnverified,
				Path:    list.Path,
				Message: fmt.Sprintf("UNVERIFIED mark on line %d", m.Line),
			})
		}
	}

	return violations
}

func agentsCanonical(facts Facts) []Violation {
	if !facts.Agents.IsPresent() || !facts.Canonical.IsPresent() {
		return nil
	}
	expected, err := agentblock.Expected(facts.Agents.Value(), facts.Canonical.Value())
	if err != nil {
		return []Violation{{
			Rule:    RuleAgentsCanonical,
			Path:    agentsPath,
			Message: err.Error(),
		}}
	}
	if facts.Agents.Value() == expected {
		return nil
	}

	return []Violation{{
		Rule:    RuleAgentsCanonical,
		Path:    agentsPath,
		Message: "AGENTS.md differs from the canonical AGENTS.md",
	}}
}

var (
	productSpecDefinition  = regexp.MustCompile(`^- ((?:FR|EL|ERR|AC)-[0-9]{3}): `)
	architectureDefinition = regexp.MustCompile(`^### (C[0-9]+)\. `)
	referencePattern       = regexp.MustCompile(`((?:FR|EL|ERR|AC)-[0-9]{3}|C[0-9]+)`)
)

func duplicateDefinition(facts Facts) []Violation {
	var violations []Violation

	for _, spec := range []struct {
		path    string
		pattern *regexp.Regexp
		text    fact.Fact[string]
	}{
		{productSpecPath, productSpecDefinition, facts.ProductSpec},
		{architecturePath, architectureDefinition, facts.Architecture},
	} {
		if !spec.text.IsPresent() {
			continue
		}

		definitions := map[string][]int{}
		for _, line := range docline.Classify(spec.text.Value()) {
			if line.InsideFence {
				continue
			}
			id, ok := matchDefinition(spec.pattern, line.Text)
			if !ok {
				continue
			}
			definitions[id] = append(definitions[id], line.Number)
		}

		ids := make([]string, 0, len(definitions))
		for id := range definitions {
			ids = append(ids, id)
		}
		slices.Sort(ids)

		for _, id := range ids {
			lines := definitions[id]
			if len(lines) < 2 {
				continue
			}

			violations = append(violations, Violation{
				Rule:    RuleDuplicateDefinition,
				Path:    spec.path,
				Message: fmt.Sprintf("identifier %s is defined on lines %s", id, joinInts(lines)),
			})
		}
	}

	return violations
}

func matchDefinition(pattern *regexp.Regexp, line string) (string, bool) {
	match := pattern.FindStringSubmatch(line)
	if match == nil {
		return "", false
	}

	return match[1], true
}

func joinInts(numbers []int) string {
	parts := make([]string, len(numbers))
	for i, number := range numbers {
		parts[i] = fmt.Sprintf("%d", number)
	}

	return strings.Join(parts, ", ")
}

func isWordChar(char byte) bool {
	switch {
	case char >= 'A' && char <= 'Z':
		return true
	case char >= 'a' && char <= 'z':
		return true
	case char >= '0' && char <= '9':
		return true
	default:
		return char == '_'
	}
}

func unresolvedReference(facts Facts) []Violation {
	if !facts.TasksText.IsPresent() {
		return nil
	}

	defined := map[string]bool{}
	if facts.ProductSpec.IsPresent() {
		for _, line := range docline.Classify(facts.ProductSpec.Value()) {
			if line.InsideFence {
				continue
			}
			if id, ok := matchDefinition(productSpecDefinition, line.Text); ok {
				defined[id] = true
			}
		}
	}

	architectureDefined := map[string]bool{}
	if facts.Architecture.IsPresent() {
		for _, line := range docline.Classify(facts.Architecture.Value()) {
			if line.InsideFence {
				continue
			}
			if id, ok := matchDefinition(architectureDefinition, line.Text); ok {
				architectureDefined[id] = true
			}
		}
	}

	firstLine := map[string]int{}
	var order []string
	for _, line := range docline.Classify(facts.TasksText.Value()) {
		if line.InsideFence {
			continue
		}
		text := line.Text
		for _, position := range referencePattern.FindAllStringIndex(text, -1) {
			start, end := position[0], position[1]
			if start > 0 && isWordChar(text[start-1]) {
				continue
			}
			if end < len(text) && isWordChar(text[end]) {
				continue
			}

			id := text[start:end]
			if _, seen := firstLine[id]; seen {
				continue
			}
			firstLine[id] = line.Number
			order = append(order, id)
		}
	}

	var violations []Violation
	for _, id := range order {
		switch {
		case strings.HasPrefix(id, "C"):
			if !facts.Architecture.IsPresent() || architectureDefined[id] {
				continue
			}
		default:
			if !facts.ProductSpec.IsPresent() || defined[id] {
				continue
			}
		}

		violations = append(violations, Violation{
			Rule:    RuleUnresolvedReference,
			Path:    tasksPath,
			Message: fmt.Sprintf("identifier %s is referenced but never defined (first appearance on line %d)", id, firstLine[id]),
		})
	}

	return violations
}
