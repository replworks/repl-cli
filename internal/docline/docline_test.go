package docline

import (
	"strings"
	"testing"
)

func TestClassifyLines(t *testing.T) {
	tests := []struct {
		name  string
		text  string
		texts []string
	}{
		{name: "empty", text: "", texts: nil},
		{name: "single line without newline", text: "a", texts: []string{"a"}},
		{name: "trailing newline ends no extra line", text: "a\nb\n", texts: []string{"a", "b"}},
		{name: "blank line is a line", text: "a\n\nb", texts: []string{"a", "", "b"}},
		{name: "crlf is stripped", text: "a\r\nb\r\n", texts: []string{"a", "b"}},
		{name: "lone lf text", text: "\n", texts: []string{""}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lines := Classify(tt.text)
			if len(lines) != len(tt.texts) {
				t.Fatalf("Classify() returned %d lines, want %d", len(lines), len(tt.texts))
			}
			for i, line := range lines {
				if line.Number != i+1 {
					t.Fatalf("line %d number = %d, want %d", i, line.Number, i+1)
				}
				if line.Text != tt.texts[i] {
					t.Fatalf("line %d text = %q, want %q", i+1, line.Text, tt.texts[i])
				}
			}
		})
	}
}

func TestClassifyFences(t *testing.T) {
	tests := []struct {
		name   string
		text   string
		inside []bool
	}{
		{
			name:   "backtick fence",
			text:   "head\n```\ninside\n```\ntail\n",
			inside: []bool{false, true, true, true, false},
		},
		{
			name:   "tilde fence",
			text:   "~~~\ninside\n~~~\n",
			inside: []bool{true, true, true},
		},
		{
			name:   "unclosed fence extends to the end",
			text:   "```\ninside\ntail\n",
			inside: []bool{true, true, true},
		},
		{
			name:   "different character does not close",
			text:   "```\n~~~\ninside\n```\n",
			inside: []bool{true, true, true, true},
		},
		{
			name:   "shorter run does not close",
			text:   "````\n```\ninside\n````\n",
			inside: []bool{true, true, true, true},
		},
		{
			name:   "indentation up to three spaces opens",
			text:   "   ```\ninside\n   ```\n",
			inside: []bool{true, true, true},
		},
		{
			name:   "four spaces do not open",
			text:   "    ```\ninside\n",
			inside: []bool{false, false},
		},
		{
			name:   "info string opens",
			text:   "```go\ninside\n```\n",
			inside: []bool{true, true, true},
		},
		{
			name:   "closing fence with trailing spaces closes",
			text:   "```\ninside\n```  \nafter\n",
			inside: []bool{true, true, true, false},
		},
		{
			name:   "two backticks do not open",
			text:   "``\ninside\n",
			inside: []bool{false, false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lines := Classify(tt.text)
			if len(lines) != len(tt.inside) {
				t.Fatalf("Classify() returned %d lines, want %d", len(lines), len(tt.inside))
			}
			for i, line := range lines {
				if line.InsideFence != tt.inside[i] {
					t.Fatalf("line %d (%q) insideFence = %v, want %v", i+1, line.Text, line.InsideFence, tt.inside[i])
				}
			}
		})
	}
}

func TestClassifyKeepsContent(t *testing.T) {
	text := strings.Join([]string{"- [ ] T-001 first", "  indented", ""}, "\n") + "\n"

	lines := Classify(text)
	if len(lines) != 3 {
		t.Fatalf("Classify() returned %d lines, want 3", len(lines))
	}
	if lines[0].Text != "- [ ] T-001 first" {
		t.Fatalf("line 1 text = %q, want the original text", lines[0].Text)
	}
	if lines[1].Text != "  indented" {
		t.Fatalf("line 2 text = %q, want leading spaces preserved", lines[1].Text)
	}
	if lines[2].Text != "" {
		t.Fatalf("line 3 text = %q, want empty", lines[2].Text)
	}
}
