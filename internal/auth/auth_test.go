package auth

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSolve(t *testing.T) {
	cases := []struct {
		secret    string
		challenge string
		exp       string
	}{
		0: {
			secret:    "foo\n",
			challenge: "ixslvvxrgkjptxmcgnnsdxsvdmvfympg",
			exp:       "455ce847f0073c7ab3b1465f74507b75d3dc064c1e7de3b71e00de9092fdc89a",
		},
		1: {
			secret:    "supersecret\n",
			challenge: "trbtjmjbpokrrdmexlukhxvfixkpswpb",
			exp:       "20e4a11aec810b1a6dddd41552570b410ee5f0bdb4e49ea1b67a2574c92397c7",
		},
	}

	var buf [Size << 1]byte
	for caseIndex := range cases {
		kase := cases[caseIndex]

		t.Run(strconv.Itoa(caseIndex), func(t *testing.T) {
			Solve(buf[:], kase.secret, []byte(kase.challenge))
			assert.Equal(t, kase.exp, string(buf[:]))
		})
	}
}
