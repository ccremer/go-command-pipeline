package pipeline

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
)

func TestNewFanOutStep(t *testing.T) {
	tests := map[string]struct {
		jobs               int
		givenResultHandler ParallelResultHandler[*testContext]
		returnErr          error
		expectedCounts     int64
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
				atomic.AddInt64(&ctx.count, 1)
				return nil
			},
			expectedCounts: 2,
		},
	}
	for name, tt := range tests {
		ctx := &testContext{context.Background(), 0}
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
						atomic.AddInt64(&ctx.count, 1)
						return tt.returnErr
					}))
				}
			}, handler)
			err := step.Action(ctx)
			assert.NoError(t, err)
			assert.Equal(t, tt.expectedCounts, ctx.count)
		})
	}
}

func TestNewFanOutStep_Cancel(t *testing.T) {
	defer goleak.VerifyNone(t)
	step := NewFanOutStep("fanout", func(ctx *testContext, pipelines chan *Pipeline[*testContext]) {
		defer close(pipelines)
		for i := 0; i < 10000; i++ {
			select {
			case <-ctx.Done():
				return
			default:
				p := NewPipeline[*testContext]()
				pipelines <- p.WithSteps(p.NewStep("increase", func(ctx *testContext) error {
					atomic.AddInt64(&ctx.count, 1)
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
	pctx := &testContext{ctx, 0}
	err := NewPipeline[*testContext]().WithSteps(step).RunWithContext(pctx)
	require.Error(t, err)
	assert.Equal(t, int64(3), pctx.count)
	assert.EqualError(t, err, `step 'fanout' failed: context deadline exceeded, collection error: some error`)
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
