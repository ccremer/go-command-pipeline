/*
Package parallel extends the command-pipeline core with concurrency steps.
*/
package parallel

import (
	"context"

	pipeline "github.com/ccremer/go-command-pipeline"
)

// PipelineSupplier is a function that spawns pipeline.Pipeline for consumption.
// Supply new pipelines by putting new pipeline.Pipeline instances into the given channel.
// The function must close the channel once all pipelines are spawned (`defer close()` recommended).
//
// The parent pipeline may get canceled, thus the given context is provided to stop putting more pipeline.Pipeline instances into the channel.
// Use
//  select { case <-ctx.Done(): return, default: pipelinesChan <- ... }
// to cancel the supply, otherwise you may leak an orphaned goroutine.
type PipelineSupplier func(ctx context.Context, pipelinesChan chan *pipeline.Pipeline)
