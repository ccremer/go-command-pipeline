//go:build examples

package examples

import (
	"context"
	"fmt"
	"testing"

	pipeline "github.com/ccremer/go-command-pipeline"
)

type ClientContext struct {
	context.Context
	client         *Client
	createClientFn pipeline.ActionFunc[*ClientContext]
	recorder       *pipeline.DependencyRecorder[*ClientContext]
}

func TestExample_DependencyRecorder(t *testing.T) {
	token := "someverysecuretoken"
	ctx := &ClientContext{Context: context.Background(), recorder: pipeline.NewDependencyRecorder[*ClientContext]()}
	ctx.createClientFn = createClientFn(token)
	pipe := pipeline.NewPipeline[*ClientContext]().WithBeforeHooks(ctx.recorder.Record)
	pipe.WithSteps(
		pipe.NewStep("create client", ctx.createClientFn),
		pipe.NewStep("connect to API", connect),
		pipe.NewStep("get resource", getResource),
	)
	err := pipe.RunWithContext(ctx)
	if err != nil {
		t.Fatal(err)
	}
}

// Client is some sort of client used to connect with some API.
type Client struct {
	Token string
}

func (c *Client) Connect() error {
	// some logic connect to API using token
	return nil
}

func (c *Client) GetResource() (string, error) {
	// some logic to get a resource from API, let's assume this is only possible after calling Connect().
	// (arguably bad design for a client but let's roll with it for demo purposes)
	return "resource", nil
}

func createClientFn(token string) func(ctx *ClientContext) error {
	return func(ctx *ClientContext) error {
		ctx.client = &Client{Token: token}
		return nil
	}
}

func connect(ctx *ClientContext) error {
	// we need to check first if the client has been created, otherwise the client remains 'nil' and we'd run into a Nil pointer panic.
	// This allows us to declare a certain order for the steps at compile time, and checking them at runtime.
	if err := ctx.recorder.RequireDependencyByFuncName(ctx.createClientFn); err != nil {
		return err
	}
	return ctx.client.Connect()
}

func getResource(ctx *ClientContext) error {
	// We can check for preconditions more easily
	ctx.recorder.MustRequireDependencyByFuncName(connect, ctx.createClientFn)
	resource, err := ctx.client.GetResource()
	fmt.Println(resource)
	return err
}
