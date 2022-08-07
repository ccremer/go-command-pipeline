package pipeline

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_Predicates(t *testing.T) {
	counter := 0
	tests := map[string]struct {
		givenPredicate Predicate[context.Context]
		expectedCounts int
		expectedResult bool
	}{
		"GivenBoolPredicate_WhenRunning_ThenExpectTrue": {
			givenPredicate: Bool[context.Context](true),
			expectedCounts: 0,
			expectedResult: true,
		},
		"GivenAnd_WhenBothFalse_ThenExpectFalseAndIgnoreSecondPredicate": {
			givenPredicate: And[context.Context](falsePredicate(&counter), falsePredicate(&counter)),
			expectedCounts: -1,
			expectedResult: false,
		},
		"GivenAndPredicate_WhenSecondFalse_ThenExpectFalse": {
			givenPredicate: And[context.Context](truePredicate(&counter), falsePredicate(&counter)),
			expectedCounts: 0,
			expectedResult: false,
		},
		"GivenAndPredicate_WhenBothTrue_ThenExpectTrue": {
			givenPredicate: And[context.Context](truePredicate(&counter), truePredicate(&counter)),
			expectedCounts: 2,
			expectedResult: true,
		},
		"GivenNotPredicate_WhenFalse_ThenExpectTrue": {
			givenPredicate: Not[context.Context](falsePredicate(&counter)),
			expectedCounts: -1,
			expectedResult: true,
		},
		"GivenNotPredicate_WhenTrue_ThenExpectFalse": {
			givenPredicate: Not[context.Context](truePredicate(&counter)),
			expectedCounts: 1,
			expectedResult: false,
		},
		"GivenOrPredicate_WhenBothFalse_ThenExpectFalse": {
			givenPredicate: Or[context.Context](falsePredicate(&counter), falsePredicate(&counter)),
			expectedCounts: -2,
			expectedResult: false,
		},
		"GivenOrPredicate_WhenFirstTrue_ThenExpectTrue": {
			givenPredicate: Or[context.Context](truePredicate(&counter), falsePredicate(&counter)),
			expectedCounts: 1,
			expectedResult: true,
		},
		"GivenOrPredicate_WhenSecondTrue_ThenExpectTrue": {
			givenPredicate: Or[context.Context](falsePredicate(&counter), truePredicate(&counter)),
			expectedCounts: 0,
			expectedResult: true,
		},
		"GivenOrPredicate_WhenBothTrue_ThenExpectTrueAfterFirstPredicate": {
			givenPredicate: Or[context.Context](truePredicate(&counter), truePredicate(&counter)),
			expectedCounts: 1,
			expectedResult: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			counter = 0
			step := Step[context.Context]{Condition: tt.givenPredicate}
			result := step.Condition(nil)
			assert.Equal(t, tt.expectedCounts, counter)
			assert.Equal(t, tt.expectedResult, result)
		})
	}
}

func TestBoolPtr(t *testing.T) {
	called := false
	b := false
	p := NewPipeline[context.Context]()
	p.WithSteps(
		p.When(BoolPtr[context.Context](&b), "boolptr", func(_ context.Context) error {
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
