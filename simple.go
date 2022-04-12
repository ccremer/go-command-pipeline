package pipeline

import (
	"context"
	"reflect"
	"runtime"
	"strings"
)

// NewAnonymous creates a pipeline from the given anonymous functions.
// This is meant to quickly set up a Pipeline that executes error-prone functions in a fail-first fashion.
// The context allows to carry over values.
//
// Note: You may want to use Pipeline.WithOptions(DisableErrorWrapping) to suppress function names as step names in errors.
func NewAnonymous(funcs ...func(ctx context.Context) error) *Pipeline {
	steps := make([]Step, len(funcs))
	for i := 0; i < len(funcs); i++ {
		fn := funcs[i]
		steps[i] = NewStepFromFunc(getFunctionName(fn), func(ctx context.Context) error {
			return fn(ctx)
		})
	}
	return NewPipeline().WithSteps(steps...)
}

func getFunctionName(temp interface{}) string {
	strs := strings.Split(runtime.FuncForPC(reflect.ValueOf(temp).Pointer()).Name(), ".")
	return strs[len(strs)-1]
}
