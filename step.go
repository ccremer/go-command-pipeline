package pipeline

import "context"

// NewStep returns a new Step with given name and action.
func NewStep[T context.Context](name string, action ActionFunc[T]) Step[T] {
	return Step[T]{
		Name:   name,
		Action: action,
	}
}

// WithErrorHandler wraps given errorHandler and sets the ResultHandler of this specific step and returns the step itself.
// The difference to WithResultHandler is that errorHandler only gets called if Result.Err is non-nil.
func (s Step[T]) WithErrorHandler(errorHandler func(ctx T, err error) error) Step[T] {
	s.Handler = func(ctx T, result Result) error {
		if result.IsFailed() {
			return errorHandler(ctx, result.Err())
		}
		return nil
	}
	return s
}
