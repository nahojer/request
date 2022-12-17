// Go package that provides syntactic sugar for sending HTTP requests with sane
// defaults.
package request

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"time"
)

// DefaultClientTimeout holds the timeout value for the default HTTP client.
var DefaultClientTimeout = time.Minute * 1

// Request sends HTTP requests.
type Request struct {
	header  http.Header
	timeout *time.Duration
	body    io.Reader
}

// New returns a new Request.
func New() *Request {
	return &Request{
		header: make(http.Header),
	}
}

// Do sends an HTTP request and returns an HTTP response.
func (r *Request) Do(ctx context.Context, method, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, r.body)
	if err != nil {
		return nil, err
	}
	req.Header = r.header

	c := clientFromContext(ctx)
	if r.timeout != nil {
		c.Timeout = *r.timeout
	}

	return c.Do(req)
}

// WithTimeout sets the request timeout.
func (r *Request) WithTimeout(d time.Duration) *Request {
	r.timeout = &d
	return r
}

// WithBody sets the body of the request r.
func (r *Request) WithBody(b io.Reader) *Request {
	r.body = b
	return r
}

// WithJSONBody sets the body of the request to the JSON representation of data and
// the Content-Type header to application/json.
func (r *Request) WithJSONBody(data any) *Request {
	pr, pw := io.Pipe()
	go func() {
		pw.CloseWithError(json.NewEncoder(pw).Encode(data))
	}()
	r.body = pr
	r.header.Set("Content-Type", "application/json")
	return r
}

// WithXMLBody sets the body of the request to the XML representation of data and
// the Content-Type header to application/xml.
func (r *Request) WithXMLBody(data any) *Request {
	pr, pw := io.Pipe()
	go func() {
		pw.CloseWithError(xml.NewEncoder(pw).Encode(data))
	}()
	r.body = pr
	r.header.Set("Content-Type", "application/xml")
	return r
}

// WithHeader sets the headers entries of the request associated with key to the
// single element value. It replaces any existing values associated with key.
// The key is case insensitive; it is conanicalized by
// textproto.CanonicalMIMEHeaderKey.
func (r *Request) WithHeader(key, value string) *Request {
	r.header.Set(key, value)
	return r
}

// WithMultiValuedHeader adds a key, value pair to the header of the request.
// It appends to any existing values associated with key. The key is case
// insensitive; it is canonicalized by http.CanonicalHeaderKey.
func (r *Request) WithMultiValuedHeader(key, value string) *Request {
	r.header.Add(key, value)
	return r
}

// WithContentType sets the Content-Type header of the request to s.
func (r *Request) WithContentType(s string) *Request {
	r.header.Set("Content-Type", s)
	return r
}

// WithBasicAuth sets the request's Authorization header to use HTTP Basic
// Authentication with the provided username and password.
func (r *Request) WithBasicAuth(username, password string) *Request {
	auth := fmt.Sprintf("%s:%s", username, password)
	r.header.Set("Authorization", fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(auth))))
	return r
}

// WithBearerAuthentication sets the request's Authorization header to use HTTP
// Bearer Authentication with the provided token.
func (r *Request) WithBearerAuthentication(token string) *Request {
	r.header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	return r
}
