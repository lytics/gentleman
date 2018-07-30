package main

import (
	"fmt"

	"github.com/lytics/gentleman"
)

func main() {
	// Create a new client
	cli := gentleman.New()

	// Define base URL
	cli.BaseURL("http://httpbin.org")

	// Create a new request based on the current client
	req := cli.Request()

	// Define the URL path at request level
	req.Path("/headers")

	// Set a new header field
	req.SetHeader("Client", "gentleman")

	// Get the wire representation of the request
	b, err := req.Dump(true)
	if err != nil {
		fmt.Printf("Dump error: %s\n", err)
		return
	}
	// dump the whole request as its HTTP/1.x wire representation
	fmt.Printf("Request HTTP/1.x wire representation:\n%s\n", string(b))

	// Perform the request
	res, err := req.Send()
	if err != nil {
		fmt.Printf("Request error: %s\n", err)
		return
	}
	if !res.Ok {
		fmt.Printf("Invalid server response: %d\n", res.StatusCode)
		return
	}

	b, err = res.Dump(true)
	if err != nil {
		fmt.Printf("Dump error: %s\n", err)
		return
	}
	// dump the whole response as its HTTP/1.x wire representation
	fmt.Printf("Response HTTP/1.x wire representation:\n%s\n", string(b))

	// Reads the whole response body and returns it as string
	fmt.Printf("Response: %s\n", res.String())

}
