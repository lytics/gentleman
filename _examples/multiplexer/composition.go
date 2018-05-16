package main

import (
	"fmt"

	"github.com/lytics/gentleman"
	"github.com/lytics/gentleman/mux"
	"github.com/lytics/gentleman/plugins/url"
)

func main() {
	// Create a new client
	cli := gentleman.New()

	// Define the server url (must be first)
	cli.Use(url.URL("http://httpbin.org"))

	// Create a new multiplexer based on multiple matchers
	mx := mux.If(mux.Method("GET"), mux.Host("httpbin.org"))

	// Attach a custom plugin on the multiplexer that will be executed if the matchers passes
	mx.Use(url.Path("/headers"))

	// Attach the multiplexer on the main client
	cli.Use(mx)

	// Perform the request
	res, err := cli.Request().Send()
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
