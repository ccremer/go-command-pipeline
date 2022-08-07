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
		p := NewPipeline[context.Context]().WithOptions(DisableErrorWrapping)
		p.WithSteps(
			NewStep[context.Context]("disabled error wrapping", func(_ context.Context) error {
				return errors.New("some error")
			}),
		)
		assert.True(t, p.options.disableErrorWrapping)
		err := p.RunWithContext(context.Background())
		require.Error(t, err)
		assert.Equal(t, "some error", err.Error())
	})
}
