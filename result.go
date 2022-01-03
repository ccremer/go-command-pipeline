package pipeline

import "errors"

// ErrAbort indicates that the pipeline should be terminated immediately without returning an error.
var ErrAbort = errors.New("abort")

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
