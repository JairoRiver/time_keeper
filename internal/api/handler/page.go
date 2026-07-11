package handler

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/JairoRiver/time_keeper/internal/controller"
	"github.com/JairoRiver/time_keeper/internal/util"
	"github.com/JairoRiver/time_keeper/internal/view/pages"
	"github.com/labstack/echo/v4"
)

// LandingPage renders the public landing page. Authenticated users are redirected to /registro.
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
		return h.internalError(c, err)
	}
	if err := issueRefreshCookie(h, c, ctx, user.UserId, user.Role); err != nil {
		return h.internalError(c, err)
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
		return h.internalError(c, err)
	}

	entries, err := h.ctrl.ListEntryTime(ctx, controller.ListEntryTimeParams{
		UserId:     userInfo.UserId,
		PageNumber: page,
	})
	if err != nil {
		return h.internalError(c, err)
	}

	activeEntry, hasActive, err := h.ctrl.GetActiveTimer(ctx, userInfo.UserId)
	if err != nil {
		return h.internalError(c, err)
	}

	groups := groupEntriesByDay(entries)
	return pages.Registro(groups, page, user.UserIdentityID == "", activeEntry, hasActive).
		Render(c.Request().Context(), c.Response().Writer)
}

// TimerStart creates a new open entry (no TimeEnd) and redirects back to /registro.
// Requires PageAuthMiddleware.
func (h *Handler) TimerStart(c echo.Context) error {
	userInfo := c.Get(util.RefreshTokenName).(UserInfo)
	tag := strings.TrimSpace(c.FormValue("tag"))
	if tag == "" {
		tag = "Sin etiqueta"
	}
	ctx := context.Background()

	_, err := h.ctrl.CreateEntryTime(ctx, controller.CreateEntryTimeParams{
		UserID:    userInfo.UserId,
		Tag:       tag,
		TimeStart: time.Now().UTC(),
	})
	if err != nil {
		return h.internalError(c, err)
	}
	return c.Redirect(http.StatusSeeOther, "/registro")
}

// TimerStop sets TimeEnd=now on the active entry and redirects back to /registro.
// Requires PageAuthMiddleware.
func (h *Handler) TimerStop(c echo.Context) error {
	userInfo := c.Get(util.RefreshTokenName).(UserInfo)
	ctx := context.Background()

	_, err := h.ctrl.StopTimer(ctx, userInfo.UserId)
	if err != nil {
		return h.internalError(c, err)
	}
	return c.Redirect(http.StatusSeeOther, "/registro")
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
		return h.internalError(c, err)
	}

	dateStart, dateEnd := parseDateRange(c)

	entries, err := h.ctrl.ListEntryTimeByDateRange(ctx, controller.ListEntryTimeByDateRangeParams{
		UserId:    userInfo.UserId,
		DateStart: dateStart,
		DateEnd:   dateEnd,
	})
	if err != nil {
		return h.internalError(c, err)
	}

	columns := buildDayColumns(entries, dateStart, dateEnd)
	return pages.Resumen(columns, dateStart, dateEnd, user.UserIdentityID == "").
		Render(c.Request().Context(), c.Response().Writer)
}

// ── helpers ──────────────────────────────────────────────────────────────────

func groupEntriesByDay(entries []controller.EntryTimeResponse) []pages.DayGroup {
	// Preserve the SQL-ordered day sequence (DESC) and tag sequence (ASC within day).
	var dayOrder []string
	dayTagOrder := make(map[string][]string)
	// date key → tag key → entries
	table := make(map[string]map[string][]controller.EntryTimeResponse)

	for _, e := range entries {
		dk := e.TimeStart.Format("2006-01-02")
		tk := e.Tag
		if _, ok := table[dk]; !ok {
			table[dk] = make(map[string][]controller.EntryTimeResponse)
			dayOrder = append(dayOrder, dk)
		}
		if _, ok := table[dk][tk]; !ok {
			dayTagOrder[dk] = append(dayTagOrder[dk], tk)
		}
		table[dk][tk] = append(table[dk][tk], e)
	}

	groups := make([]pages.DayGroup, 0, len(dayOrder))
	for _, dk := range dayOrder {
		t, _ := time.Parse("2006-01-02", dk)
		tagGroups := make([]pages.TagGroup, 0, len(dayTagOrder[dk]))
		for _, tk := range dayTagOrder[dk] {
			es := table[dk][tk]
			var total time.Duration
			hasOpen := false
			for _, e := range es {
				if e.TimeEnd.IsZero() {
					hasOpen = true
				} else {
					total += e.TimeEnd.Sub(e.TimeStart)
				}
			}
			tagGroups = append(tagGroups, pages.TagGroup{
				Tag:     tk,
				Entries: es,
				Total:   total,
				HasOpen: hasOpen,
			})
		}
		groups = append(groups, pages.DayGroup{
			Label:     t.Format("02-01-2006"),
			TagGroups: tagGroups,
		})
	}
	return groups
}

func buildDayColumns(entries []controller.EntryTimeResponse, dateStart, dateEnd time.Time) []pages.DayColumn {
	byDate := make(map[string][]controller.EntryTimeResponse)
	for _, e := range entries {
		key := e.TimeStart.Format("2006-01-02")
		byDate[key] = append(byDate[key], e)
	}

	start := truncateToDay(dateStart)
	end := truncateToDay(dateEnd)
	var columns []pages.DayColumn
	for d := start; !d.After(end); d = d.AddDate(0, 0, 1) {
		key := d.Format("2006-01-02")
		columns = append(columns, pages.DayColumn{
			Date:    d,
			Label:   d.Format("02/01/2006"),
			Entries: byDate[key],
		})
	}
	return columns
}

func parseDateRange(c echo.Context) (time.Time, time.Time) {
	now := time.Now().UTC()
	dateEnd := truncateToDay(now).Add(24*time.Hour - time.Second)
	dateStart := truncateToDay(now.AddDate(0, 0, -6))

	if s := c.QueryParam("date_start"); s != "" {
		if t, err := time.ParseInLocation("2006-01-02", s, time.UTC); err == nil {
			dateStart = t
		}
	}
	if s := c.QueryParam("date_end"); s != "" {
		if t, err := time.ParseInLocation("2006-01-02", s, time.UTC); err == nil {
			dateEnd = t.Add(24*time.Hour - time.Second)
		}
	}
	return dateStart, dateEnd
}

func truncateToDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}
