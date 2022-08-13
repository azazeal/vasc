package vasc

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"testing"
	"testing/quick"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	secure   = os.Getenv("SECURE_HOST")   // secure varnish test instance
	secret   = os.Getenv("SECRET")        // secure varnish instance's secret
	insecure = os.Getenv("INSECURE_HOST") // insecure varnish test instance
)

func TestDialError(t *testing.T) {
	client, err := Dial("tcp", "unknown-host", newConfig(""))
	assert.Error(t, err)
	assert.Nil(t, client)
}

func TestDialSecure(t *testing.T) {
	t.Parallel()

	testDial(t, true)
}

func TestDialInsecure(t *testing.T) {
	t.Parallel()

	testDial(t, false)
}

func testDial(t *testing.T, sec bool) {
	t.Helper()

	var (
		client *Client
		err    error
	)
	if sec {
		client, err = Dial("tcp", secure, newConfig(secret))
	} else {
		client, err = Dial("tcp", insecure, newConfig(""))
	}
	require.NoError(t, err)

	require.NoError(t, client.Close())
	require.NoError(t, client.Close()) // close twice intentionally

	n, err := client.conn.Read([]byte{1})
	assert.Zero(t, n)
	assert.True(t, errors.Is(err, net.ErrClosed))
}

func TestDialSecureWithoutSecret(t *testing.T) {
	t.Parallel()

	testHandshakeFailed(t, secure, "")
}

func TestDialSecureWithWrongSecret(t *testing.T) {
	t.Parallel()

	testHandshakeFailed(t, secure, secret+"!")
}

func testHandshakeFailed(t *testing.T, addr, secret string) {
	t.Helper()

	cfg := newConfig(secret)

	client, err := DialTimeout("tcp", addr, cfg, time.Second<<3)
	assert.True(t, IsHandshakeFailed(err))
	assert.Nil(t, client)
}

func TestExecute(t *testing.T) {
	t.Parallel()

	const exp = `banner
Print welcome banner.

`

	client, err := Dial("tcp", secure, newConfig(secret))
	require.NoError(t, err)

	code, data, err := client.Execute(nil, "help", `"banner"`)
	assert.NoError(t, err)
	assert.Equal(t, exp, string(data))
	assert.Equal(t, StatusOK, code)
}

func TestInvalidResponseHeader(t *testing.T) {
	t.Parallel()

	testTranscript(t, 1, IsInvalidResponseHeader)
}

func TestHandshakeChallengeTooShort(t *testing.T) {
	t.Parallel()

	testTranscript(t, 2, IsHandshakeChallengeTooShort)
}

func TestUnexpectedHandshakeStatusCode(t *testing.T) {
	t.Parallel()

	testTranscript(t, 3, IsUnexpectedHandshakeStatusCode)
}

func TestUnexpectedHandshakeStatusCodeResponse(t *testing.T) {
	t.Parallel()

	testTranscript(t, 4, IsUnexpectedHandshakeStatusCode)
}

func testTranscript(t *testing.T, id int, predicate func(error) bool) {
	t.Helper()

	path := filepath.Join("testdata", fmt.Sprintf("%d.transcript", id))
	data, err := os.ReadFile(path)
	require.NoError(t, err)

	conn, srv := net.Pipe()
	defer conn.Close()

	go func() {
		_, _ = srv.Write(data)
		srv.Close()
	}()
	go func() {
		_, _ = io.Copy(io.Discard, srv)
	}()

	client, err := From(conn, newConfig(""))
	assert.True(t, predicate(err))
	assert.Nil(t, client)
}

func TestInvalidResponseHeaderError(t *testing.T) {
	fn := func(err invalidResponseHeaderError) bool {
		return fmt.Sprintf("vasc: invalid response header %q", string(err)) ==
			err.Error()
	}
	require.NoError(t, quick.Check(fn, nil))
}

func TestUnexpectedHandshakeStatusCodeError(t *testing.T) {
	fn := func(err unexpectedHandshakeStatusCodeError) bool {
		return fmt.Sprintf("vasc: unexpected handshake status code %d", err) ==
			err.Error()
	}
	require.NoError(t, quick.Check(fn, nil))
}

func TestHandshakeChallengeTooShortError(t *testing.T) {
	fn := func(err handshakeChallengeTooShortError) bool {
		return fmt.Sprintf("vasc: handshake challenge too short (len: %d)", err) ==
			err.Error()
	}
	require.NoError(t, quick.Check(fn, nil))
}

func newConfig(secret string) Config {
	return Config{
		Secret:       secret,
		ReadTimeout:  time.Second << 1,
		WriteTimeout: time.Second << 1,
	}
}
