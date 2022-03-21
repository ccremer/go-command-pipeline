package pipeline

import (
	"context"
	"fmt"
	"sync"
)

// ParallelResultHandler is a callback that provides a Result map and expect a single, combined Result object.
// The map key is a zero-based index of n-th Pipeline spawned, e.g. pipeline number 3 will have index 2.
// Return an empty Result if you want to ignore errors, or reduce multiple errors into a single one to make the parent Pipeline fail.
type ParallelResultHandler func(ctx context.Context, results map[uint64]Result) Result

func collectResults(ctx context.Context, handler ParallelResultHandler, m *sync.Map) Result {
	collectiveResult := Result{}
	if handler != nil {
		// convert sync.Map to conventional map for easier access
		resultMap := make(map[uint64]Result)
		m.Range(func(key, value interface{}) bool {
			resultMap[key.(uint64)] = value.(Result)
			return true
		})
		collectiveResult = handler(ctx, resultMap)
	}
	return collectiveResult
}

func setResultErrorFromContext(ctx context.Context, result Result) Result {
	if ctx.Err() != nil {
		if result.Err() != nil {
			err := fmt.Errorf("%w, collection error: %v", ctx.Err(), result.Err())
			return NewResult(result.Name(), err, result.IsAborted(), true)
		}
		return NewResult(result.Name(), ctx.Err(), result.IsAborted(), true)
	}
	return result
}
