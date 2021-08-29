package vasc

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimpleErrors(t *testing.T) {
	testIs(t, "IsSyntax", IsSyntax, errSyntax)
	testIs(t, "IsUnknown", IsUnknown, errUnknown)
	testIs(t, "IsUnimplemented", IsUnimplemented, errUnimplemented)
	testIs(t, "IsTooFew", IsTooFew, errTooFew)
	testIs(t, "IsTooMany", IsTooMany, errTooMany)
	testIs(t, "IsParam", IsParam, errParam)
	testIs(t, "IsHandshakeFailed", IsHandshakeFailed, errHandshakeFailed)
	testIs(t, "IsCannot", IsCannot, errCannot)
	testIs(t, "IsComms", IsComms, errComms)
	testIs(t, "IsClosed", IsClosed, errClosed)
	testIs(t, "IsInvalidJSONResponse", IsInvalidJSONResponse, errInvalidJSONResponse)
}

func testIs(t *testing.T, name string, fn func(error) bool, err error) {
	t.Helper()

	t.Run(name, func(t *testing.T) {
		cases := []struct {
			err error
			exp bool
		}{
			0: {},
			1: {
				err: assert.AnError,
			},
			2: {
				err: fmt.Errorf("wrapped: %w", assert.AnError),
			},
			3: {
				err: err,
				exp: true,
			},
			4: {
				err: fmt.Errorf("wrapped: %w", err),
				exp: true,
			},
		}

		for caseIndex := range cases {
			kase := cases[caseIndex]

			t.Run(strconv.Itoa(caseIndex), func(t *testing.T) {
				assert.Equal(t, kase.exp, fn(kase.err))
			})
		}
	})
}
