package db

import (
	"context"
	"testing"

	"github.com/JairoRiver/time_keeper/internal/util"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
)

// Create user function for test proposes
func createRandomUser(t *testing.T) User {
	email := util.RandomEmail()
	userParam := CreateUserParams{
		Email:          pgtype.Text{String: email, Valid: true},
		Role:           util.UserDefauldRole,
		SecretTokenKey: util.RandomString(64),
	}
	user, err := testQueries.CreateUser(context.Background(), userParam)
	assert.NoError(t, err)
	assert.NotEmpty(t, user)
	assert.Equal(t, email, user.Email.String)
	assert.Equal(t, util.UserDefauldRole, user.Role)
	assert.True(t, user.IsActive)
	assert.False(t, user.EmailValidated)
	assert.Equal(t, userParam.SecretTokenKey, user.SecretTokenKey)
	assert.NotEmpty(t, user.CreatedAt)
	assert.Empty(t, user.UpdatedAt)
	assert.Zero(t, user.UserIdentityID)

	return user
}

func TestCreateUser(t *testing.T) {
	_ = createRandomUser(t)
}

func TestGetUserByEmail(t *testing.T) {
	user := createRandomUser(t)

	getUser, err := testQueries.GetUserByEmail(context.Background(), user.Email)
	assert.NoError(t, err)
	assert.Equal(t, user, getUser)
}

func TestGetUserById(t *testing.T) {
	user := createRandomUser(t)

	getUser, err := testQueries.GetUserById(context.Background(), user.ID)
	assert.NoError(t, err)
	assert.Equal(t, user, getUser)
}

func TestGetByIdentityId(t *testing.T) {
	user := createRandomUser(t)

	//Update and add IdentityId
	userIdentityId := uuid.New()
	updateParams := UpdateUserParams{
		UserIdentityID: pgtype.UUID{Bytes: userIdentityId, Valid: true},
		ID:             user.ID,
	}
	testQueries.UpdateUser(context.Background(), updateParams)

	getUser, err := testQueries.GetUserByIdentityId(context.Background(), pgtype.UUID{Bytes: userIdentityId, Valid: true})
	assert.NoError(t, err)
	assert.Equal(t, user.ID, getUser.ID)
	assert.Equal(t, userIdentityId.String(), getUser.UserIdentityID.String())
	assert.Equal(t, user.Email, getUser.Email)
	assert.Equal(t, user.Role, getUser.Role)
	assert.False(t, getUser.EmailValidated)
	assert.True(t, getUser.IsActive)
	assert.Equal(t, user.SecretTokenKey, getUser.SecretTokenKey)
}

func TestUpdateUser(t *testing.T) {
	user := createRandomUser(t)
	newEmail := util.RandomEmail()
	userIdentityId := uuid.New()
	newSecret := util.RandomString(64)
	updateParams := UpdateUserParams{
		ID:             user.ID,
		Role:           pgtype.Text{String: util.UserAdminRole, Valid: true},
		Email:          pgtype.Text{String: newEmail, Valid: true},
		EmailValidated: pgtype.Bool{Bool: true, Valid: true},
		IsActive:       pgtype.Bool{Bool: false, Valid: true},
		UserIdentityID: pgtype.UUID{Bytes: userIdentityId, Valid: true},
		SecretTokenKey: pgtype.Text{String: newSecret, Valid: true},
	}
	updatedUser, err := testQueries.UpdateUser(context.Background(), updateParams)
	assert.NoError(t, err)
	assert.Equal(t, user.ID, updatedUser.ID)
	assert.Equal(t, newEmail, updatedUser.Email.String)
	assert.Equal(t, util.UserAdminRole, updatedUser.Role)
	assert.True(t, updatedUser.EmailValidated)
	assert.False(t, updatedUser.IsActive)
	assert.Equal(t, userIdentityId.String(), updatedUser.UserIdentityID.String())
	assert.Equal(t, newSecret, updatedUser.SecretTokenKey)
}

func TestGetUserSecretById(t *testing.T) {
	user := createRandomUser(t)

	getUser, err := testQueries.GetUserSecretById(context.Background(), user.ID)
	assert.NoError(t, err)
	assert.Equal(t, user.ID, getUser.ID)
	assert.Equal(t, user.SecretTokenKey, getUser.SecretTokenKey)
}
