package handler

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/JairoRiver/time_keeper/internal/controller"
	"github.com/JairoRiver/time_keeper/internal/util"
	"github.com/JairoRiver/time_keeper/pkg/token"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// Create User Handler.
type ResponseUser struct {
	UserId      uuid.UUID
	Email       string
	AccessToken string
	ExpiredAt   time.Time
}

func parseUserResponse(user controller.UserResponse, token string, expiredAt time.Time) ResponseUser {
	response := ResponseUser{
		UserId:      user.UserId,
		Email:       user.Email,
		AccessToken: token,
		ExpiredAt:   expiredAt,
	}
	return response
}

// @Summary Create a new user
// @Description generate a new user
// @ID post-create-user
// @Produce  json
// @Success 201 {object} ResponseUser
// @Router /api/v1/user [post]
func (h *Handler) CreateUser(c echo.Context) error {
	params := controller.CreateUserParam{Role: util.UserDefauldRole}
	ctx := context.Background()
	user, err := h.ctrl.CreateUser(ctx, params)
	if err != nil {
		return h.internalError(c, err)
	}

	//create jwt
	secretKey, err := h.ctrl.GetUserSecretKey(ctx, user.UserId)
	if err != nil {
		return h.internalError(c, err)
	}

	tokenMaker, err := token.NewJWTMaker(secretKey.SecretKey)
	if err != nil {
		return h.internalError(c, err)
	}

	accessToken, payload, err := tokenMaker.CreateToken(user.UserId, user.Role, accessTokenDuration)
	if err != nil {
		return h.internalError(c, err)
	}

	//create cookies with refresh token
	refreshToken, _, err := tokenMaker.CreateToken(user.UserId, user.Role, refreshTokenDuration)
	if err != nil {
		return h.internalError(c, err)
	}
	c.SetCookie(h.sessionCookie(refreshToken))

	response := parseUserResponse(user, accessToken, payload.ExpiresAt.Time.UTC())
	return c.JSON(http.StatusCreated, response)
}

// @Summary Refresh token endpoint
// @Description generate a new Access token and refresh token if have a valid refresh token
// @ID post-refres-token
// @Produce  json
// @Success 201 {object} ResponseUser
// @Router /api/v1/refresh [post]
func (h *Handler) RefreshToken(c echo.Context) error {
	// Get userId from contex
	payloadCookie := c.Get(util.RefreshTokenName)
	userInfo, ok := payloadCookie.(UserInfo)
	if !ok {
		return h.internalError(c, errors.New("RefreshToken cant get payload from context"))
	}

	//get user info
	ctx := context.Background()
	params := controller.GetUserParams{
		GetType: util.GetUserTypeId,
		Value:   userInfo.UserId,
	}
	user, err := h.ctrl.GetUser(ctx, params)
	if err != nil {
		return h.internalError(c, err)
	}

	//create jwt
	secretKey, err := h.ctrl.GetUserSecretKey(ctx, userInfo.UserId)
	if err != nil {
		return h.internalError(c, err)
	}

	tokenMaker, err := token.NewJWTMaker(secretKey.SecretKey)
	if err != nil {
		return h.internalError(c, err)
	}

	accessToken, payload, err := tokenMaker.CreateToken(userInfo.UserId, userInfo.Role, accessTokenDuration)
	if err != nil {
		return h.internalError(c, err)
	}

	//create cookies with refresh token
	refreshToken, _, err := tokenMaker.CreateToken(userInfo.UserId, userInfo.Role, refreshTokenDuration)
	if err != nil {
		return h.internalError(c, err)
	}
	c.SetCookie(h.sessionCookie(refreshToken))

	response := parseUserResponse(user, accessToken, payload.ExpiresAt.Time.UTC())
	return c.JSON(http.StatusCreated, response)

}

//TODO: update
//TODO: login
//TODO: Logout
