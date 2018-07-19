[gentleman](https://github.com/lytics/gentleman)'s v2 plugin providing retry policy capabilities to your HTTP clients.

Constant backoff strategy will be used by default with a maximum of 3 attempts, but you use a custom or third-party retry strategies.
Request bodies will be cached in the stack in order to re-send them if needed.

By default, retry will happen in case of network error or server response error (>= 500 || = 429).
You can use a custom `Evaluator` function to determine with custom logic when should retry or not. One request may have more then one Evaluator, they will run following the order they were added.

Behind the scenes it implements a custom [http.RoundTripper](https://golang.org/pkg/net/http/#RoundTripper)
interface which acts like a proxy to `http.Transport`, in order to take full control of the response and retry the request if needed.

## Installation

```bash
go get -u gopkg.in/h2non/gentleman-retry.v2
```

## Versions

- **[v2](https://github.com/h2non/gentleman-retry/tree/master)** - Latest version, uses `gentleman@v2`.

## API

See [godoc reference](https://godoc.org/github.com/h2non/gentleman-retry) for detailed API documentation.

## Examples

#### Default retry strategy

```go
package main

import (
  "fmt"

  "gopkg.in/h2non/gentleman.v2"
  "gopkg.in/h2non/gentleman-retry.v2"
)

func main() {
  // Create a new client
  cli := gentleman.New()

  // Define base URL
  cli.URL("http://httpbin.org")

  // Register the retry plugin, using the built-in constant back off strategy
  cli.Use(retry.New(retry.ConstantBackoff, nil))

  // Create a new request based on the current client
  req := cli.Request()

  // Define the URL path at request level
  req.Path("/status/503")

  // Set a new header field
  req.SetHeader("Client", "gentleman")

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
}
```

#### Exponential retry strategy

```go
package main

import (
  "fmt"
  "time"

  "github.com/lytics/gentleman"
  "github.com/lytics/gentleman/plugins/retry"
  "github.com/lytics/gentleman/plugins/retry/retrier"
)

func main() {
  // Create a new client
  cli := gentleman.New()

  // Define base URL
  cli.URL("http://httpbin.org")

  // Register the retry plugin, using a custom exponential retry strategy
  cli.Use(retry.New(retrier.New(retrier.ExponentialBackoff(3, 100*time.Millisecond), nil), nil))

  // Create a new request based on the current client
  req := cli.Request()

  // Define the URL path at request level
  req.Path("/status/503")

  // Set a new header field
  req.SetHeader("Client", "gentleman")

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
}
```

## License

MIT - Tomas Aparicio, Jonas Xavier
