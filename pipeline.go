package pipeline

import (
	"context"
	"errors"
	"fmt"
)

// Pipeline holds and runs intermediate actions, called "steps".
type Pipeline struct {
	steps       []Step
	beforeHooks []Listener
	finalizer   ResultHandler
	options     options
}

// Result is the object that is returned after each step and after running a pipeline.
type Result struct {
	// Err contains the step's returned error, nil otherwise.
	// In an aborted pipeline with ErrAbort it will still be nil.
	Err error
	// Name is an optional identifier for a result.
	// ActionFunc may set this property before returning to help a ResultHandler with further processing.
	Name string

	aborted bool
}

// Step is an intermediary action and part of a Pipeline.
type Step struct {
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

// Listener is a simple func that listens to Pipeline events.
type Listener func(step Step)

// ActionFunc is the func that contains your business logic.
type ActionFunc func(ctx context.Context) Result

// ResultHandler is a func that gets called when a step's ActionFunc has finished with any Result.
type ResultHandler func(ctx context.Context, result Result) error

// NewPipeline returns a new Pipeline instance.
func NewPipeline() *Pipeline {
	return &Pipeline{}
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

// WithSteps appends the given array of steps to the Pipeline at the end and returns itself.
func (p *Pipeline) WithSteps(steps ...Step) *Pipeline {
	p.steps = steps
	return p
}

// WithNestedSteps is similar to AsNestedStep, but it accepts the steps given directly as parameters.
func (p *Pipeline) WithNestedSteps(name string, steps ...Step) Step {
	return NewStep(name, func(_ context.Context) Result {
		nested := &Pipeline{beforeHooks: p.beforeHooks, steps: steps, options: p.options}
		return nested.Run()
	})
}

// AsNestedStep converts the Pipeline instance into a Step that can be used in other pipelines.
// The properties are passed to the nested pipeline.
func (p *Pipeline) AsNestedStep(name string) Step {
	return NewStep(name, func(_ context.Context) Result {
		nested := &Pipeline{beforeHooks: p.beforeHooks, steps: p.steps, options: p.options}
		return nested.Run()
	})
}

// WithFinalizer returns itself while setting the finalizer for the pipeline.
// The finalizer is a handler that gets called after the last step is in the pipeline is completed.
// If a pipeline aborts early then it is also called.
func (p *Pipeline) WithFinalizer(handler ResultHandler) *Pipeline {
	p.finalizer = handler
	return p
}

// Run executes the pipeline with context.Background and returns the result.
// Steps are executed sequentially as they were added to the Pipeline.
// If a Step returns a Result with a non-nil error, the Pipeline is aborted and its Result contains the affected step's error.
// However, if Result.Err is wrapped in ErrAbort, then the pipeline is aborted, but the final Result.Err will be nil.
func (p *Pipeline) Run() Result {
	return p.RunWithContext(context.Background())
}

// RunWithContext is like Run but with a given context.Context.
func (p *Pipeline) RunWithContext(ctx context.Context) Result {
	result := p.doRun(ctx)
	if p.finalizer != nil {
		result.Err = p.finalizer(ctx, result)
	}
	return result
}

func (p *Pipeline) doRun(ctx context.Context) Result {
	for _, step := range p.steps {
		for _, hooks := range p.beforeHooks {
			hooks(step)
		}

		result := step.F(ctx)
		var err error
		if step.H != nil {
			err = step.H(ctx, result)
		} else {
			err = result.Err
		}
		if err != nil {
			if errors.Is(err, ErrAbort) {
				// Abort pipeline without error
				return Result{aborted: true}
			}
			if p.options.disableErrorWrapping {
				return Result{Err: err}
			}
			return Result{Err: fmt.Errorf("step '%s' failed: %w", step.Name, err)}
		}
	}
	return Result{}
}
