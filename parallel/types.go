/*
Package parallel extends the command-pipeline core with concurrency steps.
*/
package parallel

import (
	"sync"

	pipeline "github.com/ccremer/go-command-pipeline"
)

type (
	// ResultHandler is a callback that provides a result map and expect a single, combined pipeline.Result object.
	// The map key is a zero-based index of n-th pipeline.Pipeline spawned, e.g. pipeline number 3 will have index 2.
	// Context may be nil.
	// Return an empty pipeline.Result if you want to ignore errors, or reduce multiple errors into a single one to make the parent pipeline fail.
	ResultHandler func(ctx pipeline.Context, results map[uint64]pipeline.Result) pipeline.Result
	// PipelineSupplier is a function that spawns pipeline.Pipeline for consumption.
	// The function must close the channel once all pipelines are spawned (`defer close()` recommended).
	PipelineSupplier func(chan *pipeline.Pipeline)
)

func collectResults(ctx pipeline.Context, handler ResultHandler, m *sync.Map) pipeline.Result {
	collectiveResult := pipeline.Result{}
	if handler != nil {
		// convert sync.Map to conventional map for easier access
		resultMap := make(map[uint64]pipeline.Result)
		m.Range(func(key, value interface{}) bool {
			resultMap[key.(uint64)] = value.(pipeline.Result)
			return true
		})
		collectiveResult = handler(ctx, resultMap)
	}
	return collectiveResult
}
