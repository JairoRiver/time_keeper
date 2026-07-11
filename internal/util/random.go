package util

import (
	crand "crypto/rand"
	"fmt"
	"math/rand"
	"strings"
)

const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKL"

// secureAlphabet is the full alphanumeric set used by SecureRandomString.
const secureAlphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// RandomInt generate a random interger between min and max
func RandomInt(min, max int64) int64 {
	return min + rand.Int63n(max-min+1)
}

// RandomString generate a random string
func RandomString(n int) string {
	var sb strings.Builder
	k := len(alphabet)

	for i := 0; i < n; i++ {
		c := alphabet[rand.Intn(k)]
		sb.WriteByte(c)
	}

	return sb.String()
}

// SecureRandomString generates a cryptographically secure random string of
// length n using crypto/rand and the full alphanumeric alphabet. Use it for
// security-sensitive values such as the per-user JWT signing key. It returns an
// error if the system entropy source fails; the error must never be silently
// swallowed with a math/rand fallback.
func SecureRandomString(n int) (string, error) {
	if n <= 0 {
		return "", fmt.Errorf("util: SecureRandomString length must be positive, got %d", n)
	}

	const k = len(secureAlphabet)
	// Largest multiple of k representable in a byte. Bytes >= limit are
	// rejected so every alphabet character is equally likely (no modulo bias).
	limit := byte(256 - (256 % k))

	out := make([]byte, n)
	buf := make([]byte, n)
	filled := 0
	for filled < n {
		if _, err := crand.Read(buf); err != nil {
			return "", fmt.Errorf("util: SecureRandomString read entropy: %w", err)
		}
		for _, b := range buf {
			if b >= limit {
				continue
			}
			out[filled] = secureAlphabet[int(b)%k]
			filled++
			if filled == n {
				break
			}
		}
	}
	return string(out), nil
}

// RandomURL generate a random URL
func RandomEmail() string {
	auxInt := RandomInt(3, 8)

	return RandomString(int(auxInt)) + "@" + RandomString(5) + ".com"
}
