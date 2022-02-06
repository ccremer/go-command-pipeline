//go:build examples
// +build examples

package examples

import (
	"context"
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
	p.WithContext(context.WithValue(context.Background(), "data", &Data{}))
	p.WithSteps(
		pipeline.NewStep("define random number", defineNumber),
		pipeline.NewStepFromFunc("print number", printNumber),
	)
	result := p.Run()
	if !result.IsSuccessful() {
		t.Fatal(result.Err)
	}
}

func defineNumber(ctx context.Context) pipeline.Result {
	ctx.Value("data").(*Data).Number = rand.Int()
	return pipeline.Result{}
}

func printNumber(ctx context.Context) error {
	_, err := fmt.Println(ctx.Value("data").(*Data).Number)
	return err
}
