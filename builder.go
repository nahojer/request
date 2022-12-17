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

// Builder builds and sends HTTP requests.
type Builder struct {
	header  http.Header
	timeout *time.Duration
	body    io.Reader
}

// NewBuilder returns a new request builder.
func NewBuilder() *Builder {
	return &Builder{
		header: make(http.Header),
	}
}

// Do sends an HTTP request as configured on the builder and returns an HTTP
// response.
func (b *Builder) Do(method, url string) (*http.Response, error) {
	return b.DoWithContext(context.Background(), method, url)
}

// DoWithContext sends an HTTP request as configured on the builder and returns
// an HTTP response.
func (b *Builder) DoWithContext(ctx context.Context, method, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, b.body)
	if err != nil {
		return nil, err
	}
	req.Header = b.header

	c := clientFromContext(ctx)
	if b.timeout != nil {
		c.Timeout = *b.timeout
	}

	return c.Do(req)
}

// Timeout sets the request timeout.
func (b *Builder) Timeout(d time.Duration) *Builder {
	b.timeout = &d
	return b
}

// Body sets the body of the request.
func (b *Builder) Body(r io.Reader) *Builder {
	b.body = r
	return b
}

// JSONBody sets the body of the request to the JSON representation of data and
// the Content-Type header to application/json.
func (b *Builder) JSONBody(data any) *Builder {
	pr, pw := io.Pipe()
	go func() {
		pw.CloseWithError(json.NewEncoder(pw).Encode(data))
	}()
	b.body = pr
	b.header.Set("Content-Type", "application/json")
	return b
}

// XMLBody sets the body of the request to the XML representation of data and
// the Content-Type header to application/xml.
func (b *Builder) XMLBody(data any) *Builder {
	pr, pw := io.Pipe()
	go func() {
		pw.CloseWithError(xml.NewEncoder(pw).Encode(data))
	}()
	b.body = pr
	b.header.Set("Content-Type", "application/xml")
	return b
}

// Header sets the headers entries of the request associated with key to the
// single element value. It replaces any existing values associated with key.
// The key is case insensitive; it is conanicalized by
// textproto.CanonicalMIMEHeaderKey.
func (b *Builder) Header(key, value string) *Builder {
	b.header.Set(key, value)
	return b
}

// MultiValuedHeader adds a key, value pair to the header of the request.
// It appends to any existing values associated with key. The key is case
// insensitive; it is canonicalized by http.CanonicalHeaderKey.
func (b *Builder) MultiValuedHeader(key, value string) *Builder {
	b.header.Add(key, value)
	return b
}

// ContentType sets the Content-Type header of the request.
func (b *Builder) ContentType(s string) *Builder {
	b.header.Set("Content-Type", s)
	return b
}

// BasicAuth sets the request's Authorization header to use HTTP Basic
// Authentication with the provided username and password.
func (b *Builder) BasicAuth(username, password string) *Builder {
	auth := fmt.Sprintf("%s:%s", username, password)
	b.header.Set("Authorization", fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(auth))))
	return b
}

// BearerAuthentication sets the request's Authorization header to use HTTP
// Bearer Authentication with the provided token.
func (b *Builder) BearerAuthentication(token string) *Builder {
	b.header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	return b
}
