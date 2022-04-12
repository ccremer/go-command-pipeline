package pipeline

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContext(t *testing.T) {
	tests := map[string]struct {
		givenKey      interface{}
		givenValue    interface{}
		expectedValue interface{}
		expectedFound bool
	}{
		"GivenNonExistentKey_ThenExpectNilAndFalse": {
			givenKey:      nil,
			expectedValue: nil,
		},
		"GivenKeyWithNilValue_ThenExpectNilAndTrue": {
			givenKey:      "key",
			givenValue:    nil,
			expectedValue: nil,
			expectedFound: true,
		},
		"GivenKeyWithValue_ThenExpectValueAndTrue": {
			givenKey:      "key",
			givenValue:    "value",
			expectedValue: "value",
			expectedFound: true,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctx := VariableContext(context.Background())
			if tc.givenKey != nil {
				AddToContext(ctx, tc.givenKey, tc.givenValue)
			}
			result, found := ValueFromContext(ctx, tc.givenKey)
			assert.Equal(t, tc.expectedValue, result, "value")
			assert.Equal(t, tc.expectedFound, found, "value found")
		})
	}
}

func TestContextPanics(t *testing.T) {
	assert.PanicsWithError(t, "context was not set up with VariableContext()", func() {
		AddToContext(context.Background(), "key", "value")
	}, "AddToContext")
	assert.PanicsWithError(t, "context was not set up with VariableContext()", func() {
		ValueFromContext(context.Background(), "key")
	}, "ValueFromContext")
}

func ExampleVariableContext() {
	ctx := VariableContext(context.Background())
	p := NewPipeline().WithSteps(
		NewStepFromFunc("store value", func(ctx context.Context) error {
			AddToContext(ctx, "key", "value")
			return nil
		}),
		NewStepFromFunc("retrieve value", func(ctx context.Context) error {
			value, _ := ValueFromContext(ctx, "key")
			fmt.Println(value)
			return nil
		}),
	)
	p.RunWithContext(ctx)
	// Output: value
}
