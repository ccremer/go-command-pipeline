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
			ctx := MutableContext(context.Background())
			if tc.givenKey != nil {
				StoreInContext(ctx, tc.givenKey, tc.givenValue)
			}
			result, found := LoadFromContext(ctx, tc.givenKey)
			assert.Equal(t, tc.expectedValue, result, "value")
			assert.Equal(t, tc.expectedFound, found, "value found")
		})
	}
}

func TestContextPanics(t *testing.T) {
	assert.PanicsWithError(t, "context was not set up with MutableContext()", func() {
		StoreInContext(context.Background(), "key", "value")
	}, "StoreInContext")
	assert.PanicsWithError(t, "context was not set up with MutableContext()", func() {
		LoadFromContext(context.Background(), "key")
	}, "LoadFromContext")
}

func TestMutableContextRepeated(t *testing.T) {
	parent := context.Background()
	result := MutableContext(parent)
	assert.NotEqual(t, parent, result)
	repeated := MutableContext(result)
	assert.Equal(t, result, repeated)
}

func TestMustLoadFromContext(t *testing.T) {
	t.Run("KeyExistsWithNil", func(t *testing.T) {
		ctx := MutableContext(context.Background())
		StoreInContext(ctx, "key", nil)
		result := MustLoadFromContext(ctx, "key")
		assert.Nil(t, result)
	})
	t.Run("KeyDoesntExist", func(t *testing.T) {
		assert.PanicsWithError(t, `key "key" was not found in context`, func() {
			ctx := MutableContext(context.Background())
			_ = MustLoadFromContext(ctx, "key")
		})
	})
	t.Run("KeyExistsWithValue", func(t *testing.T) {
		ctx := MutableContext(context.Background())
		StoreInContext(ctx, "key", "value")
		result := MustLoadFromContext(ctx, "key")
		assert.Equal(t, "value", result)
	})
}

func TestLoadFromContextOrDefault(t *testing.T) {
	t.Run("KeyExists", func(t *testing.T) {
		ctx := MutableContext(context.Background())
		StoreInContext(ctx, "key", "value")
		result := LoadFromContextOrDefault(ctx, "key", "default")
		assert.Equal(t, "value", result)
	})
	t.Run("KeyDoesntExist", func(t *testing.T) {
		ctx := MutableContext(context.Background())
		result := LoadFromContextOrDefault(ctx, "key", "default")
		assert.Equal(t, "default", result)
	})
}

func ExampleMutableContext() {
	type key struct{}

	ctx := MutableContext(context.Background())
	p := NewPipeline[context.Context]().WithSteps(
		NewStep("store value", func(ctx context.Context) error {
			StoreInContext(ctx, key{}, "value")
			return nil
		}),
		NewStep("retrieve value", func(ctx context.Context) error {
			value, _ := LoadFromContext(ctx, key{})
			fmt.Println(value)
			return nil
		}),
	)
	_ = p.RunWithContext(ctx)
	// Output: value
}
