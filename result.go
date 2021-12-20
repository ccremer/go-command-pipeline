package pipeline

import "errors"

// ErrAbort indicates that the pipeline should be terminated immediately without returning an error.
var ErrAbort = errors.New("abort")

// IsSuccessful returns true if the contained error is nil.
func (r Result) IsSuccessful() bool {
	return r.Err == nil
}

// IsFailed returns true if the contained error is non-nil.
func (r Result) IsFailed() bool {
	return r.Err != nil
}
