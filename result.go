package pipeline

import "errors"

// ErrAbort indicates that the pipeline should be terminated immediately without being marked as failed (returning an error).
var ErrAbort = errors.New("abort")

// Result is the object that is returned after each step and after running a pipeline.
type Result struct {
	// Err contains the step's returned error, nil otherwise.
	// In an aborted pipeline with ErrAbort it will still be nil.
	Err error

	name     string
	aborted  bool
	canceled bool
}

// Name retrieves the name of the (last) step that has been executed.
func (r Result) Name() string {
	return r.name
}

// IsSuccessful returns true if the contained error is nil.
// Aborted pipelines (with ErrAbort) are still reported as success.
// To query if a pipeline is aborted early, use IsAborted.
func (r Result) IsSuccessful() bool {
	return r.Err == nil
}

// IsFailed returns true if the contained error is non-nil.
func (r Result) IsFailed() bool {
	return r.Err != nil
}

// IsAborted returns true if the pipeline didn't stop with an error, but just aborted early with ErrAbort.
func (r Result) IsAborted() bool {
	return r.aborted
}

// IsCanceled returns true if the pipeline's context has been canceled.
func (r Result) IsCanceled() bool {
	return r.canceled
}

// Canceled sets Result.IsCanceled to true.
func Canceled(result Result) Result {
	result.canceled = true
	return result
}
