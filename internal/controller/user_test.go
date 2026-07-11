package controller

import (
	"context"
	"errors"
	"testing"

	"github.com/JairoRiver/time_keeper/internal/util"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
)

// Create user function for test proposes on controller package
func createRandomUser(t *testing.T, email string) UserResponse {
	//invalid role params
	invalidUserParams := CreateUserParam{
		Role: util.RandomString(4),
	}
	invalidRoleUser, err := testControl.CreateUser(context.Background(), invalidUserParams)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidRoleValue))
	assert.Zero(t, invalidRoleUser)

	//create valid user
	createUserParams := CreateUserParam{
		Role:  util.UserDefauldRole,
		Email: email,
	}
	user, err := testControl.CreateUser(context.Background(), createUserParams)
	assert.NoError(t, err)
	assert.NotZero(t, user)
	assert.Equal(t, user.Role, createUserParams.Role)
	assert.Equal(t, user.Email, createUserParams.Email)
	return user
}

func TestCreateUser(t *testing.T) {
	_ = createRandomUser(t, "")
}

func TestGetUser(t *testing.T) {
	email := util.RandomEmail()
	user := createRandomUser(t, email)

	//test get by Id
	//type id invalid id format
	idTypeInvalidIdParams := GetUserParams{GetType: util.GetUserTypeId, Value: util.RandomString(8)}
	userTypeIdInvalidId, err := testControl.GetUser(context.Background(), idTypeInvalidIdParams)
	assert.Zero(t, userTypeIdInvalidId)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidIdType))

	//type id zero UUID
	idTypeZeroIdParams := GetUserParams{GetType: util.GetUserTypeId, Value: uuid.Nil}
	userTypeIdZeroId, err := testControl.GetUser(context.Background(), idTypeZeroIdParams)
	assert.Zero(t, userTypeIdZeroId)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrEmptyId))

	//type id get user
	idTypeParams := GetUserParams{GetType: util.GetUserTypeId, Value: user.UserId}
	userTypeId, err := testControl.GetUser(context.Background(), idTypeParams)
	assert.NoError(t, err)
	assert.Equal(t, user.UserId, userTypeId.UserId)
	assert.Equal(t, user.Email, userTypeId.Email)
	assert.Equal(t, user.Role, userTypeId.Role)
	assert.Equal(t, user.EmailValidated, userTypeId.EmailValidated)
	assert.Equal(t, user.IsActive, userTypeId.IsActive)

	//test get by email
	//type id invalid email format
	emailTypeInvalidIdParams := GetUserParams{GetType: util.GetUserTypeEmail, Value: uuid.New()}
	userTypeEmailInvalidId, err := testControl.GetUser(context.Background(), emailTypeInvalidIdParams)
	assert.Zero(t, userTypeEmailInvalidId)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidEmailType))

	//type email zero string
	emailTypeZeroEmailParams := GetUserParams{GetType: util.GetUserTypeEmail, Value: ""}
	userTypeEmailZeroEmail, err := testControl.GetUser(context.Background(), emailTypeZeroEmailParams)
	assert.Zero(t, userTypeEmailZeroEmail)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrEmptyEmail))

	//type email get user
	emailTypeParams := GetUserParams{GetType: util.GetUserTypeEmail, Value: user.Email}
	userTypeEmail, err := testControl.GetUser(context.Background(), emailTypeParams)
	assert.NoError(t, err)
	assert.Equal(t, user.UserId, userTypeEmail.UserId)
	assert.Equal(t, user.Email, userTypeEmail.Email)
	assert.Equal(t, user.Role, userTypeEmail.Role)
	assert.Equal(t, user.EmailValidated, userTypeEmail.EmailValidated)
	assert.Equal(t, user.IsActive, userTypeEmail.IsActive)
}

func TestUpdateUser(t *testing.T) {
	user := createRandomUser(t, "")
	//error id are empty
	errorUpdateParams := UpdateUserParams{}
	errorUpdatedUser, err := testControl.UpdateUser(context.Background(), errorUpdateParams)
	assert.Zero(t, errorUpdatedUser)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrEmptyId))

	//update valid inputs
	newEmail := util.RandomEmail()
	newRole := util.UserAdminRole
	newSecretKey := util.RandomString(64)

	updateParams := UpdateUserParams{
		Id:             user.UserId,
		Email:          newEmail,
		Role:           newRole,
		EmailValidated: pgtype.Bool{Bool: true, Valid: true},
		IsActive:       pgtype.Bool{Bool: false, Valid: true},
		SecretKey:      newSecretKey,
	}
	updatedUser, err := testControl.UpdateUser(context.Background(), updateParams)
	assert.NoError(t, err)
	assert.NotZero(t, updatedUser)
	assert.Equal(t, user.UserId, updatedUser.UserId)
	assert.Equal(t, newEmail, updatedUser.Email)
	assert.Equal(t, newRole, updatedUser.Role)
	assert.True(t, updatedUser.EmailValidated)
	assert.False(t, updatedUser.IsActive)

	//check secret updated
	checkSecret, err := testControl.GetUserSecretKey(context.Background(), user.UserId)
	assert.NoError(t, err)
	assert.NotZero(t, checkSecret)
	assert.Equal(t, user.UserId, checkSecret.UserId)
	assert.Equal(t, newSecretKey, checkSecret.SecretKey)
}

func TestGetUserSecretKey(t *testing.T) {
	user := createRandomUser(t, "")

	//check id zero UUID
	userZeroId, err := testControl.GetUserSecretKey(context.Background(), uuid.Nil)
	assert.Zero(t, userZeroId)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrEmptyId))

	//check valid id
	userSecret, err := testControl.GetUserSecretKey(context.Background(), user.UserId)
	assert.NoError(t, err)
	assert.NotNil(t, userSecret)
	assert.Equal(t, user.UserId, userSecret.UserId)
	assert.NotZero(t, userSecret.SecretKey)
}
