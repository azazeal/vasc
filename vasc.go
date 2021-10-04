// Package vasc implements a Varnish Administrative Socket client.
package vasc

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/azazeal/vasc/internal/auth"
)

// The set of Varnish response status codes.
const (
	StatusSyntax        = 100
	StatusUnknown       = 101
	StatusUnimplemented = 102
	StatusTooFew        = 104
	StatusTooMany       = 105
	StatusParam         = 106
	statusAuth          = 107
	StatusOK            = 200
	StatusContinue      = 201
	StatusCannot        = 300
	StatusComms         = 400
	StatusClose         = 500
)

// Config holds the configuration for instances of Client.
type Config struct {
	// Secret is the secret the Varnish instance uses. Leave empty when
	// connecting to insecure Varnish instances.
	Secret string

	// ReadTimeout denotes the maximum amount of time to wait while reading from
	// the remote Varnish instance. Values less than 1 mean no read timeout.
	ReadTimeout time.Duration

	// WriteTimeout denotes the maximum amount of time to wait while writing to
	// the remote Varnish instance. Values less than 1 mean no write timeout.
	WriteTimeout time.Duration
}

// Client implements a Varnish administrative socket client.
//
// Instances of Client are safe for concurrent use by multiple callers.
type Client struct {
	mu         sync.Mutex
	conn       net.Conn
	cfg        Config
	in         []byte   // buffer for incoming messages
	out        []byte   // buffer for outgoing messages
	header     [13]byte // buffer for headers
	closeOnce  sync.Once
	closeError error
}

// Close implements io.Closer for Client.
func (c *Client) Close() error {
	c.closeOnce.Do(func() {
		c.closeError = c.conn.Close()
	})

	return c.closeError
}

// Dial is shorthand for DialTimeout(network, address, time.Minute>>1).
func Dial(network, address string, cfg Config) (*Client, error) {
	return DialTimeout(network, address, cfg, time.Minute>>1)
}

// DialTimeout calls DialContext with the given arguments and a context with
// the given timeout.
func DialTimeout(network, address string, cfg Config, timeout time.Duration) (*Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return DialContext(ctx, network, address, cfg)
}

var dialer net.Dialer

// DialContext acts like Dial but takes a context. It behaves similarly to the
// net.Dialer's DialContext function.
func DialContext(ctx context.Context, network, address string, cfg Config) (*Client, error) {
	conn, err := dialer.DialContext(ctx, network, address)
	if err != nil {
		return nil, err
	}

	client, err := From(conn, cfg)
	if err != nil {
		_ = conn.Close()
	}

	return client, err
}

// From returns a new Client running on conn.
//
// In case of an error callers should close conn. Alternatively, callers must
// not retain access to it.
func From(conn net.Conn, cfg Config) (*Client, error) {
	client := &Client{
		conn: conn,
		cfg:  cfg,
	}

	if err := handshake(client); err != nil {
		return nil, err
	}

	return client, nil
}

var errHandshakeFailed = errors.New("vasc: handshake failed")

// IsHandshakeFailed reports whether any error in err's chain occured due to the
// handshake with the Varnish instance failing.
func IsHandshakeFailed(err error) bool {
	return errors.Is(err, errHandshakeFailed)
}

type errUnexpectedHandshakeStatusCode int

func (err errUnexpectedHandshakeStatusCode) Error() string {
	return fmt.Sprintf("vasc: unexpected handshake status code %d", err)
}

// IsUnexpectedHandshakeStatusCode reports whether any error in err's chain
// occured due to the Varnish instance returning an unexpected status code
// during handshake.
func IsUnexpectedHandshakeStatusCode(err error) (is bool) {
	for err != nil {
		if _, is = err.(errUnexpectedHandshakeStatusCode); is {
			break
		}

		err = errors.Unwrap(err)
	}

	return
}

type errHandshakeChallengeTooShort int

func (err errHandshakeChallengeTooShort) Error() string {
	return fmt.Sprintf("vasc: handshake challenge too short (len: %d)", err)
}

// IsHandshakeChallengeTooShort reports whether any error in err's chain occured
// due to the Varnish instance returning an authentication challenge that's too
// short during handshake.
func IsHandshakeChallengeTooShort(err error) (is bool) {
	for err != nil {
		if _, is = err.(errHandshakeChallengeTooShort); is {
			break
		}

		err = errors.Unwrap(err)
	}

	return
}

func handshake(c *Client) error {
	switch code, err := c.response(); {
	case err != nil:
		return err
	case code == StatusOK:
		return nil // no need to login
	case code != statusAuth:
		return errUnexpectedHandshakeStatusCode(code)
	case len(c.in) < auth.Size:
		return errHandshakeChallengeTooShort(len(c.in))
	}

	// we have to login
	if err := c.auth(c.in[:auth.Size]); err != nil {
		return err
	}

	switch code, err := c.response(); {
	case err != nil:
		return err
	case code == StatusClose:
		return errHandshakeFailed
	case code != StatusOK:
		return errUnexpectedHandshakeStatusCode(code)
	default:
		return nil
	}
}

func (c *Client) auth(challenge []byte) error {
	const size = 6 + auth.Size<<1

	if cap(c.out) < size {
		c.out = make([]byte, size)
	} else {
		c.out = c.out[:size]
	}

	const prefix = "auth "
	copy(c.out, prefix)
	auth.Solve(c.out[len(prefix):], c.cfg.Secret, challenge)
	c.out[size-1] = '\n'

	return c.flush()
}

func (c *Client) flush() (err error) {
	defer func() {
		for i := range c.out {
			c.out[i] = 0
		}
		c.out = c.out[:0]
	}()

	if err = c.setWriteTimeout(); err == nil {
		_, err = c.conn.Write(c.out)
	}

	return
}

func (c *Client) setReadTimeout() (err error) {
	if v := c.cfg.ReadTimeout; v > 0 {
		d := time.Now().Add(v)
		err = c.conn.SetReadDeadline(d)
	}

	return
}

func (c *Client) setWriteTimeout() (err error) {
	if v := c.cfg.WriteTimeout; v > 0 {
		d := time.Now().Add(v)
		err = c.conn.SetWriteDeadline(d)
	}

	return
}

type errInvalidResponseHeader string

func (err errInvalidResponseHeader) Error() string {
	return fmt.Sprintf("vasc: invalid response header %q", string(err))
}

// IsInvalidResponseHeader reports whether any error in err's chain occured due
// to the remote Varnish instance returning a response with an invalid header.
func IsInvalidResponseHeader(err error) (is bool) {
	for err != nil {
		if _, is = err.(errInvalidResponseHeader); is {
			break
		}

		err = errors.Unwrap(err)
	}

	return
}

func (c *Client) readHeader() (code, size int, err error) {
	switch _, err = io.ReadFull(c.conn, c.header[:]); {
	case err != nil:
		return
	case c.header[12] != '\n', c.header[3] != ' ':
		err = errInvalidResponseHeader(c.header[:])

		return
	}

	var ok bool
	if code, ok = parseHeaderInt(c.header[:3], false); ok {
		size, ok = parseHeaderInt(c.header[4:], true)
	}

	if !ok {
		err = errInvalidResponseHeader(c.header[:])
	}

	return
}

func parseHeaderInt(b []byte, first bool) (int, bool) {
	var v int
	for _, c := range b {
		switch {
		case first && c == ' ', !first && c == '\n':
			return v, true
		case c < '0', c > '9':
			return v, false
		default:
			v *= 10
			v += int(c - '0')
		}
	}

	return v, true
}

// Execute instructs the remote Varnish instance to execute then given command
// with the given arguments, appends the response into the given slice and
// returns the response's status code, the returning slice and the first error
// encountered, if any.
func (c *Client) Execute(dst []byte, command string, args ...string) (code int, data []byte, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.out = append(c.out, command...)
	for _, p := range args {
		c.out = append(c.out, ' ')
		c.out = append(c.out, p...)
	}
	c.out = append(c.out, '\n')

	if err = c.flush(); err == nil {
		code, err = c.response()
		data = append(dst, c.in...)
	}

	return
}

func (c *Client) response() (code int, err error) {
	c.in = c.in[:0]

	if err = c.setReadTimeout(); err != nil {
		return
	}

	var size int
	if code, size, err = c.readHeader(); err != nil {
		return
	}

	if s := size + 1; cap(c.in) < s {
		c.in = make([]byte, s)
	} else {
		c.in = c.in[:s]
	}

	size, err = io.ReadFull(c.conn, c.in)
	c.in = c.in[:size]

	return
}
