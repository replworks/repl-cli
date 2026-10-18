package diff

import (
	"reflect"
	"testing"
)

func TestCompareIdenticalTexts(t *testing.T) {
	text := "# AGENTS.md\n\n## DOCUMENT_ORDER\n"

	if got := Compare(text, text); len(got) != 0 {
		t.Fatalf("Compare() = %+v, want no entries", got)
	}
}

func TestCompareEmptyTexts(t *testing.T) {
	if got := Compare("", ""); len(got) != 0 {
		t.Fatalf("Compare() = %+v, want no entries", got)
	}
}

func TestCompareAddedLine(t *testing.T) {
	got := Compare("a\nb\n", "a\nx\nb\n")
	want := []Entry{{Kind: Added, Line: "x", Position: 2}}

	assertEntries(t, got, want)
}

func TestCompareRemovedLine(t *testing.T) {
	got := Compare("a\nx\nb\n", "a\nb\n")
	want := []Entry{{Kind: Removed, Line: "x", Position: 2}}

	assertEntries(t, got, want)
}

func TestCompareReplacedLine(t *testing.T) {
	got := Compare("a\nold\nb\n", "a\nnew\nb\n")
	want := []Entry{
		{Kind: Removed, Line: "old", Position: 2},
		{Kind: Added, Line: "new", Position: 2},
	}

	assertEntries(t, got, want)
}

func TestCompareSeveralChangesInDocumentOrder(t *testing.T) {
	got := Compare("1\n2\n3\n4\n5\n", "1\nX\n3\n4\nY\n")
	want := []Entry{
		{Kind: Removed, Line: "2", Position: 2},
		{Kind: Added, Line: "X", Position: 2},
		{Kind: Removed, Line: "5", Position: 5},
		{Kind: Added, Line: "Y", Position: 5},
	}

	assertEntries(t, got, want)
}

func TestCompareAddedAndRemovedLinesAtTheSamePosition(t *testing.T) {
	got := Compare("a\nb\n", "a\nx\ny\n")
	want := []Entry{
		{Kind: Removed, Line: "b", Position: 2},
		{Kind: Added, Line: "x", Position: 2},
		{Kind: Added, Line: "y", Position: 3},
	}

	assertEntries(t, got, want)
}

func TestCompareMissingFinalNewline(t *testing.T) {
	got := Compare("a\nb", "a\nb\n")
	want := []Entry{
		{Kind: Removed, Line: "b", Position: 2},
		{Kind: Added, Line: "b", Position: 2},
	}

	assertEntries(t, got, want)
}

func TestCompareIgnoresUnchangedLines(t *testing.T) {
	got := Compare("a\nb\nc\nd\n", "a\nB\nc\nD\n")

	for _, entry := range got {
		if entry.Line == "a" || entry.Line == "c" {
			t.Fatalf("Compare() = %+v, want unchanged lines to be omitted", got)
		}
	}
}

func TestCompareIsDeterministic(t *testing.T) {
	local := "1\n2\n3\n"
	canonical := "1\nx\n3\n"

	first := Compare(local, canonical)
	for i := 0; i < 5; i++ {
		if got := Compare(local, canonical); !reflect.DeepEqual(got, first) {
			t.Fatalf("Compare() = %+v, want the same result as %+v", got, first)
		}
	}
}

func TestCalculatorMatchesCompare(t *testing.T) {
	local := "a\nb\n"
	canonical := "a\nc\n"

	if got, want := (Calculator{}).Compare(local, canonical), Compare(local, canonical); !reflect.DeepEqual(got, want) {
		t.Fatalf("Calculator.Compare() = %+v, want %+v", got, want)
	}
}

func assertEntries(t *testing.T, got, want []Entry) {
	t.Helper()

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Compare() = %+v, want %+v", got, want)
	}
}
