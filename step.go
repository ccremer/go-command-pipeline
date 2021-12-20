package pipeline

// NewStep returns a new Step with given name and action.
func NewStep(name string, action ActionFunc) Step {
	return Step{
		Name: name,
		F:    action,
	}
}

// NewStepFromFunc returns a new Step with given name using a function that expects an error.
func NewStepFromFunc(name string, fn func(ctx Context) error) Step {
	return NewStep(name, func(ctx Context) Result {
		err := fn(ctx)
		return Result{Err: err, Name: name}
	})
}

// WithResultHandler sets the ResultHandler of this specific step and returns the step itself.
func (s Step) WithResultHandler(handler ResultHandler) Step {
	s.H = handler
	return s
}

// WithErrorHandler wraps given errorHandler and sets the ResultHandler of this specific step and returns the step itself.
// The difference to WithResultHandler is that errorHandler only gets called if Result.Err is non-nil.
func (s Step) WithErrorHandler(errorHandler func(ctx Context, err error) error) Step {
	s.H = func(ctx Context, result Result) error {
		if result.IsFailed() {
			return errorHandler(ctx, result.Err)
		}
		return nil
	}
	return s
}
