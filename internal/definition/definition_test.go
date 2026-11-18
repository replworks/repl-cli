package definition

import (
	"slices"
	"testing"
)

const spec = "" +
	"# PRODUCT_SPEC.md\n" +
	"\n" +
	"```text\n" +
	"- FR-001: fenced, not a definition\n" +
	"```\n" +
	"- FR-001: first definition\n" +
	"  - continuation one\n" +
	"\t- continuation two\n" +
	"- FR-002: next definition\n" +
	"- FR-001: second definition\n" +
	"- EL-001: element\n" +
	"  - Purpose: p\n" +
	"\n" +
	"  - after empty line\n" +
	"- ERR-001: error\n" +
	"- AC-001: criterion\n"

const architecture = "" +
	"# ARCHITECTURE.md\n" +
	"\n" +
	"### C2. Second\n" +
	"\n" +
	"- body line\n" +
	"```text\n" +
	"## fenced heading\n" +
	"```\n" +
	"#### deeper heading stays\n" +
	"\n" +
	"\n" +
	"### C3. Third\n" +
	"- body\n" +
	"## 8. Architectural Rules\n" +
	"\n" +
	"- AR-03: Pure components.\n" +
	"  continued\n" +
	"- AR-04: Next rule.\n"

const tasks = "" +
	"# TASKS.md\n" +
	"\n" +
	"- [x] T-001 One\n" +
	"  - Satisfies: FR-001\n" +
	"\n" +
	"- [ ] T-002 Two\n" +
	"  - Satisfies: C2\n"

func TestValid(t *testing.T) {
	for _, id := range []string{"T-1", "T-033", "FR-001", "EL-009", "ERR-012", "AC-086", "C1", "C19", "AR-03", "AR-123"} {
		if !Valid(id) {
			t.Errorf("Valid(%q) = false, want true", id)
		}
	}
	for _, id := range []string{"", "foo", "FR-1", "FR-0001", "C", "C-1", "T-", "AR-", "fr-001", "X-001", "FR-001:"} {
		if Valid(id) {
			t.Errorf("Valid(%q) = true, want false", id)
		}
	}
}

func TestDocumentOf(t *testing.T) {
	cases := map[string]string{
		"T-001": TasksPath, "FR-001": ProductSpecPath, "EL-001": ProductSpecPath,
		"ERR-001": ProductSpecPath, "AC-001": ProductSpecPath, "C3": ArchitecturePath, "AR-01": ArchitecturePath,
	}
	for id, want := range cases {
		if got := DocumentOf(id); got != want {
			t.Errorf("DocumentOf(%q) = %q, want %q", id, got, want)
		}
	}
}

func TestFind(t *testing.T) {
	tests := []struct {
		id, text  string
		wantLine  int
		wantLines []string
	}{
		{"FR-001", spec, 6, []string{"- FR-001: first definition", "  - continuation one", "\t- continuation two"}},
		{"FR-002", spec, 9, []string{"- FR-002: next definition"}},
		{"EL-001", spec, 11, []string{"- EL-001: element", "  - Purpose: p"}},
		{"ERR-001", spec, 15, []string{"- ERR-001: error"}},
		{"AC-001", spec, 16, []string{"- AC-001: criterion"}},
		{"C2", architecture, 3, []string{"### C2. Second", "", "- body line", "```text", "## fenced heading", "```", "#### deeper heading stays"}},
		{"C3", architecture, 12, []string{"### C3. Third", "- body"}},
		{"AR-03", architecture, 16, []string{"- AR-03: Pure components.", "  continued"}},
		{"T-001", tasks, 3, []string{"- [x] T-001 One", "  - Satisfies: FR-001"}},
		{"T-002", tasks, 6, []string{"- [ ] T-002 Two", "  - Satisfies: C2"}},
	}
	for _, tt := range tests {
		t.Run(tt.id, func(t *testing.T) {
			got, ok := Find(tt.id, tt.text)
			if !ok {
				t.Fatalf("Find(%q) not found", tt.id)
			}
			if got.ID != tt.id || got.Path != DocumentOf(tt.id) || got.Line != tt.wantLine || !slices.Equal(got.Lines, tt.wantLines) {
				t.Fatalf("Find(%q) = %+v, want line %d lines %q", tt.id, got, tt.wantLine, tt.wantLines)
			}
		})
	}
}

func TestFindNotFound(t *testing.T) {
	for _, tt := range []struct{ id, text string }{
		{"FR-999", spec},
		{"C9", architecture},
		{"T-009", tasks},
		{"FR-001", "```\n- FR-001: fenced\n```\n"},
		{"C1", "### C10. Ten\n"},
		{"FR-001", "- FR-0011: longer\n"},
	} {
		if got, ok := Find(tt.id, tt.text); ok {
			t.Errorf("Find(%q) = %+v, want not found", tt.id, got)
		}
	}
}

func TestFindKeepsTextExactly(t *testing.T) {
	got, ok := Find("FR-001", "- FR-001: a  b \r\n  - c\t\r\nx\n")
	if !ok || !slices.Equal(got.Lines, []string{"- FR-001: a  b ", "  - c\t"}) {
		t.Fatalf("Find() = %+v", got)
	}
	if got.Text() != "- FR-001: a  b \n  - c\t" {
		t.Fatalf("Text() = %q", got.Text())
	}
}
