package pipeline

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

const stringKey = "stringKey"
const boolKey = "boolKey"
const intKey = "intKey"
const valueKey = "value"

func TestDefaultContext_Implements_Context(t *testing.T) {
	assert.Implements(t, (*Context)(nil), new(DefaultContext))
}

type valueTestCase struct {
	givenValues map[interface{}]interface{}

	defaultBool   bool
	defaultString string
	defaultInt    int

	expectedBool   bool
	expectedString string
	expectedInt    int
}

var valueTests = map[string]valueTestCase{
	"GivenNilValues_ThenExpectDefaults": {
		givenValues: nil,
	},
	"GivenNonExistingKey_ThenExpectDefaults": {
		givenValues:    map[interface{}]interface{}{},
		defaultBool:    true,
		expectedBool:   true,
		defaultString:  "default",
		expectedString: "default",
		defaultInt:     10,
		expectedInt:    10,
	},
	"GivenExistingKey_WhenInvalidType_ThenExpectDefaults": {
		givenValues: map[interface{}]interface{}{
			boolKey:   "invalid",
			stringKey: 0,
			intKey:    "invalid",
		},
		defaultBool:    true,
		expectedBool:   true,
		defaultString:  "default",
		expectedString: "default",
		defaultInt:     10,
		expectedInt:    10,
	},
	"GivenExistingKey_WhenValidType_ThenExpectValues": {
		givenValues: map[interface{}]interface{}{
			boolKey:   true,
			stringKey: "string",
			intKey:    10,
		},
		expectedBool:   true,
		expectedString: "string",
		expectedInt:    10,
	},
}

func TestDefaultContext_BoolValue(t *testing.T) {
	for name, tt := range valueTests {
		t.Run(name, func(t *testing.T) {
			ctx := DefaultContext{values: tt.givenValues}
			result := ctx.BoolValue(boolKey, tt.defaultBool)
			assert.Equal(t, tt.expectedBool, result)
		})
	}
}

func TestDefaultContext_StringValue(t *testing.T) {
	for name, tt := range valueTests {
		t.Run(name, func(t *testing.T) {
			ctx := DefaultContext{values: tt.givenValues}
			result := ctx.StringValue(stringKey, tt.defaultString)
			assert.Equal(t, tt.expectedString, result)
		})
	}
}

func TestDefaultContext_IntValue(t *testing.T) {
	for name, tt := range valueTests {
		t.Run(name, func(t *testing.T) {
			ctx := DefaultContext{values: tt.givenValues}
			result := ctx.IntValue(intKey, tt.defaultInt)
			assert.Equal(t, tt.expectedInt, result)
		})
	}
}

func TestDefaultContext_SetValue(t *testing.T) {
	ctx := DefaultContext{values: map[interface{}]interface{}{}}
	ctx.SetValue(stringKey, "string")
	assert.Equal(t, "string", ctx.values[stringKey])
}

func TestDefaultContext_Value(t *testing.T) {
	t.Run("GivenNilValues_ThenExpectNil", func(t *testing.T) {
		ctx := DefaultContext{values: nil}
		result := ctx.Value(valueKey)
		assert.Nil(t, result)
	})
	t.Run("GivenNonExistingKey_ThenExpectNil", func(t *testing.T) {
		ctx := DefaultContext{values: map[interface{}]interface{}{}}
		result := ctx.Value(valueKey)
		assert.Nil(t, result)
	})
	t.Run("GivenExistingKey_WhenKeyContainsNil_ThenExpectNil", func(t *testing.T) {
		ctx := DefaultContext{values: map[interface{}]interface{}{
			valueKey: nil,
		}}
		result := ctx.Value(valueKey)
		assert.Nil(t, result)
	})
}

func TestDefaultContext_ValueOrDefault(t *testing.T) {
	t.Run("GivenNilValues_ThenExpectDefault", func(t *testing.T) {
		ctx := DefaultContext{values: nil}
		result := ctx.ValueOrDefault(valueKey, valueKey)
		assert.Equal(t, result, valueKey)
	})
	t.Run("GivenNonExistingKey_ThenExpectDefault", func(t *testing.T) {
		ctx := DefaultContext{values: map[interface{}]interface{}{}}
		result := ctx.ValueOrDefault(valueKey, valueKey)
		assert.Equal(t, result, valueKey)
	})
	t.Run("GivenExistingKey_ThenExpectValue", func(t *testing.T) {
		ctx := DefaultContext{values: map[interface{}]interface{}{
			valueKey: valueKey,
		}}
		result := ctx.ValueOrDefault(valueKey, "default")
		assert.Equal(t, result, valueKey)
	})
}

func ExampleDefaultContext_BoolValue() {
	ctx := DefaultContext{}
	ctx.SetValue("key", true)
	fmt.Println(ctx.BoolValue("key", false))
	// Output: true
}

func ExampleDefaultContext_StringValue() {
	ctx := DefaultContext{}
	ctx.SetValue("key", "string")
	fmt.Println(ctx.StringValue("key", "default"))
	// Output: string
}

func ExampleDefaultContext_IntValue() {
	ctx := DefaultContext{}
	ctx.SetValue("key", 1)
	fmt.Println(ctx.IntValue("key", 0))
	// Output: 1
}
