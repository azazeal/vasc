package vasc

import (
	"bytes"
	"encoding/json"
	"time"
)

// Banner returns the banner of the Varnish instance.
func (c *Client) Banner() (banner string, err error) {
	if err = c.request("banner"); err == nil {
		banner = string(c.in)
	}

	return
}

// Panic returns the last panic of the Varnish instance, if any.
func (c *Client) Panic() (message string, err error) {
	err = c.json(&message, "panic", "-j")
	return
}

// ClearPanic clears the last panic of the Varnish instance. If zero counters
// is set ClearPanic will also clear the related varnishstat counters.
func (c *Client) ClearPanic(includeCounters bool) (err error) {
	if !includeCounters {
		err = c.request("panic.clear")
	} else {
		err = c.request("panic.clear", "-z")
	}

	return
}

// Backend is the set of backends.
type Backend struct {
	Name    string    `json:"name"`
	Healthy bool      `json:"healthy"`
	Stamp   time.Time `json:"time"`
}

// BackendList returns the list of backends the varnish instance is configured
// with.
func (c *Client) BackendList() (backends []Backend, err error) {
	err = c.json(&backends, "backend.list", "-j", "-p")
	return
}

func (c *Client) json(dst interface{}, params ...string) (err error) {
	if err = c.request(params...); err != nil {
		return
	}

	var resp struct {
		version int
		params  []string
		stamp   float64
		body    json.RawMessage
	}

	tokens := []interface{}{
		&resp.version,
		&resp.params,
		&resp.stamp,
		&resp.body,
	}

	switch err = json.NewDecoder(bytes.NewReader(c.in)).Decode(&tokens); err {
	default:
		err = errInvalidJSONResponse
	case nil:
		panic(string(resp.body))
		err = json.Unmarshal(resp.body, dst)
	}

	return
}

func (c *Client) request(params ...string) (err error) {
	if len(params) == 0 {
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.out = append(c.out, params[0]...)
	for _, p := range params[1:] {
		c.out = append(c.out, ' ')
		c.out = append(c.out, p...)
	}
	c.out = append(c.out, '\n')

	if err = c.flush(); err != nil {
		return
	}

	if err = c.next(); err != nil {
		return
	}

	return
}
