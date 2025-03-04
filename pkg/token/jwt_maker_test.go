package token

import (
	"testing"
	"time"

	"github.com/JairoRiver/time_keeper/internal/util"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestJWTMaker(t *testing.T) {
	maker, err := NewJWTMaker(util.RandomString(64))
	assert.NoError(t, err)

	userId := uuid.New()
	role := util.UserDefauldRole
	duration := time.Minute

	issuedAt := time.Now()
	expiredAt := issuedAt.Add(duration)

	token, payload, err := maker.CreateToken(userId, role, duration)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.NotEmpty(t, payload)

	payload, err = maker.VerifyToken(token)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	assert.NotZero(t, payload.ID)
	assert.Equal(t, userId, payload.UserId)
	assert.Equal(t, role, payload.Role)
	assert.WithinDuration(t, issuedAt, payload.IssuedAt.Time.Local(), time.Second)
	assert.WithinDuration(t, expiredAt, payload.ExpiresAt.Time.Local(), time.Second)
}

func TestExpiredJWTToken(t *testing.T) {
	maker, err := NewJWTMaker(util.RandomString(64))
	assert.NoError(t, err)

	token, payload, err := maker.CreateToken(uuid.New(), util.UserDefauldRole, -time.Minute)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
	assert.NotEmpty(t, payload)

	payload, err = maker.VerifyToken(token)
	assert.Error(t, err)
	assert.EqualError(t, err, ErrExpiredToken.Error())
	assert.Nil(t, payload)
}

func TestInvalidJWTTokenAlgNone(t *testing.T) {
	payload, err := NewPayload(uuid.New(), util.UserDefauldRole, time.Minute)
	assert.NoError(t, err)

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodNone, payload)
	token, err := jwtToken.SignedString(jwt.UnsafeAllowNoneSignatureType)
	assert.NoError(t, err)

	maker, err := NewJWTMaker(util.RandomString(64))
	assert.NoError(t, err)

	payload2, err := maker.VerifyToken(token)
	assert.Error(t, err)
	assert.EqualError(t, err, ErrInvalidToken.Error())
	assert.Nil(t, payload2)
}
