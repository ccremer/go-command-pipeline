package pipeline

import (
	"context"
	"fmt"
	"sync"
)

// ParallelResultHandler is a callback that provides a Result map and expect a single, combined Result object.
// The map key is a zero-based index of n-th Pipeline spawned, e.g. pipeline number 3 will have index 2.
// Return an empty error if you want to ignore errors, or reduce multiple errors into a single one to make the parent Pipeline fail.
type ParallelResultHandler[T context.Context] func(ctx T, results map[uint64]error) error

func collectResults[T context.Context](ctx T, handler ParallelResultHandler[T], m *sync.Map) error {
	if handler != nil {
		// convert sync.Map to conventional map for easier access
		resultMap := make(map[uint64]error)
		m.Range(func(key, value interface{}) bool {
			if value == nil {
				resultMap[key.(uint64)] = nil
			} else {
				resultMap[key.(uint64)] = value.(error)
			}
			return true
		})
		return handler(ctx, resultMap)
	}
	return nil
}

func setResultErrorFromContext(ctx context.Context, name string, err error) error {
	if ctx.Err() != nil {
		if err != nil {
			wrapped := fmt.Errorf("%w, collection error: %v", ctx.Err(), err)
			return newResult(name, wrapped)
		}
		return newResult(name, ctx.Err())
	}
	return err
}
