package main

import (
	"fmt"

	"github.com/lytics/gentleman"
	"github.com/lytics/gentleman/plugins/cookies"
)

func main() {
	// Create a new client
	cli := gentleman.New()

	// Define cookies
	cli.Use(cookies.Set("foo", "bar"))

	// Configure cookie jar store
	cli.Use(cookies.Jar())

	// Perform the request
	res, err := cli.Request().URL("http://httpbin.org/cookies").Send()
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
