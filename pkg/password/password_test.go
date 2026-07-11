package password

import (
	"encoding/base64"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/argon2"
)

func TestHashVerifyRoundTrip(t *testing.T) {
	const plain = "correct horse battery staple"

	encoded, err := Hash(plain)
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(encoded, "$argon2id$v=19$m=19456,t=2,p=1$"))

	ok, err := Verify(plain, encoded)
	require.NoError(t, err)
	assert.True(t, ok)
}

func TestVerifyWrongPassword(t *testing.T) {
	encoded, err := Hash("the right password")
	require.NoError(t, err)

	ok, err := Verify("the WRONG password", encoded)
	require.NoError(t, err) // wrong password is not an error
	assert.False(t, ok)
}

func TestHashUsesRandomSalt(t *testing.T) {
	h1, err := Hash("same password")
	require.NoError(t, err)
	h2, err := Hash("same password")
	require.NoError(t, err)
	assert.NotEqual(t, h1, h2, "each hash must use a fresh salt")
}

func TestVerifyMalformedHash(t *testing.T) {
	cases := map[string]string{
		"empty":            "",
		"not phc":          "just-a-string",
		"wrong algo":       "$bcrypt$v=19$m=19456,t=2,p=1$c2FsdA$aGFzaA",
		"missing sections": "$argon2id$v=19$m=19456,t=2,p=1$c2FsdA",
		"bad params":       "$argon2id$v=19$m=foo,t=2,p=1$c2FsdA$aGFzaA",
		"bad salt b64":     "$argon2id$v=19$m=19456,t=2,p=1$!!!$aGFzaA",
	}
	for name, encoded := range cases {
		t.Run(name, func(t *testing.T) {
			ok, err := Verify("whatever", encoded)
			assert.Error(t, err)
			assert.False(t, ok)
		})
	}
}

func TestVerifyIncompatibleVersion(t *testing.T) {
	ok, err := Verify("whatever", "$argon2id$v=16$m=19456,t=2,p=1$c2FsdA$aGFzaA")
	assert.ErrorIs(t, err, ErrIncompatibleVersion)
	assert.False(t, ok)
}

// TestVerifyReadsParamsFromHash proves Verify uses the parameters embedded in
// the encoded string, not the current defaults: a hash built with weaker params
// still verifies.
func TestVerifyReadsParamsFromHash(t *testing.T) {
	const plain = "legacy password"
	// Deliberately different from the current defaults (m=19456, t=2).
	const m, tt, p = 8192, 1, 1

	salt := []byte("0123456789abcdef")
	key := argon2.IDKey([]byte(plain), salt, tt, m, p, argonKeyLen)
	b64 := base64.RawStdEncoding
	encoded := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, m, tt, p, b64.EncodeToString(salt), b64.EncodeToString(key))

	ok, err := Verify(plain, encoded)
	require.NoError(t, err)
	assert.True(t, ok, "hash with non-default params must still verify")
}

func TestValidate(t *testing.T) {
	cases := []struct {
		name    string
		input   string
		wantErr error
	}{
		{"empty", "", ErrPasswordTooShort},
		{"too short", "1234567", ErrPasswordTooShort},
		{"min length", "12345678", nil},
		{"max length", strings.Repeat("a", 72), nil},
		{"too long", strings.Repeat("a", 73), ErrPasswordTooLong},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			err := Validate(c.input)
			if c.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, c.wantErr)
			}
		})
	}
}
