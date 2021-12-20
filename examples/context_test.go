//go:build examples
// +build examples

package examples

import (
	"fmt"
	"math/rand"
	"testing"

	pipeline "github.com/ccremer/go-command-pipeline"
)

type Data struct {
	Number int
}

func TestExample_Context(t *testing.T) {
	// Create pipeline with defaults
	p := pipeline.NewPipeline()
	p.WithContext(&Data{})
	p.WithSteps(
		pipeline.NewStep("define random number", defineNumber),
		pipeline.NewStepFromFunc("print number", printNumber),
	)
	result := p.Run()
	if !result.IsSuccessful() {
		t.Fatal(result.Err)
	}
}

func defineNumber(ctx pipeline.Context) pipeline.Result {
	ctx.(*Data).Number = rand.Int()
	return pipeline.Result{}
}

func printNumber(ctx pipeline.Context) error {
	_, err := fmt.Println(ctx.(*Data).Number)
	return err
}
