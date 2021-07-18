//+build examples

package git

import (
	"log"
	"os"
	"os/exec"

	pipeline "github.com/ccremer/go-command-pipeline"
	"github.com/ccremer/go-command-pipeline/predicate"
)

func main() {
	p := pipeline.NewPipeline()
	p.WithSteps(
		predicate.ToStep("clone repository", CloneGitRepository(), predicate.Not(DirExists("my-repo"))),
		pipeline.NewStep("checkout branch", CheckoutBranch()),
		pipeline.NewStep("pull", Pull()).WithResultHandler(logSuccess),
	)
	result := p.Run()
	if !result.IsSuccessful() {
		log.Fatal(result.Err)
	}
}

func logSuccess(result pipeline.Result) error {
	log.Println("handler called")
	return result.Err
}

func CloneGitRepository() pipeline.ActionFunc {
	return func() pipeline.Result {
		err := execGitCommand("clone", "git@github.com/ccremer/go-command-pipeline")
		if err != nil {
			return pipeline.Result{Err: err}
		}
		return pipeline.Result{}
	}
}

func Pull() pipeline.ActionFunc {
	return func() pipeline.Result {
		err := execGitCommand("pull")
		if err != nil {
			return pipeline.Result{Err: err}
		}
		return pipeline.Result{}
	}
}

func CheckoutBranch() pipeline.ActionFunc {
	return func() pipeline.Result {
		err := execGitCommand("checkout", "master")
		if err != nil {
			return pipeline.Result{Err: err}
		}
		return pipeline.Result{}
	}
}

func execGitCommand(args ...string) error {
	cmd := exec.Command("git", args...)
	return cmd.Run()
}

func DirExists(path string) predicate.Predicate {
	return func(step pipeline.Step) bool {
		if info, err := os.Stat(path); err != nil || !info.IsDir() {
			return false
		}
		return true
	}
}
