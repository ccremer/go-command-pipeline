package pipeline

// Option configures the given Pipeline with a behaviour-altering setting.
type Option func(pipeline *Pipeline)

// WithOptions configures the Pipeline with settings.
// The options are applied immediately.
// Options are applied to nested pipelines provided they are set before building the nested pipeline.
// Nested pipelines can be configured with their own options.
func (p *Pipeline) WithOptions(options ...Option) *Pipeline {
	for _, option := range options {
		option(p)
	}
	return p
}

// DisableErrorWrapping disables the wrapping of errors that are emitted from pipeline steps.
// This effectively causes Result.Err to be exactly the error as returned from a step.
var DisableErrorWrapping Option = func(pipeline *Pipeline) {
	pipeline.disableErrorWrapping = true
}
