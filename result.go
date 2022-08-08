package pipeline

// Result is the object that is returned after each step and after running a pipeline.
type Result interface {
	error
	// Name retrieves the name of the (last) step that has been executed.
	Name() string
}

type resultImpl struct {
	err  error
	name string
}

// newResult is the constructor for all properties.
func newResult(stepName string, err error) Result {
	if err == nil {
		panic("error cannot be nil: " + stepName)
	}
	return resultImpl{
		name: stepName,
		err:  err,
	}
}

func (r resultImpl) Error() string {
	return r.err.Error()
}

func (r resultImpl) Name() string {
	return r.name
}

// Unwrap implements xerrors.Wrapper.
func (r resultImpl) Unwrap() error {
	return r.err
}
