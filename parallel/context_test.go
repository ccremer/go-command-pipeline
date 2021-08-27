package parallel

import (
	"testing"

	pipeline "github.com/ccremer/go-command-pipeline"
	"github.com/stretchr/testify/assert"
)

func TestConcurrentContext_Implements_Context(t *testing.T) {
	assert.Implements(t, (*pipeline.Context)(nil), new(ConcurrentContext))
}

func TestConcurrentContext_Constructor(t *testing.T) {
	wrapped := pipeline.DefaultContext{}
	ctx := NewConcurrentContext(&wrapped)
	ctx.SetValue("key", "value")
	assert.Equal(t, "value", ctx.StringValue("key", "default"))
}
