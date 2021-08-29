// Package auth implements the authentication challenge mechanism.
package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
)

// Size denotes the size of cryptographic challenges challenge.
const Size = sha256.Size

var nl = []byte{'\n'}

// Solve writes the hex solution for the given secret & challenge combination
// into w and returns the number of bytes written along with the first error
// encountered.
func Solve(dst []byte, secret string, challenge []byte) {
	_ = dst[Size<<1-1]

	h := sha256.New()
	_, _ = h.Write(challenge)
	_, _ = h.Write(nl)
	_, _ = io.WriteString(h, secret)
	_, _ = h.Write(challenge)
	_, _ = h.Write(nl)

	var sum [Size]byte
	_ = h.Sum(sum[:0])
	_ = hex.Encode(dst[:Size<<1], sum[:Size])
}
