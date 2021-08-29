package vasc

import (
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	addr         = "varnish:8080"
	insecureAddr = "varnish:8081"
)

func TestHandshakeFailed(t *testing.T) {
	cfg := config(t)
	cfg.Secret += "12"

	conn := dial(t, addr)
	defer conn.Close()

	client, err := Handshake(conn, cfg)
	assert.Same(t, err, errHandshakeFailed)
	assert.Nil(t, client)
}

func TestHandshake(t *testing.T) {
	client := setup(t)

	assertCloseAfterHandshake(t, client)
}

func TestHandshakeSkipped(t *testing.T) {
	client, err := Handshake(dial(t, insecureAddr), config(t))
	require.NoError(t, err)

	assertCloseAfterHandshake(t, client)
}

func assertCloseAfterHandshake(t *testing.T, client *Client) {
	t.Helper()

	assert.NoError(t, client.Close())

	buf := make([]byte, 1)
	n, err := client.conn.Read(buf)
	assert.Zero(t, n)
	assert.ErrorIs(t, err, net.ErrClosed)
}

func config(t *testing.T) Config {
	t.Helper()

	secret, err := os.ReadFile(filepath.Join("testdata", "secret"))
	require.NoError(t, err)

	return Config{
		Secret: string(secret),
	}
}

func dial(t *testing.T, addr string) (conn net.Conn) {
	t.Helper()

	var err error
	if conn, err = net.DialTimeout("tcp", addr, time.Second<<1); err != nil {
		t.Fatalf("failed dialing %s: %v", addr, err)
	}

	return
}

func setup(t *testing.T) *Client {
	t.Helper()

	client, err := Handshake(dial(t, addr), config(t))
	require.NoError(t, err)

	return client
}
