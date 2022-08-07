package pipeline

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
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
			goleak.VerifyNone(t)
			counts = 0
			if tt.expectPanic {
				assert.Panics(t, func() {
					NewWorkerPoolStep[context.Context]("pool", 0, nil, nil)
				})
				return
			}
			pipes := []*Pipeline[*testContext]{
				NewPipeline[*testContext]().AddStep(NewStep[*testContext]("step", func(_ *testContext) error {
					atomic.AddUint64(&counts, 1)
					return newResult("step", tt.expectedError)
				})),
			}
			step := NewWorkerPoolStep("pool", 1, SupplierFromSlice(pipes),
				func(ctx *testContext, results map[uint64]error) error {
					assert.Error(t, results[0])
					return results[0]
				})
			ctx := &testContext{Context: context.Background()}
			result := step.Action(ctx)
			assert.Error(t, result)
		})
	}
}

func TestNewWorkerPoolStep_Cancel(t *testing.T) {
	defer goleak.VerifyNone(t)
	step := NewWorkerPoolStep[*testContext]("workerpool", 2, func(ctx *testContext, pipelines chan *Pipeline[*testContext]) {
		defer close(pipelines)
		for i := 0; i < 10000; i++ {
			select {
			case <-ctx.Done():
				return
			default:
				p := NewPipeline[*testContext]()
				pipelines <- p.WithSteps(
					p.NewStep("noop", func(_ *testContext) error { return nil }),
					p.NewStep("increase", func(ctx *testContext) error {
						atomic.AddInt64(&ctx.count, 1)
						time.Sleep(10 * time.Millisecond)
						return nil
					}))
			}
		}
		t.Fail() // should not reach this
	}, func(ctx *testContext, results map[uint64]error) error {
		require.Len(t, results, 9)
		for r := uint64(0); r < 6; r++ {
			var result Result
			// The first 6 jobs are successful
			require.NoError(t, result)
		}
		for r := uint64(6); r < 9; r++ {
			require.Error(t, results[r])
			var result Result
			if errors.As(results[r], &result) {
				// remaining jobs were cancelled
				assert.EqualError(t, result, `step 'noop' failed: context deadline exceeded`)
				assert.Equal(t, "noop", result.Name())
			} else {
				require.Fail(t, "errors are not of type Result")
			}
		}
		return nil
	})
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Millisecond)
	defer cancel()
	pctx := &testContext{Context: ctx}
	err := NewPipeline[*testContext]().WithSteps(step).RunWithContext(pctx)
	require.Error(t, err)
	assert.Equal(t, int64(6), pctx.count, "successful increments")
	assert.EqualError(t, err, `step 'workerpool' failed: context deadline exceeded`)
}

func ExampleNewWorkerPoolStep() {
	p := NewPipeline[*testContext]()
	pool := NewWorkerPoolStep[*testContext]("pool", 2, func(ctx *testContext, pipelines chan *Pipeline[*testContext]) {
		defer close(pipelines)
		// create some pipelines
		for i := 0; i < 3; i++ {
			n := i
			select {
			case <-ctx.Done():
				return // parent pipeline has been canceled, let's not create more pipelines.
			default:
				newP := NewPipeline[*testContext]()
				pipelines <- newP.AddStep(newP.NewStep(fmt.Sprintf("i = %d", n), func(_ *testContext) error {
					time.Sleep(time.Duration(n * 100000000)) // fake some load
					fmt.Println(fmt.Sprintf("This is job item %d", n))
					return nil
				}))
			}
		}
	}, func(ctx *testContext, results map[uint64]error) error {
		for jobIndex, err := range results {
			var result Result
			if errors.As(err, &result) {
				fmt.Println(fmt.Sprintf("Job %d failed: %v", jobIndex, result))
			}
		}
		return nil
	})
	p.AddStep(pool)
	p.RunWithContext(&testContext{Context: context.Background()})
	// Output: This is job item 0
	// This is job item 1
	// This is job item 2
}
