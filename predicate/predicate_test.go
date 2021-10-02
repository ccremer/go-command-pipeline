package predicate

import (
	"testing"

	pipeline "github.com/ccremer/go-command-pipeline"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Predicates(t *testing.T) {
	counter := 0
	tests := map[string]struct {
		givenPredicate Predicate
		expectedCounts int
	}{
		"GivenBoolPredicate_WhenRunning_ThenInvoke": {
			givenPredicate: Bool(true),
			expectedCounts: 1,
		},
		"GivenAnd_WhenBothFalse_ThenIgnoreSecondPredicate": {
			givenPredicate: And(falsePredicate(&counter), falsePredicate(&counter)),
			expectedCounts: -1,
		},
		"GivenAndPredicate_WhenFirstTrue_ThenIgnore": {
			givenPredicate: And(truePredicate(&counter), falsePredicate(&counter)),
			expectedCounts: 0,
		},
		"GivenAndPredicate_WhenBothTrue_ThenRunAction": {
			givenPredicate: And(truePredicate(&counter), truePredicate(&counter)),
			expectedCounts: 3,
		},
		"GivenNotPredicate_WhenFalse_ThenRunAction": {
			givenPredicate: Not(falsePredicate(&counter)),
			expectedCounts: 0,
		},
		"GivenNotPredicate_WhenTrue_ThenIgnoreAction": {
			givenPredicate: Not(truePredicate(&counter)),
			expectedCounts: 1,
		},
		"GivenOrPredicate_WhenBothFalse_ThenIgnoreAction": {
			givenPredicate: Or(falsePredicate(&counter), falsePredicate(&counter)),
			expectedCounts: -2,
		},
		"GivenOrPredicate_WhenFirstTrue_ThenRunAction": {
			givenPredicate: Or(truePredicate(&counter), falsePredicate(&counter)),
			expectedCounts: 2,
		},
		"GivenOrPredicate_WhenSecondTrue_ThenRunAction": {
			givenPredicate: Or(falsePredicate(&counter), truePredicate(&counter)),
			expectedCounts: 1,
		},
		"GivenOrPredicate_WhenBothTrue_ThenRunActionAfterFirstPredicate": {
			givenPredicate: Or(truePredicate(&counter), truePredicate(&counter)),
			expectedCounts: 2,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			counter = 0
			step := ToStep("name", func(_ pipeline.Context) pipeline.Result {
				counter += 1
				return pipeline.Result{}
			}, tt.givenPredicate)
			result := step.F(nil)
			assert.Equal(t, tt.expectedCounts, counter)
			assert.NoError(t, result.Err)
		})
	}
}

func TestToNestedStep(t *testing.T) {
	counter := 0
	tests := map[string]struct {
		givenPredicate Predicate
		expectedCounts int
	}{
		"GivenPipeline_WhenPredicateEvalsTrue_ThenRunPipeline": {
			givenPredicate: truePredicate(&counter),
			expectedCounts: 2,
		},
		"GivenPipeline_WhenPredicateEvalsFalse_ThenIgnorePipeline": {
			givenPredicate: falsePredicate(&counter),
			expectedCounts: -1,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			counter = 0
			p := pipeline.NewPipeline().AddStep(pipeline.NewStep("nested step", func(_ pipeline.Context) pipeline.Result {
				counter++
				return pipeline.Result{}
			}))
			step := ToNestedStep("super step", p, tt.givenPredicate)
			_ = step.F(nil)
			assert.Equal(t, tt.expectedCounts, counter)
		})
	}
}

func TestWrapIn(t *testing.T) {
	counter := 0
	tests := map[string]struct {
		givenPredicate Predicate
		expectedCalls  int
	}{
		"GivenWrappedStep_WhenPredicateEvalsTrue_ThenRunAction": {
			givenPredicate: truePredicate(&counter),
			expectedCalls:  2,
		},
		"GivenWrappedStep_WhenPredicateEvalsFalse_ThenIgnoreAction": {
			givenPredicate: falsePredicate(&counter),
			expectedCalls:  -1,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			counter = 0
			step := pipeline.NewStep("step", func(_ pipeline.Context) pipeline.Result {
				counter++
				return pipeline.Result{}
			})
			wrapped := WrapIn(step, tt.givenPredicate)
			result := wrapped.F(nil)
			require.NoError(t, result.Err)
			assert.Equal(t, tt.expectedCalls, counter)
			assert.Equal(t, step.Name, wrapped.Name)
		})
	}
}

func truePredicate(counter *int) Predicate {
	return func(_ pipeline.Context, step pipeline.Step) bool {
		*counter++
		return true
	}
}

func falsePredicate(counter *int) Predicate {
	return func(_ pipeline.Context, step pipeline.Step) bool {
		*counter--
		return false
	}
}
