package pipeline

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
	"strings"
)

// Recorder Records the steps executed in a pipeline.
type Recorder[T context.Context] interface {
	// Record adds the step to the execution Records.
	Record(step Step[T])
}

// DependencyResolver provides means to query if a pipeline Step is satisfied as a dependency for another Step.
// It is used together with Recorder.
type DependencyResolver[T context.Context] interface {
	Recorder[T]
	// RequireDependencyByStepName checks if any of the given step names are present in the Records.
	// It returns nil if all given step names are in the Records in any order.
	RequireDependencyByStepName(stepNames ...string) error
	// MustRequireDependencyByStepName is RequireDependencyByStepName but any non-nil errors result in a panic.
	MustRequireDependencyByStepName(stepNames ...string)
	// RequireDependencyByFuncName checks if any of the given action functions are present in the Records.
	// It returns nil if all given functions are in the Records in any order.
	// Since functions aren't comparable for equality, the resolver attempts to compare them by name through reflection.
	RequireDependencyByFuncName(actions ...ActionFunc[T]) error
	// MustRequireDependencyByFuncName is RequireDependencyByFuncName but any non-nil errors result in a panic.
	MustRequireDependencyByFuncName(actions ...ActionFunc[T])
}

// DependencyRecorder is a Recorder and DependencyResolver that tracks each Step executed and can be used to query if certain steps are in the Records.
type DependencyRecorder[T context.Context] struct {
	// Records contains a slice of Steps that were run.
	// It contains also the last Step that failed with an error.
	Records []Step[T]
}

// NewDependencyRecorder returns a new instance of DependencyRecorder.
func NewDependencyRecorder[T context.Context]() *DependencyRecorder[T] {
	return &DependencyRecorder[T]{Records: []Step[T]{}}
}

// Record implements Recorder.
func (s *DependencyRecorder[T]) Record(step Step[T]) {
	s.Records = append(s.Records, step)
}

// RequireDependencyByStepName implements DependencyResolver.RequireDependencyByStepName.
// A DependencyError is returned with a list of names that aren't in the Records.
// Steps that share the same name are not distinguishable.
func (s *DependencyRecorder[T]) RequireDependencyByStepName(stepNames ...string) error {
	if len(stepNames) == 0 {
		return nil
	}
	missing := make([]string, 0)
	for _, desiredName := range stepNames {
		found := false
		for _, step := range s.Records {
			if step.Name == desiredName {
				found = true
				break
			}
		}
		if !found {
			missing = append(missing, desiredName)
		}
	}
	if len(missing) == 0 {
		return nil
	}
	return fmt.Errorf("%w", &DependencyError{MissingSteps: missing})
}

// MustRequireDependencyByStepName implements DependencyResolver.MustRequireDependencyByStepName.
func (s *DependencyRecorder[T]) MustRequireDependencyByStepName(stepNames ...string) {
	err := s.RequireDependencyByStepName(stepNames...)
	if err != nil {
		panic(err)
	}
}

// RequireDependencyByFuncName implements DependencyResolver.RequireDependencyByFuncName.
//
// Direct function pointers can easily be compared:
//
//  func myFunc(ctx context.Context) error {
//    return nil
//  }
//  ...
//  pipe.AddStep("test", myFunc)
//  ...
//  recorder.RequireDependencyByFuncName(myFunc)
//
// Note that you may experience unexpected behaviour when dealing with generative functions.
// For example, the following snippet will not work, since the function names from 2 different call locations are different:
//
//  generateFunc() func(ctx context.Context) error {
//    return func(ctx context.Context) error {
//      return nil
//    }
//  }
//  ...
//  pipe.AddStep("test", generateFunc())
//  ...
//  recorder.RequireDependencyByFuncName(generateFunc()) // will end in an error
//
// As an alternative, you may store the generated function in a variable that is accessible from multiple locations:
//
//  var genFunc = generateFunc()
//  ...
//  pipe.AddStep("test", genFunc())
//  ...
//  recorder.RequireDependencyByFuncName(genFunc()) // works
func (s *DependencyRecorder[T]) RequireDependencyByFuncName(actions ...ActionFunc[T]) error {
	if len(actions) == 0 {
		return nil
	}
	missing := make([]string, 0)
	for _, desiredAction := range actions {
		found := false
		desiredActionName := getFunctionName(desiredAction)
		for _, step := range s.Records {
			actionName := getFunctionName(step.Action)
			if actionName == desiredActionName {
				found = true
				break
			}
		}
		if !found {
			missing = append(missing, desiredActionName)
		}
	}
	if len(missing) == 0 {
		return nil
	}
	return fmt.Errorf("%w", &DependencyError{MissingSteps: missing})
}

// MustRequireDependencyByFuncName implements DependencyResolver.MustRequireDependencyByFuncName.
func (s *DependencyRecorder[T]) MustRequireDependencyByFuncName(actions ...ActionFunc[T]) {
	err := s.RequireDependencyByFuncName(actions...)
	if err != nil {
		panic(err)
	}
}

func getFunctionName(temp interface{}) string {
	value := reflect.ValueOf(temp)
	if value.Kind() != reflect.Func {
		panic(fmt.Errorf("given value is not a function: %v", temp))
	}
	strs := runtime.FuncForPC(value.Pointer()).Name()
	return strs
}

// DependencyError is an error that indicates which steps did not satisfy dependency requirements.
type DependencyError struct {
	// MissingSteps returns a slice of Step or ActionFunc names.
	MissingSteps []string
}

// Error returns a stringed list of steps that did not run either by Step or ActionFunc name.
func (d *DependencyError) Error() string {
	joined := strings.Join(d.MissingSteps, ", ")
	return fmt.Sprintf("required steps did not run: [%s]", joined)
}
