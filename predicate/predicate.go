package predicate

import (
	pipeline "github.com/ccremer/go-command-pipeline"
)

type (
	// Predicate is a function that expects 'true' if a pipeline.ActionFunc should run.
	// It is evaluated lazily resp. only when needed.
	Predicate func(ctx pipeline.Context, step pipeline.Step) bool
)

// ToStep wraps the given action func in its own step.
// When the step's function is called, the given Predicate will evaluate whether the action should actually run.
// It returns the action's pipeline.Result, otherwise an empty (successful) pipeline.Result struct.
// The pipeline.Context from the pipeline is passed through the given action.
func ToStep(name string, action pipeline.ActionFunc, predicate Predicate) pipeline.Step {
	step := pipeline.Step{Name: name}
	step.F = func(ctx pipeline.Context) pipeline.Result {
		if predicate(ctx, step) {
			return action(ctx)
		}
		return pipeline.Result{}
	}
	return step
}

// ToNestedStep wraps the given pipeline in its own step.
// When the step's function is called, the given Predicate will evaluate whether the nested pipeline.Pipeline should actually run.
// It returns the pipeline's pipeline.Result, otherwise an empty (successful) pipeline.Result struct.
// The given pipeline has to define its own pipeline.Context, it's not passed "down".
func ToNestedStep(name string, p *pipeline.Pipeline, predicate Predicate) pipeline.Step {
	step := pipeline.Step{Name: name}
	step.F = func(ctx pipeline.Context) pipeline.Result {
		if predicate(ctx, step) {
			return p.Run()
		}
		return pipeline.Result{}
	}
	return step
}

// If returns a new step that wraps the given step and executes its action only if the given Predicate evaluates true.
// The pipeline.Context from the pipeline is passed through the given action.
func If(predicate Predicate, originalStep pipeline.Step) pipeline.Step {
	wrappedStep := pipeline.Step{Name: originalStep.Name}
	wrappedStep.F = func(ctx pipeline.Context) pipeline.Result {
		if predicate(ctx, wrappedStep) {
			return originalStep.F(ctx)
		}
		return pipeline.Result{}
	}
	return wrappedStep
}

// Bool returns a Predicate that simply returns v when evaluated.
func Bool(v bool) Predicate {
	return func(_ pipeline.Context, step pipeline.Step) bool {
		return v
	}
}

// Not returns a Predicate that evaluates, but then negates the given Predicate.
func Not(predicate Predicate) Predicate {
	return func(ctx pipeline.Context, step pipeline.Step) bool {
		return !predicate(ctx, step)
	}
}

// And returns a Predicate that does logical AND of the given predicates.
// p2 is not evaluated if p1 evaluates already to false.
func And(p1, p2 Predicate) Predicate {
	return func(ctx pipeline.Context, step pipeline.Step) bool {
		return p1(ctx, step) && p2(ctx, step)
	}
}

// Or returns a Predicate that does logical OR of the given predicates.
// p2 is not evaluated if p1 evaluates already to true.
func Or(p1, p2 Predicate) Predicate {
	return func(ctx pipeline.Context, step pipeline.Step) bool {
		return p1(ctx, step) || p2(ctx, step)
	}
}
