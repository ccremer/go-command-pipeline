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

func (h *hook) Accept(_ Step) {
	h.calls += 1
}

func TestPipeline_Run(t *testing.T) {
	callCount := 0
	hook := &hook{}
	tests := map[string]struct {
		givenSteps           []Step
		givenBeforeHook      Listener
		givenFinalizer       ResultHandler
		expectErrorString    string
		expectedCalls        int
		additionalAssertions func(t *testing.T, result Result)
	}{
		"GivenSingleStep_WhenRunning_ThenCallStep": {
			givenSteps: []Step{
				NewStep("test-step", func(_ context.Context) Result {
					callCount += 1
					return newEmptyResult("test-step")
				}),
			},
			expectedCalls: 1,
		},
		"GivenSingleStep_WhenBeforeHookGiven_ThenCallBeforeHook": {
			givenSteps: []Step{
				NewStepFromFunc("test-step", func(_ context.Context) error {
					callCount += hook.calls + 1
					return nil
				}),
			},
			givenBeforeHook: hook.Accept,
			expectedCalls:   2,
		},
		"GivenPipelineWithFinalizer_WhenRunning_ThenCallHandler": {
			givenFinalizer: func(_ context.Context, result Result) error {
				callCount += 1
				return nil
			},
			expectedCalls: 1,
		},
		"GivenSingleStepWithoutHandler_WhenRunningWithError_ThenReturnError": {
			givenSteps: []Step{
				NewStepFromFunc("test-step", func(_ context.Context) error {
					callCount += 1
					return errors.New("step failed")
				}),
			},
			expectedCalls:     1,
			expectErrorString: "step failed",
		},
		"GivenStepWithErrAbort_WhenRunningWithErrAbort_ThenDoNotExecuteNextSteps": {
			givenSteps: []Step{
				NewStepFromFunc("test-step", func(_ context.Context) error {
					callCount += 1
					return ErrAbort
				}),
				NewStepFromFunc("step-should-not-execute", func(_ context.Context) error {
					callCount += 1
					return errors.New("should not execute")
				}),
			},
			expectedCalls: 1,
			additionalAssertions: func(t *testing.T, result Result) {
				assert.True(t, result.IsAborted())
			},
		},
		"GivenSingleStepWithHandler_WhenRunningWithError_ThenAbortWithError": {
			givenSteps: []Step{
				NewStepFromFunc("test-step", func(_ context.Context) error {
					callCount += 1
					return nil
				}).WithResultHandler(func(_ context.Context, result Result) error {
					callCount += 1
					return errors.New("handler")
				}),
				NewStepFromFunc("don't run this step", func(_ context.Context) error {
					callCount += 1
					return nil
				}),
			},
			expectedCalls:     2,
			expectErrorString: "handler",
		},
		"GivenSingleStepWithHandler_WhenNullifyingError_ThenContinuePipeline": {
			givenSteps: []Step{
				NewStepFromFunc("test-step", func(_ context.Context) error {
					callCount += 1
					return errors.New("failed step")
				}).WithResultHandler(func(_ context.Context, result Result) error {
					callCount += 1
					return nil
				}),
				NewStepFromFunc("continue", func(_ context.Context) error {
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
			givenSteps: []Step{
				NewStepFromFunc("test-step", func(_ context.Context) error {
					callCount += 1
					return nil
				}),
				NewPipeline().
					AddStep(NewStepFromFunc("nested-step", func(_ context.Context) error {
						callCount += 1
						return nil
					})).AsNestedStep("nested-pipeline"),
			},
			expectedCalls: 2,
		},
		"GivenNestedPipeline_WhenParentPipelineRuns_ThenRunNestedAsWell_Variant2": {
			givenSteps: []Step{
				NewPipeline().
					WithNestedSteps("nested-pipeline",
						NewStepFromFunc("nested-step", func(_ context.Context) error {
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
			p := &Pipeline{}
			p.WithSteps(tt.givenSteps...)
			p.WithFinalizer(tt.givenFinalizer)
			if tt.givenBeforeHook != nil {
				p.AddBeforeHook(tt.givenBeforeHook)
			}
			actualResult := p.Run()
			if tt.expectErrorString != "" {
				require.Error(t, actualResult.Err())
				assert.True(t, actualResult.IsFailed())
				assert.Contains(t, actualResult.Err().Error(), tt.expectErrorString)
			} else {
				assert.NoError(t, actualResult.Err())
				assert.True(t, actualResult.IsCompleted())
			}
			assert.Equal(t, tt.expectedCalls, callCount)
			if tt.additionalAssertions != nil {
				tt.additionalAssertions(t, actualResult)
			}
		})
	}
}

func TestPipeline_RunWithContext_CancelLongRunningStep(t *testing.T) {
	p := NewPipeline().AddStepFromFunc("long running", func(ctx context.Context) error {
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
	result := p.RunWithContext(ctx)
	assert.True(t, result.IsCanceled(), "IsCanceled()")
	assert.Equal(t, "long running", result.Name())
	assert.EqualError(t, result.Err(), "step \"long running\" failed: context canceled")
}

func ExamplePipeline_RunWithContext() {
	// prepare pipeline
	p := NewPipeline().WithSteps(
		NewStepFromFunc("short step", func(ctx context.Context) error {
			fmt.Println("short step")
			return nil
		}),
		NewStepFromFunc("long running step", func(ctx context.Context) error {
			time.Sleep(100 * time.Millisecond)
			return nil
		}),
		NewStepFromFunc("canceled step", func(ctx context.Context) error {
			return errors.New("shouldn't execute")
		}),
	)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	result := p.RunWithContext(ctx)
	// inspect the result
	fmt.Println(result.IsCanceled())
	fmt.Println(result.Err())
	// Output: short step
	// true
	// step "canceled step" failed: context deadline exceeded
}

func TestNewStepFromFunc(t *testing.T) {
	called := false
	step := NewStepFromFunc("name", func(ctx context.Context) error {
		called = true
		return nil
	})
	_ = step.F(nil)
	assert.True(t, called)
}
