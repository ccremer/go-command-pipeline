package pipeline

import (
	"fmt"
)

type (
	// Pipeline holds and runs intermediate actions, called "steps".
	Pipeline struct {
		log   Logger
		steps []Step
	}
	// Result is the object that is returned after each step and after running a pipeline.
	Result struct {
		// Err contains the step's returned error, nil otherwise.
		Err error
		// Name is an optional identifier for a result.
		// ActionFunc may set this property before returning to help an ResultHandler with further processing.
		Name string
	}
	// Step is an intermediary action and part of a Pipeline.
	Step struct {
		// Name describes the step's human-readable name.
		// It has no other uses other than easily identifying a step for debugging or logging.
		Name string
		// F is the ActionFunc assigned to a pipeline Step.
		// This is required.
		F ActionFunc
		// H is the ResultHandler assigned to a pipeline Step.
		// This is optional and it will be called in any case if it is set after F completed.
		// Use cases could be logging, updating a GUI or handle errors while continuing the pipeline.
		// The function may return nil even if the Result contains an error, in which case the pipeline will continue.
		// This function is called before the next step's F is invoked.
		H ResultHandler
	}
	// Logger is a simple interface that enables logging certain actions with your favourite logging framework.
	Logger interface {
		// Log is expected to write the given message to a logging framework or similar.
		// The name contains the step's name.
		Log(message, name string)
	}
	// ActionFunc is the func that contains your business logic.
	ActionFunc func() Result
	// ResultHandler is a func that gets called when a step's ActionFunc has finished with any Result.
	ResultHandler func(result Result) error

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

// WithNestedSteps is similar to AsNestedStep but it accepts the steps given directly as parameters.
func (p *Pipeline) WithNestedSteps(name string, steps ...Step) Step {
	return NewStep(name, func() Result {
		nested := &Pipeline{log: p.log, steps: steps}
		return nested.Run()
	})
}

// AsNestedStep converts the Pipeline instance into a Step that can be used in other pipelines.
// The logger and abort handler are passed to the nested pipeline.
func (p *Pipeline) AsNestedStep(name string) Step {
	return NewStep(name, func() Result {
		nested := &Pipeline{log: p.log, steps: p.steps}
		return nested.Run()
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
	for _, step := range p.steps {
		p.log.Log("executing step", step.Name)

		r := step.F()
		if step.H != nil {
			if handlerErr := step.H(r); handlerErr != nil {
				return Result{Err: fmt.Errorf("step '%s' failed: %w", step.Name, handlerErr)}
			}
		} else {
			if r.Err != nil {
				return Result{Err: fmt.Errorf("step '%s' failed: %w", step.Name, r.Err)}
			}
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

// WithResultHandler sets the ResultHandler of this specific step and returns the step itself.
func (s Step) WithResultHandler(handler ResultHandler) Step {
	s.H = handler
	return s
}
