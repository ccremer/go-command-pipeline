package pipeline

import (
	"fmt"
)

type (
	// Pipeline holds and runs intermediate actions, called "steps".
	Pipeline struct {
		log          Logger
		steps        []Step
		abortHandler Handler
	}
	// Result is the object that is returned after each step and after running a pipeline.
	Result struct {
		Abort bool
		// Err contains the step's returned error, nil otherwise.
		Err error
	}
	// Step is an intermediary action and part of a Pipeline.
	Step struct {
		// Name describes the step's human-readable name.
		// It has no other uses other than easily identifying a step for debugging or logging.
		Name string
		// F is the ActionFunc assigned to a pipeline Step.
		F ActionFunc
	}
	// Logger is a simple interface that enables logging certain actions with your favourite logging framework.
	Logger interface {
		// Log is expected to write the given message to a logging framework or similar.
		// The name contains the step's name.
		Log(message, name string)
	}
	// ActionFunc is the func that contains your business logic.
	ActionFunc func() Result
	Handler    func(result Result)

	nullLogger struct{}
)

func (n nullLogger) Log(_, _ string) {}

// NewPipeline returns a new Pipeline instance that doesn't log anything.
func NewPipeline() *Pipeline {
	return NewPipelineWithLogger(nullLogger{})
}

// NewPipelineWithLogger returns a new Pipeline instance with the given logger that shouldn't be nil.
func NewPipelineWithLogger(logger Logger) *Pipeline {
	return &Pipeline{log: logger}
}

// AddStep appends the given step to the Pipeline at the end and returns itself.
func (p *Pipeline) AddStep(step Step) *Pipeline {
	p.steps = append(p.steps, step)
	return p
}

// WithSteps appends the given arrway of steps to the Pipeline at the end and returns itself.
func (p *Pipeline) WithSteps(steps ...Step) *Pipeline {
	p.steps = steps
	return p
}

func (p *Pipeline) WithAbortHandler(handler Handler) *Pipeline {
	p.abortHandler = handler
	return p
}

// AsNestedStep converts the Pipeline instance into a Step that can be used in other pipelines.
// The logger and abort handler are passed to the nested pipeline.
func (p *Pipeline) AsNestedStep(name string) Step {
	return NewStep(name, func() Result {
		nested := &Pipeline{log: p.log, abortHandler: p.abortHandler, steps: p.steps}
		return nested.runPipeline()
	})
}

// IsSuccessful returns true if the contained error is nil.
func (r Result) IsSuccessful() bool {
	return r.Err == nil
}

// IsFailed returns true if the contained error is non-nil.
func (r Result) IsFailed() bool {
	return r.Err != nil
}

// Run executes the pipeline and returns the result.
// Steps are executed sequentially as they were added to the Pipeline.
// If a Step returns a Result with a non-nil error, the Pipeline is aborted its Result contains the affected step's error.
func (p *Pipeline) Run() Result {
	result := p.runPipeline()
	return result
}

func (p *Pipeline) runPipeline() Result {
	for _, step := range p.steps {
		p.log.Log("executing step", step.Name)

		if r := step.F(); r.Abort || r.Err != nil {
			if p.abortHandler != nil {
				p.abortHandler(r)
			}
			if r.Err == nil {
				p.log.Log("aborting pipeline by step", step.Name)
				return Result{Abort: r.Abort}
			}

			return Result{Err: fmt.Errorf("step '%s' failed: %w", step.Name, r.Err), Abort: r.Abort}
		}
	}
	return Result{}
}

// NewStep returns a new Step with given name and action.
func NewStep(name string, action ActionFunc) Step {
	return Step{
		Name: name,
		F:    action,
	}
}

func Abort() ActionFunc {
	return func() Result {
		return Result{Abort: true}
	}
}
