package handler

import (
	"context"
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
		return c.JSON(http.StatusInternalServerError, err)
	}

	//create jwt
	secretKey, err := h.ctrl.GetUserSecretKey(ctx, user.UserId)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	tokenMaker, err := token.NewJWTMaker(secretKey.SecretKey)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	accessToken, payload, err := tokenMaker.CreateToken(user.UserId, user.Role, accessTokenDuration)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	//create cookies with refresh token
	refreshToken, _, err := tokenMaker.CreateToken(user.UserId, user.Role, refreshTokenDuration)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	cookie := http.Cookie{
		Name:     util.RefreshTokenName,
		Value:    refreshToken,
		Expires:  time.Now().UTC().Add(refreshTokenDuration),
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	}
	c.SetCookie(&cookie)

	response := parseUserResponse(user, accessToken, payload.ExpiresAt.Time.UTC())
	return c.JSON(http.StatusCreated, response)
}

//TODO: update
//TODO: login
//TODO: Logout
