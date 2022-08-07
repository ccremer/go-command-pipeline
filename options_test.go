package pipeline

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPipeline_WithOptions(t *testing.T) {
	t.Run("DisableErrorWrapping", func(t *testing.T) {
		p := NewPipeline[*testContext]().WithOptions(Options{DisableErrorWrapping: true})
		p.WithSteps(
			NewStep[*testContext]("disabled error wrapping", func(_ *testContext) error {
				return errors.New("some error")
			}),
		)
		assert.True(t, p.options.DisableErrorWrapping)
		err := p.RunWithContext(&testContext{Context: context.Background()})
		require.Error(t, err)
		assert.Equal(t, "some error", err.Error())
	})
}
