package parallel

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	pipeline "github.com/ccremer/go-command-pipeline"
	"github.com/stretchr/testify/assert"
)

func TestNewFanOutStep(t *testing.T) {
	var counts uint64
	tests := map[string]struct {
		jobs               int
		givenResultHandler ResultHandler
		returnErr          error
		expectedCounts     int
	}{
		"GivenSinglePipeline_WhenRunningStep_ThenReturnSuccess": {
			jobs:           1,
			expectedCounts: 1,
		},
		"GivenMultiplePipeline_WhenRunningStep_ThenReturnSuccess": {
			jobs:           200,
			expectedCounts: 200,
		},
		"GivenPipelineWith_WhenRunningStep_ThenReturnSuccessButRunErrorHandler": {
			jobs:      1,
			returnErr: fmt.Errorf("should be called"),
			givenResultHandler: func(ctx pipeline.Context, _ map[uint64]pipeline.Result) pipeline.Result {
				atomic.AddUint64(&counts, 1)
				return pipeline.Result{}
			},
			expectedCounts: 2,
		},
	}
	for name, tt := range tests {
		counts = 0
		t.Run(name, func(t *testing.T) {
			handler := tt.givenResultHandler
			if handler == nil {
				handler = func(ctx pipeline.Context, results map[uint64]pipeline.Result) pipeline.Result {
					assert.NoError(t, results[0].Err)
					return pipeline.Result{}
				}
			}
			step := NewFanOutStep("fanout", func(funcs chan *pipeline.Pipeline) {
				defer close(funcs)
				for i := 0; i < tt.jobs; i++ {
					funcs <- pipeline.NewPipeline().WithSteps(pipeline.NewStep("step", func(_ pipeline.Context) pipeline.Result {
						atomic.AddUint64(&counts, 1)
						return pipeline.Result{Err: tt.returnErr}
					}))
				}
			}, handler)
			result := step.F(nil)
			assert.NoError(t, result.Err)
			assert.Equal(t, uint64(tt.expectedCounts), counts)
		})
	}
}

func ExampleNewFanOutStep() {
	p := pipeline.NewPipeline()
	fanout := NewFanOutStep("fanout", func(pipelines chan *pipeline.Pipeline) {
		defer close(pipelines)
		// create some pipelines
		for i := 0; i < 3; i++ {
			n := i
			pipelines <- pipeline.NewPipeline().AddStep(pipeline.NewStep(fmt.Sprintf("i = %d", n), func(_ pipeline.Context) pipeline.Result {
				time.Sleep(time.Duration(n * 10000000)) // fake some load
				fmt.Println(fmt.Sprintf("I am worker %d", n))
				return pipeline.Result{}
			}))
		}
	}, func(ctx pipeline.Context, results map[uint64]pipeline.Result) pipeline.Result {
		for worker, result := range results {
			if result.IsFailed() {
				fmt.Println(fmt.Sprintf("Worker %d failed: %v", worker, result.Err))
			}
		}
		return pipeline.Result{}
	})
	p.AddStep(fanout)
	p.Run()
	// Output: I am worker 0
	// I am worker 1
	// I am worker 2
}
