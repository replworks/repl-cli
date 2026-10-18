package mark

import (
	"reflect"
	"testing"
)

const path = ".replworks/ARCHITECTURE.md"

func TestScanReportsAMarkInAHeading(t *testing.T) {
	text := "# ARCHITECTURE.md\n\n## 1. Purpose [UNVERIFIED]\n"

	got := Scanner{}.Scan(path, text)

	want := []Mark{{Path: path, Line: 3, Heading: "## 1. Purpose [UNVERIFIED]"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("marks = %v, want %v", got, want)
	}
}

func TestScanReportsEveryOccurrenceInOneHeading(t *testing.T) {
	text := "## Purpose [UNVERIFIED] and [UNVERIFIED]\n"

	got := Scanner{}.Scan(path, text)

	want := []Mark{
		{Path: path, Line: 1, Heading: "## Purpose [UNVERIFIED] and [UNVERIFIED]"},
		{Path: path, Line: 1, Heading: "## Purpose [UNVERIFIED] and [UNVERIFIED]"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("marks = %v, want %v", got, want)
	}
}

func TestScanReportsMarksInDocumentOrder(t *testing.T) {
	text := "" +
		"### Third [UNVERIFIED]\n" +
		"## First [UNVERIFIED]\n" +
		"\n" +
		"#### Fourth [UNVERIFIED]\n"

	got := Scanner{}.Scan(path, text)

	want := []Mark{
		{Path: path, Line: 1, Heading: "### Third [UNVERIFIED]"},
		{Path: path, Line: 2, Heading: "## First [UNVERIFIED]"},
		{Path: path, Line: 4, Heading: "#### Fourth [UNVERIFIED]"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("marks = %v, want %v", got, want)
	}
}

func TestScanIgnoresMarksOutsideHeadings(t *testing.T) {
	text := "" +
		"Add [UNVERIFIED] to the heading.\n" +
		"- [ ] T-014 Report [UNVERIFIED] marks\n" +
		"| section | [UNVERIFIED] |\n" +
		"#No space after the hash [UNVERIFIED]\n" +
		"####### Seven hashes [UNVERIFIED]\n" +
		"  ## Indented heading [UNVERIFIED]\n" +
		"#\n" +
		"# Valid [UNVERIFIED]\n"

	got := Scanner{}.Scan(path, text)

	want := []Mark{{Path: path, Line: 8, Heading: "# Valid [UNVERIFIED]"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("marks = %v, want %v", got, want)
	}
}

func TestScanIgnoresMarksInsideFencedBlocks(t *testing.T) {
	text := "" +
		"```\n" +
		"## Inside [UNVERIFIED]\n" +
		"```\n" +
		"## Outside [UNVERIFIED]\n" +
		"~~~\n" +
		"## Inside again [UNVERIFIED]\n" +
		"~~~~\n" +
		"## Outside again [UNVERIFIED]\n"

	got := Scanner{}.Scan(path, text)

	want := []Mark{
		{Path: path, Line: 4, Heading: "## Outside [UNVERIFIED]"},
		{Path: path, Line: 8, Heading: "## Outside again [UNVERIFIED]"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("marks = %v, want %v", got, want)
	}
}

func TestScanReportsNoMarkForAnUnclosedFence(t *testing.T) {
	text := "## Heading [UNVERIFIED]\n```\n## Inside [UNVERIFIED]\n"

	got := Scanner{}.Scan(path, text)

	want := []Mark{{Path: path, Line: 1, Heading: "## Heading [UNVERIFIED]"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("marks = %v, want %v", got, want)
	}
}

func TestScanCountsLinesOfTextWithCarriageReturns(t *testing.T) {
	text := "# Title\r\n\r\n## Purpose [UNVERIFIED]\r\n"

	got := Scanner{}.Scan(path, text)

	want := []Mark{{Path: path, Line: 3, Heading: "## Purpose [UNVERIFIED]"}}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("marks = %v, want %v", got, want)
	}
}

func TestScanReportsNoMarkForEmptyText(t *testing.T) {
	if got := (Scanner{}).Scan(path, ""); len(got) != 0 {
		t.Fatalf("marks = %v, want none", got)
	}
}
