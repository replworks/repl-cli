// Package fact models one piece of information obtained from outside a pure
// component. Every fact is exactly one of present, absent, or failed.
package fact

type Kind int

const (
	KindPresent Kind = iota
	KindAbsent
	KindFailed
)

type Fact[T any] struct {
	kind   Kind
	value  T
	reason error
}

func Present[T any](value T) Fact[T] {
	return Fact[T]{kind: KindPresent, value: value}
}

func Absent[T any]() Fact[T] {
	return Fact[T]{kind: KindAbsent}
}

func Failed[T any](reason error) Fact[T] {
	return Fact[T]{kind: KindFailed, reason: reason}
}

func (f Fact[T]) Kind() Kind {
	return f.kind
}

func (f Fact[T]) IsPresent() bool {
	return f.kind == KindPresent
}

func (f Fact[T]) IsAbsent() bool {
	return f.kind == KindAbsent
}

func (f Fact[T]) IsFailed() bool {
	return f.kind == KindFailed
}

func (f Fact[T]) Value() T {
	return f.value
}

func (f Fact[T]) Reason() error {
	return f.reason
}
