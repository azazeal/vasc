// Package vasc implements a Varnish Administrative Socket client.
package vasc

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"sync"
	"time"

	"github.com/azazeal/vasc/internal/auth"
)

const (
	// DefaultConnectTimeout denotes the default connect timeout.
	DefaultConnectTimeout = time.Second << 3

	// DefaultReadTimeout denotes the default read timeout.
	DefaultReadTimeout = time.Minute >> 1

	// DefaultWriteTimeout denotes the default write timeout.
	DefaultWriteTimeout = time.Minute
)

// Config wraps the configuration for instances of Client.
type Config struct {
	// Secret is the
	Secret string

	// ConnectTimeout denotes the maximum amount of time to wait while
	// connecting to the varnish instance.
	ConnectTimeout time.Duration

	// ReadTimeout denotes the maximum amount of time to wait while reading from
	// the varnish instance. Values less than 1 mean no read timeout.
	ReadTimeout time.Duration

	// WriteTimeout denotes the maximum amount of time to wait while writing to
	// the varnish instance. Values less than 1 mean no read timeout.
	WriteTimeout time.Duration
}

// Client wraps the functionality of a Varnish Administrative Socket Client.
//
// Instances of Client returned by Handshake are safe for concurrent use.
type Client struct {
	mu         sync.Mutex
	conn       net.Conn
	cfg        Config
	in         []byte // buffer for incoming messages
	out        []byte // buffer for outgoing messages
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

// Handshake returns a new Client running on the given Conn after performing
// the initial handshake.
//
// In case of an error callers should close conn. Alternatively, callers must
// not retain access to conn.
func Handshake(conn net.Conn, cfg Config) (*Client, error) {
	c := &Client{
		conn: conn,
		cfg:  cfg,
	}

	if err := c.handshake(); err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Client) handshake() (err error) {
	switch err = c.next(); {
	default:
		return
	case err == nil:
		return nil // no auth required
	case errors.Is(err, errAuth):
		break // we have to login
	}

	if len(c.in) < auth.Size {
		err = errHandshakeFailed

		return
	}

	if err = c.auth(c.in[:auth.Size]); err != nil {
		return
	}

	if err = c.next(); err != nil {
		err = errHandshakeFailed
	}

	return
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

	if err = c.setWriteTimeout(); err != nil {
		return
	}

	var n int
	n, err = c.conn.Write(c.out)
	sent(c.out[:n])

	return
}

func (c *Client) next() error {
	c.in = c.in[:0]

	if err := c.setReadTimeout(); err != nil {
		return err
	}

	var size int
	code, size, err := c.header()
	if err != nil {
		return err
	}

	if s := size + 1; cap(c.in) < s {
		c.in = make([]byte, s)
	} else {
		c.in = c.in[:s]
	}

	size, err = io.ReadFull(c.conn, c.in)
	c.in = c.in[:size]
	read(c.in)

	if err == nil {
		err = codeToError(code, size)
	}
	return err
}

func (c *Client) setReadTimeout() error {
	v := c.cfg.ReadTimeout
	if v < 1 {
		return nil
	}

	d := time.Now().Add(v)
	return c.conn.SetReadDeadline(d)
}

func (c *Client) setWriteTimeout() error {
	v := c.cfg.WriteTimeout
	if v < 1 {
		return nil
	}

	d := time.Now().Add(v)
	return c.conn.SetWriteDeadline(d)
}

func (c *Client) header() (code, size int, err error) {
	var data [13]byte
	switch err = fillHeader(&data, c.conn); {
	case err != nil:
		return
	case data[12] != '\n', data[3] != ' ':
		err = ErrInvalidHeader(data[:])

		return
	}

	var ok bool
	if code, ok = parseHeaderInt(data[:3], false); ok {
		size, ok = parseHeaderInt(data[4:], true)
	}

	if !ok {
		err = ErrInvalidHeader(data[:])
	}

	return
}

func fillHeader(h *[13]byte, r io.Reader) error {
	n, err := io.ReadFull(r, h[:])
	read(h[:n])

	return err
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

func sent(b []byte) {
	if len(b) == 0 {
		return
	}

	fmt.Fprintf(os.Stderr, "|> %q\n", b)
}

func read(b []byte) {
	if len(b) == 0 {
		return
	}

	fmt.Fprintf(os.Stderr, "<| %q\n", b)
}
