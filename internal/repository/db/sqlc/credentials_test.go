package db

import (
	"context"
	"strings"
	"testing"

	"github.com/JairoRiver/time_keeper/internal/util"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// a syntactically valid-looking argon2id PHC string (contents are irrelevant to
// the DB layer, which only stores/returns it verbatim).
const testHash = "$argon2id$v=19$m=19456,t=2,p=1$c2FsdHNhbHQwMTIzNDU2$aGFzaGhhc2hoYXNoaGFzaGhhc2hoYXNo"

func TestCreateUserWithPasswordHash(t *testing.T) {
	email := util.RandomEmail()
	user, err := testQueries.CreateUser(context.Background(), CreateUserParams{
		Email:          pgtype.Text{String: email, Valid: true},
		Role:           util.UserDefauldRole,
		SecretTokenKey: util.RandomString(64),
		PasswordHash:   pgtype.Text{String: testHash, Valid: true},
	})
	require.NoError(t, err)
	assert.Equal(t, testHash, user.PasswordHash.String)

	// Lookup is case-insensitive (unique index on lower(email)).
	creds, err := testQueries.GetUserCredentialsByEmail(context.Background(), strings.ToUpper(email))
	require.NoError(t, err)
	assert.Equal(t, user.ID, creds.ID)
	assert.Equal(t, testHash, creds.PasswordHash.String)
	assert.Equal(t, util.UserDefauldRole, creds.Role)
	assert.True(t, creds.IsActive)
}

func TestUpdateUserPasswordHash(t *testing.T) {
	// createRandomUser creates a user with a NULL password_hash.
	user := createRandomUser(t)
	assert.False(t, user.PasswordHash.Valid, "new user should have no password hash")

	updated, err := testQueries.UpdateUser(context.Background(), UpdateUserParams{
		ID:           user.ID,
		PasswordHash: pgtype.Text{String: testHash, Valid: true},
	})
	require.NoError(t, err)
	assert.Equal(t, testHash, updated.PasswordHash.String)

	creds, err := testQueries.GetUserCredentialsByEmail(context.Background(), user.Email.String)
	require.NoError(t, err)
	assert.Equal(t, testHash, creds.PasswordHash.String)
}

func TestGetUserCredentialsByEmailNotFound(t *testing.T) {
	_, err := testQueries.GetUserCredentialsByEmail(context.Background(), "nobody-"+util.RandomString(8)+"@nowhere.test")
	assert.ErrorIs(t, err, pgx.ErrNoRows)
}
