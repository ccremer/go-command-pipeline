package pipeline

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_Predicates(t *testing.T) {
	counter := 0
	tests := map[string]struct {
		givenPredicate Predicate[context.Context]
		expectedCounts int
	}{
		"GivenBoolPredicate_WhenRunning_ThenInvoke": {
			givenPredicate: Bool[context.Context](true),
			expectedCounts: 1,
		},
		"GivenAnd_WhenBothFalse_ThenIgnoreSecondPredicate": {
			givenPredicate: And[context.Context](falsePredicate(&counter), falsePredicate(&counter)),
			expectedCounts: -1,
		},
		"GivenAndPredicate_WhenFirstTrue_ThenIgnore": {
			givenPredicate: And[context.Context](truePredicate(&counter), falsePredicate(&counter)),
			expectedCounts: 0,
		},
		"GivenAndPredicate_WhenBothTrue_ThenRunAction": {
			givenPredicate: And[context.Context](truePredicate(&counter), truePredicate(&counter)),
			expectedCounts: 3,
		},
		"GivenNotPredicate_WhenFalse_ThenRunAction": {
			givenPredicate: Not[context.Context](falsePredicate(&counter)),
			expectedCounts: 0,
		},
		"GivenNotPredicate_WhenTrue_ThenIgnoreAction": {
			givenPredicate: Not[context.Context](truePredicate(&counter)),
			expectedCounts: 1,
		},
		"GivenOrPredicate_WhenBothFalse_ThenIgnoreAction": {
			givenPredicate: Or[context.Context](falsePredicate(&counter), falsePredicate(&counter)),
			expectedCounts: -2,
		},
		"GivenOrPredicate_WhenFirstTrue_ThenRunAction": {
			givenPredicate: Or[context.Context](truePredicate(&counter), falsePredicate(&counter)),
			expectedCounts: 2,
		},
		"GivenOrPredicate_WhenSecondTrue_ThenRunAction": {
			givenPredicate: Or[context.Context](falsePredicate(&counter), truePredicate(&counter)),
			expectedCounts: 1,
		},
		"GivenOrPredicate_WhenBothTrue_ThenRunActionAfterFirstPredicate": {
			givenPredicate: Or[context.Context](truePredicate(&counter), truePredicate(&counter)),
			expectedCounts: 2,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			counter = 0
			step := ToStep("name", func(_ context.Context) error {
				counter += 1
				return nil
			}, tt.givenPredicate)
			result := step.Action(nil)
			assert.Equal(t, tt.expectedCounts, counter)
			assert.NoError(t, result)
		})
	}
}

func TestToNestedStep(t *testing.T) {
	counter := 0
	tests := map[string]struct {
		givenPredicate Predicate[context.Context]
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
			p := NewPipeline[context.Context]().AddStep(NewStep("nested step", func(_ context.Context) error {
				counter++
				return nil
			}))
			step := ToNestedStep("super step", tt.givenPredicate, p)
			_ = step.Action(context.Background())
			assert.Equal(t, tt.expectedCounts, counter)
		})
	}
}

func TestIf(t *testing.T) {
	counter := 0
	tests := map[string]struct {
		givenPredicate Predicate[context.Context]
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
			step := NewStep("step", func(_ context.Context) error {
				counter++
				return nil
			})
			wrapped := If(tt.givenPredicate, step)
			result := wrapped.Action(nil)
			require.NoError(t, result)
			assert.Equal(t, tt.expectedCalls, counter)
			assert.Equal(t, step.Name, wrapped.Name)
		})
	}
}

func TestIfOrElse(t *testing.T) {
	counter := 0
	tests := map[string]struct {
		givenPredicate Predicate[context.Context]
		expectedCalls  int
	}{
		"GivenWrappedStep_WhenPredicateEvalsTrue_ThenRunMainAction": {
			givenPredicate: truePredicate(&counter),
			expectedCalls:  2,
		},
		"GivenWrappedStep_WhenPredicateEvalsFalse_ThenRunAlternativeAction": {
			givenPredicate: falsePredicate(&counter),
			expectedCalls:  -2,
		},
	}

	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			counter = 0
			trueStep := NewStep("true", func(_ context.Context) error {
				counter++
				return nil
			})
			falseStep := NewStep("false", func(ctx context.Context) error {
				counter--
				return nil
			})
			wrapped := IfOrElse(tt.givenPredicate, trueStep, falseStep)
			result := wrapped.Action(nil)
			require.NoError(t, result)
			assert.Equal(t, tt.expectedCalls, counter)
			assert.Equal(t, trueStep.Name, wrapped.Name)
		})
	}
}
func TestBoolPtr(t *testing.T) {
	called := false
	b := false
	p := NewPipeline[context.Context]()
	p.WithSteps(
		p.If(BoolPtr[context.Context](&b), "boolptr", func(_ context.Context) error {
			called = true
			return nil
		}),
	)
	b = true
	_ = p.RunWithContext(context.Background())
	assert.True(t, called)
}

func truePredicate(counter *int) Predicate[context.Context] {
	return func(_ context.Context) bool {
		*counter++
		return true
	}
}

func falsePredicate(counter *int) Predicate[context.Context] {
	return func(_ context.Context) bool {
		*counter--
		return false
	}
}
