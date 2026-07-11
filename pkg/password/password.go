// Package password hashes and verifies user passwords with argon2id, using the
// standard PHC string format so the cost parameters travel with each hash.
package password

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/argon2"
)

// argon2id cost parameters (OWASP recommendation, 2024).
const (
	argonMemory      = 19456 // KiB (19 MiB)
	argonTime        = 2
	argonParallelism = 1
	argonSaltLen     = 16 // bytes
	argonKeyLen      = 32 // bytes

	minPasswordLen = 8
	maxPasswordLen = 72
)

var (
	// ErrPasswordTooShort / ErrPasswordTooLong are returned by Validate.
	ErrPasswordTooShort = fmt.Errorf("password must be at least %d characters", minPasswordLen)
	ErrPasswordTooLong  = fmt.Errorf("password must be at most %d characters", maxPasswordLen)

	// ErrInvalidHash is returned by Verify when the encoded hash is malformed.
	ErrInvalidHash = errors.New("invalid argon2id hash format")
	// ErrIncompatibleVersion is returned when the hash uses an argon2 version
	// this build cannot compute.
	ErrIncompatibleVersion = errors.New("incompatible argon2 version")
)

// Validate checks a plaintext password against the length policy. It must run
// before Hash at registration time.
func Validate(plain string) error {
	if len(plain) < minPasswordLen {
		return ErrPasswordTooShort
	}
	if len(plain) > maxPasswordLen {
		return ErrPasswordTooLong
	}
	return nil
}

// Hash derives an argon2id hash of plain and returns it in PHC string format:
//
//	$argon2id$v=19$m=19456,t=2,p=1$<b64-salt>$<b64-hash>
func Hash(plain string) (string, error) {
	salt := make([]byte, argonSaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("password: read salt: %w", err)
	}

	key := argon2.IDKey([]byte(plain), salt, argonTime, argonMemory, argonParallelism, argonKeyLen)

	b64 := base64.RawStdEncoding
	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, argonMemory, argonTime, argonParallelism,
		b64.EncodeToString(salt), b64.EncodeToString(key),
	), nil
}

// Verify reports whether plain matches the encoded argon2id hash. The cost
// parameters and salt are read from the encoded string, so hashes produced with
// older parameters keep verifying after the defaults are bumped. Comparison is
// constant-time. A false result with a nil error means "wrong password"; a
// non-nil error means the encoded hash could not be parsed.
func Verify(plain, encoded string) (bool, error) {
	p, salt, hash, err := decode(encoded)
	if err != nil {
		return false, err
	}

	other := argon2.IDKey([]byte(plain), salt, p.time, p.memory, p.parallelism, uint32(len(hash)))
	return subtle.ConstantTimeCompare(hash, other) == 1, nil
}

type params struct {
	memory      uint32
	time        uint32
	parallelism uint8
}

// decode parses a PHC-format argon2id hash into its parameters, salt and hash.
func decode(encoded string) (params, []byte, []byte, error) {
	// $argon2id$v=19$m=19456,t=2,p=1$<salt>$<hash>
	// -> ["", "argon2id", "v=19", "m=19456,t=2,p=1", "<salt>", "<hash>"]
	parts := strings.Split(encoded, "$")
	if len(parts) != 6 || parts[1] != "argon2id" {
		return params{}, nil, nil, ErrInvalidHash
	}

	var version int
	if _, err := fmt.Sscanf(parts[2], "v=%d", &version); err != nil {
		return params{}, nil, nil, ErrInvalidHash
	}
	if version != argon2.Version {
		return params{}, nil, nil, ErrIncompatibleVersion
	}

	var p params
	if _, err := fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &p.memory, &p.time, &p.parallelism); err != nil {
		return params{}, nil, nil, ErrInvalidHash
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return params{}, nil, nil, ErrInvalidHash
	}
	hash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return params{}, nil, nil, ErrInvalidHash
	}

	return p, salt, hash, nil
}
