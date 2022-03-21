package pipeline

import (
	"context"
	"sync"
)

/*
NewFanOutStep creates a pipeline step that runs nested pipelines in their own Go routines.
The function provided as Supplier is expected to close the given channel when no more pipelines should be executed, otherwise this step blocks forever.
The step waits until all pipelines are finished.
If the given ParallelResultHandler is non-nil it will be called after all pipelines were run, otherwise the step is considered successful.

If the context is canceled, no new pipelines will be retrieved from the channel and the Supplier is expected to stop supplying new instances.
Also, once canceled, the step waits for the remaining children pipelines and collects their result via given ParallelResultHandler.
However, the error returned from ParallelResultHandler is wrapped in context.Canceled.
*/
func NewFanOutStep(name string, pipelineSupplier Supplier, handler ParallelResultHandler) Step {
	step := Step{Name: name}
	step.F = func(ctx context.Context) Result {
		pipelineChan := make(chan *Pipeline)
		m := sync.Map{}
		var wg sync.WaitGroup
		i := uint64(0)

		go pipelineSupplier(ctx, pipelineChan)
		for pipe := range pipelineChan {
			p := pipe
			wg.Add(1)
			n := i
			i++
			go func() {
				defer wg.Done()
				m.Store(n, p.RunWithContext(ctx))
			}()
		}
		wg.Wait()
		res := collectResults(ctx, handler, &m)
		return setResultErrorFromContext(ctx, res)
	}
	return step
}
