package predicate

import (
	"testing"

	pipeline "github.com/ccremer/go-command-pipeline"
	"github.com/stretchr/testify/assert"
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
			step := ToStep("name", func() pipeline.Result {
				counter += 1
				return pipeline.Result{}
			}, tt.givenPredicate)
			result := step.F()
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
			p := pipeline.NewPipeline().AddStep(pipeline.NewStep("nested step", func() pipeline.Result {
				counter++
				return pipeline.Result{}
			}))
			step := ToNestedStep("super step", p, tt.givenPredicate)
			_ = step.F()
			assert.Equal(t, tt.expectedCounts, counter)
		})
	}
}

func truePredicate(counter *int) Predicate {
	return func(step pipeline.Step) bool {
		*counter++
		return true
	}
}

func falsePredicate(counter *int) Predicate {
	return func(step pipeline.Step) bool {
		*counter--
		return false
	}
}
