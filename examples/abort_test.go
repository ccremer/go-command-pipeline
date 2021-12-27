//go:build examples
// +build examples

package examples

import (
	"errors"
	"testing"

	pipeline "github.com/ccremer/go-command-pipeline"
	"github.com/stretchr/testify/assert"
)

func TestExample_Abort(t *testing.T) {
	p := pipeline.NewPipeline()
	p.WithSteps(
		pipeline.NewStepFromFunc("abort demo", abort),
		pipeline.NewStepFromFunc("never executed", doNotExecute),
	)
	result := p.Run()
	assert.True(t, result.IsSuccessful())
}

func doNotExecute(_ pipeline.Context) error {
	return errors.New("should not execute")
}

func abort(_ pipeline.Context) error {
	// some logic that can handle errors, but you don't want to bubble up the error.

	// terminate pipeline gracefully
	return pipeline.ErrAbort
}