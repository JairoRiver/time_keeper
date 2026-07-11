package handler

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/JairoRiver/time_keeper/internal/controller"
	"github.com/JairoRiver/time_keeper/internal/util"
	"github.com/JairoRiver/time_keeper/internal/view/pages"
	"github.com/JairoRiver/time_keeper/pkg/password"
	"github.com/JairoRiver/time_keeper/pkg/token"
	"github.com/a-h/templ"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// genericAuthError is shown for unexpected server-side failures during auth.
const genericAuthError = "Algo salió mal. Inténtalo de nuevo."

// redirectDashboard is where users land after a successful login/register.
const redirectDashboard = "/registro"

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

// clearCookie expires a cookie immediately. Path must match the one used when
// setting the cookie.
func clearCookie(c echo.Context, name string) {
	c.SetCookie(&http.Cookie{
		Name:    name,
		Value:   "",
		Path:    "/",
		Expires: time.Unix(0, 0),
		MaxAge:  -1,
	})
}

// renderPage renders a templ component with an explicit HTTP status. WriteHeader
// must be called on the echo Response (not the raw Writer) so the status is
// actually committed before templ writes the body.
func renderPage(c echo.Context, status int, component templ.Component) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
	c.Response().WriteHeader(status)
	return component.Render(c.Request().Context(), c.Response().Writer)
}

// hasValidSession reports whether the request carries a valid refresh-token cookie.
func (h *Handler) hasValidSession(c echo.Context) bool {
	cookie, err := c.Cookie(util.RefreshTokenName)
	if err != nil {
		return false
	}
	userId, err := getUserIdFromToken(cookie.Value)
	if err != nil {
		return false
	}
	_, err = auxVerifyToken(h, userId, cookie.Value)
	return err == nil
}

// LoginPage renders the sign-in form. Already-authenticated users skip to the dashboard.
func (h *Handler) LoginPage(c echo.Context) error {
	if h.hasValidSession(c) {
		return c.Redirect(http.StatusSeeOther, redirectDashboard)
	}
	return renderPage(c, http.StatusOK, pages.Login("", ""))
}

// LoginSubmit authenticates an email/password pair and starts a session.
func (h *Handler) LoginSubmit(c echo.Context) error {
	email := strings.TrimSpace(c.FormValue("email"))
	pass := c.FormValue("password")
	ctx := context.Background()

	user, err := h.ctrl.AuthenticateUser(ctx, controller.AuthenticateUserParams{
		Email:    email,
		Password: pass,
	})
	if err != nil {
		if errors.Is(err, controller.ErrInvalidCredentials) {
			return renderPage(c, http.StatusUnauthorized, pages.Login(email, "Email o contraseña incorrectos."))
		}
		h.log.Error().Err(err).Msg("LoginSubmit: AuthenticateUser error")
		return renderPage(c, http.StatusInternalServerError, pages.Login(email, genericAuthError))
	}

	if err := issueRefreshCookie(h, c, ctx, user.UserId, user.Role); err != nil {
		h.log.Error().Err(err).Msg("LoginSubmit: issueRefreshCookie error")
		return renderPage(c, http.StatusInternalServerError, pages.Login(email, genericAuthError))
	}
	return c.Redirect(http.StatusSeeOther, redirectDashboard)
}

// RegisterPage renders the sign-up form. Already-authenticated users skip to the dashboard.
func (h *Handler) RegisterPage(c echo.Context) error {
	if h.hasValidSession(c) {
		return c.Redirect(http.StatusSeeOther, redirectDashboard)
	}
	return renderPage(c, http.StatusOK, pages.Register("", ""))
}

// RegisterSubmit creates an email/password account and starts a session.
func (h *Handler) RegisterSubmit(c echo.Context) error {
	email := strings.TrimSpace(c.FormValue("email"))
	pass := c.FormValue("password")
	confirm := c.FormValue("password_confirm")
	ctx := context.Background()

	if pass != confirm {
		return renderPage(c, http.StatusBadRequest, pages.Register(email, "Las contraseñas no coinciden."))
	}

	user, err := h.ctrl.RegisterUser(ctx, controller.RegisterUserParams{
		Email:    email,
		Password: pass,
		Role:     util.UserDefauldRole,
	})
	if err != nil {
		switch {
		case errors.Is(err, controller.ErrEmailTaken):
			return renderPage(c, http.StatusConflict, pages.Register(email, "Ese email ya está registrado."))
		case errors.Is(err, controller.ErrEmptyEmail):
			return renderPage(c, http.StatusBadRequest, pages.Register(email, "Introduce un email válido."))
		case errors.Is(err, password.ErrPasswordTooShort), errors.Is(err, password.ErrPasswordTooLong):
			return renderPage(c, http.StatusBadRequest, pages.Register(email, "La contraseña debe tener entre 8 y 72 caracteres."))
		default:
			h.log.Error().Err(err).Msg("RegisterSubmit: RegisterUser error")
			return renderPage(c, http.StatusInternalServerError, pages.Register(email, genericAuthError))
		}
	}

	if err := issueRefreshCookie(h, c, ctx, user.UserId, user.Role); err != nil {
		h.log.Error().Err(err).Msg("RegisterSubmit: issueRefreshCookie error")
		return renderPage(c, http.StatusInternalServerError, pages.Register(email, genericAuthError))
	}
	return c.Redirect(http.StatusSeeOther, redirectDashboard)
}

// Logout clears the local session cookie and returns to the landing page.
func (h *Handler) Logout(c echo.Context) error {
	clearCookie(c, util.RefreshTokenName)
	return c.Redirect(http.StatusSeeOther, "/")
}

// LinkPage renders the form for an anonymous user to attach email/password
// credentials to their current session. Requires PageAuthMiddleware.
func (h *Handler) LinkPage(c echo.Context) error {
	return renderPage(c, http.StatusOK, pages.Link("", ""))
}

// LinkSubmit attaches email/password credentials to the current user, keeping
// the same user id (and therefore their existing time entries). The active
// session cookie stays valid. Requires PageAuthMiddleware.
func (h *Handler) LinkSubmit(c echo.Context) error {
	userInfo := c.Get(util.RefreshTokenName).(UserInfo)
	email := strings.TrimSpace(c.FormValue("email"))
	pass := c.FormValue("password")
	confirm := c.FormValue("password_confirm")
	ctx := context.Background()

	if pass != confirm {
		return renderPage(c, http.StatusBadRequest, pages.Link(email, "Las contraseñas no coinciden."))
	}

	_, err := h.ctrl.SetPassword(ctx, controller.SetPasswordParams{
		UserId:   userInfo.UserId,
		Email:    email,
		Password: pass,
	})
	if err != nil {
		switch {
		case errors.Is(err, controller.ErrEmailTaken):
			return renderPage(c, http.StatusConflict, pages.Link(email, "Ese email ya está registrado."))
		case errors.Is(err, controller.ErrEmptyEmail):
			return renderPage(c, http.StatusBadRequest, pages.Link(email, "Introduce un email válido."))
		case errors.Is(err, password.ErrPasswordTooShort), errors.Is(err, password.ErrPasswordTooLong):
			return renderPage(c, http.StatusBadRequest, pages.Link(email, "La contraseña debe tener entre 8 y 72 caracteres."))
		default:
			h.log.Error().Err(err).Msg("LinkSubmit: SetPassword error")
			return renderPage(c, http.StatusInternalServerError, pages.Link(email, genericAuthError))
		}
	}
	return c.Redirect(http.StatusSeeOther, redirectDashboard)
}
