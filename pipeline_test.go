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

func (h *hook) Accept(_ Step[*testContext]) {
	h.calls += 1
}

func TestPipeline_RunWithContext(t *testing.T) {
	hook := &hook{}
	tests := map[string]struct {
		givenSteps           []Step[*testContext]
		givenBeforeHook      Listener[*testContext]
		givenFinalizer       ErrorHandler[*testContext]
		expectErrorString    string
		expectedCalls        int64
		additionalAssertions func(t *testing.T, err error)
	}{
		"GivenSingleStep_WhenRunning_ThenCallStep": {
			givenSteps: []Step[*testContext]{
				NewStep[*testContext]("test-step", func(ctx *testContext) error {
					ctx.count += 1
					return nil
				}),
			},
			expectedCalls: 1,
		},
		"GivenSingleStep_WhenBeforeHookGiven_ThenCallBeforeHook": {
			givenSteps: []Step[*testContext]{
				NewStep[*testContext]("test-step", func(ctx *testContext) error {
					ctx.count += int64(hook.calls + 1)
					return nil
				}),
			},
			givenBeforeHook: hook.Accept,
			expectedCalls:   2,
		},
		"GivenPipelineWithFinalizer_WhenRunning_ThenCallHandler": {
			givenFinalizer: func(ctx *testContext, err error) error {
				ctx.count += 1
				return nil
			},
			expectedCalls: 1,
		},
		"GivenSingleStepWithoutHandler_WhenRunningWithError_ThenReturnError": {
			givenSteps: []Step[*testContext]{
				NewStep("test-step", func(ctx *testContext) error {
					ctx.count += 1
					return errors.New("step failed")
				}),
			},
			expectedCalls:     1,
			expectErrorString: "step failed",
		},
		"GivenSingleStepWithHandler_WhenRunningWithError_ThenAbortWithError": {
			givenSteps: []Step[*testContext]{
				NewStep[*testContext]("test-step", func(ctx *testContext) error {
					ctx.count += 1
					return nil
				}).WithErrorHandler(func(ctx *testContext, _ error) error {
					ctx.count += 1
					return errors.New("handler")
				}),
				NewStep[*testContext]("don't run this step", func(ctx *testContext) error {
					ctx.count += 1
					return nil
				}),
			},
			expectedCalls:     2,
			expectErrorString: "handler",
		},
		"GivenSingleStepWithHandler_WhenNullifyingError_ThenContinuePipeline": {
			givenSteps: []Step[*testContext]{
				NewStep[*testContext]("test-step", func(ctx *testContext) error {
					ctx.count += 1
					return errors.New("failed step")
				}).WithErrorHandler(func(ctx *testContext, _ error) error {
					ctx.count += 1
					return nil
				}),
				NewStep[*testContext]("continue", func(ctx *testContext) error {
					ctx.count += 1
					return nil
				}),
			},
			additionalAssertions: func(t *testing.T, err error) {
				assert.NoError(t, err)
			},
			expectedCalls: 3,
		},
		"GivenNestedPipeline_WhenParentPipelineRuns_ThenRunNestedAsWell": {
			givenSteps: []Step[*testContext]{
				NewStep[*testContext]("test-step", func(ctx *testContext) error {
					ctx.count += 1
					return nil
				}),
				NewPipeline[*testContext]().
					AddStep(NewStep[*testContext]("nested-step", func(ctx *testContext) error {
						ctx.count += 1
						return nil
					})).AsNestedStep("nested-pipeline"),
			},
			expectedCalls: 2,
		},
		"GivenNestedPipeline_WhenParentPipelineRuns_ThenRunNestedAsWell_Variant2": {
			givenSteps: []Step[*testContext]{
				NewPipeline[*testContext]().
					WithNestedSteps("nested-pipeline", nil,
						NewStep[*testContext]("nested-step", func(ctx *testContext) error {
							ctx.count += 1
							return nil
						})),
			},
			expectedCalls: 1,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			p := &Pipeline[*testContext]{}
			p.WithSteps(tt.givenSteps...)
			p.WithFinalizer(tt.givenFinalizer)
			if tt.givenBeforeHook != nil {
				p.WithBeforeHooks(tt.givenBeforeHook)
			}
			pctx := &testContext{Context: context.Background(), count: 0}
			err := p.RunWithContext(pctx)
			if tt.expectErrorString != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectErrorString)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.expectedCalls, pctx.count)
			if tt.additionalAssertions != nil {
				tt.additionalAssertions(t, err)
			}
		})
	}
}

func TestPipeline_RunWithContext_CancelLongRunningStep(t *testing.T) {
	p := NewPipeline[context.Context]()
	p.AddStepFromFunc("long running", func(ctx context.Context) error {
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
	err := p.RunWithContext(ctx).(Result)
	assert.Equal(t, "long running", err.Name())
	assert.EqualError(t, err, "step 'long running' failed: context canceled")
}

func TestPipeline_RunWithContext_ErrorAs(t *testing.T) {
	p := NewPipeline[context.Context]()
	p.WithSteps(p.NewStep("error-as", func(ctx context.Context) error {
		return errors.New("error")
	}))
	err := p.RunWithContext(context.Background())
	var result Result
	if errors.As(err, &result) {
		assert.EqualError(t, result, `step 'error-as' failed: error`)
		assert.Equal(t, "error-as", result.Name())
	}
}

func ExamplePipeline_RunWithContext() {
	// prepare pipeline
	type exampleContext struct {
		context.Context
		field string
	}

	p := NewPipeline[*exampleContext]()
	p.WithSteps(
		p.NewStep("short step", func(ctx *exampleContext) error {
			fmt.Println(ctx.field)
			return nil
		}),
		p.NewStep("long running step", func(ctx *exampleContext) error {
			time.Sleep(100 * time.Millisecond)
			return nil
		}),
		p.NewStep("canceled step", func(ctx *exampleContext) error {
			return errors.New("shouldn't execute")
		}),
	)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	pctx := &exampleContext{ctx, "hello world"}
	err := p.RunWithContext(pctx)
	// inspect the result
	fmt.Println(err)
	// Output: hello world
	// step 'canceled step' failed: context deadline exceeded
}

func ExamplePipeline_When() {
	p := NewPipeline[context.Context]()
	p.WithSteps(
		p.When(Bool[context.Context](true), "run", func(ctx context.Context) error {
			return nil
		}),
	)
}

type testContext struct {
	context.Context
	count int64
}
