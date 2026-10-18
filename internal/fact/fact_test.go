package fact

import (
	"errors"
	"testing"
)

func TestConstructorsSetKind(t *testing.T) {
	present := Present("value")
	if !present.IsPresent() || present.IsAbsent() || present.IsFailed() {
		t.Fatalf("Present() kind = %v, want present", present.Kind())
	}
	if got := present.Value(); got != "value" {
		t.Fatalf("Present() value = %q, want %q", got, "value")
	}

	absent := Absent[string]()
	if absent.IsPresent() || !absent.IsAbsent() || absent.IsFailed() {
		t.Fatalf("Absent() kind = %v, want absent", absent.Kind())
	}
	if got := absent.Value(); got != "" {
		t.Fatalf("Absent() value = %q, want the zero value", got)
	}

	reason := errors.New("boom")
	failed := Failed[string](reason)
	if failed.IsPresent() || failed.IsAbsent() || !failed.IsFailed() {
		t.Fatalf("Failed() kind = %v, want failed", failed.Kind())
	}
	if !errors.Is(failed.Reason(), reason) {
		t.Fatalf("Failed() reason = %v, want %v", failed.Reason(), reason)
	}
}
