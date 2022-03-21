package pipeline

import (
	"context"
	"sync"
	"sync/atomic"
)

/*
NewWorkerPoolStep creates a pipeline step that runs nested pipelines in a thread pool.
The function provided as Supplier is expected to close the given channel when no more pipelines should be executed, otherwise this step blocks forever.
The step waits until all pipelines are finished.
 * If the given ParallelResultHandler is non-nil it will be called after all pipelines were run, otherwise the step is considered successful.
 * The pipelines are executed in a pool of a number of Go routines indicated by size.
 * If size is 1, the pipelines are effectively run in sequence.
 * If size is 0 or less, the function panics.
*/
func NewWorkerPoolStep(name string, size int, pipelineSupplier Supplier, handler ParallelResultHandler) Step {
	if size < 1 {
		panic("pool size cannot be lower than 1")
	}
	step := Step{Name: name}
	step.F = func(ctx context.Context) Result {
		pipelineChan := make(chan *Pipeline, size)
		m := sync.Map{}
		var wg sync.WaitGroup
		count := uint64(0)

		go pipelineSupplier(ctx, pipelineChan)
		for i := 0; i < size; i++ {
			wg.Add(1)
			go poolWork(ctx, pipelineChan, &wg, &count, &m)
		}

		wg.Wait()
		res := collectResults(ctx, handler, &m)
		return setResultErrorFromContext(ctx, name, res)
	}
	return step
}

func poolWork(ctx context.Context, pipelineChan chan *Pipeline, wg *sync.WaitGroup, i *uint64, m *sync.Map) {
	defer wg.Done()
	for pipe := range pipelineChan {
		n := atomic.AddUint64(i, 1) - 1
		m.Store(n, pipe.RunWithContext(ctx))
	}
}
