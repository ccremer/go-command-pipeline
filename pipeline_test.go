package pipeline

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type hook struct {
	calls int
}

func (h *hook) Accept(_ Step[*PipelineContext]) {
	h.calls += 1
}

func TestPipeline_Run(t *testing.T) {
	callCount := 0
	hook := &hook{}
	tests := map[string]struct {
		givenSteps           []Step[*PipelineContext]
		givenBeforeHook      Listener[*PipelineContext]
		givenFinalizer       ResultHandler[*PipelineContext]
		expectErrorString    string
		expectedCalls        int
		additionalAssertions func(t *testing.T, result Result)
	}{
		"GivenSingleStep_WhenRunning_ThenCallStep": {
			givenSteps: []Step[*PipelineContext]{
				NewStep[*PipelineContext]("test-step", func(_ *PipelineContext) error {
					callCount += 1
					return newEmptyResult("test-step")
				}),
			},
			expectedCalls: 1,
		},
		"GivenSingleStep_WhenBeforeHookGiven_ThenCallBeforeHook": {
			givenSteps: []Step[*PipelineContext]{
				NewStep[*PipelineContext]("test-step", func(_ *PipelineContext) error {
					callCount += hook.calls + 1
					return nil
				}),
			},
			givenBeforeHook: hook.Accept,
			expectedCalls:   2,
		},
		"GivenPipelineWithFinalizer_WhenRunning_ThenCallHandler": {
			givenFinalizer: func(_ *PipelineContext, result Result) error {
				callCount += 1
				return nil
			},
			expectedCalls: 1,
		},
		"GivenSingleStepWithoutHandler_WhenRunningWithError_ThenReturnError": {
			givenSteps: []Step[*PipelineContext]{
				NewStep("test-step", func(_ *PipelineContext) error {
					callCount += 1
					return errors.New("step failed")
				}),
			},
			expectedCalls:     1,
			expectErrorString: "step failed",
		},
		"GivenSingleStepWithHandler_WhenRunningWithError_ThenAbortWithError": {
			givenSteps: []Step[*PipelineContext]{
				NewStep[*PipelineContext]("test-step", func(_ *PipelineContext) error {
					callCount += 1
					return nil
				}).WithResultHandler(func(_ *PipelineContext, result Result) error {
					callCount += 1
					return errors.New("handler")
				}),
				NewStep[*PipelineContext]("don't run this step", func(_ *PipelineContext) error {
					callCount += 1
					return nil
				}),
			},
			expectedCalls:     2,
			expectErrorString: "handler",
		},
		"GivenSingleStepWithHandler_WhenNullifyingError_ThenContinuePipeline": {
			givenSteps: []Step[*PipelineContext]{
				NewStep[*PipelineContext]("test-step", func(_ *PipelineContext) error {
					callCount += 1
					return errors.New("failed step")
				}).WithResultHandler(func(_ *PipelineContext, result Result) error {
					callCount += 1
					return nil
				}),
				NewStep[*PipelineContext]("continue", func(_ *PipelineContext) error {
					callCount += 1
					return nil
				}),
			},
			additionalAssertions: func(t *testing.T, result Result) {
				assert.True(t, result.IsSuccessful())
			},
			expectedCalls: 3,
		},
		"GivenNestedPipeline_WhenParentPipelineRuns_ThenRunNestedAsWell": {
			givenSteps: []Step[*PipelineContext]{
				NewStep[*PipelineContext]("test-step", func(_ *PipelineContext) error {
					callCount += 1
					return nil
				}),
				NewPipeline[*PipelineContext]().
					AddStep(NewStep[*PipelineContext]("nested-step", func(_ *PipelineContext) error {
						callCount += 1
						return nil
					})).AsNestedStep("nested-pipeline"),
			},
			expectedCalls: 2,
		},
		"GivenNestedPipeline_WhenParentPipelineRuns_ThenRunNestedAsWell_Variant2": {
			givenSteps: []Step[*PipelineContext]{
				NewPipeline[*PipelineContext]().
					WithNestedSteps("nested-pipeline",
						NewStep[*PipelineContext]("nested-step", func(_ *PipelineContext) error {
							callCount += 1
							return nil
						})),
			},
			expectedCalls: 1,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			callCount = 0
			p := &Pipeline[*PipelineContext]{}
			p.WithSteps(tt.givenSteps...)
			p.WithFinalizer(tt.givenFinalizer)
			if tt.givenBeforeHook != nil {
				p.AddBeforeHook(tt.givenBeforeHook)
			}
			pctx := &PipelineContext{Context: context.Background()}
			actualResult := p.RunWithContext(pctx)
			if tt.expectErrorString != "" {
				require.Error(t, actualResult.Err())
				assert.True(t, actualResult.IsFailed())
				assert.Contains(t, actualResult.Err().Error(), tt.expectErrorString)
			} else {
				assert.NoError(t, actualResult.Err())
				assert.True(t, actualResult.IsSuccessful())
			}
			assert.Equal(t, tt.expectedCalls, callCount)
			if tt.additionalAssertions != nil {
				tt.additionalAssertions(t, actualResult)
			}
		})
	}
}

func TestPipeline_RunWithContext_CancelLongRunningStep(t *testing.T) {
	p := NewPipeline[PipelineContext]()
	p.AddStepFromFunc("long running", func(ctx PipelineContext) error {
		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				// doing nothing
			}
		}
	})

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	go func() {
		time.Sleep(5 * time.Millisecond)
		cancel()
	}()
	pctx := PipelineContext{Context: ctx}
	result := p.RunWithContext(pctx)
	assert.True(t, result.IsCanceled(), "IsCanceled()")
	assert.Equal(t, "long running", result.Name())
	assert.EqualError(t, result.Err(), "step \"long running\" failed: context canceled")
}

func ExamplePipeline_RunWithContext() {
	// prepare pipeline
	p := NewPipeline[PipelineContext]()
	p.WithSteps(
		p.NewStep("short step", func(ctx PipelineContext) error {
			fmt.Println("short step")
			return nil
		}),
		p.NewStep("long running step", func(ctx PipelineContext) error {
			time.Sleep(100 * time.Millisecond)
			return nil
		}),
		p.NewStep("canceled step", func(ctx PipelineContext) error {
			return errors.New("shouldn't execute")
		}),
	)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	pctx := PipelineContext{Context: ctx}
	result := p.RunWithContext(pctx)
	// inspect the result
	fmt.Println(result.IsCanceled())
	fmt.Println(result.Err())
	// Output: short step
	// true
	// step "canceled step" failed: context deadline exceeded
}
