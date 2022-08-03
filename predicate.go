package pipeline

import (
	"context"
)

// Predicate is a function that expects 'true' if a ActionFunc should run.
// It is evaluated lazily resp. only when needed.
type Predicate func(ctx context.Context) bool

// ToStep wraps the given action func in its own step.
// When the step's function is called, the given Predicate will evaluate whether the action should actually run.
// It returns the action's Result, otherwise an empty (successful) Result.
// The context.Context from the pipeline is passed through the given action.
func ToStep[T context.Context](name string, action ActionFunc[T], predicate Predicate) Step[T] {
	step := Step[T]{Name: name}
	step.Action = func(ctx T) error {
		if predicate(ctx) {
			return newResultWithError(name, action(ctx))
		}
		return newEmptyResult(name)
	}
	return step
}

// ToNestedStep wraps the given pipeline in its own step.
// When the step's function is called, the given Predicate will evaluate whether the nested Pipeline should actually run.
// It returns the pipeline's Result, otherwise an empty (successful) Result struct.
// The given pipeline has to define its own context.Context, it's not passed "down".
func ToNestedStep[T context.Context](name string, predicate Predicate, p *Pipeline[T]) Step[T] {
	step := Step[T]{Name: name}
	step.Action = func(ctx T) error {
		if predicate(ctx) {
			return p.RunWithContext(ctx)
		}
		return newEmptyResult(name)
	}
	return step
}

// If returns a new step that wraps the given step and executes its action only if the given Predicate evaluates true.
// The context.Context from the pipeline is passed through the given action.
func If[T context.Context](predicate Predicate, originalStep Step[T]) Step[T] {
	wrappedStep := Step[T]{Name: originalStep.Name}
	wrappedStep.Action = func(ctx T) error {
		if predicate(ctx) {
			return originalStep.Action(ctx)
		}
		return newEmptyResult(originalStep.Name)
	}
	return wrappedStep
}

// IfOrElse returns a new step that wraps the given steps and executes its action based on the given Predicate.
// The name of the step is taken from `trueStep`.
// The context.Context from the pipeline is passed through the given actions.
func IfOrElse[T context.Context](predicate Predicate, trueStep Step[T], falseStep Step[T]) Step[T] {
	wrappedStep := Step[T]{Name: trueStep.Name}
	wrappedStep.Action = func(ctx T) error {
		if predicate(ctx) {
			return trueStep.Action(ctx)
		} else {
			return falseStep.Action(ctx)
		}
	}
	return wrappedStep
}

// Bool returns a Predicate that simply returns v when evaluated.
// Use BoolPtr() over Bool() if the value can change between setting up the pipeline and evaluating the predicate.
func Bool(v bool) Predicate {
	return func(_ context.Context) bool {
		return v
	}
}

// BoolPtr returns a Predicate that returns *v when evaluated.
// Use BoolPtr() over Bool() if the value can change between setting up the pipeline and evaluating the predicate.
func BoolPtr(v *bool) Predicate {
	return func(_ context.Context) bool {
		return *v
	}
}

// Not returns a Predicate that evaluates, but then negates the given Predicate.
func Not(predicate Predicate) Predicate {
	return func(ctx context.Context) bool {
		return !predicate(ctx)
	}
}

// And returns a Predicate that does logical AND of the given predicates.
// p2 is not evaluated if p1 evaluates already to false.
func And(p1, p2 Predicate) Predicate {
	return func(ctx context.Context) bool {
		return p1(ctx) && p2(ctx)
	}
}

// Or returns a Predicate that does logical OR of the given predicates.
// p2 is not evaluated if p1 evaluates already to true.
func Or(p1, p2 Predicate) Predicate {
	return func(ctx context.Context) bool {
		return p1(ctx) || p2(ctx)
	}
}
