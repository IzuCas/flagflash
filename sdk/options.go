package sdk

import (
	"net/http"
	"time"
)

// Option is a functional option for configuring the Client.
type Option func(*Client)

// WithHTTPClient replaces the default HTTP client.
func WithHTTPClient(hc *http.Client) Option {
	return func(cl *Client) { cl.http = hc }
}

// WithTimeout sets a default HTTP request timeout applied to requests that do
// not carry a shorter context deadline (default: no fixed client-level timeout).
func WithTimeout(d time.Duration) Option {
	return func(c *Client) { c.http.Timeout = d }
}
