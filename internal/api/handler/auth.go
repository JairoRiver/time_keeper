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

	redirectDashboard = "/registro"
	redirectLogin     = "/auth/login"
)

// LinkAccount starts the Logto OIDC flow for an already-authenticated user
// who wants to link their anonymous account to a Logto identity.
// Requires a valid refresh token cookie (via CookieMiddleware).
func (h *Handler) LinkAccount(c echo.Context) error {
	state := util.RandomString(32)
	h.setOAuthCookie(c, oauthStateCookie, state)
	h.setOAuthCookie(c, oauthModeCookie, oauthModeLinkVal)
	return c.Redirect(http.StatusTemporaryRedirect, h.identity.BuildAuthURL(state))
}

// Callback handles the redirect from Logto after the user authenticates.
// Depending on the oauth_mode cookie it either logs in / creates a user (login
// mode) or attaches the Logto identity to the current anonymous user (link mode).
func (h *Handler) Callback(c echo.Context) error {
	// Verify CSRF state.
	stateCookie, err := c.Cookie(oauthStateCookie)
	if err != nil {
		h.log.Warn().Msg("auth callback: oauth state cookie missing")
		return c.Redirect(http.StatusSeeOther, redirectLogin)
	}
	if stateCookie.Value != c.QueryParam("state") {
		h.log.Warn().Msg("auth callback: oauth state mismatch")
		return c.Redirect(http.StatusSeeOther, redirectLogin)
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
		h.log.Error().Err(err).Msg("auth callback: exchange code failed")
		return c.Redirect(http.StatusSeeOther, redirectLogin)
	}

	if len(claims.Sub) == 0 {
		h.log.Error().Msg("auth callback: identity provider returned empty sub")
		return c.Redirect(http.StatusSeeOther, redirectLogin)
	}

	if isLinkMode {
		return h.handleLink(c, ctx, claims.Sub, claims.Email)
	}
	return h.handleLogin(c, ctx, claims.Sub, claims.Email)
}

// handleLink attaches a Logto identity to the current anonymous user.
func (h *Handler) handleLink(c echo.Context, ctx context.Context, sub string, email string) error {
	cookie, err := c.Cookie(util.RefreshTokenName)
	if err != nil {
		h.log.Warn().Msg("handleLink: no refresh token cookie")
		return c.Redirect(http.StatusSeeOther, redirectLogin)
	}
	userId, err := getUserIdFromToken(cookie.Value)
	if err != nil {
		h.log.Warn().Err(err).Msg("handleLink: invalid token")
		return c.Redirect(http.StatusSeeOther, redirectLogin)
	}
	payload, err := auxVerifyToken(h, userId, cookie.Value)
	if err != nil {
		h.log.Warn().Err(err).Msg("handleLink: token verification failed")
		return c.Redirect(http.StatusSeeOther, redirectLogin)
	}

	existingUser, err := h.ctrl.GetUser(ctx, controller.GetUserParams{
		GetType: util.GetUserTypeIndetityId,
		Value:   sub,
	})
	if err != nil && !errors.Is(err, controller.ErrUserNotFound) {
		h.log.Error().Err(err).Msg("handleLink: GetUser error")
		return c.Redirect(http.StatusSeeOther, redirectDashboard)
	}

	// Identity already linked to another internal user → log in as that user.
	if err == nil {
		h.log.Info().Str("sub", sub).Msg("handleLink: identity exists, switching session to existing user")
		if err := issueRefreshCookie(h, c, ctx, existingUser.UserId, existingUser.Role); err != nil {
			h.log.Error().Err(err).Msg("handleLink: issueRefreshCookie for existing user failed")
		}
		return c.Redirect(http.StatusSeeOther, redirectDashboard)
	}

	// New link — attach Logto identity to the current anonymous user.
	_, err = h.ctrl.UpdateUser(ctx, controller.UpdateUserParams{
		Id:             payload.UserId,
		UserIdentityID: sub,
		Email:          email,
	})
	if err != nil {
		h.log.Error().Err(err).Msg("handleLink: UpdateUser error")
		return c.Redirect(http.StatusSeeOther, redirectDashboard)
	}

	return c.Redirect(http.StatusSeeOther, redirectDashboard)
}

// handleLogin finds an existing user by Logto sub, or creates one if new.
func (h *Handler) handleLogin(c echo.Context, ctx context.Context, sub string, email string) error {
	user, err := h.ctrl.GetUser(ctx, controller.GetUserParams{
		GetType: util.GetUserTypeIndetityId,
		Value:   sub,
	})
	if err != nil {
		if !errors.Is(err, controller.ErrUserNotFound) {
			h.log.Error().Err(err).Msg("handleLogin: GetUser error")
			return c.Redirect(http.StatusSeeOther, redirectLogin)
		}
		user, err = h.ctrl.CreateUser(ctx, controller.CreateUserParam{
			Email: email,
			Role:  util.UserDefauldRole,
		})
		if err != nil {
			h.log.Error().Err(err).Msg("handleLogin: CreateUser error")
			return c.Redirect(http.StatusSeeOther, redirectLogin)
		}
		_, err = h.ctrl.UpdateUser(ctx, controller.UpdateUserParams{
			Id:             user.UserId,
			UserIdentityID: sub,
		})
		if err != nil {
			h.log.Error().Err(err).Msg("handleLogin: UpdateUser error")
			return c.Redirect(http.StatusSeeOther, redirectLogin)
		}
	}

	if err := issueRefreshCookie(h, c, ctx, user.UserId, user.Role); err != nil {
		h.log.Error().Err(err).Msg("handleLogin: issueRefreshCookie error")
		return c.Redirect(http.StatusSeeOther, redirectLogin)
	}

	return c.Redirect(http.StatusSeeOther, redirectDashboard)
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
	c.SetCookie(h.sessionCookie(refreshToken))
	return nil
}

// setOAuthCookie sets a short-lived HttpOnly cookie for the OAuth CSRF flow.
func (h *Handler) setOAuthCookie(c echo.Context, name, value string) {
	c.SetCookie(&http.Cookie{
		Name:     name,
		Value:    value,
		Expires:  time.Now().UTC().Add(oauthCookieMaxAge),
		HttpOnly: true,
		Secure:   h.secureCookies,
		SameSite: http.SameSiteLaxMode,
	})
}

// clearCookie expires a cookie immediately.
// Path must match the path used when setting the cookie.
func clearCookie(c echo.Context, name string) {
	c.SetCookie(&http.Cookie{
		Name:    name,
		Value:   "",
		Path:    "/",
		Expires: time.Unix(0, 0),
		MaxAge:  -1,
	})
}
