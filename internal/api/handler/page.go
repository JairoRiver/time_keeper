package handler

import (
	"github.com/JairoRiver/time_keeper/internal/view/pages"
	"github.com/labstack/echo/v4"
)

// HelloPage renders the hello test page. Remove once real pages exist.
func (h *Handler) HelloPage(c echo.Context) error {
	return pages.Hello("World").Render(c.Request().Context(), c.Response().Writer)
}
