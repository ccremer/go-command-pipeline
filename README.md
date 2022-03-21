# go-command-pipeline

![Go version](https://img.shields.io/github/go-mod/go-version/ccremer/go-command-pipeline)
[![Version](https://img.shields.io/github/v/release/ccremer/go-command-pipeline)][releases]
[![Go Report Card](https://goreportcard.com/badge/github.com/ccremer/go-command-pipeline)][goreport]
[![Codecov](https://img.shields.io/codecov/c/github/ccremer/go-command-pipeline?token=XGOC4XUMJ5)][codecov]

Small Go utility that executes business actions in a pipeline.

## Usage

```go
import (
    "context"
    pipeline "github.com/ccremer/go-command-pipeline"
)

type Data struct {
    Number int
}

func main() {
	data := &Data // define arbitrary data to pass around in the steps.
	p := pipeline.NewPipeline()
	p.WithSteps(
		pipeline.NewStepFromFunc("define random number", defineNumber),
		pipeline.NewStepFromFunc("print number", printNumber),
	)
	result := p.RunWithContext(context.WithValue(context.Background, "data", data))
	if !result.IsSuccessful() {
		log.Fatal(result.Err)
	}
}

func defineNumber(ctx context.Context) error {
	ctx.Value("data").(*Data).Number = 10
	return nil
}

// Let's assume this is a business function that can fail.
// You can enable "automatic" fail-on-first-error pipelines by having more small functions that return errors.
func printNumber(ctx context.Context) error {
	number := ctx.Value("data").(*Data).Number
	fmt.Println(number)
	return nil
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
    p := pipeline.NewPipeline().WithSteps(
        pipeline.NewStepFromFunc("prepareTransaction", prepareTransaction()),
        pipeline.NewStepFromFunc("executeQuery", executeQuery()),
        pipeline.NewStepFromFunc("commitTransaction", commit()),
    )
    return p.RunWithContext(context.WithValue(context.Background(), myKey, data).Err
}

func executeQuery() error {
	return func(ctx context.Context) error {
		data := ctx.Value(myKey).(*Data)
		err := database.executeQuery("SOME QUERY", data)
		return err
	)
}
...
```
While it seems to add more lines in order to set up a pipeline, it makes it very easily understandable what `Persist()` does without all the error handling.

[releases]: https://github.com/ccremer/go-command-pipeline/releases
[codecov]: https://app.codecov.io/gh/ccremer/go-command-pipeline
[goreport]: https://goreportcard.com/report/github.com/ccremer/go-command-pipeline
