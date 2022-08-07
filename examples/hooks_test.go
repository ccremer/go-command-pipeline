//go:build examples

package examples

import (
	"context"
	"fmt"
	"testing"

	pipeline "github.com/ccremer/go-command-pipeline"
)

func TestExample_Hooks(t *testing.T) {
	p := pipeline.NewPipeline[context.Context]()
	p.WithBeforeHooks(func(step pipeline.Step[context.Context]) {
		fmt.Println(fmt.Sprintf("Entering step: %s", step.Name))
	})
	p.WithSteps(
		p.NewStep("hook demo", AfterHookAction),
	)
	err := p.RunWithContext(context.Background())
	if err != nil {
		t.Fatal(err)
	}
}

func AfterHookAction(_ context.Context) error {
	fmt.Println("I am called in an action after the hooks")
	return nil
}
