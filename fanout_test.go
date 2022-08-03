package pipeline

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/goleak"
)

func TestNewFanOutStep(t *testing.T) {
	var counts uint64
	tests := map[string]struct {
		jobs               int
		givenResultHandler ParallelResultHandler[*testContext]
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
			givenResultHandler: func(ctx *testContext, _ map[uint64]error) error {
				atomic.AddUint64(&counts, 1)
				return nil
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
				handler = func(ctx *testContext, results map[uint64]error) error {
					assert.NoError(t, results[0])
					return nil
				}
			}
			step := NewFanOutStep("fanout", func(_ *testContext, funcs chan *Pipeline[*testContext]) {
				defer close(funcs)
				for i := 0; i < tt.jobs; i++ {
					p := NewPipeline[*testContext]()
					funcs <- p.WithSteps(p.NewStep("step", func(_ *testContext) error {
						atomic.AddUint64(&counts, 1)
						return tt.returnErr
					}))
				}
			}, handler)
			result := step.Action(&testContext{context.Background()})
			assert.NoError(t, result)
			assert.Equal(t, tt.expectedCounts, int(counts))
		})
	}
}

func TestNewFanOutStep_Cancel(t *testing.T) {
	defer goleak.VerifyNone(t)
	var counts uint64
	step := NewFanOutStep("fanout", func(ctx *testContext, pipelines chan *Pipeline[*testContext]) {
		defer close(pipelines)
		for i := 0; i < 10000; i++ {
			select {
			case <-ctx.Done():
				return
			default:
				p := NewPipeline[*testContext]()
				pipelines <- p.WithSteps(p.NewStep("increase", func(_ *testContext) error {
					atomic.AddUint64(&counts, 1)
					return nil
				}))
				time.Sleep(10 * time.Millisecond)
			}
		}
		t.Fail() // should not reach this
	}, func(ctx *testContext, results map[uint64]error) error {
		assert.Len(t, results, 3)
		return fmt.Errorf("some error")
	})
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Millisecond)
	defer cancel()
	result := NewPipeline[*testContext]().WithSteps(step).RunWithContext(&testContext{ctx})
	assert.Equal(t, 3, int(counts))
	assert.True(t, result.IsCanceled(), "canceled flag")
	assert.EqualError(t, result.Err(), `step "fanout" failed: context deadline exceeded, collection error: some error`)
}

func ExampleNewFanOutStep() {
	p := NewPipeline[context.Context]()
	fanout := NewFanOutStep[context.Context]("fanout", func(ctx context.Context, pipelines chan *Pipeline[context.Context]) {
		defer close(pipelines)
		// create some pipelines
		for i := 0; i < 3; i++ {
			n := i
			select {
			case <-ctx.Done():
				return // parent pipeline has been canceled, let's not create more pipelines.
			default:
				p := NewPipeline[context.Context]()
				pipelines <- p.AddStep(p.NewStep(fmt.Sprintf("i = %d", n), func(_ context.Context) error {
					time.Sleep(time.Duration(n * 10000000)) // fake some load
					fmt.Println(fmt.Sprintf("I am worker %d", n))
					return nil
				}))
			}
		}
	}, func(ctx context.Context, results map[uint64]error) error {
		for worker, result := range results {
			if result != nil {
				fmt.Println(fmt.Sprintf("Worker %d failed: %v", worker, result))
			}
		}
		return nil
	})
	p.AddStep(fanout)
	p.RunWithContext(context.Background())
	// Output: I am worker 0
	// I am worker 1
	// I am worker 2
}
