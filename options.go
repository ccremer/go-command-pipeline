package pipeline

import "context"

type options struct {
	disableErrorWrapping bool
}

// Option configures the given Pipeline with a behaviour-altering setting.
type Option[T context.Context] func(pipeline *Pipeline[T])

// WithOptions configures the Pipeline with settings.
// The options are applied immediately.
// Options are applied to nested pipelines provided they are set before building the nested pipeline.
// Nested pipelines can be configured with their own options.
func (p *Pipeline[T]) WithOptions(options ...Option[T]) *Pipeline[T] {
	for _, option := range options {
		option(p)
	}
	return p
}

// DisableErrorWrapping disables the wrapping of errors that are emitted from pipeline steps.
// This effectively causes Result.Err to be exactly the error as returned from a step.
var DisableErrorWrapping Option[context.Context] = func(pipeline *Pipeline[context.Context]) {
	pipeline.options.disableErrorWrapping = true
}
