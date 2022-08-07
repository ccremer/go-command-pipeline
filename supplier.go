package pipeline

import (
	"context"
)

// Supplier is a function that spawns Pipeline for consumption.
// Supply new pipelines by putting new Pipeline instances into the given channel.
// The function must close the channel once all pipelines are spawned (`defer close()` recommended).
//
// The parent pipeline may get canceled, thus the given context is provided to stop putting more Pipeline instances into the channel.
// Use
//  select { case <-ctx.Done(): return, default: pipelinesChan <- ... }
// to cancel the supply, otherwise you may leak an orphaned goroutine.
type Supplier[T context.Context] func(ctx T, pipelinesChan chan *Pipeline[T])

// SupplierFromSlice returns a Supplier that accepts the given slice of Pipeline and iterates over it to feed the channel.
//
// Context cancellation is only effective if the channel is limited in size.
// All pipelines may get executed even if the parent pipeline has been canceled, unless each child Pipeline listens for context.Done() in their steps.
func SupplierFromSlice[T context.Context](pipelines []*Pipeline[T]) Supplier[T] {
	return func(ctx T, pipelinesChan chan *Pipeline[T]) {
		defer close(pipelinesChan)
		for _, pipe := range pipelines {
			select {
			case <-ctx.Done():
				return
			default:
				pipelinesChan <- pipe
			}
		}
	}
}
