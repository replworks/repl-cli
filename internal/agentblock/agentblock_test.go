package agentblock

import (
	"errors"
	"testing"
)

const canonical = "# AGENTS.md\n\nRule one.\n"

func TestExpectedWithoutBlocksIsCanonical(t *testing.T) {
	got, err := Expected("# old AGENTS.md\n", canonical)
	if err != nil {
		t.Fatalf("Expected() error = %v", err)
	}
	if got != canonical {
		t.Fatalf("Expected() = %q, want %q", got, canonical)
	}
}

func TestExpectedKeepsTagBlock(t *testing.T) {
	block := "<laravel-boost-guidelines>\n=== foundation rules ===\n\n```php\n</laravel-boost-guidelines>\n```\n</laravel-boost-guidelines>"
	local := "# old\n\n" + block + "\nTrailing text.\n"

	got, err := Expected(local, canonical)
	if err != nil {
		t.Fatalf("Expected() error = %v", err)
	}
	want := canonical + "\n" + block + "\n"
	if got != want {
		t.Fatalf("Expected() = %q, want %q", got, want)
	}
}

func TestExpectedKeepsCommentBlocksInOrder(t *testing.T) {
	first := "<!-- BEGIN:nextjs-agent-rules -->\n\n## Next.js\n\n<!-- END:nextjs-agent-rules -->"
	second := "<tool.rules>\nignored name\n</tool.rules>"
	third := "<x-tool_1>\nbody\n</x-tool_1>"
	local := third + "\n" + first + "\n" + second + "\n"

	got, err := Expected(local, canonical)
	if err != nil {
		t.Fatalf("Expected() error = %v", err)
	}
	want := canonical + "\n" + third + "\n" + "\n" + first + "\n"
	if got != want {
		t.Fatalf("Expected() = %q, want %q", got, want)
	}
}

func TestExpectedIgnoresLinesThatAreNotBlockLines(t *testing.T) {
	local := "```text\n<inside>\n</inside>\n```\n" +
		" <indented>\n</indented>\n" +
		"Commit author: <name> <noreply email>\n" +
		"<details open>\n</details>\n" +
		"<!-- BEGIN:x-->\n<!-- END:x -->\n"

	got, err := Expected(local, canonical)
	if err != nil {
		t.Fatalf("Expected() error = %v", err)
	}
	if got != canonical {
		t.Fatalf("Expected() = %q, want %q", got, canonical)
	}
}

func TestExpectedIgnoresClosingLineInsideFence(t *testing.T) {
	local := "<rules>\n```\n</rules>\n```\n"

	_, err := Expected(local, canonical)
	var unclosed *UnclosedError
	if !errors.As(err, &unclosed) {
		t.Fatalf("Expected() error = %v, want *UnclosedError", err)
	}
	if unclosed.Line != 1 || unclosed.Opening != "<rules>" {
		t.Fatalf("UnclosedError = %+v, want line 1 <rules>", unclosed)
	}
}

func TestExpectedReportsUnclosedOpeningLine(t *testing.T) {
	local := "# AGENTS.md\n\n<laravel-boost-guidelines>\nrules\n"

	_, err := Expected(local, canonical)
	var unclosed *UnclosedError
	if !errors.As(err, &unclosed) {
		t.Fatalf("Expected() error = %v, want *UnclosedError", err)
	}
	if unclosed.Line != 3 {
		t.Fatalf("UnclosedError.Line = %d, want 3", unclosed.Line)
	}
	if got := err.Error(); got != "AGENTS.md line 3 opens <laravel-boost-guidelines> without a matching closing line; add the closing line or remove the block" {
		t.Fatalf("Error() = %q", got)
	}
}

func TestExpectedSkipsOpeningLinesOfCanonical(t *testing.T) {
	withTag := "# AGENTS.md\n\n<example>\n"
	local := withTag + "<kept>\nbody\n</kept>\n"

	got, err := Expected(local, withTag)
	if err != nil {
		t.Fatalf("Expected() error = %v", err)
	}
	want := withTag + "\n<kept>\nbody\n</kept>\n"
	if got != want {
		t.Fatalf("Expected() = %q, want %q", got, want)
	}
}

func TestExpectedIsIdempotent(t *testing.T) {
	local := "old\n<rules>\nbody\n</rules>\n"

	first, err := Expected(local, canonical)
	if err != nil {
		t.Fatalf("Expected() error = %v", err)
	}
	second, err := Expected(first, canonical)
	if err != nil {
		t.Fatalf("Expected() error = %v", err)
	}
	if first != second {
		t.Fatalf("Expected(Expected()) = %q, want %q", second, first)
	}
}

func TestExpectedKeepsCRLFAndAddsLineBreakToCanonical(t *testing.T) {
	block := "<rules>\r\nbody\r\n</rules>\r"
	local := block + "\n"

	got, err := Expected(local, "# AGENTS.md")
	if err != nil {
		t.Fatalf("Expected() error = %v", err)
	}
	want := "# AGENTS.md\n\n" + block + "\n"
	if got != want {
		t.Fatalf("Expected() = %q, want %q", got, want)
	}
}
