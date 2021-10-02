package parallel

import (
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	pipeline "github.com/ccremer/go-command-pipeline"
	"github.com/stretchr/testify/assert"
)

func TestNewWorkerPoolStep(t *testing.T) {
	counts := uint64(0)
	tests := map[string]struct {
		expectPanic       bool
		expectedFailIndex uint64
		expectedError     error
	}{
		"GivenInvalidSize_WhenCreatingStep_ThenPanic": {
			expectPanic: true,
		},
		"GivenErrorHandler_WhenPipelineFails_ThenCallWithNestedError": {
			expectedFailIndex: 0,
			expectedError:     errors.New("should fail"),
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			counts = 0
			if tt.expectPanic {
				assert.Panics(t, func() {
					NewWorkerPoolStep("pool", 0, nil, nil)
				})
				return
			}
			step := NewWorkerPoolStep("pool", 1, func(pipelines chan *pipeline.Pipeline) {
				defer close(pipelines)
				pipelines <- pipeline.NewPipeline().AddStep(pipeline.NewStep("step", func(_ pipeline.Context) pipeline.Result {
					atomic.AddUint64(&counts, 1)
					return pipeline.Result{Err: tt.expectedError}
				}))
			}, func(ctx pipeline.Context, results map[uint64]pipeline.Result) pipeline.Result {
				assert.Error(t, results[0].Err)
				return pipeline.Result{Err: results[0].Err}
			})
			result := step.F(nil)
			assert.Error(t, result.Err)
		})
	}
}

func ExampleNewWorkerPoolStep() {
	p := pipeline.NewPipeline()
	pool := NewWorkerPoolStep("pool", 2, func(pipelines chan *pipeline.Pipeline) {
		defer close(pipelines)
		// create some pipelines
		for i := 0; i < 3; i++ {
			n := i
			pipelines <- pipeline.NewPipeline().AddStep(pipeline.NewStep(fmt.Sprintf("i = %d", n), func(_ pipeline.Context) pipeline.Result {
				time.Sleep(time.Duration(n * 100000000)) // fake some load
				fmt.Println(fmt.Sprintf("This is job item %d", n))
				return pipeline.Result{}
			}))
		}
	}, func(ctx pipeline.Context, results map[uint64]pipeline.Result) pipeline.Result {
		for jobIndex, result := range results {
			if result.IsFailed() {
				fmt.Println(fmt.Sprintf("Job %d failed: %v", jobIndex, result.Err))
			}
		}
		return pipeline.Result{}
	})
	p.AddStep(pool)
	p.Run()
	// Output: This is job item 0
	// This is job item 1
	// This is job item 2
}
