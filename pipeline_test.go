package pipeline

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPipeline_runPipeline(t *testing.T) {
	callCount := 0
	tests := map[string]struct {
		givenSteps        []Step
		givenAbortHandler Handler
		expectedResult    Result
		expectedCalls     int
	}{
		"GivenSingleStep_WhenRunning_ThenCallStep": {
			givenSteps: []Step{
				NewStep("test-step", func() Result {
					callCount += 1
					return Result{}
				}),
			},
			expectedCalls:  1,
			expectedResult: Result{},
		},
		"GivenStepThatAborts_WhenRunning_ThenAbortPipeline": {
			givenSteps: []Step{
				NewStep("abort", func() Result {
					return Result{Abort: true}
				}),
				NewStep("test-step", func() Result {
					callCount += 1
					return Result{}
				}),
			},
			expectedCalls:  0,
			expectedResult: Result{Abort: true},
		},
		"GivenAbortHandler_WhenAborted_ThenRunHandler": {
			givenSteps: []Step{
				NewStep("test-step", func() Result {
					callCount += 1
					return Result{Abort: true}
				}),
			},
			givenAbortHandler: func(result Result) {
				callCount += 1
				assert.True(t, result.Abort)
			},
			expectedCalls:  2,
			expectedResult: Result{Abort: true},
		},
		"GivenNestedPipeline_WhenParentPipelineRuns_ThenRunNestedAsWell": {
			givenSteps: []Step{
				NewStep("test-step", func() Result {
					callCount += 1
					return Result{}
				}),
				NewPipeline().
					WithSteps(NewStep("nested-step", func() Result {
						callCount += 1
						return Result{}
					})).AsNestedStep("nested-pipeline"),
			},
			expectedCalls:  2,
			expectedResult: Result{},
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			callCount = 0
			p := &Pipeline{
				log:          nullLogger{},
				steps:        tt.givenSteps,
				abortHandler: tt.givenAbortHandler,
			}
			actualResult := p.Run()
			assert.Equal(t, tt.expectedResult, actualResult)
			assert.Equal(t, tt.expectedCalls, callCount)
		})
	}
}
