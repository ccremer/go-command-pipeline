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
		pipeline.ToStep("clone repository", CloneGitRepository(), pipeline.Not(DirExists("my-repo"))),
		pipeline.NewStep("checkout branch", CheckoutBranch()),
		pipeline.NewStep("pull", Pull()).WithResultHandler(logSuccess),
	)
	result := p.Run()
	if !result.IsSuccessful() {
		t.Fatal(result.Err())
	}
}

func logSuccess(_ context.Context, result pipeline.Result) error {
	log.Println("handler called")
	return result.Err()
}

func CloneGitRepository() pipeline.ActionFunc {
	return func(_ context.Context) pipeline.Result {
		err := execGitCommand("clone", "git@github.com/ccremer/go-command-pipeline")
		return pipeline.NewResultWithError("clone repository", err)
	}
}

func Pull() pipeline.ActionFunc {
	return func(_ context.Context) pipeline.Result {
		err := execGitCommand("pull")
		return pipeline.NewResultWithError("pull", err)
	}
}

func CheckoutBranch() pipeline.ActionFunc {
	return func(_ context.Context) pipeline.Result {
		err := execGitCommand("checkout", "master")
		return pipeline.NewResultWithError("checkout branch", err)
	}
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
