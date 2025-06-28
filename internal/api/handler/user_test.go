package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/JairoRiver/time_keeper/internal/controller"
	"github.com/JairoRiver/time_keeper/internal/util"
	"github.com/JairoRiver/time_keeper/pkg/token"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateUser_Success(t *testing.T) {
	e := echo.New()

	mockCtrl := new(MockController)
	h := newTestHandler(mockCtrl)

	userID := uuid.New()
	secretKey := util.RandomString(64)

	expectedUser := controller.UserResponse{
		UserId: userID,
		Role:   util.UserDefauldRole,
	}

	// Mock expectations
	mockCtrl.On("CreateUser", mock.Anything, controller.CreateUserParam{Role: util.UserDefauldRole}).Return(expectedUser, nil)
	mockCtrl.On("GetUserSecretKey", mock.Anything, userID).Return(controller.UserKeyResponse{UserId: userID, SecretKey: secretKey}, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/user", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if assert.NoError(t, h.CreateUser(c)) {
		assert.Equal(t, http.StatusCreated, rec.Code)

		var resp ResponseUser
		if assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp)) {
			assert.Equal(t, expectedUser.UserId, resp.UserId)
			assert.NotEmpty(t, resp.AccessToken)
			assert.WithinDuration(t, time.Now().Add(accessTokenDuration), resp.ExpiredAt, time.Minute)
		}

		// Verify refresh token cookie exists
		cookie := rec.Result().Cookies()[0]
		assert.Equal(t, util.RefreshTokenName, cookie.Name)
		assert.NotEmpty(t, cookie.Value)
	}

	mockCtrl.AssertExpectations(t)
}

func TestRefreshToken_Success(t *testing.T) {
	e := echo.New()

	mockCtrl := new(MockController)
	h := newTestHandler(mockCtrl)

	userID := uuid.New()
	role := util.UserDefauldRole
	secretKey := util.RandomString(64)

	// create initial refresh token to add as cookie (not verified in handler but good for completeness)
	tokenMaker, _ := token.NewJWTMaker(secretKey)
	refreshToken, _, _ := tokenMaker.CreateToken(userID, role, refreshTokenDuration)

	// Mock expectations
	mockCtrl.On("GetUserSecretKey", mock.Anything, userID).Return(controller.UserKeyResponse{UserId: userID, SecretKey: secretKey}, nil)
	mockCtrl.On("GetUser", mock.Anything, controller.GetUserParams{GetType: util.GetUserTypeId, Value: userID}).Return(controller.UserResponse{UserId: userID, Role: role}, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/refresh", nil)
	req.AddCookie(&http.Cookie{Name: util.RefreshTokenName, Value: refreshToken})

	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	addCookiePayload(c, userID, role)

	if assert.NoError(t, h.RefreshToken(c)) {
		assert.Equal(t, http.StatusCreated, rec.Code)

		var resp ResponseUser
		if assert.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp)) {
			assert.Equal(t, userID, resp.UserId)
			assert.NotEmpty(t, resp.AccessToken)
		}
		// Verify new refresh token cookie exists
		cookie := rec.Result().Cookies()[0]
		assert.Equal(t, util.RefreshTokenName, cookie.Name)
		assert.NotEmpty(t, cookie.Value)
	}

	mockCtrl.AssertExpectations(t)
}
