package vasc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBanner(t *testing.T) {
	const exp = `-----------------------------
Varnish Cache CLI 1.0
-----------------------------
Linux,5.4.0-84-generic,x86_64,-junix,-smalloc,-sdefault,-hcritbit
varnish-6.6.1 revision e6a8c860944c4f6a7e1af9f40674ea78bbdcdc66

Type 'help' for command list.
Type 'quit' to close CLI session.

`

	client := setup(t)

	got, err := client.Banner()
	assert.NoError(t, err)
	assert.Equal(t, exp, got)
}

func TestPanic(t *testing.T) {
	client := setup(t)

	msg, err := client.Panic()
	assert.NoError(t, err)
	assert.Empty(t, msg)
}

func TestClearPanic(t *testing.T) {
	client := setup(t)

	err := client.ClearPanic(false)
	assert.NoError(t, err)
}

func TestBackendList(t *testing.T) {
	client := setup(t)

	got, err := client.BackendList()
	assert.NoError(t, err)
	assert.NotEmpty(t, got)
}
