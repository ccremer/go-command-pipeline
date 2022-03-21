package pipeline

import "errors"

// ErrAbort indicates that the pipeline should be terminated immediately without being marked as failed (returning an error).
var ErrAbort = errors.New("abort")

// Result is the object that is returned after each step and after running a pipeline.
type Result interface {
	// Err contains the step's returned error, nil otherwise.
	// In an aborted pipeline with ErrAbort it will still be nil.
	Err() error
	// Name retrieves the name of the (last) step that has been executed.
	Name() string
	// IsAborted returns true if the pipeline didn't stop with an error, but just aborted early with ErrAbort.
	IsAborted() bool
	// IsCanceled returns true if the pipeline's context has been canceled.
	IsCanceled() bool
	// IsSuccessful returns true if the contained error is nil and not aborted.
	IsSuccessful() bool
	// IsCompleted returns true if the contained error is nil.
	// Aborted pipelines (with ErrAbort) are still reported as completed.
	// To query if a pipeline is aborted early, use IsAborted.
	IsCompleted() bool
	// IsFailed returns true if the contained error is non-nil.
	IsFailed() bool
}

type resultImpl struct {
	err      error
	name     string
	aborted  bool
	canceled bool
}

// newEmptyResult returns a Result with just the name.
func newEmptyResult(stepName string) Result {
	return resultImpl{
		name: stepName,
	}
}

// newResult is the constructor for all properties.
func newResult(stepName string, err error, aborted, canceled bool) Result {
	return resultImpl{
		name:     stepName,
		err:      err,
		aborted:  aborted,
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
	return r.err == nil && r.aborted == false
}

func (r resultImpl) IsCompleted() bool {
	return r.err == nil
}

func (r resultImpl) IsFailed() bool {
	return r.err != nil
}

func (r resultImpl) IsAborted() bool {
	return r.aborted
}

func (r resultImpl) IsCanceled() bool {
	return r.canceled
}
