package pipeline

import (
	"context"
	"fmt"
	"sync"
)

type contextKey struct{}

// MutableContext adds a map to the given context that can be used to store mutable values in the context.
// It uses sync.Map under the hood.
// Repeated calls to MutableContext with the same parent has no effect and returns the same context.
//
// See also StoreInContext and LoadFromContext.
func MutableContext(parent context.Context) context.Context {
	if parent.Value(contextKey{}) == nil {
		return context.WithValue(parent, contextKey{}, &sync.Map{})
	}
	return parent
}

// StoreInContext adds the given key and value to ctx.
// Any keys or values added during pipeline execution is available in the next steps, provided the pipeline runs synchronously.
// In parallel executed pipelines you may encounter race conditions.
// Use LoadFromContext to retrieve values.
//
// Note: This method is thread-safe, but panics if ctx has not been set up with MutableContext first.
func StoreInContext(ctx context.Context, key, value interface{}) {
	m := ctx.Value(contextKey{})
	if m == nil {
		panic(fmt.Errorf("context was not set up with MutableContext()"))
	}
	m.(*sync.Map).Store(key, value)
}

// LoadFromContext returns the value from the given context with the given key.
// It returns the value and true, or nil and false if the key doesn't exist.
// It returns nil and true if the key exists and the value actually is nil.
// Use StoreInContext to store values.
//
// Note: This method is thread-safe, but panics if the ctx has not been set up with MutableContext first.
func LoadFromContext(ctx context.Context, key interface{}) (interface{}, bool) {
	m := ctx.Value(contextKey{})
	if m == nil {
		panic(fmt.Errorf("context was not set up with MutableContext()"))
	}
	mp := m.(*sync.Map)
	val, found := mp.Load(key)
	return val, found
}

// MustLoadFromContext is similar to LoadFromContext, except it doesn't return a bool to indicate whether the key exists.
// It panics if the key doesn't exist.
// Use StoreInContext to store values.
//
// Note: This method is thread-safe, but panics if the ctx has not been set up with MutableContext first.
func MustLoadFromContext(ctx context.Context, key interface{}) interface{} {
	val, found := LoadFromContext(ctx, key)
	if !found {
		panic(fmt.Errorf("key %q was not found in context", key))
	}
	return val
}
