# go-command-pipeline

![Go version](https://img.shields.io/github/go-mod/go-version/ccremer/go-command-pipeline)
[![Version](https://img.shields.io/github/v/release/ccremer/go-command-pipeline)][releases]
[![Go Reference](https://pkg.go.dev/badge/github.com/ccremer/go-command-pipeline.svg)](https://pkg.go.dev/github.com/ccremer/go-command-pipeline)
[![Go Report Card](https://goreportcard.com/badge/github.com/ccremer/go-command-pipeline)][goreport]
[![Codecov](https://img.shields.io/codecov/c/github/ccremer/go-command-pipeline?token=XGOC4XUMJ5)][codecov]

Small Go utility that executes business actions (functions) in a pipeline.
This utility is for you if you think that the business logic is distracted by Go's error handling with `if err != nil` all over the place.

## Usage

```go
import (
    "context"
    pipeline "github.com/ccremer/go-command-pipeline"
)

type Data struct {
    context.Context
    Number int
}

func main() {
	data := &Data{context.Background(), 0} // define arbitrary data to pass around in the steps.
	p := pipeline.NewPipeline[*Data]()
	// define business steps neatly in one place:
	p.WithSteps(
		p.NewStep("define random number", defineNumber),
		p.NewStep("print number", printNumber),
	)
	err := p.RunWithContext(data)
	if err != nil {
		log.Fatal(result)
	}
}

func defineNumber(ctx *Data) error {
	ctx.Number = 10
	return nil
}

// Let's assume this is a business function that can fail.
// You can enable "automatic" fail-on-first-error pipelines by having more small functions that return errors.
func printNumber(ctx *Data) error {
	_, err := fmt.Println(ctx.Number)
	return err
}
```

See more usage in the `examples` dir

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
func Persist(data *Data) error {
    p := pipeline.NewPipeline[*Data]()
    p.WithSteps(
        p.NewStep("prepareTransaction", prepareTransaction()),
        p.NewStep("executeQuery", executeQuery()),
        p.NewStep("commitTransaction", commit()),
    )
    return p.RunWithContext(data)
}

func executeQuery() pipeline.ActionFunc[*Data] {
	return func(data *Data) error {
		err := database.executeQuery("SOME QUERY", data)
		return err
	)
}
...
```
While it seems to add more lines in order to set up a pipeline, it makes it very easily understandable what `Persist()` does without all the error handling.
Plus, each small step might get easier to unit test.

[releases]: https://github.com/ccremer/go-command-pipeline/releases
[codecov]: https://app.codecov.io/gh/ccremer/go-command-pipeline
[goreport]: https://goreportcard.com/report/github.com/ccremer/go-command-pipeline
