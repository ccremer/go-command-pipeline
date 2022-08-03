package pipeline

import "context"

// NewStep returns a new Step with given name and action.
func NewStep[T context.Context](name string, action ActionFunc[T]) Step[T] {
	return Step[T]{
		Name:   name,
		Action: action,
	}
}

// WithErrorHandler sets the ErrorHandler of this specific step and returns the step itself.
func (s Step[T]) WithErrorHandler(errorHandler ErrorHandler[T]) Step[T] {
	s.Handler = errorHandler
	return s
}
