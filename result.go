package pipeline

import "errors"

// ErrAbort indicates that the pipeline should be terminated immediately without being marked as failed (returning an error).
var ErrAbort = errors.New("abort")

// Result is the object that is returned after each step and after running a pipeline.
type Result struct {
	err      error
	name     string
	aborted  bool
	canceled bool
}

// NewEmptyResult returns a Result with just the name.
func NewEmptyResult(stepName string) Result {
	return Result{
		name: stepName,
	}
}

// NewResult is the constructor for all properties.
func NewResult(stepName string, err error, aborted, canceled bool) Result {
	return Result{
		name:     stepName,
		err:      err,
		aborted:  aborted,
		canceled: canceled,
	}
}

// NewResultWithError constructs a Result with given name and error.
func NewResultWithError(stepName string, err error) Result {
	return Result{
		name: stepName,
		err:  err,
	}
}

// Name retrieves the name of the (last) step that has been executed.
func (r Result) Name() string {
	return r.name
}

// Err contains the step's returned error, nil otherwise.
// In an aborted pipeline with ErrAbort it will still be nil.
func (r Result) Err() error {
	return r.err
}

// IsSuccessful returns true if the contained error is nil and not aborted.
func (r Result) IsSuccessful() bool {
	return r.err == nil && r.aborted == false
}

// IsCompleted returns true if the contained error is nil.
// Aborted pipelines (with ErrAbort) are still reported as completed.
// To query if a pipeline is aborted early, use IsAborted.
func (r Result) IsCompleted() bool {
	return r.err == nil
}

// IsFailed returns true if the contained error is non-nil.
func (r Result) IsFailed() bool {
	return r.err != nil
}

// IsAborted returns true if the pipeline didn't stop with an error, but just aborted early with ErrAbort.
func (r Result) IsAborted() bool {
	return r.aborted
}

// IsCanceled returns true if the pipeline's context has been canceled.
func (r Result) IsCanceled() bool {
	return r.canceled
}
