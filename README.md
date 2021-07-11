# go-command-pipeline

![Go version](https://img.shields.io/github/go-mod/go-version/ccremer/go-command-pipeline)
[![Version](https://img.shields.io/github/v/release/ccremer/go-command-pipeline)][releases]
[![GitHub downloads](https://img.shields.io/github/downloads/ccremer/go-command-pipeline/total)][releases]
[![Go Report Card](https://goreportcard.com/badge/github.com/ccremer/go-command-pipeline)][goreport]
[![Codecov](https://img.shields.io/codecov/c/github/ccremer/go-command-pipeline?token=XGOC4XUMJ5)][codecov]

Small Go utility that executes high-level actions in a pipeline fashion.
Especially useful when combined with the Facade design pattern.

## Usage

```go
import (
    pipeline "github.com/ccremer/go-command-pipeline"
    "github.com/ccremer/go-command-pipeline/predicate"
)

func main() {
	p := pipeline.NewPipeline()
	p.WithSteps(
		predicate.ToStep("clone repository", CloneGitRepository(), predicate.Not(DirExists("my-repo"))),
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
[codecov]: https://app.codecov.io/gh/ccremer/go-command-pipeline
[goreport]: https://goreportcard.com/report/github.com/ccremer/go-command-pipeline
