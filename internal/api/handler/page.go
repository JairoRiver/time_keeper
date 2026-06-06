package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/JairoRiver/time_keeper/internal/controller"
	"github.com/JairoRiver/time_keeper/internal/util"
	"github.com/JairoRiver/time_keeper/internal/view/pages"
	"github.com/labstack/echo/v4"
)

// LandingPage renders the public landing page. If the user already has a valid
// session cookie they are sent straight to /registro.
func (h *Handler) LandingPage(c echo.Context) error {
	if cookie, err := c.Cookie(util.RefreshTokenName); err == nil {
		if userId, err := getUserIdFromToken(cookie.Value); err == nil {
			if _, err := auxVerifyToken(h, userId, cookie.Value); err == nil {
				return c.Redirect(http.StatusSeeOther, "/registro")
			}
		}
	}
	return pages.Landing().Render(c.Request().Context(), c.Response().Writer)
}

// Try creates an anonymous user, issues a session cookie, and redirects to /registro.
func (h *Handler) Try(c echo.Context) error {
	ctx := context.Background()
	user, err := h.ctrl.CreateUser(ctx, controller.CreateUserParam{
		Role: util.UserDefauldRole,
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	if err := issueRefreshCookie(h, c, ctx, user.UserId, user.Role); err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}
	return c.Redirect(http.StatusSeeOther, "/registro")
}

// RegistroPage renders the time registration page.
// Requires PageAuthMiddleware.
func (h *Handler) RegistroPage(c echo.Context) error {
	userInfo := c.Get(util.RefreshTokenName).(UserInfo)
	ctx := context.Background()

	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page == 0 {
		page = 1
	}

	user, err := h.ctrl.GetUser(ctx, controller.GetUserParams{
		GetType: util.GetUserTypeId,
		Value:   userInfo.UserId,
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	entries, err := h.ctrl.ListEntryTime(ctx, controller.ListEntryTimeParams{
		UserId:     userInfo.UserId,
		PageNumber: page,
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	return pages.Registro(entries, page, user.UserIdentityID == "").Render(c.Request().Context(), c.Response().Writer)
}

// ResumenPage renders the summary calendar page.
// Requires PageAuthMiddleware.
func (h *Handler) ResumenPage(c echo.Context) error {
	userInfo := c.Get(util.RefreshTokenName).(UserInfo)
	ctx := context.Background()

	user, err := h.ctrl.GetUser(ctx, controller.GetUserParams{
		GetType: util.GetUserTypeId,
		Value:   userInfo.UserId,
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	entries, err := h.ctrl.ListEntryTime(ctx, controller.ListEntryTimeParams{
		UserId:     userInfo.UserId,
		PageNumber: 1,
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	return pages.Resumen(entries, user.UserIdentityID == "").Render(c.Request().Context(), c.Response().Writer)
}

// HelloPage renders the hello test page. Remove once real pages exist.
func (h *Handler) HelloPage(c echo.Context) error {
	return pages.Hello("World").Render(c.Request().Context(), c.Response().Writer)
}
