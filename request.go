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

// HTTP MIME types.
const (
	wildcardMIME = "*/*"
	jsonMIME     = "application/json"
	xmlMIME      = "application/xml"
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
func (r *Request) WithJSONBody(v any) *Request {
	pr, pw := io.Pipe()
	go func() {
		pw.CloseWithError(json.NewEncoder(pw).Encode(v))
	}()
	r.body = pr
	r.header.Set("Content-Type", jsonMIME)
	return r
}

// WithXMLBody sets the body of the request to the XML representation of data and
// the Content-Type header to application/xml.
func (r *Request) WithXMLBody(v any) *Request {
	pr, pw := io.Pipe()
	go func() {
		pw.CloseWithError(xml.NewEncoder(pw).Encode(v))
	}()
	r.body = pr
	r.header.Set("Content-Type", xmlMIME)
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

// WithResult returns a request client who's Do func returns a [Result] instead
// of the raw HTTP response.
func (r *Request) WithResult() *withResult {
	return &withResult{req: r}
}

// WithJSONResult sets the Accept header of the request to application/json
// if the header is not already set, and returns a HTTP client who's Do func
// returns a [Result] instead of the raw HTTP response and decodes the JSON
// response body into v.
func (r *Request) WithJSONResult(v any) *withResult {
	if accept := r.header.Get("Accept"); accept == "" {
		r.header.Set("Accept", jsonMIME)
	}
	return &withResult{
		req: r,
		unmarshal: func(data []byte) error {
			if err := json.Unmarshal(data, v); err != nil {
				return fmt.Errorf("failed to unmarshal JSON: %w", err)
			}
			return nil
		},
	}
}

// WithXMLResult sets the Accept header of the request to application/xml
// if the header is not already set, and returns a HTTP client who's Do func
// returns a [Result] instead of the raw HTTP response and decodes the XML
// response body into v.
func (r *Request) WithXMLResult(v any) *withResult {
	if accept := r.header.Get("Accept"); accept == "" {
		r.header.Set("Accept", xmlMIME)
	}
	return &withResult{
		req: r,
		unmarshal: func(data []byte) error {
			if err := xml.Unmarshal(data, v); err != nil {
				return fmt.Errorf("failed to unmarshal JSON: %w", err)
			}
			return nil
		},
	}
}
