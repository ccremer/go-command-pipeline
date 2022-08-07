package pipeline

import (
	"context"
	"fmt"
)

// Pipeline holds and runs intermediate actions, called "steps".
type Pipeline[T context.Context] struct {
	steps       []Step[T]
	beforeHooks []Listener[T]
	finalizer   ErrorHandler[T]
	options     options
}

// Listener is a simple func that listens to Pipeline events.
type Listener[T context.Context] func(step Step[T])

// ActionFunc is the func that contains your business logic.
type ActionFunc[T context.Context] func(ctx T) error

// ErrorHandler is a func that gets called when a step's ActionFunc has finished with an error.
type ErrorHandler[T context.Context] func(ctx T, err error) error

// NewPipeline returns a new Pipeline instance.
func NewPipeline[T context.Context]() *Pipeline[T] {
	return &Pipeline[T]{}
}

// WithBeforeHooks takes a list of listeners.
// Each Listener is called once in the given order just before the ActionFunc is invoked.
// The listeners should return as fast as possible, as they are not intended to do actual business logic.
func (p *Pipeline[T]) WithBeforeHooks(listeners ...Listener[T]) *Pipeline[T] {
	p.beforeHooks = listeners
	return p
}

// AddStep appends the given step to the Pipeline at the end and returns itself.
func (p *Pipeline[T]) AddStep(step Step[T]) *Pipeline[T] {
	p.steps = append(p.steps, step)
	return p
}

// AddStepFromFunc appends the given function to the Pipeline at the end and returns itself.
func (p *Pipeline[T]) AddStepFromFunc(name string, fn ActionFunc[T]) *Pipeline[T] {
	return p.AddStep(NewStep[T](name, fn))
}

// WithSteps appends the given array of steps to the Pipeline at the end and returns itself.
func (p *Pipeline[T]) WithSteps(steps ...Step[T]) *Pipeline[T] {
	p.steps = steps
	return p
}

// WithNestedSteps is similar to AsNestedStep, but it accepts the steps given directly as parameters.
func (p *Pipeline[T]) WithNestedSteps(name string, steps ...Step[T]) Step[T] {
	return NewStep[T](name, func(ctx T) error {
		nested := &Pipeline[T]{beforeHooks: p.beforeHooks, steps: steps, options: p.options}
		return nested.RunWithContext(ctx)
	})
}

// AsNestedStep converts the Pipeline instance into a Step that can be used in other pipelines.
// The properties are passed to the nested pipeline.
func (p *Pipeline[T]) AsNestedStep(name string) Step[T] {
	return NewStep[T](name, func(ctx T) error {
		nested := &Pipeline[T]{beforeHooks: p.beforeHooks, steps: p.steps, options: p.options}
		return nested.RunWithContext(ctx)
	})
}

// WithFinalizer returns itself while setting the finalizer for the pipeline.
// The finalizer is a handler that gets called after the last step is in the pipeline is completed.
// If a pipeline aborts early or gets canceled then it is also called.
func (p *Pipeline[T]) WithFinalizer(handler ErrorHandler[T]) *Pipeline[T] {
	p.finalizer = handler
	return p
}

// NewStep is syntactic sugar for NewStep but with T already set.
func (p *Pipeline[T]) NewStep(name string, action ActionFunc[T]) Step[T] {
	return NewStep[T](name, action)
}

// If is syntactic sugar for If combined with NewStep.
func (p *Pipeline[T]) If(predicate Predicate[T], name string, action ActionFunc[T]) Step[T] {
	return If[T](predicate, p.NewStep(name, action))
}

// RunWithContext executes the Pipeline.
// Steps are executed sequentially as they were added to the Pipeline.
// Upon cancellation of the context, the pipeline does not terminate a currently running step, instead it skips the remaining steps in the execution order.
// The context is passed to each Step.Action and each Step may need to listen to the context cancellation event to truly cancel a long-running step.
// If the pipeline gets canceled, the context's error is returned.
//
// All non-nil errors, except the error returned from the pipeline's finalizer, are wrapped in Result.
// This can be used to retrieve the metadata of the step that returned the error with errors.As:
//  err := p.RunWithContext(ctx)
//  var result pipeline.Result
//  if errors.As(err, &result) {
//    fmt.Println(result.Name())
//  }
func (p *Pipeline[T]) RunWithContext(ctx T) error {
	result := p.doRun(ctx)
	if p.finalizer != nil {
		err := p.finalizer(ctx, result)
		return err
	}
	return result
}

func (p *Pipeline[T]) doRun(ctx T) Result {
	for _, step := range p.steps {
		select {
		case <-ctx.Done():
			result := p.fail(ctx.Err(), step)
			return result
		default:
			for _, hooks := range p.beforeHooks {
				hooks(step)
			}

			err := step.Action(ctx)
			if step.Handler != nil {
				err = step.Handler(ctx, err)
			}
			if err != nil {
				return p.fail(err, step)
			}
		}
	}
	return nil
}

func (p *Pipeline[T]) fail(err error, step Step[T]) Result {
	var resultErr error
	if p.options.disableErrorWrapping {
		resultErr = err
	} else {
		resultErr = fmt.Errorf("step '%s' failed: %w", step.Name, err)
	}
	return newResult(step.Name, resultErr)
}
