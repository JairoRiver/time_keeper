package controller

import (
	"context"
	"errors"
	"fmt"

	db "github.com/JairoRiver/time_keeper/internal/repository/db/sqlc"
	"github.com/JairoRiver/time_keeper/internal/util"
	"github.com/JairoRiver/time_keeper/pkg/password"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

type UserResponse struct {
	UserId         uuid.UUID
	Email          string
	Role           string
	EmailValidated bool
	IsActive       bool
}

func formatUserResponse(user db.User) UserResponse {
	userResponse := UserResponse{
		UserId:         user.ID,
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
	//generate a cryptographically secure per-user JWT signing key
	secretKey, err := util.SecureRandomString(64)
	if err != nil {
		return UserResponse{}, fmt.Errorf("control CreateUser generate secret key error: %w", err)
	}

	//check if role have a valid value
	dbUserParams := db.CreateUserParams{SecretTokenKey: secretKey}
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
	switch params.GetType {
	case util.GetUserTypeId:
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

	case util.GetUserTypeEmail:
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

	default:
		return UserResponse{}, fmt.Errorf("control GetUser invalid get param type: %w", ErrInvalidGetParamType)
	}
}

// update user control method
type UpdateUserParams struct {
	Id             uuid.UUID
	Email          string
	Role           string
	EmailValidated pgtype.Bool
	IsActive       pgtype.Bool
	SecretKey      string
}

func (c *Control) UpdateUser(ctx context.Context, params UpdateUserParams) (UserResponse, error) {
	//check if the id are empty
	if params.Id == uuid.Nil {
		return UserResponse{}, fmt.Errorf("control UpdateUser Id are empty error: %w", ErrEmptyId)
	}

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
	if len(params.SecretKey) > 0 {
		dbUdateParam.SecretTokenKey = pgtype.Text{String: params.SecretKey, Valid: true}
	}

	user, err := c.repo.UpdateUser(ctx, dbUdateParam)
	if err != nil {
		return UserResponse{}, fmt.Errorf("control UpdateUser repo UpdateUser error: %w", err)
	}
	userResponse := formatUserResponse(user)
	return userResponse, nil
}

type UserKeyResponse struct {
	UserId    uuid.UUID
	SecretKey string
}

func (c *Control) GetUserSecretKey(ctx context.Context, userId uuid.UUID) (UserKeyResponse, error) {
	if userId == uuid.Nil {
		return UserKeyResponse{}, fmt.Errorf("control GetUserSecretKey userId are empty error: %w", ErrEmptyId)
	}

	user, err := c.repo.GetUserSecretById(ctx, userId)
	if err != nil {
		return UserKeyResponse{}, fmt.Errorf("control GetUserSecretKey GetUserSecretById error: %w", err)
	}
	userResponse := UserKeyResponse{UserId: user.ID, SecretKey: user.SecretTokenKey}
	return userResponse, nil
}

// dummyHash keeps AuthenticateUser's timing roughly constant when the email
// does not exist or has no local password, mitigating account enumeration.
var dummyHash, _ = password.Hash("timing-equalizer-not-a-real-password")

// isUniqueViolation reports whether err is a Postgres unique-constraint error
// (SQLSTATE 23505), e.g. the users_email_key index firing.
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

// Register user control method
type RegisterUserParams struct {
	Email    string
	Password string
	Role     string
}

// RegisterUser creates a user with an email/password credential. Returns
// ErrEmailTaken if the email is already registered.
func (c *Control) RegisterUser(ctx context.Context, params RegisterUserParams) (UserResponse, error) {
	if len(params.Email) == 0 {
		return UserResponse{}, ErrEmptyEmail
	}
	if params.Role != util.UserAdminRole && params.Role != util.UserDefauldRole {
		return UserResponse{}, fmt.Errorf("control RegisterUser invalid role error: %w", ErrInvalidRoleValue)
	}
	if err := password.Validate(params.Password); err != nil {
		return UserResponse{}, err
	}

	hash, err := password.Hash(params.Password)
	if err != nil {
		return UserResponse{}, fmt.Errorf("control RegisterUser hash password error: %w", err)
	}
	secretKey, err := util.SecureRandomString(64)
	if err != nil {
		return UserResponse{}, fmt.Errorf("control RegisterUser generate secret key error: %w", err)
	}

	user, err := c.repo.CreateUser(ctx, db.CreateUserParams{
		Email:          pgtype.Text{String: params.Email, Valid: true},
		Role:           params.Role,
		SecretTokenKey: secretKey,
		PasswordHash:   pgtype.Text{String: hash, Valid: true},
	})
	if err != nil {
		if isUniqueViolation(err) {
			return UserResponse{}, ErrEmailTaken
		}
		return UserResponse{}, fmt.Errorf("control RegisterUser repo CreateUser error: %w", err)
	}
	return formatUserResponse(user), nil
}

// Authenticate user control method
type AuthenticateUserParams struct {
	Email    string
	Password string
}

// AuthenticateUser verifies an email/password pair. Any failure (unknown email,
// wrong password, missing local password, inactive user) returns
// ErrInvalidCredentials, and equal work is always spent so the caller cannot
// distinguish the cases by timing.
func (c *Control) AuthenticateUser(ctx context.Context, params AuthenticateUserParams) (UserResponse, error) {
	creds, err := c.repo.GetUserCredentialsByEmail(ctx, params.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			_, _ = password.Verify(params.Password, dummyHash)
			return UserResponse{}, ErrInvalidCredentials
		}
		return UserResponse{}, fmt.Errorf("control AuthenticateUser GetUserCredentialsByEmail error: %w", err)
	}

	// User exists but has no local password (anonymous account).
	if !creds.PasswordHash.Valid {
		_, _ = password.Verify(params.Password, dummyHash)
		return UserResponse{}, ErrInvalidCredentials
	}

	ok, err := password.Verify(params.Password, creds.PasswordHash.String)
	if err != nil {
		return UserResponse{}, fmt.Errorf("control AuthenticateUser verify password error: %w", err)
	}
	if !ok || !creds.IsActive {
		return UserResponse{}, ErrInvalidCredentials
	}

	return c.GetUser(ctx, GetUserParams{GetType: util.GetUserTypeId, Value: creds.ID})
}

// Set password control method
type SetPasswordParams struct {
	UserId   uuid.UUID
	Email    string
	Password string
}

// SetPassword attaches an email/password credential to an existing user (e.g. an
// anonymous user turning their session into a real account). Returns
// ErrEmailTaken if the email belongs to another user.
func (c *Control) SetPassword(ctx context.Context, params SetPasswordParams) (UserResponse, error) {
	if params.UserId == uuid.Nil {
		return UserResponse{}, ErrEmptyId
	}
	if len(params.Email) == 0 {
		return UserResponse{}, ErrEmptyEmail
	}
	if err := password.Validate(params.Password); err != nil {
		return UserResponse{}, err
	}

	hash, err := password.Hash(params.Password)
	if err != nil {
		return UserResponse{}, fmt.Errorf("control SetPassword hash password error: %w", err)
	}

	user, err := c.repo.UpdateUser(ctx, db.UpdateUserParams{
		ID:           params.UserId,
		Email:        pgtype.Text{String: params.Email, Valid: true},
		PasswordHash: pgtype.Text{String: hash, Valid: true},
	})
	if err != nil {
		if isUniqueViolation(err) {
			return UserResponse{}, ErrEmailTaken
		}
		return UserResponse{}, fmt.Errorf("control SetPassword repo UpdateUser error: %w", err)
	}
	return formatUserResponse(user), nil
}
