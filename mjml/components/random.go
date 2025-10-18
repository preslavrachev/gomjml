package components

import (
	"math/rand"
	"strings"
)

// genRandomHexString generates a random hexadecimal string of the specified length.
// This replicates MJML's genRandomHexString function which uses Math.random()
// to generate random hex digits for unique element IDs.
func genRandomHexString(length int) string {
	var sb strings.Builder
	sb.Grow(length)

	for i := 0; i < length; i++ {
		sb.WriteByte("0123456789abcdef"[rand.Intn(16)])
	}

	return sb.String()
}
