package pipeline

import (
	"errors"
	"fmt"
)

type (
	// Pipeline holds and runs intermediate actions, called "steps".
	Pipeline struct {
		steps                []Step
		context              Context
		beforeHooks          []Listener
		finalizer            ResultHandler
		disableErrorWrapping bool
	}
	// Result is the object that is returned after each step and after running a pipeline.
	Result struct {
		// Err contains the step's returned error, nil otherwise.
		Err error
		// Name is an optional identifier for a result.
		// ActionFunc may set this property before returning to help a ResultHandler with further processing.
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
		// This is optional, and it will be called in any case if it is set after F completed.
		// Use cases could be logging, updating a GUI or handle errors while continuing the pipeline.
		// The function may return nil even if the Result contains an error, in which case the pipeline will continue.
		// This function is called before the next step's F is invoked.
		H ResultHandler
	}
	// Context contains arbitrary data relevant for the pipeline execution.
	Context interface{}
	// Listener is a simple func that listens to Pipeline events.
	Listener func(step Step)
	// ActionFunc is the func that contains your business logic.
	// The context is a user-defined arbitrary data of type interface{} that gets provided in every Step, but may be nil if not set.
	ActionFunc func(ctx Context) Result
	// ResultHandler is a func that gets called when a step's ActionFunc has finished with any Result.
	// Context may be nil.
	ResultHandler func(ctx Context, result Result) error
)

// NewPipeline returns a new quiet Pipeline instance with KeyValueContext.
func NewPipeline() *Pipeline {
	return &Pipeline{}
}

// NewPipelineWithContext returns a new Pipeline instance with the given context.
func NewPipelineWithContext(ctx Context) *Pipeline {
	return &Pipeline{context: ctx}
}

// WithBeforeHooks takes a list of listeners.
// Each Listener.Accept is called once in the given order just before the ActionFunc is invoked.
// The listeners should return as fast as possible, as they are not intended to do actual business logic.
func (p *Pipeline) WithBeforeHooks(listeners []Listener) *Pipeline {
	p.beforeHooks = listeners
	return p
}

// AddBeforeHook adds the given listener to the list of hooks.
// See WithBeforeHooks.
func (p *Pipeline) AddBeforeHook(listener Listener) *Pipeline {
	return p.WithBeforeHooks(append(p.beforeHooks, listener))
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

// WithNestedSteps is similar to AsNestedStep, but it accepts the steps given directly as parameters.
func (p *Pipeline) WithNestedSteps(name string, steps ...Step) Step {
	return NewStep(name, func(_ Context) Result {
		nested := &Pipeline{beforeHooks: p.beforeHooks, steps: steps, context: p.context, disableErrorWrapping: p.disableErrorWrapping}
		return nested.Run()
	})
}

// AsNestedStep converts the Pipeline instance into a Step that can be used in other pipelines.
// The properties are passed to the nested pipeline.
func (p *Pipeline) AsNestedStep(name string) Step {
	return NewStep(name, func(_ Context) Result {
		nested := &Pipeline{beforeHooks: p.beforeHooks, steps: p.steps, context: p.context}
		return nested.Run()
	})
}

// WithContext returns itself while setting the context for the pipeline steps.
func (p *Pipeline) WithContext(ctx Context) *Pipeline {
	p.context = ctx
	return p
}

// WithFinalizer returns itself while setting the finalizer for the pipeline.
// The finalizer is a handler that gets called after the last step is in the pipeline is completed.
// If a pipeline aborts early then it is also called.
func (p *Pipeline) WithFinalizer(handler ResultHandler) *Pipeline {
	p.finalizer = handler
	return p
}

// Run executes the pipeline and returns the result.
// Steps are executed sequentially as they were added to the Pipeline.
// If a Step returns a Result with a non-nil error, the Pipeline is aborted and its Result contains the affected step's error.
// However, if Result.Err is wrapped in ErrAbort, then the pipeline is aborted, but the final Result.Err will be nil.
func (p *Pipeline) Run() Result {
	result := p.doRun()
	if p.finalizer != nil {
		result.Err = p.finalizer(p.context, result)
	}
	return result
}

func (p *Pipeline) doRun() Result {
	for _, step := range p.steps {
		for _, hooks := range p.beforeHooks {
			hooks(step)
		}

		result := step.F(p.context)
		var err error
		if step.H != nil {
			err = step.H(p.context, result)
		} else {
			err = result.Err
		}
		if err != nil {
			if errors.Is(err, ErrAbort) {
				// Abort pipeline without error
				return Result{}
			}
			if p.disableErrorWrapping {
				return Result{Err: err}
			}
			return Result{Err: fmt.Errorf("step '%s' failed: %w", step.Name, err)}
		}
	}
	return Result{}
}
