package util

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecureRandomString(t *testing.T) {
	const n = 64

	s, err := SecureRandomString(n)
	require.NoError(t, err)
	require.Len(t, s, n)

	// Every character must come from the secure alphabet.
	for _, c := range s {
		assert.True(t, strings.ContainsRune(secureAlphabet, c),
			"unexpected character %q not in secure alphabet", c)
	}

	// Two consecutive calls must not collide.
	s2, err := SecureRandomString(n)
	require.NoError(t, err)
	assert.NotEqual(t, s, s2)
}

func TestSecureRandomStringInvalidLength(t *testing.T) {
	for _, n := range []int{0, -1} {
		s, err := SecureRandomString(n)
		assert.Error(t, err)
		assert.Empty(t, s)
	}
}
