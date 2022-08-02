package pipeline

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPipeline_WithOptions(t *testing.T) {
	t.Run("DisableErrorWrapping", func(t *testing.T) {
		p := NewPipeline[*PipelineContext]().WithOptions(DisableErrorWrapping)
		p.WithSteps(
			NewStep[*PipelineContext]("disabled error wrapping", func(_ *PipelineContext) error {
				return errors.New("some error")
			}),
		)
		assert.True(t, p.options.disableErrorWrapping)
		result := p.RunWithContext(&PipelineContext{Context: context.Background()})
		assert.Equal(t, "some error", result.Err().Error())
	})
}
