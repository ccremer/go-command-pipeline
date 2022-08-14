package pipeline

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDependencyRecorder_ImplementsInterface(t *testing.T) {
	assert.Implements(t, (*DependencyResolver[context.Context])(nil), new(DependencyRecorder[context.Context]))
	assert.Implements(t, (*Recorder[context.Context])(nil), new(DependencyRecorder[context.Context]))
}

func TestDependencyRecorder_Record(t *testing.T) {
	recorder := DependencyRecorder[context.Context]{}
	step := NewStep("test", func(ctx context.Context) error {
		return nil
	})
	recorder.Record(step)
	assert.Equal(t, step.Name, recorder.Records[0].Name)
}

func TestDependencyRecorder_RequireByStepName(t *testing.T) {
	tests := map[string]struct {
		givenRecordedSteps     []Step[context.Context]
		givenRequiredStepNames []string
		expectedError          string
	}{
		"GivenNoStepNames_ThenReturnNil": {
			givenRecordedSteps:     []Step[context.Context]{newTestStep("step 1")},
			givenRequiredStepNames: []string{},
			expectedError:          "",
		},
		"GivenNoStepsRecorded_ThenReturnError": {
			givenRecordedSteps:     []Step[context.Context]{},
			givenRequiredStepNames: []string{"step 1"},
			expectedError:          "required steps did not run: [step 1]",
		},
		"GivenWrongStepsRecorded_ThenReturnError": {
			givenRecordedSteps:     []Step[context.Context]{newTestStep("another")},
			givenRequiredStepNames: []string{"step 1"},
			expectedError:          "required steps did not run: [step 1]",
		},
		"GivenStepsRecorded_WhenOneStepMissing_ThenReturnError": {
			givenRecordedSteps:     []Step[context.Context]{newTestStep("step 1")},
			givenRequiredStepNames: []string{"step 2"},
			expectedError:          "required steps did not run: [step 2]",
		},
		"GivenStepsRecorded_WhenMultipleStepsMissing_ThenReturnError": {
			givenRecordedSteps:     []Step[context.Context]{newTestStep("step 1")},
			givenRequiredStepNames: []string{"step 2", "step 3"},
			expectedError:          "required steps did not run: [step 2, step 3]",
		},
		"GivenStepsRecorded_WhenNoStepsMissing_ThenReturnNil": {
			givenRecordedSteps:     []Step[context.Context]{newTestStep("step 1"), newTestStep("step 2")},
			givenRequiredStepNames: []string{"step 2", "step 1"},
			expectedError:          "",
		},
		"GivenDuplicateStepsRecorded_WhenStepMissing_ThenReturnError": {
			givenRecordedSteps:     []Step[context.Context]{newTestStep("step 1"), newTestStep("step 1")},
			givenRequiredStepNames: []string{"step 2", "step 1"},
			expectedError:          "required steps did not run: [step 2]",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			recorder := DependencyRecorder[context.Context]{Records: tc.givenRecordedSteps}
			err := recorder.RequireDependencyByStepName(tc.givenRequiredStepNames...)
			if tc.expectedError != "" {
				assert.EqualError(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDependencyRecorder_RequireDependencyByFuncName(t *testing.T) {
	tests := map[string]struct {
		givenRecordedSteps   []Step[context.Context]
		givenRequiredActions []ActionFunc[context.Context]
		expectedError        string
	}{
		"GivenNoStepNames_ThenReturnNil": {
			givenRecordedSteps:   []Step[context.Context]{newTestStep("step 1")},
			givenRequiredActions: []ActionFunc[context.Context]{},
			expectedError:        "",
		},
		"GivenNoStepsRecorded_ThenReturnError": {
			givenRecordedSteps:   []Step[context.Context]{},
			givenRequiredActions: []ActionFunc[context.Context]{func(ctx context.Context) error { return nil }},
			expectedError:        "required steps did not run: [github.com/ccremer/go-command-pipeline.TestDependencyRecorder_RequireDependencyByFuncName.func1]",
		},
		"GivenWrongStepsRecorded_ThenReturnError": {
			givenRecordedSteps:   []Step[context.Context]{newTestStep("another")},
			givenRequiredActions: []ActionFunc[context.Context]{func(ctx context.Context) error { return nil }},
			expectedError:        "required steps did not run: [github.com/ccremer/go-command-pipeline.TestDependencyRecorder_RequireDependencyByFuncName.func2]",
		},
		"GivenStepsRecorded_WhenOneStepMissing_ThenReturnError": {
			givenRecordedSteps:   []Step[context.Context]{newTestStep("step 1")},
			givenRequiredActions: []ActionFunc[context.Context]{func(ctx context.Context) error { return nil }},
			expectedError:        "required steps did not run: [github.com/ccremer/go-command-pipeline.TestDependencyRecorder_RequireDependencyByFuncName.func3]",
		},
		"GivenStepsRecorded_WhenMultipleStepsMissing_ThenReturnError": {
			givenRecordedSteps:   []Step[context.Context]{newTestStep("step 1")},
			givenRequiredActions: []ActionFunc[context.Context]{func(ctx context.Context) error { return nil }, func(ctx context.Context) error { return nil }},
			expectedError:        "required steps did not run: [github.com/ccremer/go-command-pipeline.TestDependencyRecorder_RequireDependencyByFuncName.func4, github.com/ccremer/go-command-pipeline.TestDependencyRecorder_RequireDependencyByFuncName.func5]",
		},
		"GivenStepsRecorded_WhenNoStepsMissing_ThenReturnNil": {
			givenRecordedSteps:   []Step[context.Context]{newTestStep("step 1"), newTestStep("step 2")},
			givenRequiredActions: []ActionFunc[context.Context]{newTestStep("step 2").Action, newTestStep("step 1").Action},
			expectedError:        "",
		},
		"GivenDuplicateStepsRecorded_WhenStepMissing_ThenReturnError": {
			givenRecordedSteps:   []Step[context.Context]{newTestStep("step 1"), newTestStep("step 1")},
			givenRequiredActions: []ActionFunc[context.Context]{func(ctx context.Context) error { return nil }, newTestStep("step 1").Action},
			expectedError:        "required steps did not run: [github.com/ccremer/go-command-pipeline.TestDependencyRecorder_RequireDependencyByFuncName.func6]",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			recorder := DependencyRecorder[context.Context]{Records: tc.givenRecordedSteps}
			err := recorder.RequireDependencyByFuncName(tc.givenRequiredActions...)
			if tc.expectedError != "" {
				assert.EqualError(t, err, tc.expectedError)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDependencyRecorder_MustRequireDependencyByStepName(t *testing.T) {
	assert.PanicsWithError(t, "required steps did not run: [test]", func() {
		recorder := NewDependencyRecorder[context.Context]()
		recorder.MustRequireDependencyByStepName("test")
	})
	assert.NotPanics(t, func() {
		recorder := NewDependencyRecorder[context.Context]()
		recorder.Record(newTestStep("test"))
		recorder.MustRequireDependencyByStepName("test")
	})
}

func TestDependencyRecorder_MustRequireDependencyByFuncName(t *testing.T) {
	assert.PanicsWithError(t, "required steps did not run: [github.com/ccremer/go-command-pipeline.TestDependencyRecorder_MustRequireDependencyByFuncName.func1.1]", func() {
		recorder := NewDependencyRecorder[context.Context]()
		recorder.MustRequireDependencyByFuncName(func(_ context.Context) error { return nil })
	})
	assert.NotPanics(t, func() {
		recorder := NewDependencyRecorder[context.Context]()
		step := newTestStep("test")
		step.Action = func(_ context.Context) error { return nil }
		recorder.Record(step)
		recorder.MustRequireDependencyByFuncName(step.Action)
	})
}

func newTestStep(name string) Step[context.Context] {
	return NewStep[context.Context](name, func(_ context.Context) error {
		fmt.Println(name) // do something with the name to make functions between steps not the same
		return nil
	})
}

func ExampleDependencyRecorder_RequireDependencyByStepName() {
	recorder := NewDependencyRecorder[context.Context]()
	p := NewPipeline[context.Context]().WithBeforeHooks(recorder.Record)
	p.WithSteps(
		p.When(Bool[context.Context](false), "step 1", func(_ context.Context) error {
			fmt.Println("running step 1")
			return nil
		}),
		p.NewStep("step 2", func(_ context.Context) error {
			if err := recorder.RequireDependencyByStepName("step 1"); err != nil {
				return err
			}
			// this won't run
			fmt.Println("running step 2")
			return nil
		}),
	)
	err := p.RunWithContext(context.Background())
	fmt.Println(err)
	// Output:
	// step 'step 2' failed: required steps did not run: [step 1]
}

func ExampleDependencyRecorder_RequireDependencyByFuncName() {
	action := func(_ context.Context) error {
		fmt.Println("running step 1")
		return nil
	}
	recorder := NewDependencyRecorder[context.Context]()
	p := NewPipeline[context.Context]().WithBeforeHooks(recorder.Record)
	p.WithSteps(
		p.When(Bool[context.Context](false), "step 1", action),
		p.NewStep("step 2", func(_ context.Context) error {
			if err := recorder.RequireDependencyByFuncName(action); err != nil {
				return err
			}
			// this won't run
			fmt.Println("running step 2")
			return nil
		}),
	)
	err := p.RunWithContext(context.Background())
	fmt.Println(err)
	// Output:
	// step 'step 2' failed: required steps did not run: [github.com/ccremer/go-command-pipeline.ExampleDependencyRecorder_RequireDependencyByFuncName.func1]
}
