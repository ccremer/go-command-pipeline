package pipeline

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPipeline_WithOptions(t *testing.T) {
	t.Run("DisableErrorWrapping", func(t *testing.T) {
		p := NewPipeline().WithOptions(DisableErrorWrapping)
		p.WithSteps(
			NewStepFromFunc("disabled error wrapping", func(_ Context) error {
				return errors.New("some error")
			}),
		)
		assert.True(t, p.disableErrorWrapping)
		result := p.Run()
		assert.Equal(t, "some error", result.Err.Error())
	})
}
