package pipelinetest

import (
	"context"

	pipeline "github.com/ccremer/go-command-pipeline"
)

// NewNoResolver returns a pipeline.DependencyResolver that doesn't actually resolve anything.
// This can be used for testing.
func NewNoResolver[T context.Context]() pipeline.DependencyResolver[T] {
	return &NoResolver[T]{}
}

// NoResolver is a pipeline.DependencyResolver that doesn't actually resolve anything.
// This can be used for testing.
type NoResolver[T context.Context] struct{}

func (d NoResolver[T]) Record(_ pipeline.Step[T]) {
	// noop
}

func (d NoResolver[T]) RequireDependencyByStepName(_ ...string) error {
	// noop
	return nil
}

func (d NoResolver[T]) MustRequireDependencyByStepName(_ ...string) {
	// noop
}

func (d NoResolver[T]) RequireDependencyByFuncName(_ ...pipeline.ActionFunc[T]) error {
	// noop
	return nil
}

func (d NoResolver[T]) MustRequireDependencyByFuncName(_ ...pipeline.ActionFunc[T]) {
	// noop
}
