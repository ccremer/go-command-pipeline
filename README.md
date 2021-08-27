# go-command-pipeline

![Go version](https://img.shields.io/github/go-mod/go-version/ccremer/go-command-pipeline)
[![Version](https://img.shields.io/github/v/release/ccremer/go-command-pipeline)][releases]
[![Go Report Card](https://goreportcard.com/badge/github.com/ccremer/go-command-pipeline)][goreport]
[![Codecov](https://img.shields.io/codecov/c/github/ccremer/go-command-pipeline?token=XGOC4XUMJ5)][codecov]

Small Go utility that executes business actions in a pipeline.

## Usage

```go
import (
    pipeline "github.com/ccremer/go-command-pipeline"
    "github.com/ccremer/go-command-pipeline/predicate"
)

func main() {
	p := pipeline.NewPipeline()
	p.WithSteps(
		pipeline.NewStep("define random number", defineNumber),
		pipeline.NewStepFromFunc("print number", printNumber),
	)
	result := p.Run()
	if !result.IsSuccessful() {
		log.Fatal(result.Err)
	}
}

func defineNumber(ctx pipeline.Context) pipeline.Result {
	ctx.SetValue("number", rand.Int())
	return pipeline.Result{}
}

// Let's assume this is a business function that can fail.
// You can enable "automatic" fail-on-first-error pipelines by having more small functions that return errors.
func printNumber(ctx pipeline.Context) error {
	_, err := fmt.Println(ctx.IntValue("number", 0))
	return err
}
```

## Who is it for

This utility is interesting for you if you have many business functions that are executed sequentially, each with their own error handling.
Do you grow tired of the tedious error handling in Go when all you do is passing the error "up" in the stack in over 90% of the cases, only to log it at the root?
This utility helps you focus on the business logic by dividing each failure-prone action into small steps since pipeline aborts on first error.

Consider the following prose example:
```go
func Persist(data Data) error {
    err := database.prepareTransaction()
    if err != nil {
        return err
    }
    err = database.executeQuery("SOME QUERY", data)
    if err != nil {
        return err
    }
    err = database.commit()
    return err
}
```
We have tons of `if err != nil` that bloats the function with more error handling than actual interesting business logic.

It could be simplified to something like this:
```go
func Persist(data Data) error {
    p := pipeline.NewPipeline().WithSteps(
        pipeline.NewStep("prepareTransaction", prepareTransaction()),
        pipeline.NewStep("executeQuery", executeQuery(data)),
        pipeline.NewStep("commitTransaction", commit()),
    )
    return p.Run().Err
}

func executeQuery(data Data) pipeline.ActionFunc {
	return func(_ pipeline.Context) pipeline.Result {
        err := database.executeQuery("SOME QUERY", data)
        return pipeline.Result{Err: err}
	}
}
...
```
While it seems to add more lines in order to set up a pipeline, it makes it very easily understandable what `Persist()` does without all the error handling.

[releases]: https://github.com/ccremer/go-command-pipeline/releases
[codecov]: https://app.codecov.io/gh/ccremer/go-command-pipeline
[goreport]: https://goreportcard.com/report/github.com/ccremer/go-command-pipeline
