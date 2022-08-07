package pipeline

import (
	"context"
)

// Predicate is a function that expects 'true' if an ActionFunc should run.
// It is evaluated lazily resp. only when needed.
// Predicate should be idempotent, meaning multiple invocations return the same result and without side effects.
type Predicate[T context.Context] func(ctx T) bool

// Bool returns a Predicate that simply returns v when evaluated.
// Use BoolPtr() over Bool() if the value can change between setting up the pipeline and evaluating the predicate.
func Bool[T context.Context](v bool) Predicate[T] {
	return func(_ T) bool {
		return v
	}
}

// BoolPtr returns a Predicate that returns *v when evaluated.
// Use BoolPtr() over Bool() if the value can change between setting up the pipeline and evaluating the predicate.
func BoolPtr[T context.Context](v *bool) Predicate[T] {
	return func(_ T) bool {
		return *v
	}
}

// Not returns a Predicate that evaluates, but then negates the given Predicate.
func Not[T context.Context](predicate Predicate[T]) Predicate[T] {
	return func(ctx T) bool {
		return !predicate(ctx)
	}
}

// And returns a Predicate that does logical AND of the given predicates.
// p2 is not evaluated if p1 evaluates already to false.
func And[T context.Context](p1, p2 Predicate[T]) Predicate[T] {
	return func(ctx T) bool {
		return p1(ctx) && p2(ctx)
	}
}

// Or returns a Predicate that does logical OR of the given predicates.
// p2 is not evaluated if p1 evaluates already to true.
func Or[T context.Context](p1, p2 Predicate[T]) Predicate[T] {
	return func(ctx T) bool {
		return p1(ctx) || p2(ctx)
	}
}
