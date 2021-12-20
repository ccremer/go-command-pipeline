//go:build examples
// +build examples

package examples

import (
	"fmt"
	"testing"

	pipeline "github.com/ccremer/go-command-pipeline"
)

type PipelineLogger struct{}

func (l *PipelineLogger) Accept(step pipeline.Step) {
	fmt.Println(fmt.Sprintf("Executing step: %s", step.Name))
}

func TestExample_Hooks(t *testing.T) {
	p := pipeline.NewPipeline()
	p.AddBeforeHook(&PipelineLogger{})
	p.WithSteps(
		pipeline.NewStep("hook demo", AfterHookAction()),
	)
	result := p.Run()
	if !result.IsSuccessful() {
		t.Fatal(result.Err)
	}
}

func AfterHookAction() pipeline.ActionFunc {
	return func(ctx pipeline.Context) pipeline.Result {
		fmt.Println("I am called in an action after the hooks")
		return pipeline.Result{}
	}
}
