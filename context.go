package pipeline

import (
	"context"
	"errors"
	"sync"
)

type contextKey struct{}

// VariableContext adds a map to the given context that can be used to store intermediate values in the context.
// It uses sync.Map under the hood.
//
// See also AddToContext() and ValueFromContext.
func VariableContext(parent context.Context) context.Context {
	return context.WithValue(parent, contextKey{}, &sync.Map{})
}

// AddToContext adds the given key and value to ctx.
// Any keys or values added during pipeline execution is available in the next steps, provided the pipeline runs synchronously.
// In parallel executed pipelines you may encounter race conditions.
// Use ValueFromContext to retrieve values.
//
// Note: This method is thread-safe, but panics if ctx has not been set up with VariableContext first.
func AddToContext(ctx context.Context, key, value interface{}) {
	m := ctx.Value(contextKey{})
	if m == nil {
		panic(errors.New("context was not set up with VariableContext()"))
	}
	m.(*sync.Map).Store(key, value)
}

// ValueFromContext returns the value from the given context with the given key.
// It returns the value and true, or nil and false if the key doesn't exist.
// It may return nil and true if the key exists, but the value actually is nil.
// Use AddToContext to store values.
//
// Note: This method is thread-safe, but panics if the ctx has not been set up with VariableContext first.
func ValueFromContext(ctx context.Context, key interface{}) (interface{}, bool) {
	m := ctx.Value(contextKey{})
	if m == nil {
		panic(errors.New("context was not set up with VariableContext()"))
	}
	mp := m.(*sync.Map)
	val, found := mp.Load(key)
	return val, found
}
