package parallel

import (
	"sync"

	pipeline "github.com/ccremer/go-command-pipeline"
)

// ConcurrentContext implements pipeline.Context by wrapping an existing pipeline.Context in a sync.Mutex.
type ConcurrentContext struct {
	WrappedContext pipeline.Context
	m              sync.Mutex
}

// NewConcurrentContext wraps the given pipeline.Context in a new ConcurrentContext.
func NewConcurrentContext(wrapped pipeline.Context) ConcurrentContext {
	return ConcurrentContext{
		WrappedContext: wrapped,
	}
}

// Value implements pipeline.Context:Value
func (ctx *ConcurrentContext) Value(key interface{}) interface{} {
	ctx.m.Lock()
	defer ctx.m.Unlock()
	return ctx.WrappedContext.Value(key)
}

// ValueOrDefault implements pipeline.Context:Value.
func (ctx *ConcurrentContext) ValueOrDefault(key interface{}, defaultValue interface{}) interface{} {
	ctx.m.Lock()
	defer ctx.m.Unlock()
	return ctx.WrappedContext.ValueOrDefault(key, defaultValue)
}

// StringValue implements pipeline.Context:StringValue.
func (ctx *ConcurrentContext) StringValue(key interface{}, defaultValue string) string {
	ctx.m.Lock()
	defer ctx.m.Unlock()
	return ctx.WrappedContext.StringValue(key, defaultValue)
}

// BoolValue implements pipeline.Context:BoolValue.
func (ctx *ConcurrentContext) BoolValue(key interface{}, defaultValue bool) bool {
	ctx.m.Lock()
	defer ctx.m.Unlock()
	return ctx.WrappedContext.BoolValue(key, defaultValue)
}

// IntValue implements pipeline.Context:IntValue.
func (ctx *ConcurrentContext) IntValue(key interface{}, defaultValue int) int {
	ctx.m.Lock()
	defer ctx.m.Unlock()
	return ctx.WrappedContext.IntValue(key, defaultValue)
}

// SetValue implements pipeline.Context:SetValue.
func (ctx *ConcurrentContext) SetValue(key interface{}, value interface{}) {
	ctx.m.Lock()
	defer ctx.m.Unlock()
	ctx.WrappedContext.SetValue(key, value)
}
