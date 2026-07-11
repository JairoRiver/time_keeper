package handler

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/JairoRiver/time_keeper/internal/controller"
	"github.com/JairoRiver/time_keeper/internal/util"
	"github.com/JairoRiver/time_keeper/internal/view/pages"
	"github.com/JairoRiver/time_keeper/pkg/password"
	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
)

// genericAuthError is shown for unexpected server-side failures during auth.
const genericAuthError = "Algo salió mal. Inténtalo de nuevo."

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
