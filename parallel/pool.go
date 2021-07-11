package parallel

import (
	"sync"
	"sync/atomic"

	pipeline "github.com/ccremer/go-command-pipeline"
)

/*
NewWorkerPoolStep creates a pipeline step that runs nested pipelines in a thread pool.
The function provided as PipelineSupplier is expected to close the given channel when no more pipelines should be executed, otherwise this step blocks forever.
The step waits until all pipelines are finished.
If the given ResultHandler is non-nil it will be called after all pipelines were run, otherwise the step is considered successful.
The pipelines are executed in a pool of a number of Go routines indicated by size.
If size is 1, the pipelines are effectively run in sequence.
If size is 0 or less, the function panics.
*/
func NewWorkerPoolStep(name string, size int, pipelineSupplier PipelineSupplier, handler ResultHandler) pipeline.Step {
	if size < 1 {
		panic("pool size cannot be lower than 1")
	}
	step := pipeline.Step{Name: name}
	step.F = func() pipeline.Result {
		pipelineChan := make(chan *pipeline.Pipeline, size)
		m := sync.Map{}
		var wg sync.WaitGroup
		count := uint64(0)

		go pipelineSupplier(pipelineChan)
		for i := 0; i < size; i++ {
			wg.Add(1)
			go poolWork(pipelineChan, &wg, &count, &m)
		}

		wg.Wait()
		return collectResults(handler, &m)
	}
	return step
}

func poolWork(pipelineChan chan *pipeline.Pipeline, wg *sync.WaitGroup, i *uint64, m *sync.Map) {
	defer wg.Done()
	for pipe := range pipelineChan {
		n := atomic.AddUint64(i, 1) - 1
		m.Store(n, pipe.Run())
	}
}
