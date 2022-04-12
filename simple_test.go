package pipeline

import (
	"context"
	"fmt"
)

func ExampleNewAnonymous() {
	p := NewAnonymous(
		func(ctx context.Context) error {
			return nil
		},
		func(ctx context.Context) error {
			return fmt.Errorf("fail")
		}).WithOptions(DisableErrorWrapping)
	result := p.Run()
	fmt.Printf("Name: %s, Err: %v", result.Name(), result.Err())
	// Output: Name: func2, Err: fail
}
