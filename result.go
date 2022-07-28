package pipeline

// Result is the object that is returned after each step and after running a pipeline.
type Result interface {
	// Err contains the step's returned error, nil otherwise.
	Err() error
	// Name retrieves the name of the (last) step that has been executed.
	Name() string
	// IsCanceled returns true if the pipeline's context has been canceled.
	IsCanceled() bool
	// IsSuccessful returns true if the contained error is nil.
	IsSuccessful() bool
	// IsFailed returns true if the contained error is non-nil.
	IsFailed() bool
}

type resultImpl struct {
	err      error
	name     string
	canceled bool
}

// newEmptyResult returns a Result with just the name.
func newEmptyResult(stepName string) Result {
	return resultImpl{
		name: stepName,
	}
}

// newResult is the constructor for all properties.
func newResult(stepName string, err error, canceled bool) Result {
	return resultImpl{
		name:     stepName,
		err:      err,
		canceled: canceled,
	}
}

// newResultWithError constructs a Result with given name and error.
func newResultWithError(stepName string, err error) Result {
	return resultImpl{
		name: stepName,
		err:  err,
	}
}

func (r resultImpl) Name() string {
	return r.name
}

func (r resultImpl) Err() error {
	return r.err
}

func (r resultImpl) IsSuccessful() bool {
	return r.err == nil
}

func (r resultImpl) IsFailed() bool {
	return r.err != nil
}

func (r resultImpl) IsCanceled() bool {
	return r.canceled
}
