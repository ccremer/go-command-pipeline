package pipeline

import (
	"errors"
	"testing"

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
		givenSteps        []Step
		givenBeforeHook   Listener
		givenFinalizer    ResultHandler
		expectErrorString string
		expectedCalls     int
	}{
		"GivenSingleStep_WhenRunning_ThenCallStep": {
			givenSteps: []Step{
				NewStep("test-step", func(_ Context) Result {
					callCount += 1
					return Result{}
				}),
			},
			expectedCalls: 1,
		},
		"GivenSingleStep_WhenBeforeHookGiven_ThenCallBeforeHook": {
			givenSteps: []Step{
				NewStep("test-step", func(_ Context) Result {
					callCount += hook.calls + 1
					return Result{}
				}),
			},
			givenBeforeHook: hook.Accept,
			expectedCalls:   2,
		},
		"GivenPipelineWithFinalizer_WhenRunning_ThenCallHandler": {
			givenFinalizer: func(_ Context, result Result) error {
				callCount += 1
				return nil
			},
			expectedCalls: 1,
		},
		"GivenSingleStepWithoutHandler_WhenRunningWithError_ThenReturnError": {
			givenSteps: []Step{
				NewStep("test-step", func(_ Context) Result {
					callCount += 1
					return Result{Err: errors.New("step failed")}
				}),
			},
			expectedCalls:     1,
			expectErrorString: "step failed",
		},
		"GivenStepWithErrAbort_WhenRunningWithErrAbort_ThenDoNotExecuteNextSteps": {
			givenSteps: []Step{
				NewStepFromFunc("test-step", func(_ Context) error {
					callCount += 1
					return ErrAbort
				}),
				NewStepFromFunc("step-should-not-execute", func(_ Context) error {
					callCount += 1
					return errors.New("should not execute")
				}),
			},
			expectedCalls: 1,
		},
		"GivenSingleStepWithHandler_WhenRunningWithError_ThenAbortWithError": {
			givenSteps: []Step{
				NewStep("test-step", func(_ Context) Result {
					callCount += 1
					return Result{}
				}).WithResultHandler(func(_ Context, result Result) error {
					callCount += 1
					return errors.New("handler")
				}),
				NewStep("don't run this step", func(_ Context) Result {
					callCount += 1
					return Result{}
				}),
			},
			expectedCalls:     2,
			expectErrorString: "handler",
		},
		"GivenSingleStepWithHandler_WhenNullifyingError_ThenContinuePipeline": {
			givenSteps: []Step{
				NewStep("test-step", func(_ Context) Result {
					callCount += 1
					return Result{Err: errors.New("failed step")}
				}).WithResultHandler(func(_ Context, result Result) error {
					callCount += 1
					return nil
				}),
				NewStep("continue", func(_ Context) Result {
					callCount += 1
					return Result{}
				}),
			},
			expectedCalls: 3,
		},
		"GivenNestedPipeline_WhenParentPipelineRuns_ThenRunNestedAsWell": {
			givenSteps: []Step{
				NewStep("test-step", func(_ Context) Result {
					callCount += 1
					return Result{}
				}),
				NewPipeline().
					AddStep(NewStep("nested-step", func(_ Context) Result {
						callCount += 1
						return Result{}
					})).AsNestedStep("nested-pipeline"),
			},
			expectedCalls: 2,
		},
		"GivenNestedPipeline_WhenParentPipelineRuns_ThenRunNestedAsWell_Variant2": {
			givenSteps: []Step{
				NewPipeline().
					WithNestedSteps("nested-pipeline",
						NewStep("nested-step", func(_ Context) Result {
							callCount += 1
							return Result{}
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
				require.Error(t, actualResult.Err)
				assert.True(t, actualResult.IsFailed())
				assert.Contains(t, actualResult.Err.Error(), tt.expectErrorString)
			} else {
				assert.NoError(t, actualResult.Err)
				assert.True(t, actualResult.IsSuccessful())
			}
			assert.Equal(t, tt.expectedCalls, callCount)
		})
	}
}

func TestPipeline_RunWithContext(t *testing.T) {
	t.Run("custom type", func(t *testing.T) {
		context := "some type"
		p := NewPipelineWithContext(context)
		p.AddStep(NewStep("context", func(ctx Context) Result {
			assert.Equal(t, context, ctx)
			return Result{}
		}))
		result := p.Run()
		require.NoError(t, result.Err)
	})
	t.Run("nil context", func(t *testing.T) {
		p := NewPipeline().WithContext(nil)
		p.AddStep(NewStep("context", func(ctx Context) Result {
			assert.Nil(t, ctx)
			return Result{}
		}))
		result := p.Run()
		require.NoError(t, result.Err)
	})
}

func TestNewStepFromFunc(t *testing.T) {
	called := false
	step := NewStepFromFunc("name", func(ctx Context) error {
		called = true
		return nil
	})
	_ = step.F(nil)
	assert.True(t, called)
}
