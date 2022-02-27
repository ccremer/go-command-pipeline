package pipeline

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStep_WithErrorHandler(t *testing.T) {
	tests := map[string]struct {
		givenError        error
		expectedExecution bool
	}{
		"GivenHandler_WhenErrorIsNil_ThenDoNotRunHandler": {
			givenError:        nil,
			expectedExecution: false,
		},
		"GivenHandler_WhenErrorGiven_ThenExecuteHandler": {
			givenError:        errors.New("error"),
			expectedExecution: true,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			executed := false
			s := NewStepFromFunc("test", func(_ context.Context) error {
				return nil
			}).WithErrorHandler(func(_ context.Context, err error) error {
				executed = true
				return err
			})
			err := s.H(nil, Result{err: tt.givenError})
			assert.Equal(t, tt.givenError, err)
			assert.Equal(t, tt.expectedExecution, executed)
		})
	}
}
