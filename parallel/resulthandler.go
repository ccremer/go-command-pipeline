package parallel

import (
	"context"
	"fmt"
	"sync"

	pipeline "github.com/ccremer/go-command-pipeline"
)

// ResultHandler is a callback that provides a result map and expect a single, combined pipeline.Result object.
// The map key is a zero-based index of n-th pipeline.Pipeline spawned, e.g. pipeline number 3 will have index 2.
// Return an empty pipeline.Result if you want to ignore errors, or reduce multiple errors into a single one to make the parent pipeline fail.
type ResultHandler func(ctx context.Context, results map[uint64]pipeline.Result) pipeline.Result

func collectResults(ctx context.Context, handler ResultHandler, m *sync.Map) pipeline.Result {
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

func setResultErrorFromContext(ctx context.Context, result pipeline.Result) pipeline.Result {
	if ctx.Err() != nil {
		if result.Err() != nil {
			err := fmt.Errorf("%w, collection error: %v", ctx.Err(), result.Err())
			return pipeline.NewResult(result.Name(), err, result.IsAborted(), true)
		}
		return pipeline.NewResult(result.Name(), ctx.Err(), result.IsAborted(), true)
	}
	return result
}
