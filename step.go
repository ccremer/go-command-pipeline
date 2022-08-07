package pipeline

import (
	"context"
	"fmt"
)

// Step is an intermediary action and part of a Pipeline.
type Step[T context.Context] struct {
	// Name describes the step's human-readable name.
	// It has no other uses other than easily identifying a step for debugging or logging.
	Name string
	// Action is the ActionFunc assigned to a pipeline Step.
	// This is required.
	Action ActionFunc[T]
	// Handler is the ErrorHandler assigned to a pipeline Step.
	// This is optional, and it will be called if it is set after Action completed.
	// Use cases could be logging, updating a GUI or handle errors while continuing the pipeline.
	// The function may return nil even if the given error is non-nil, in which case the pipeline will continue.
	// This function is called before the next step's Action is invoked.
	Handler ErrorHandler[T]
	// Condition determines if the Step's Action is actually going to be executed in the pipeline.
	// When nil, the Action is executed.
	Condition Predicate[T]
}

// NewStep returns a new Step with given name and action.
func NewStep[T context.Context](name string, action ActionFunc[T]) Step[T] {
	if action == nil {
		panic(fmt.Errorf("action cannot be empty for step %q", name))
	}
	return Step[T]{
		Name:   name,
		Action: action,
	}
}

// NewStepIf is syntactic sugar for NewStep with Step.When.
func NewStepIf[T context.Context](predicate Predicate[T], name string, actionFunc ActionFunc[T]) Step[T] {
	return NewStep[T](name, actionFunc).When(predicate)
}

// WithErrorHandler sets the ErrorHandler of this specific step and returns the step itself.
func (s Step[T]) WithErrorHandler(errorHandler ErrorHandler[T]) Step[T] {
	s.Handler = errorHandler
	return s
}

// When sets Step.Condition.
// When the given predicate returns false, the step is skipped without error.
func (s Step[T]) When(predicate Predicate[T]) Step[T] {
	s.Condition = predicate
	return s
}
