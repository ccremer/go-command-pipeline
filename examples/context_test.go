//go:build examples

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

var key = struct{}{}

func TestExample_Context(t *testing.T) {
	// Create pipeline with defaults
	p := pipeline.NewPipeline()
	p.WithSteps(
		pipeline.NewStepFromFunc("define random number", defineNumber),
		pipeline.NewStepFromFunc("print number", printNumber),
	)
	result := p.RunWithContext(context.WithValue(context.Background(), key, &Data{}))
	if !result.IsSuccessful() {
		t.Fatal(result.Err())
	}
}

func defineNumber(ctx context.Context) error {
	ctx.Value(key).(*Data).Number = rand.Int()
	return nil
}

func printNumber(ctx context.Context) error {
	_, err := fmt.Println(ctx.Value(key).(*Data).Number)
	return err
}
