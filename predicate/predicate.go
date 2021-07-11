package predicate

import (
	pipeline "github.com/ccremer/go-command-pipeline"
)

type (
	// Predicate is a function that expects 'true' if a pipeline.ActionFunc should run.
	// Is is evaluated lazily resp. only when needed.
	Predicate func(step pipeline.Step) bool
)

// ToStep wraps the given action func in its own step.
// When the step's function is called, the given Predicate will evaluate whether the action should actually run.
// It returns the action's pipeline.Result, otherwise an empty (successful) pipeline.Result struct.
func ToStep(name string, action pipeline.ActionFunc, predicate Predicate) pipeline.Step {
	step := pipeline.Step{Name: name}
	step.F = func() pipeline.Result {
		if predicate(step) {
			return action()
		}
		return pipeline.Result{}
	}
	return step
}

// ToNestedStep wraps the given pipeline in its own step.
// When the step's function is called, the given Predicate will evaluate whether the nested pipeline.Pipeline should actually run.
// It returns the pipeline's pipeline.Result, otherwise an empty (successful) pipeline.Result struct.
func ToNestedStep(name string, p *pipeline.Pipeline, predicate Predicate) pipeline.Step {
	step := pipeline.Step{Name: name}
	step.F = func() pipeline.Result {
		if predicate(step) {
			return p.Run()
		}
		return pipeline.Result{}
	}
	return step
}

// Bool returns a Predicate that simply returns v when evaluated.
func Bool(v bool) Predicate {
	return func(step pipeline.Step) bool {
		return v
	}
}

// Not returns a Predicate that evaluates, but then negates the given Predicate.
func Not(predicate Predicate) Predicate {
	return func(step pipeline.Step) bool {
		return !predicate(step)
	}
}

// And returns a Predicate that does logical AND of the given predicates.
// p2 is not evaluated if p1 evaluates already to false.
func And(p1, p2 Predicate) Predicate {
	return func(step pipeline.Step) bool {
		return p1(step) && p2(step)
	}
}

// Or returns a Predicate that does logical OR of the given predicates.
// p2 is not evaluated if p1 evaluates already to true.
func Or(p1, p2 Predicate) Predicate {
	return func(step pipeline.Step) bool {
		return p1(step) || p2(step)
	}
}
