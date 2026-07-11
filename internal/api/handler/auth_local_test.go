package handler

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/JairoRiver/time_keeper/internal/controller"
	"github.com/JairoRiver/time_keeper/internal/util"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func postForm(h *Handler, handler echo.HandlerFunc, form url.Values) *httptest.ResponseRecorder {
	return postFormAs(handler, form, nil)
}

// postFormAs posts a form. When userID is non-nil, an authenticated UserInfo is
// injected into the context (as PageAuthMiddleware/CookieMiddleware would).
func postFormAs(handler echo.HandlerFunc, form url.Values, userID *uuid.UUID) *httptest.ResponseRecorder {
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(form.Encode()))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationForm)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	if userID != nil {
		addCookiePayload(c, *userID, util.UserDefauldRole)
	}
	_ = handler(c)
	return rec
}

func TestLoginSubmit_Success(t *testing.T) {
	mockCtrl := new(MockController)
	h := newTestHandler(mockCtrl)

	userID := uuid.New()
	secretKey := util.RandomString(64)
	mockCtrl.On("AuthenticateUser", mock.Anything, controller.AuthenticateUserParams{
		Email: "user@test.com", Password: "sup3rsecret123",
	}).Return(controller.UserResponse{UserId: userID, Role: util.UserDefauldRole}, nil)
	mockCtrl.On("GetUserSecretKey", mock.Anything, userID).
		Return(controller.UserKeyResponse{UserId: userID, SecretKey: secretKey}, nil)

	rec := postForm(h, h.LoginSubmit, url.Values{
		"email":    {"user@test.com"},
		"password": {"sup3rsecret123"},
	})

	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, "/registro", rec.Header().Get("Location"))
	assert.Equal(t, util.RefreshTokenName, rec.Result().Cookies()[0].Name)
	mockCtrl.AssertExpectations(t)
}

func TestLoginSubmit_InvalidCredentials(t *testing.T) {
	mockCtrl := new(MockController)
	h := newTestHandler(mockCtrl)

	mockCtrl.On("AuthenticateUser", mock.Anything, mock.Anything).
		Return(controller.UserResponse{}, controller.ErrInvalidCredentials)

	rec := postForm(h, h.LoginSubmit, url.Values{
		"email":    {"user@test.com"},
		"password": {"wrongpass1"},
	})

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.Empty(t, rec.Result().Cookies())
	assert.Contains(t, rec.Body.String(), "incorrectos")
	mockCtrl.AssertExpectations(t)
}

func TestRegisterSubmit_Success(t *testing.T) {
	mockCtrl := new(MockController)
	h := newTestHandler(mockCtrl)

	userID := uuid.New()
	secretKey := util.RandomString(64)
	mockCtrl.On("RegisterUser", mock.Anything, controller.RegisterUserParams{
		Email: "new@test.com", Password: "sup3rsecret123", Role: util.UserDefauldRole,
	}).Return(controller.UserResponse{UserId: userID, Role: util.UserDefauldRole}, nil)
	mockCtrl.On("GetUserSecretKey", mock.Anything, userID).
		Return(controller.UserKeyResponse{UserId: userID, SecretKey: secretKey}, nil)

	rec := postForm(h, h.RegisterSubmit, url.Values{
		"email":            {"new@test.com"},
		"password":         {"sup3rsecret123"},
		"password_confirm": {"sup3rsecret123"},
	})

	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, "/registro", rec.Header().Get("Location"))
	assert.Equal(t, util.RefreshTokenName, rec.Result().Cookies()[0].Name)
	mockCtrl.AssertExpectations(t)
}

func TestRegisterSubmit_PasswordMismatch(t *testing.T) {
	mockCtrl := new(MockController)
	h := newTestHandler(mockCtrl)

	rec := postForm(h, h.RegisterSubmit, url.Values{
		"email":            {"new@test.com"},
		"password":         {"sup3rsecret123"},
		"password_confirm": {"different123"},
	})

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "no coinciden")
	// RegisterUser must not be called when passwords don't match.
	mockCtrl.AssertNotCalled(t, "RegisterUser", mock.Anything, mock.Anything)
}

func TestRegisterSubmit_EmailTaken(t *testing.T) {
	mockCtrl := new(MockController)
	h := newTestHandler(mockCtrl)

	mockCtrl.On("RegisterUser", mock.Anything, mock.Anything).
		Return(controller.UserResponse{}, controller.ErrEmailTaken)

	rec := postForm(h, h.RegisterSubmit, url.Values{
		"email":            {"taken@test.com"},
		"password":         {"sup3rsecret123"},
		"password_confirm": {"sup3rsecret123"},
	})

	assert.Equal(t, http.StatusConflict, rec.Code)
	assert.Contains(t, rec.Body.String(), "ya está registrado")
	mockCtrl.AssertExpectations(t)
}

func TestLinkSubmit_Success(t *testing.T) {
	mockCtrl := new(MockController)
	h := newTestHandler(mockCtrl)
	userID := uuid.New()

	mockCtrl.On("SetPassword", mock.Anything, controller.SetPasswordParams{
		UserId: userID, Email: "anon@test.com", Password: "sup3rsecret123",
	}).Return(controller.UserResponse{UserId: userID}, nil)

	rec := postFormAs(h.LinkSubmit, url.Values{
		"email":            {"anon@test.com"},
		"password":         {"sup3rsecret123"},
		"password_confirm": {"sup3rsecret123"},
	}, &userID)

	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, "/registro", rec.Header().Get("Location"))
	mockCtrl.AssertExpectations(t)
}

func TestLinkSubmit_PasswordMismatch(t *testing.T) {
	mockCtrl := new(MockController)
	h := newTestHandler(mockCtrl)
	userID := uuid.New()

	rec := postFormAs(h.LinkSubmit, url.Values{
		"email":            {"anon@test.com"},
		"password":         {"sup3rsecret123"},
		"password_confirm": {"nope1234"},
	}, &userID)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
	assert.Contains(t, rec.Body.String(), "no coinciden")
	mockCtrl.AssertNotCalled(t, "SetPassword", mock.Anything, mock.Anything)
}

func TestLinkSubmit_EmailTaken(t *testing.T) {
	mockCtrl := new(MockController)
	h := newTestHandler(mockCtrl)
	userID := uuid.New()

	mockCtrl.On("SetPassword", mock.Anything, mock.Anything).
		Return(controller.UserResponse{}, controller.ErrEmailTaken)

	rec := postFormAs(h.LinkSubmit, url.Values{
		"email":            {"taken@test.com"},
		"password":         {"sup3rsecret123"},
		"password_confirm": {"sup3rsecret123"},
	}, &userID)

	assert.Equal(t, http.StatusConflict, rec.Code)
	assert.Contains(t, rec.Body.String(), "ya está registrado")
	mockCtrl.AssertExpectations(t)
}
