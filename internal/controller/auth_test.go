package controller

import (
	"context"
	"strings"
	"testing"

	"github.com/JairoRiver/time_keeper/internal/util"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testPassword = "sup3r-s3cret-passw0rd"

func TestRegisterUser(t *testing.T) {
	email := util.RandomEmail()
	user, err := testControl.RegisterUser(context.Background(), RegisterUserParams{
		Email:    email,
		Password: testPassword,
		Role:     util.UserDefauldRole,
	})
	require.NoError(t, err)
	assert.Equal(t, email, user.Email)
	assert.Equal(t, util.UserDefauldRole, user.Role)
	assert.NotEqual(t, uuid.Nil, user.UserId)
}

func TestRegisterUserDuplicateEmail(t *testing.T) {
	email := util.RandomEmail()
	_, err := testControl.RegisterUser(context.Background(), RegisterUserParams{
		Email: email, Password: testPassword, Role: util.UserDefauldRole,
	})
	require.NoError(t, err)

	// Same email, different case → still rejected.
	_, err = testControl.RegisterUser(context.Background(), RegisterUserParams{
		Email: strings.ToUpper(email), Password: testPassword, Role: util.UserDefauldRole,
	})
	assert.ErrorIs(t, err, ErrEmailTaken)
}

func TestRegisterUserInvalidRole(t *testing.T) {
	_, err := testControl.RegisterUser(context.Background(), RegisterUserParams{
		Email: util.RandomEmail(), Password: testPassword, Role: "bogus",
	})
	assert.ErrorIs(t, err, ErrInvalidRoleValue)
}

func TestRegisterUserWeakPassword(t *testing.T) {
	_, err := testControl.RegisterUser(context.Background(), RegisterUserParams{
		Email: util.RandomEmail(), Password: "short", Role: util.UserDefauldRole,
	})
	assert.Error(t, err) // password.ErrPasswordTooShort
}

func TestAuthenticateUserSuccess(t *testing.T) {
	email := util.RandomEmail()
	registered, err := testControl.RegisterUser(context.Background(), RegisterUserParams{
		Email: email, Password: testPassword, Role: util.UserDefauldRole,
	})
	require.NoError(t, err)

	// Case-insensitive login.
	got, err := testControl.AuthenticateUser(context.Background(), AuthenticateUserParams{
		Email: strings.ToUpper(email), Password: testPassword,
	})
	require.NoError(t, err)
	assert.Equal(t, registered.UserId, got.UserId)
}

func TestAuthenticateUserWrongPassword(t *testing.T) {
	email := util.RandomEmail()
	_, err := testControl.RegisterUser(context.Background(), RegisterUserParams{
		Email: email, Password: testPassword, Role: util.UserDefauldRole,
	})
	require.NoError(t, err)

	_, err = testControl.AuthenticateUser(context.Background(), AuthenticateUserParams{
		Email: email, Password: "wrong-password",
	})
	assert.ErrorIs(t, err, ErrInvalidCredentials)
}

func TestAuthenticateUserUnknownEmail(t *testing.T) {
	_, err := testControl.AuthenticateUser(context.Background(), AuthenticateUserParams{
		Email: "ghost-" + util.RandomString(8) + "@nowhere.test", Password: testPassword,
	})
	assert.ErrorIs(t, err, ErrInvalidCredentials)
}

func TestAuthenticateUserNoLocalPassword(t *testing.T) {
	// A user with an email but no password_hash (e.g. anonymous) must not
	// be able to authenticate with a password.
	email := util.RandomEmail()
	_, err := testControl.CreateUser(context.Background(), CreateUserParam{
		Email: email, Role: util.UserDefauldRole,
	})
	require.NoError(t, err)

	_, err = testControl.AuthenticateUser(context.Background(), AuthenticateUserParams{
		Email: email, Password: testPassword,
	})
	assert.ErrorIs(t, err, ErrInvalidCredentials)
}

func TestAuthenticateUserInactive(t *testing.T) {
	email := util.RandomEmail()
	registered, err := testControl.RegisterUser(context.Background(), RegisterUserParams{
		Email: email, Password: testPassword, Role: util.UserDefauldRole,
	})
	require.NoError(t, err)

	_, err = testControl.UpdateUser(context.Background(), UpdateUserParams{
		Id:       registered.UserId,
		IsActive: pgtype.Bool{Bool: false, Valid: true},
	})
	require.NoError(t, err)

	_, err = testControl.AuthenticateUser(context.Background(), AuthenticateUserParams{
		Email: email, Password: testPassword,
	})
	assert.ErrorIs(t, err, ErrInvalidCredentials)
}

func TestSetPassword(t *testing.T) {
	// Anonymous user (no email, no password).
	anon, err := testControl.CreateUser(context.Background(), CreateUserParam{
		Role: util.UserDefauldRole,
	})
	require.NoError(t, err)

	email := util.RandomEmail()
	_, err = testControl.SetPassword(context.Background(), SetPasswordParams{
		UserId: anon.UserId, Email: email, Password: testPassword,
	})
	require.NoError(t, err)

	// The same user can now authenticate and keeps its ID (entries preserved).
	got, err := testControl.AuthenticateUser(context.Background(), AuthenticateUserParams{
		Email: email, Password: testPassword,
	})
	require.NoError(t, err)
	assert.Equal(t, anon.UserId, got.UserId)
}

func TestSetPasswordDuplicateEmail(t *testing.T) {
	email := util.RandomEmail()
	_, err := testControl.RegisterUser(context.Background(), RegisterUserParams{
		Email: email, Password: testPassword, Role: util.UserDefauldRole,
	})
	require.NoError(t, err)

	anon, err := testControl.CreateUser(context.Background(), CreateUserParam{Role: util.UserDefauldRole})
	require.NoError(t, err)

	_, err = testControl.SetPassword(context.Background(), SetPasswordParams{
		UserId: anon.UserId, Email: email, Password: testPassword,
	})
	assert.ErrorIs(t, err, ErrEmailTaken)
}
