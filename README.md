# go-command-pipeline

![Go version](https://img.shields.io/github/go-mod/go-version/ccremer/go-command-pipeline)
[![Version](https://img.shields.io/github/v/release/ccremer/go-command-pipeline)][releases]
[![GitHub downloads](https://img.shields.io/github/downloads/ccremer/go-command-pipeline/total)][releases]
![Codecov](https://img.shields.io/codecov/c/github/ccremer/go-command-pipeline?token=XGOC4XUMJ5)

Small Go utility that executes high-level actions in a pipeline fashion

## Usage

```go
import pipeline "github.com/ccremer/go-command-pipeline"

func main() {
	p := pipeline.NewPipeline()
	p.WithSteps(
		pipeline.NewStepWithPredicate("clone repository", CloneGitRepository(), pipeline.Not(DirExists("my-repo"))),
		pipeline.NewStep("checkout branch", CheckoutBranch()),
		pipeline.NewStep("pull", Pull()),
	)
	result := p.Run()
	if !result.IsSuccessful() {
		log.Fatal(result.Err)
	}
}
```

[releases]: https://github.com/ccremer/go-command-pipeline/releases
