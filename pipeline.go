package pipeline

import (
	"fmt"
)

type (
	Pipeline struct {
		log          Logger
		steps        []Step
		abortHandler Handler
	}
	Result struct {
		Abort bool
		Err   error
	}
	Step struct {
		Name string
		F    ActionFunc
		P    Predicate
	}
	Logger interface {
		Log(message, name string)
	}
	ActionFunc func() Result
	Predicate  func(step Step) bool
	Handler    func(result Result)

	nullLogger struct{}
)

func (n nullLogger) Log(_, _ string) {}

func NewPipeline() *Pipeline {
	return NewPipelineWithLogger(nullLogger{})
}

func NewPipelineWithLogger(logger Logger) *Pipeline {
	return &Pipeline{log: logger}
}

func (p *Pipeline) AddStep(step Step) *Pipeline {
	p.steps = append(p.steps, step)
	return p
}

func (p *Pipeline) WithSteps(steps ...Step) *Pipeline {
	p.steps = steps
	return p
}

func (p *Pipeline) WithAbortHandler(handler Handler) *Pipeline {
	p.abortHandler = handler
	return p
}

func (p *Pipeline) AsNestedStep(name string, predicate Predicate) Step {
	return NewStepWithPredicate(name, func() Result {
		nested := NewPipeline()
		nested.steps = p.steps
		nested.abortHandler = p.abortHandler
		return nested.runPipeline()
	}, predicate)
}

func (r Result) IsSuccessful() bool {
	return r.Err == nil
}

func (p *Pipeline) Run() Result {
	result := p.runPipeline()
	return result
}

func (p *Pipeline) runPipeline() Result {
	for _, step := range p.steps {
		if step.P != nil && !step.P(step) {
			p.log.Log("ignoring step", step.Name)
			continue
		}

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

func NewStep(name string, action ActionFunc) Step {
	return Step{
		Name: name,
		F:    action,
	}
}

func NewStepWithPredicate(name string, action ActionFunc, predicate Predicate) Step {
	return Step{
		Name: name,
		F:    action,
		P:    predicate,
	}
}

func NewIfElseStep(name string, predicate Predicate, trueAction, falseAction ActionFunc) Step {
	step := Step{
		Name: name,
	}
	fn := func() Result {
		if predicate(step) {
			return trueAction()
		} else {
			return falseAction()
		}
	}
	step.F = fn
	return step
}

func Abort() ActionFunc {
	return func() Result {
		return Result{Abort: true}
	}
}
