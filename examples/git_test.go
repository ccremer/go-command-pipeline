//go:build examples
// +build examples

package examples

import (
	"context"
	"log"
	"os"
	"os/exec"
	"testing"

	pipeline "github.com/ccremer/go-command-pipeline"
)

func TestExample_Git(t *testing.T) {
	p := pipeline.NewPipeline()
	p.WithSteps(
		CloneGitRepository(),
		CheckoutBranch(),
		Pull().WithResultHandler(logSuccess),
	)
	result := p.Run()
	if !result.IsSuccessful() {
		t.Fatal(result.Err())
	}
}

func logSuccess(_ context.Context, result pipeline.Result) error {
	log.Println("handler called", result.Name())
	return result.Err()
}

func CloneGitRepository() pipeline.Step {
	return pipeline.ToStep("clone repository", func(_ context.Context) error {
		err := execGitCommand("clone", "git@github.com/ccremer/go-command-pipeline")
		return err
	}, pipeline.Not(DirExists("my-repo")))
}

func Pull() pipeline.Step {
	return pipeline.NewStepFromFunc("pull", func(_ context.Context) error {
		err := execGitCommand("pull")
		return err
	})
}

func CheckoutBranch() pipeline.Step {
	return pipeline.NewStepFromFunc("checkout branch", func(_ context.Context) error {
		err := execGitCommand("checkout", "master")
		return err
	})
}

func execGitCommand(args ...string) error {
	// replace 'echo' with actual 'git' binary
	cmd := exec.Command("echo", args...)
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	return err
}

func DirExists(path string) pipeline.Predicate {
	return func(_ context.Context) bool {
		if info, err := os.Stat(path); err != nil || !info.IsDir() {
			return false
		}
		return true
	}
}
