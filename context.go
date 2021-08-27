package pipeline

// Context contains data relevant for the pipeline execution.
// It's primary purpose is to store and retrieve data within an ActionFunc.
type Context interface {
	// Value returns the raw value identified by key.
	// Returns nil if the key doesn't exist.
	Value(key interface{}) interface{}
	// ValueOrDefault returns the value identified by key if it exists.
	// If not, then the given default value is returned.
	ValueOrDefault(key interface{}, defaultValue interface{}) interface{}
	// StringValue is a sugared accessor like ValueOrDefault, but converts the value to string.
	// If the key cannot be found or if the value is not of type string, then the defaultValue is returned.
	StringValue(key interface{}, defaultValue string) string
	// BoolValue is a sugared accessor like ValueOrDefault, but converts the value to bool.
	// If the key cannot be found or if the value is not of type bool, then the defaultValue is returned.
	BoolValue(key interface{}, defaultValue bool) bool
	// IntValue is a sugared accessor like ValueOrDefault, but converts the value to int.
	// If the key cannot be found or if the value is not of type int, then the defaultValue is returned.
	IntValue(key interface{}, defaultValue int) int
	// SetValue sets the value at the given key.
	SetValue(key interface{}, value interface{})
}

// DefaultContext implements Context using a Map internally.
type DefaultContext struct {
	values map[interface{}]interface{}
}

// Value implements Context.Value.
func (ctx *DefaultContext) Value(key interface{}) interface{} {
	if ctx.values == nil {
		return nil
	}
	return ctx.values[key]
}

// ValueOrDefault implements Context.ValueOrDefault.
func (ctx *DefaultContext) ValueOrDefault(key interface{}, defaultValue interface{}) interface{} {
	if ctx.values == nil {
		return defaultValue
	}
	if raw, exists := ctx.values[key]; exists {
		return raw
	}
	return defaultValue
}

// StringValue implements Context.StringValue.
func (ctx *DefaultContext) StringValue(key interface{}, defaultValue string) string {
	if ctx.values == nil {
		return defaultValue
	}
	raw, exists := ctx.values[key]
	if !exists {
		return defaultValue
	}
	if strValue, isString := raw.(string); isString {
		return strValue
	}
	return defaultValue
}

// BoolValue implements Context.BoolValue.
func (ctx *DefaultContext) BoolValue(key interface{}, defaultValue bool) bool {
	if ctx.values == nil {
		return defaultValue
	}
	raw, exists := ctx.values[key]
	if !exists {
		return defaultValue
	}
	if boolValue, isBool := raw.(bool); isBool {
		return boolValue
	}
	return defaultValue
}

// IntValue implements Context.IntValue.
func (ctx *DefaultContext) IntValue(key interface{}, defaultValue int) int {
	if ctx.values == nil {
		return defaultValue
	}
	raw, exists := ctx.values[key]
	if !exists {
		return defaultValue
	}
	if intValue, isInt := raw.(int); isInt {
		return intValue
	}
	return defaultValue
}

// SetValue implements Context.SetValue.
func (ctx *DefaultContext) SetValue(key interface{}, value interface{}) {
	if ctx.values == nil {
		ctx.values = map[interface{}]interface{}{}
	}
	ctx.values[key] = value
}
