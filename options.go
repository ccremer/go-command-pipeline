package pipeline

// Options configures the given Pipeline with a behaviour-altering settings.
type Options struct {
	// DisableErrorWrapping disables the wrapping of errors that are emitted from pipeline steps.
	// This effectively causes error to be exactly the error as returned from a step.
	// The step's name is omitted from the error message.
	DisableErrorWrapping bool
}

// WithOptions configures the Pipeline with settings.
// The Options are applied immediately.
// Options are applied to nested pipelines provided they are set before building the nested pipeline.
// Nested pipelines can be configured with their own Options.
func (p *Pipeline[T]) WithOptions(options Options) *Pipeline[T] {
	p.options = options
	return p
}
