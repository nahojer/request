# request

Go package that provides syntactic sugar for sending HTTP requests with sane defaults.

All of the documentation can be found on the [go.dev](https://pkg.go.dev/github.com/nahojer/request?tab=doc) website.

Is it Good? [Yes](https://news.ycombinator.com/item?id=3067434).

## Examples

### Performing a simple POST request

Below example shows how one could perform a simple POST request:

```go
type NewMessage struct {
	Text string `json:"text"`
}

type Message struct {
	ID string `json:"id"`
	Text string `json:"text"`
}

response, err := request.New().
	WithTimeout(time.Second*10).
	WithBasicAuth("username", "password").
	WithAccept("application/json").
	WithJSONBody(&NewMessage{"Hello world!"}).
	WithHeader("x-trace-id", "477cd6fa-758c-4f85-97b9-6f180a703039").
	Do(context.Background(), http.MethodPost, "http://localhost/api/v1/messages")
if err != nil {
	panic(err)
}
defer response.Body.Close()

data, err := io.ReadAll(response.Body)
if err != nil {
	panic(err)
}

var msg Message
if err := json.Unmarshal(data, &msg); err != nil {
	panic(err)
}
```

### Results

\*\*TODO\*\*

```go
var msg Message
result, err := request.New().
	// ... options in previous example omitted for brevity
	WithJSONResult(&msg).
	Do(context.Background(), http.MethodPost, "http://localhost/api/v1/messages")
if err != nil {
	panic(fmt.Errorf("request failed with status %d and response body %q: %w",
		result.Response.Status, string(result.RawData), err))
}
```

### Overriding the underlying HTTP client via context

We can set the underlying `*http.Client` via the `context.Context` passed to the `Do` function:

```go
ctx := request.AttachClientToContext(context.Background(), &http.Client{})
```

This is great for testing as we can override the `Transport` field on the client and thus mock the external system without needing to create an interface. Consider the following example.

```go
// RoundTripperFunc simplifies creating a http.RoundTripper, which is used to
// intercept the transport of custom *http.Client in tests.
type RoundTripperFunc func(*http.Request) (*http.Response, error)

// Rountrip implements http.RoundTripper.
func (f RoundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

mockedClient := http.Client{
	Timeout: time.Second * 5,
	Transport: RoundTripperFunc(func(r *http.Request) (*http.Response, error) {
		b, err := io.ReadAll(r.Body)
		if err != nil {
			return nil, err
		}
		body := string(b)
		if strings.Contains(body, "bad request") {
			return &http.Response{StatusCode: http.StatusBadRequest}, nil
		}
		return &http.Response{StatusCode: http.StatusOK}, nil
	}),
}

// We attach our custom client to ctx for the request builder to use.
ctx := request.AttachClientToContext(context.Background(), &mockedClient)

request.New().
	// ... options in previous examples omitted for brevity
	Do(ctx, http.MethodPost, "http://localhost/api/v1/messages")
if err != nil {
	panic(err)
}
```

In above example we create a HTTP client that responds with status 400 bad request if the HTTP request body contains the string "bad request" and with status 200 OK otherwise.
