package parallel

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	pipeline "github.com/ccremer/go-command-pipeline"
	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
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
			givenResultHandler: func(ctx context.Context, _ map[uint64]pipeline.Result) pipeline.Result {
				atomic.AddUint64(&counts, 1)
				return pipeline.Result{}
			},
			expectedCounts: 2,
		},
	}
	for name, tt := range tests {
		counts = 0
		t.Run(name, func(t *testing.T) {
			goleak.VerifyNone(t)
			handler := tt.givenResultHandler
			if handler == nil {
				handler = func(ctx context.Context, results map[uint64]pipeline.Result) pipeline.Result {
					assert.NoError(t, results[0].Err)
					return pipeline.Result{}
				}
			}
			step := NewFanOutStep("fanout", func(_ context.Context, funcs chan *pipeline.Pipeline) {
				defer close(funcs)
				for i := 0; i < tt.jobs; i++ {
					funcs <- pipeline.NewPipeline().WithSteps(pipeline.NewStep("step", func(_ context.Context) pipeline.Result {
						atomic.AddUint64(&counts, 1)
						return pipeline.Result{Err: tt.returnErr}
					}))
				}
			}, handler)
			result := step.F(context.Background())
			assert.NoError(t, result.Err)
			assert.Equal(t, tt.expectedCounts, int(counts))
		})
	}
}

func TestNewFanOutStep_Cancel(t *testing.T) {
	defer goleak.VerifyNone(t)
	var counts uint64
	step := NewFanOutStep("fanout", func(ctx context.Context, pipelines chan *pipeline.Pipeline) {
		defer close(pipelines)
		for i := 0; i < 10000; i++ {
			select {
			case <-ctx.Done():
				return
			default:
				pipelines <- pipeline.NewPipeline().WithSteps(pipeline.NewStepFromFunc("increase", func(_ context.Context) error {
					atomic.AddUint64(&counts, 1)
					return nil
				}))
				time.Sleep(10 * time.Millisecond)
			}
		}
		t.Fail() // should not reach this
	}, func(ctx context.Context, results map[uint64]pipeline.Result) pipeline.Result {
		assert.Len(t, results, 3)
		return pipeline.Result{Err: fmt.Errorf("some error")}
	})
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Millisecond)
	defer cancel()
	result := pipeline.NewPipeline().WithSteps(step).RunWithContext(ctx)
	assert.Equal(t, 3, int(counts))
	assert.True(t, result.IsCanceled(), "canceled flag")
	assert.EqualError(t, result.Err, `step "fanout" failed: context deadline exceeded, collection error: some error`)
}

func ExampleNewFanOutStep() {
	p := pipeline.NewPipeline()
	fanout := NewFanOutStep("fanout", func(ctx context.Context, pipelines chan *pipeline.Pipeline) {
		defer close(pipelines)
		// create some pipelines
		for i := 0; i < 3; i++ {
			n := i
			select {
			case <-ctx.Done():
				return // parent pipeline has been canceled, let's not create more pipelines.
			default:
				pipelines <- pipeline.NewPipeline().AddStep(pipeline.NewStep(fmt.Sprintf("i = %d", n), func(_ context.Context) pipeline.Result {
					time.Sleep(time.Duration(n * 10000000)) // fake some load
					fmt.Println(fmt.Sprintf("I am worker %d", n))
					return pipeline.Result{}
				}))
			}
		}
	}, func(ctx context.Context, results map[uint64]pipeline.Result) pipeline.Result {
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
