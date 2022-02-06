//go:build examples
// +build examples

package examples

import (
	"context"
	"fmt"
	"testing"

	pipeline "github.com/ccremer/go-command-pipeline"
)

func TestExample_Hooks(t *testing.T) {
	p := pipeline.NewPipeline()
	p.AddBeforeHook(func(step pipeline.Step) {
		fmt.Println(fmt.Sprintf("Executing step: %s", step.Name))
	})
	p.WithSteps(
		pipeline.NewStep("hook demo", AfterHookAction()),
	)
	result := p.Run()
	if !result.IsSuccessful() {
		t.Fatal(result.Err)
	}
}

func AfterHookAction() pipeline.ActionFunc {
	return func(ctx context.Context) pipeline.Result {
		fmt.Println("I am called in an action after the hooks")
		return pipeline.Result{}
	}
}
