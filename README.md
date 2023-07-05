# httpr

**httpr** is yet another dependency-free library for building, 
sending HTTP requests and handling responses.

httpr tries to follow Go language idioms by providing more convenient 
(read "less boilerplate") means to work with HTTP requests, without replacing
\or abstracting standard net/http types too much, where it isn't really needed.

It also provides custom HTTP client component, which can be configured with
different options such as backoff strategy, rate-limiting, automatic response 
decompression and more.

## Installation
    
    $ go get github.com/hickar/httpr

Note: httpr is dependency-free library.

## Examples

### Simple GET request 
```go
client := httpr.New()
    
resp, err := client.Get(context.TODO(), "https://mysite.com", nil)
if err != nil {
	// handle error...
}
    
fmt.Println(resp.String())
```

### POST request with body
```go
client := httpr.New()

// body can be one of following types (or types implementing interfaces):
// string, []byte, map[string]any, io.Reader or nil.
body := "this is request body"
resp, err := client.Post(context.TODO(), "https://mysite.com", body)
if err != nil {
	// handle error...
}
```

### Request building with various options
```go
req, err := httpr.
	NewRequest().
	Post("https://mysite.com", []byte("sample body content")).
	SetHeaders(map[string]string{
		"Accept": "application/json",
	}).
	Build()

// req is a pointer to standard net/http Request instance.
```

### Building custom client with features like external ratelimiter (like uber's [ratelimit](https://github.com/uber-go/ratelimit)).
```go
client := httpr.New(
    httpr.WithRateLimiter(ratelimit.New()),
	httpr.WithTransport(httpr.BasicAuthTransport("user", "password")),
	httpr.WithRetryCount(3),
	httpr.WithTimeout(time.Minute * 3),
)
resp, err := client.Get(ctx, "https://mysite.com", nil)
// ...
```