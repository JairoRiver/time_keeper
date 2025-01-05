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
	assert.Equal(t, user.UserIdentityID, userTypeId.UserIdentityID)
	assert.Equal(t, user.Email, userTypeId.Email)
	assert.Equal(t, user.Role, userTypeId.Role)
	assert.Equal(t, user.EmailValidated, userTypeId.EmailValidated)
	assert.Equal(t, user.IsActive, userTypeId.IsActive)

	//test get by identityId
	//type id invalid id format
	identityTypeInvalidIdParams := GetUserParams{GetType: util.GetUserTypeIndetityId, Value: util.RandomString(8)}
	userTypeIdentityInvalidId, err := testControl.GetUser(context.Background(), identityTypeInvalidIdParams)
	assert.Zero(t, userTypeIdentityInvalidId)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrInvalidIdType))

	//type id zero UUID
	identityTypeZeroIdParams := GetUserParams{GetType: util.GetUserTypeIndetityId, Value: uuid.Nil}
	userTypeIdentityZeroId, err := testControl.GetUser(context.Background(), identityTypeZeroIdParams)
	assert.Zero(t, userTypeIdentityZeroId)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, ErrEmptyId))

	//type identity id get user
	identityId := uuid.New()
	updateUserParams := UpdateUserParams{Id: user.UserId, UserIdentityID: identityId}
	updatedUser, err := testControl.UpdateUser(context.Background(), updateUserParams)
	assert.NoError(t, err)
	assert.Equal(t, user.UserId, updatedUser.UserId)
	assert.Equal(t, identityId, updatedUser.UserIdentityID)
	assert.Equal(t, user.Email, updatedUser.Email)
	assert.Equal(t, user.Role, updatedUser.Role)
	assert.Equal(t, user.EmailValidated, updatedUser.EmailValidated)
	assert.Equal(t, user.IsActive, updatedUser.IsActive)

	identityTypeParams := GetUserParams{GetType: util.GetUserTypeIndetityId, Value: identityId}
	userTypeIdentity, err := testControl.GetUser(context.Background(), identityTypeParams)
	assert.NoError(t, err)
	assert.Equal(t, user.UserId, userTypeIdentity.UserId)
	assert.Equal(t, identityId, userTypeIdentity.UserIdentityID)
	assert.Equal(t, user.Email, userTypeIdentity.Email)
	assert.Equal(t, user.Role, userTypeIdentity.Role)
	assert.Equal(t, user.EmailValidated, userTypeIdentity.EmailValidated)
	assert.Equal(t, user.IsActive, userTypeIdentity.IsActive)

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
	assert.Equal(t, identityId, userTypeEmail.UserIdentityID)
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

	newEmail := util.RandomEmail()
	newRole := util.UserAdminRole
	newIdentityId := uuid.New()

	updateParams := UpdateUserParams{
		Id:             user.UserId,
		Email:          newEmail,
		Role:           newRole,
		UserIdentityID: newIdentityId,
		EmailValidated: pgtype.Bool{Bool: true, Valid: true},
		IsActive:       pgtype.Bool{Bool: false, Valid: true},
	}
	updatedUser, err := testControl.UpdateUser(context.Background(), updateParams)
	assert.NoError(t, err)
	assert.NotZero(t, updatedUser)
	assert.Equal(t, user.UserId, updatedUser.UserId)
	assert.Equal(t, newEmail, updatedUser.Email)
	assert.Equal(t, newRole, updatedUser.Role)
	assert.Equal(t, newIdentityId, updatedUser.UserIdentityID)
	assert.True(t, updatedUser.EmailValidated)
	assert.False(t, updatedUser.IsActive)
}
