package main

import (
	"fmt"

	"github.com/lytics/gentleman"
	"github.com/lytics/gentleman/context"
	"github.com/lytics/gentleman/plugin"
	"github.com/lytics/gentleman/plugins/headers"
)

func main() {
	// Create a new client
	cli := gentleman.New()

	// Define a custom header
	cli.Use(headers.Set("Token", "s3cr3t"))

	// Create a plugin for the response phase
	cli.Use(plugin.NewPhasePlugin("response", func(ctx *context.Context, h context.Handler) {
		ctx.Response.StatusCode = 201 // change the status code
		h.Next(ctx)
	}))

	// Perform the request
	res, err := cli.Request().URL("http://httpbin.org/headers").Send()
	if err != nil {
		fmt.Printf("Request error: %s\n", err)
		return
	}
	if !res.Ok {
		fmt.Printf("Invalid server response: %d\n", res.StatusCode)
		return
	}

	fmt.Printf("Status: %d\n", res.StatusCode)
	fmt.Printf("Body: %s", res.String())
}
