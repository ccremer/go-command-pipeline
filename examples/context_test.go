//go:build examples

package examples

import (
	"context"
	"fmt"
	"math/rand"
	"testing"

	pipeline "github.com/ccremer/go-command-pipeline"
)

type Context struct {
	context.Context
	Number int
}

func TestExample_Context(t *testing.T) {
	// Create pipeline with defaults
	p := pipeline.NewPipeline[*Context]()
	p.WithSteps(
		p.NewStep("define random number", defineNumber),
		p.NewStep("print number", printNumber),
	)
	err := p.RunWithContext(&Context{Context: context.Background()})
	if err != nil {
		t.Fatal(err)
	}
}

func defineNumber(ctx *Context) error {
	ctx.Number = rand.Int()
	return nil
}

func printNumber(ctx *Context) error {
	_, err := fmt.Println(ctx.Number)
	return err
}
