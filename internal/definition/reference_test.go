package definition

import (
	"slices"
	"testing"
)

func TestReferencesOrderAndBoundaries(t *testing.T) {
	excerpt := "" +
		"- [ ] T-002 Two (see T-001, T-002)\n" +
		"  - Satisfies: (FR-001, C3); `AR-03` and FR-001 again; EL-009.\n" +
		"  - Not references: XFR-001 FR-0011 FR-001x t-033-slug C3a UTF-8 AC-1 ERR-01\n"

	got := References(excerpt, "T-002", nil)

	want := []string{"T-001", "FR-001", "C3", "AR-03", "EL-009"}
	if !slices.Equal(got, want) {
		t.Fatalf("References() = %q, want %q", got, want)
	}
}

func TestReferencesIgnoreFencedCode(t *testing.T) {
	excerpt := "- [ ] T-001 One\n  ```text\n  FR-001\n  ```\n  - FR-002\n"

	got := References(excerpt, "T-001", nil)

	if !slices.Equal(got, []string{"FR-002"}) {
		t.Fatalf("References() = %q, want [FR-002]", got)
	}
}

func TestReferencesExpandRanges(t *testing.T) {
	docs := map[string]string{
		ProductSpecPath:  "- AC-012: c\n- AC-010: a\n```\n- AC-013: fenced\n```\n- AC-011: b\n- AC-014: d\n- FR-011: other form\n",
		ArchitecturePath: "### C2. Two\n### C4. Four\n### C3. Three\n",
	}

	tests := []struct {
		excerpt string
		want    []string
	}{
		{"AC-010 through AC-014", []string{"AC-010", "AC-011", "AC-012", "AC-014"}},
		{"AC-009 through AC-012", []string{"AC-009", "AC-010", "AC-011", "AC-012"}},
		{"C1 through C9", []string{"C1", "C2", "C3", "C4", "C9"}},
		{"AC-014 through AC-010", []string{"AC-014", "AC-010"}},
		{"AC-010 through FR-020", []string{"AC-010", "FR-020"}},
		{"AC-010  through AC-014", []string{"AC-010", "AC-014"}},
		{"AC-010 to AC-014", []string{"AC-010", "AC-014"}},
	}
	for _, tt := range tests {
		if got := References(tt.excerpt, "", docs); !slices.Equal(got, tt.want) {
			t.Errorf("References(%q) = %q, want %q", tt.excerpt, got, tt.want)
		}
	}

	if got := References("AC-010 through AC-014", "", nil); !slices.Equal(got, []string{"AC-010", "AC-014"}) {
		t.Errorf("References() without documents = %q, want endpoints only", got)
	}
}

func TestReferencesExpandTaskRanges(t *testing.T) {
	docs := map[string]string{TasksPath: "- [x] T-001 a\n- [ ] T-002 b\n- [ ] T-003 c\n"}

	got := References("after T-001 through T-003", "T-009", docs)

	if !slices.Equal(got, []string{"T-001", "T-002", "T-003"}) {
		t.Fatalf("References() = %q", got)
	}
}

func TestFindTaskAcceptsAnyTaskIDForm(t *testing.T) {
	got, ok := FindTask("T-001:", "- [ ] T-001: odd\n  - body\n")
	if !ok || got.Path != TasksPath || got.Line != 1 || len(got.Lines) != 2 {
		t.Fatalf("FindTask() = %+v, %v", got, ok)
	}
}
