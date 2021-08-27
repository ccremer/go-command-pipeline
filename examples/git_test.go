//+build examples

package examples

import (
	"log"
	"os"
	"os/exec"
	"testing"

	pipeline "github.com/ccremer/go-command-pipeline"
	"github.com/ccremer/go-command-pipeline/predicate"
)

func TestExample_Git(t *testing.T) {
	p := pipeline.NewPipeline()
	p.WithSteps(
		predicate.ToStep("clone repository", CloneGitRepository(), predicate.Not(DirExists("my-repo"))),
		pipeline.NewStep("checkout branch", CheckoutBranch()),
		pipeline.NewStep("pull", Pull()).WithResultHandler(logSuccess),
	)
	result := p.Run()
	if !result.IsSuccessful() {
		t.Fatal(result.Err)
	}
}

func logSuccess(result pipeline.Result) error {
	log.Println("handler called")
	return result.Err
}

func CloneGitRepository() pipeline.ActionFunc {
	return func(_ pipeline.Context) pipeline.Result {
		err := execGitCommand("clone", "git@github.com/ccremer/go-command-pipeline")
		return pipeline.Result{Err: err}
	}
}

func Pull() pipeline.ActionFunc {
	return func(_ pipeline.Context) pipeline.Result {
		err := execGitCommand("pull")
		return pipeline.Result{Err: err}
	}
}

func CheckoutBranch() pipeline.ActionFunc {
	return func(_ pipeline.Context) pipeline.Result {
		err := execGitCommand("checkout", "master")
		return pipeline.Result{Err: err}
	}
}

func execGitCommand(args ...string) error {
	// replace 'echo' with actual 'git' binary
	cmd := exec.Command("echo", args...)
	cmd.Stdout = os.Stdout
	err := cmd.Run()
	return err
}

func DirExists(path string) predicate.Predicate {
	return func(step pipeline.Step) bool {
		if info, err := os.Stat(path); err != nil || !info.IsDir() {
			return false
		}
		return true
	}
}
