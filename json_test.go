package vasc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnmarshalJSONResponse(t *testing.T) {
	const data = `[ 2, ["pid", "-j"], 1633377196.960,
  {"master": 1, "worker": 18}
]
`

	var res struct {
		Master int `json:"master"`
		Worker int `json:"worker"`
	}

	assert.NoError(t, UnmarshalJSONResponse([]byte(data), &res))
	assert.Equal(t, 1, res.Master)
	assert.Equal(t, 18, res.Worker)
}
