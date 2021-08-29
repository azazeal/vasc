package vasc

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFromContext(t *testing.T) {
	exp := new(Client)

	got := FromContext(NewContext(context.Background(), exp))
	assert.Same(t, exp, got)
}

func TestFromContextPanics(t *testing.T) {
	assert.Panics(t, func() { FromContext(context.Background()) })
}
