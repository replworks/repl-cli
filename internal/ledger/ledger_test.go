package ledger

import "testing"

func TestNewDerivations(t *testing.T) {
	text := "" +
		"- [X] T-001 first\n" +
		"- [x] T-002 second\n" +
		"- [ ] T-003 third\n" +
		"- [ ] T-004 fourth\n" +
		"- [ ] T-005 fifth\n"

	led := New(text)
	if led.Total != 5 {
		t.Fatalf("Total = %d, want 5", led.Total)
	}
	if led.Complete != 2 {
		t.Fatalf("Complete = %d, want 2", led.Complete)
	}
	if led.FirstIncompleteID != "T-003" {
		t.Fatalf("FirstIncompleteID = %q, want %q", led.FirstIncompleteID, "T-003")
	}
}

func TestNewEmptyLedger(t *testing.T) {
	led := New("")
	if led.Total != 0 || led.Complete != 0 || led.FirstIncompleteID != "" {
		t.Fatalf("New(\"\") = %+v, want an empty ledger", led)
	}
}

func TestNewAllComplete(t *testing.T) {
	led := New("- [x] T-001 one\n- [X] T-002 two\n")
	if led.Total != 2 || led.Complete != 2 || led.FirstIncompleteID != "" {
		t.Fatalf("New() = %+v, want 2 total, 2 complete, no incomplete", led)
	}
}

func TestNewIgnoresNonEntries(t *testing.T) {
	text := "" +
		"# TASKS.md\n" +
		"\n" +
		"- [ ] T-001 counted\n" +
		"- [ ] T-002 counted\n" +
		"  - [ ] T-900 indented\n" +
		"- [x]T-901 no space after bracket\n" +
		"- [] T-902 empty box\n" +
		"- [xx] T-903 two characters\n" +
		"- [-] T-904 invalid box\n" +
		"* [ ] T-905 bullet\n" +
		"- [ ]\n" +
		"- [ ]    \n" +
		"    - [ ] T-906 four spaces\n" +
		"text - [ ] T-907 inline\n"

	led := New(text)
	if led.Total != 2 {
		t.Fatalf("Total = %d, want 2", led.Total)
	}
	if led.Complete != 0 {
		t.Fatalf("Complete = %d, want 0", led.Complete)
	}
	if led.FirstIncompleteID != "T-001" {
		t.Fatalf("FirstIncompleteID = %q, want %q", led.FirstIncompleteID, "T-001")
	}
}

func TestNewIgnoresFencedEntries(t *testing.T) {
	text := "" +
		"```\n" +
		"- [ ] T-900 fenced\n" +
		"```\n" +
		"- [ ] T-001 real\n"

	led := New(text)
	if led.Total != 1 {
		t.Fatalf("Total = %d, want 1", led.Total)
	}
	if led.FirstIncompleteID != "T-001" {
		t.Fatalf("FirstIncompleteID = %q, want %q", led.FirstIncompleteID, "T-001")
	}
}

func TestNewTaskID(t *testing.T) {
	tests := []struct {
		name string
		line string
		want string
	}{
		{name: "id only", line: "- [ ] T-001", want: "T-001"},
		{name: "id then title", line: "- [ ] T-001 title", want: "T-001"},
		{name: "id then tab", line: "- [ ] T-001\ttitle", want: "T-001"},
		{name: "id then multiple spaces", line: "- [ ] T-001   title", want: "T-001"},
		{name: "id without separator", line: "- [ ] T-001:", want: "T-001:"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			led := New(tt.line + "\n")
			if led.Total != 1 {
				t.Fatalf("Total = %d, want 1", led.Total)
			}
			if led.FirstIncompleteID != tt.want {
				t.Fatalf("FirstIncompleteID = %q, want %q", led.FirstIncompleteID, tt.want)
			}
		})
	}
}

func TestNewHandlesCRLF(t *testing.T) {
	led := New("- [x] T-001 one\r\n- [ ] T-002 two\r\n")
	if led.Total != 2 || led.Complete != 1 {
		t.Fatalf("New() = %+v, want 2 total and 1 complete", led)
	}
	if led.FirstIncompleteID != "T-002" {
		t.Fatalf("FirstIncompleteID = %q, want %q", led.FirstIncompleteID, "T-002")
	}
}
