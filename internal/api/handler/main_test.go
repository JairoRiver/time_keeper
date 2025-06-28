package handler

import (
	"context"

	"github.com/JairoRiver/time_keeper/internal/controller"
	"github.com/JairoRiver/time_keeper/internal/util"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/mock"
)

// ----------------------- MOCKS ----------------------- //
// MockController is a testify mock for controller.Controller interface

type MockController struct {
	mock.Mock
}

func (m *MockController) CreateEntryTime(ctx context.Context, params controller.CreateEntryTimeParams) (controller.EntryTimeResponse, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(controller.EntryTimeResponse), args.Error(1)
}

func (m *MockController) GetEntryTime(ctx context.Context, id uuid.UUID) (controller.EntryTimeResponse, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(controller.EntryTimeResponse), args.Error(1)
}

func (m *MockController) ListEntryTime(ctx context.Context, params controller.ListEntryTimeParams) ([]controller.EntryTimeResponse, error) {
	args := m.Called(ctx, params)
	return args.Get(0).([]controller.EntryTimeResponse), args.Error(1)
}

func (m *MockController) UpdateEntryTime(ctx context.Context, params controller.UpdateEntryTimeParams) (controller.EntryTimeResponse, error) {
	args := m.Called(ctx, params)
	return args.Get(0).(controller.EntryTimeResponse), args.Error(1)
}

func (m *MockController) DeleteEntryTime(ctx context.Context, id uuid.UUID) (controller.EntryTimeResponse, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(controller.EntryTimeResponse), args.Error(1)
}

// The following methods are required to satisfy the controller.Controller interface
// but are not used in these entry time handler tests; therefore they return zero values.

// User-related stubs to satisfy the interface
func (m *MockController) CreateUser(ctx context.Context, params controller.CreateUserParam) (controller.UserResponse, error) {
	args := m.Called(ctx, params)
	// if no expectation set, return zero values
	if args.Get(0) != nil {
		return args.Get(0).(controller.UserResponse), args.Error(1)
	}
	return controller.UserResponse{}, args.Error(1)
}

func (m *MockController) GetUser(ctx context.Context, params controller.GetUserParams) (controller.UserResponse, error) {
	args := m.Called(ctx, params)
	if args.Get(0) != nil {
		return args.Get(0).(controller.UserResponse), args.Error(1)
	}
	return controller.UserResponse{}, args.Error(1)
}

func (m *MockController) GetUserSecretKey(ctx context.Context, userId uuid.UUID) (controller.UserKeyResponse, error) {
	args := m.Called(ctx, userId)
	if args.Get(0) != nil {
		return args.Get(0).(controller.UserKeyResponse), args.Error(1)
	}
	return controller.UserKeyResponse{}, args.Error(1)
}

func (m *MockController) UpdateUser(ctx context.Context, params controller.UpdateUserParams) (controller.UserResponse, error) {
	args := m.Called(ctx, params)
	if args.Get(0) != nil {
		return args.Get(0).(controller.UserResponse), args.Error(1)
	}
	return controller.UserResponse{}, args.Error(1)
}

func (m *MockController) GetEntryTimeOwner(ctx context.Context, entryTimeId uuid.UUID) (controller.EntryTimeOwnerResponse, error) {
	args := m.Called(ctx, entryTimeId)
	if args.Get(0) != nil {
		return args.Get(0).(controller.EntryTimeOwnerResponse), args.Error(1)
	}
	return controller.EntryTimeOwnerResponse{}, args.Error(1)
}

// ----------------------- HELPERS ----------------------- //

// newTestHandler returns a Handler with a mocked Controller.
func newTestHandler(mockCtrl *MockController) *Handler {
	return &Handler{ctrl: mockCtrl}
}

// addAuthPayload sets a fake authorization payload into the echo context so that
// the handler can retrieve the user information without running the real middleware.
func addAuthPayload(c echo.Context, userID uuid.UUID) {
	userInfo := UserInfo{
		UserId: userID,
		Role:   "user",
	}
	c.Set(authorizationPayloadKey, userInfo)
}

func addCookiePayload(c echo.Context, userID uuid.UUID, role string) {
	userInfo := UserInfo{
		UserId: userID,
		Role:   role,
	}
	c.Set(util.RefreshTokenName, userInfo)
}
