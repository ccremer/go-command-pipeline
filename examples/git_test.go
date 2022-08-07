//go:build examples

package examples

import (
	"context"
	"errors"
	"log"
	"os"
	"os/exec"
	"testing"

	pipeline "github.com/ccremer/go-command-pipeline"
)

type GitContext struct {
	context.Context
}

func TestExample_Git(t *testing.T) {
	p := pipeline.NewPipeline[context.Context]()
	p.WithSteps(
		p.If(pipeline.Not(DirExists("my-repo")),
			"clone repository", CloneGitRepository(),
		),
		p.NewStep("checkout branch", CheckoutBranch()),
		p.NewStep("pull", Pull()).WithErrorHandler(logSuccess),
	)
	err := p.RunWithContext(&GitContext{context.Background()})
	if err != nil {
		t.Fatal(err)
	}
}

func logSuccess(_ context.Context, err error) error {
	if err != nil {
		var result pipeline.Result
		if errors.As(err, &result) {
			log.Println("handler called", result.Name())
		}
	}
	return err
}

func CloneGitRepository() pipeline.ActionFunc[context.Context] {
	return func(_ context.Context) error {
		err := execGitCommand("clone", "git@github.com/ccremer/go-command-pipeline")
		return err
	}
}

func CheckoutBranch() pipeline.ActionFunc[context.Context] {
	return func(_ context.Context) error {
		err := execGitCommand("checkout", "master")
		return err
	}
}

func Pull() pipeline.ActionFunc[context.Context] {
	return func(_ context.Context) error {
		err := execGitCommand("pull")
		return err
	}
}

func execGitCommand(args ...string) error {
	// replace 'echo' with actual 'git' binary
	cmd := exec.Command("echo", args...)
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	return err
}

func DirExists(path string) pipeline.Predicate[context.Context] {
	return func(_ context.Context) bool {
		if info, err := os.Stat(path); err != nil || !info.IsDir() {
			return false
		}
		return true
	}
}
