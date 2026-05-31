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

const (
	oauthStateCookie  = "oauth_state"
	oauthModeCookie   = "oauth_mode"
	oauthModeLinkVal  = "link"
	oauthCookieMaxAge = 10 * time.Minute

	// TODO: update once templ pages exist.
	redirectDashboard = "/"
	redirectLogin     = "/auth/login"
)

// Login redirects the user to the Logto authorization endpoint.
func (h *Handler) Login(c echo.Context) error {
	state := util.RandomString(32)
	setOAuthCookie(c, oauthStateCookie, state)
	return c.Redirect(http.StatusTemporaryRedirect, h.identity.BuildAuthURL(state))
}

// LinkAccount starts the Logto OIDC flow for an already-authenticated user
// who wants to link their anonymous account to a Logto identity.
// Requires a valid refresh token cookie (via CookieMiddleware).
func (h *Handler) LinkAccount(c echo.Context) error {
	state := util.RandomString(32)
	setOAuthCookie(c, oauthStateCookie, state)
	setOAuthCookie(c, oauthModeCookie, oauthModeLinkVal)
	return c.Redirect(http.StatusTemporaryRedirect, h.identity.BuildAuthURL(state))
}

// Callback handles the redirect from Logto after the user authenticates.
// Depending on the oauth_mode cookie it either logs in / creates a user (login
// mode) or attaches the Logto identity to the current anonymous user (link mode).
func (h *Handler) Callback(c echo.Context) error {
	// Verify CSRF state.
	stateCookie, err := c.Cookie(oauthStateCookie)
	if err != nil || stateCookie.Value != c.QueryParam("state") {
		return c.JSON(http.StatusUnauthorized, errors.New("invalid oauth state"))
	}
	clearCookie(c, oauthStateCookie)

	isLinkMode := false
	if modeCookie, err := c.Cookie(oauthModeCookie); err == nil {
		isLinkMode = modeCookie.Value == oauthModeLinkVal
		clearCookie(c, oauthModeCookie)
	}

	ctx := context.Background()

	claims, err := h.identity.ExchangeCode(ctx, c.QueryParam("code"))
	if err != nil {
		return c.JSON(http.StatusUnauthorized, err)
	}

	sub, err := uuid.Parse(claims.Sub)
	if err != nil {
		return c.JSON(http.StatusBadRequest, errors.New("identity provider returned an invalid user id"))
	}

	if isLinkMode {
		return h.handleLink(c, ctx, sub, claims.Email)
	}
	return h.handleLogin(c, ctx, sub, claims.Email)
}

// Logout clears auth cookies and redirects to the login page.
func (h *Handler) Logout(c echo.Context) error {
	clearCookie(c, util.RefreshTokenName)
	return c.Redirect(http.StatusTemporaryRedirect, redirectLogin)
}

// handleLink attaches a Logto identity to the current anonymous user.
func (h *Handler) handleLink(c echo.Context, ctx context.Context, sub uuid.UUID, email string) error {
	// Verify the user is authenticated via refresh token cookie.
	cookie, err := c.Cookie(util.RefreshTokenName)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, errors.New("must be logged in to link an account"))
	}
	userId, err := getUserIdFromToken(cookie.Value)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, err)
	}
	payload, err := auxVerifyToken(h, userId, cookie.Value)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, err)
	}

	// Guard: reject if the Logto identity is already claimed by another user.
	_, err = h.ctrl.GetUser(ctx, controller.GetUserParams{
		GetType: util.GetUserTypeIndetityId,
		Value:   sub,
	})
	if err != nil && !errors.Is(err, controller.ErrUserNotFound) {
		return c.JSON(http.StatusInternalServerError, err)
	}
	if err == nil {
		return c.JSON(http.StatusConflict, errors.New("this identity is already linked to another account"))
	}

	_, err = h.ctrl.UpdateUser(ctx, controller.UpdateUserParams{
		Id:             payload.UserId,
		UserIdentityID: sub,
		Email:          email,
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.Redirect(http.StatusTemporaryRedirect, redirectDashboard)
}

// handleLogin finds an existing user by Logto sub, or creates one if new.
func (h *Handler) handleLogin(c echo.Context, ctx context.Context, sub uuid.UUID, email string) error {
	user, err := h.ctrl.GetUser(ctx, controller.GetUserParams{
		GetType: util.GetUserTypeIndetityId,
		Value:   sub,
	})
	if err != nil {
		if !errors.Is(err, controller.ErrUserNotFound) {
			return c.JSON(http.StatusInternalServerError, err)
		}
		// First login — create internal user and link to Logto identity.
		user, err = h.ctrl.CreateUser(ctx, controller.CreateUserParam{
			Email: email,
			Role:  util.UserDefauldRole,
		})
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err)
		}
		_, err = h.ctrl.UpdateUser(ctx, controller.UpdateUserParams{
			Id:             user.UserId,
			UserIdentityID: sub,
		})
		if err != nil {
			return c.JSON(http.StatusInternalServerError, err)
		}
	}

	if err := issueRefreshCookie(h, c, ctx, user.UserId, user.Role); err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	return c.Redirect(http.StatusTemporaryRedirect, redirectDashboard)
}

// issueRefreshCookie mints a refresh token and sets it as an HttpOnly cookie.
func issueRefreshCookie(h *Handler, c echo.Context, ctx context.Context, userId uuid.UUID, role string) error {
	secretKey, err := h.ctrl.GetUserSecretKey(ctx, userId)
	if err != nil {
		return err
	}
	maker, err := token.NewJWTMaker(secretKey.SecretKey)
	if err != nil {
		return err
	}
	refreshToken, _, err := maker.CreateToken(userId, role, refreshTokenDuration)
	if err != nil {
		return err
	}
	c.SetCookie(&http.Cookie{
		Name:     util.RefreshTokenName,
		Value:    refreshToken,
		Expires:  time.Now().UTC().Add(refreshTokenDuration),
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	return nil
}

// setOAuthCookie sets a short-lived HttpOnly cookie for the OAuth CSRF flow.
func setOAuthCookie(c echo.Context, name, value string) {
	c.SetCookie(&http.Cookie{
		Name:     name,
		Value:    value,
		Expires:  time.Now().UTC().Add(oauthCookieMaxAge),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

// clearCookie expires a cookie immediately.
func clearCookie(c echo.Context, name string) {
	c.SetCookie(&http.Cookie{
		Name:    name,
		Value:   "",
		Expires: time.Unix(0, 0),
		MaxAge:  -1,
	})
}
