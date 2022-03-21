package pipeline

import (
	"context"
)

// Predicate is a function that expects 'true' if a ActionFunc should run.
// It is evaluated lazily resp. only when needed.
type Predicate func(ctx context.Context) bool

// ToStep wraps the given action func in its own step.
// When the step's function is called, the given Predicate will evaluate whether the action should actually run.
// It returns the action's Result, otherwise an empty (successful) Result struct.
// The context.Context from the pipeline is passed through the given action.
func ToStep(name string, action ActionFunc, predicate Predicate) Step {
	step := Step{Name: name}
	step.F = func(ctx context.Context) Result {
		if predicate(ctx) {
			return action(ctx)
		}
		return NewEmptyResult(name)
	}
	return step
}

// ToNestedStep wraps the given pipeline in its own step.
// When the step's function is called, the given Predicate will evaluate whether the nested Pipeline should actually run.
// It returns the pipeline's Result, otherwise an empty (successful) Result struct.
// The given pipeline has to define its own context.Context, it's not passed "down".
func ToNestedStep(name string, predicate Predicate, p *Pipeline) Step {
	step := Step{Name: name}
	step.F = func(ctx context.Context) Result {
		if predicate(ctx) {
			return p.Run()
		}
		return NewEmptyResult(name)
	}
	return step
}

// If returns a new step that wraps the given step and executes its action only if the given Predicate evaluates true.
// The context.Context from the pipeline is passed through the given action.
func If(predicate Predicate, originalStep Step) Step {
	wrappedStep := Step{Name: originalStep.Name}
	wrappedStep.F = func(ctx context.Context) Result {
		if predicate(ctx) {
			return originalStep.F(ctx)
		}
		return NewEmptyResult(originalStep.Name)
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
