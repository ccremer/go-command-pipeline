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
					NewWorkerPoolStep("pool", 0, nil, nil)
				})
				return
			}
			pipes := []*Pipeline[*PipelineContext]{
				NewPipeline[*PipelineContext]().AddStep(NewStep[*PipelineContext]("step", func(_ *PipelineContext) error {
					atomic.AddUint64(&counts, 1)
					return newResultWithError("step", tt.expectedError)
				})),
			}
			step := NewWorkerPoolStep("pool", 1, SupplierFromSlice(pipes),
				func(ctx *PipelineContext, results map[uint64]Result) error {
					assert.Error(t, results[0].Err())
					return results[0].Err()
				})
			ctx := &PipelineContext{Context: context.Background()}
			result := step.F(ctx)
			assert.Error(t, result)
		})
	}
}

func TestNewWorkerPoolStep_Cancel(t *testing.T) {
	defer goleak.VerifyNone(t)
	var counts uint64
	step := NewWorkerPoolStep[*PipelineContext]("workerpool", 2, func(ctx *PipelineContext, pipelines chan *Pipeline[*PipelineContext]) {
		defer close(pipelines)
		for i := 0; i < 10000; i++ {
			select {
			case <-ctx.Done():
				return
			default:
				p := NewPipeline[*PipelineContext]()
				pipelines <- p.WithSteps(
					p.NewStep("noop", func(_ *PipelineContext) error { return nil }),
					p.NewStep("increase", func(_ *PipelineContext) error {
						atomic.AddUint64(&counts, 1)
						time.Sleep(10 * time.Millisecond)
						return nil
					}))
			}
		}
		t.Fail() // should not reach this
	}, func(ctx *PipelineContext, results map[uint64]Result) error {
		require.Len(t, results, 9)
		for r := uint64(0); r < 6; r++ {
			// The first 6 jobs are successful
			assert.Equal(t, "increase", results[r].Name())
			assert.False(t, results[r].IsCanceled())
			assert.NoError(t, results[r].Err())
		}
		for r := uint64(6); r < 9; r++ {
			// remaining jobs were cancelled
			assert.Equal(t, "noop", results[r].Name())
			assert.True(t, results[r].IsCanceled())
			assert.EqualError(t, results[r].Err(), `step "noop" failed: context deadline exceeded`)
		}
		return nil
	})
	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Millisecond)
	defer cancel()
	pctx := &PipelineContext{Context: ctx}
	result := NewPipeline[*PipelineContext]().WithSteps(step).RunWithContext(pctx)
	assert.Equal(t, 6, int(counts), "successful increments")
	assert.True(t, result.IsCanceled(), "overall canceled flag")
	assert.False(t, result.IsSuccessful(), "overall success flag")
	assert.EqualError(t, result.Err(), `step "workerpool" failed: context deadline exceeded`)
}

func ExampleNewWorkerPoolStep() {
	p := NewPipeline[*PipelineContext]()
	pool := NewWorkerPoolStep[*PipelineContext]("pool", 2, func(ctx *PipelineContext, pipelines chan *Pipeline[*PipelineContext]) {
		defer close(pipelines)
		// create some pipelines
		for i := 0; i < 3; i++ {
			n := i
			select {
			case <-ctx.Done():
				return // parent pipeline has been canceled, let's not create more pipelines.
			default:
				newP := NewPipeline[*PipelineContext]()
				pipelines <- newP.AddStep(newP.NewStep(fmt.Sprintf("i = %d", n), func(_ *PipelineContext) error {
					time.Sleep(time.Duration(n * 100000000)) // fake some load
					fmt.Println(fmt.Sprintf("This is job item %d", n))
					return nil
				}))
			}
		}
	}, func(ctx *PipelineContext, results map[uint64]Result) error {
		for jobIndex, result := range results {
			if result.IsFailed() {
				fmt.Println(fmt.Sprintf("Job %d failed: %v", jobIndex, result.Err()))
			}
		}
		return nil
	})
	p.AddStep(pool)
	p.RunWithContext(&PipelineContext{Context: context.Background()})
	// Output: This is job item 0
	// This is job item 1
	// This is job item 2
}
