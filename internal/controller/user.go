package controller

import (
	"context"
	"fmt"

	db "github.com/JairoRiver/time_keeper/internal/repository/db/sqlc"
	"github.com/JairoRiver/time_keeper/internal/util"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type UserResponse struct {
	UserId         uuid.UUID
	UserIdentityID uuid.UUID
	Email          string
	Role           string
	EmailValidated bool
	IsActive       bool
}

func formatUserResponse(user db.User) UserResponse {
	userResponse := UserResponse{
		UserId:         user.ID,
		UserIdentityID: user.UserIdentityID.Bytes,
		Email:          user.Email.String,
		Role:           user.Role,
		EmailValidated: user.EmailValidated,
		IsActive:       user.IsActive,
	}
	return userResponse
}

// Create User Control method
type CreateUserParam struct {
	Email string
	Role  string
}

func (c *Control) CreateUser(ctx context.Context, params CreateUserParam) (UserResponse, error) {
	//check if role have a valid value
	var dbUserParams db.CreateUserParams
	if params.Role == util.UserAdminRole || params.Role == util.UserDefauldRole {
		dbUserParams.Role = params.Role
	} else {
		return UserResponse{}, fmt.Errorf("control CreateUser invalid role error: %w", ErrInvalidRoleValue)
	}

	//check if email is empty string
	if len(params.Email) == 0 {
		dbUserParams.Email = pgtype.Text{Valid: false}
	} else {
		dbUserParams.Email = pgtype.Text{String: params.Email, Valid: true}
	}

	//create user
	user, err := c.repo.CreateUser(ctx, dbUserParams)
	if err != nil {
		return UserResponse{}, fmt.Errorf("control CreateUser repo CreateUser error: %w", err)
	}
	userResponse := formatUserResponse(user)
	return userResponse, nil
}

// Get User Control Method
type GetUserParams struct {
	GetType string
	Value   interface{}
}

func (c *Control) GetUser(ctx context.Context, params GetUserParams) (UserResponse, error) {
	//get user by user ID
	if params.GetType == util.GetUserTypeId {
		//validated if the param value are an UUID
		if id, ok := params.Value.(uuid.UUID); ok {
			//check if the id are empty
			if id == uuid.Nil {
				return UserResponse{}, fmt.Errorf("control GetUser Id type Id are empty error: %w", ErrEmptyId)
			}

			user, err := c.repo.GetUserById(ctx, id)
			if err != nil {
				return UserResponse{}, fmt.Errorf("control GetUser Id type repo GetUserById error: %w", err)
			}
			userResponse := formatUserResponse(user)
			return userResponse, nil
		}
		return UserResponse{}, fmt.Errorf("control GetUser Id type invalid Id type: %w", ErrInvalidIdType)

	} else if params.GetType == util.GetUserTypeIndetityId {
		//validated if the param value are an UUID
		if id, ok := params.Value.(uuid.UUID); ok {
			//check if the IdentityId are empty
			if id == uuid.Nil {
				return UserResponse{}, fmt.Errorf("control GetUser IdentityId Type IdIdentityId are empty error: %w", ErrEmptyId)
			}

			user, err := c.repo.GetUserByIdentityId(ctx, pgtype.UUID{Bytes: id, Valid: true})
			if err != nil {
				return UserResponse{}, fmt.Errorf("control GetUser IdentityId type repo GetUserByIdentityId error: %w", err)
			}
			userResponse := formatUserResponse(user)
			return userResponse, nil
		}
		return UserResponse{}, fmt.Errorf("control GetUser IdentityId type invalid Id type: %w", ErrInvalidIdType)

	} else if params.GetType == util.GetUserTypeEmail {
		// validated if the param value are a string type
		if email, ok := params.Value.(string); ok {
			//check if the email are empty
			if len(email) == 0 {
				return UserResponse{}, fmt.Errorf("control GetUser email type email are empty error: %w", ErrEmptyEmail)
			}

			user, err := c.repo.GetUserByEmail(ctx, pgtype.Text{String: email, Valid: true})
			if err != nil {
				return UserResponse{}, fmt.Errorf("control GetUser email type repo GetUserByEmail error: %w", err)
			}
			userResponse := formatUserResponse(user)
			return userResponse, nil
		}
		return UserResponse{}, fmt.Errorf("control GetUser email type invalid email type: %w", ErrInvalidEmailType)

	}
	return UserResponse{}, fmt.Errorf("control GetUser invalid get param type: %w", ErrInvalidGetParamType)
}

// update user control method
type UpdateUserParams struct {
	Id             uuid.UUID
	Email          string
	Role           string
	UserIdentityID uuid.UUID
	EmailValidated pgtype.Bool
	IsActive       pgtype.Bool
}

func (c *Control) UpdateUser(ctx context.Context, params UpdateUserParams) (UserResponse, error) {
	dbUdateParam := db.UpdateUserParams{
		ID:             params.Id,
		EmailValidated: params.EmailValidated,
		IsActive:       params.IsActive,
	}

	// check if the values are not empty
	if len(params.Email) > 0 {
		dbUdateParam.Email = pgtype.Text{String: params.Email, Valid: true}
	}
	if len(params.Role) > 0 {
		dbUdateParam.Role = pgtype.Text{String: params.Role, Valid: true}
	}
	if params.UserIdentityID != uuid.Nil {
		dbUdateParam.UserIdentityID = pgtype.UUID{Bytes: params.UserIdentityID, Valid: true}
	}

	user, err := c.repo.UpdateUser(ctx, dbUdateParam)
	if err != nil {
		return UserResponse{}, fmt.Errorf("control UpdateUser repo UpdateUser error: %w", err)
	}
	userResponse := formatUserResponse(user)
	return userResponse, nil
}
