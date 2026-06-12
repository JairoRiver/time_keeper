package pages

import (
	"fmt"
	"hash/fnv"
	"time"

	"github.com/JairoRiver/time_keeper/internal/controller"
)

// TagGroup aggregates all entries that share the same tag within a single day.
type TagGroup struct {
	Tag     string
	Entries []controller.EntryTimeResponse // individual entries, chronological
	Total   time.Duration                  // sum of completed entry durations
	HasOpen bool                           // at least one entry has no TimeEnd
}

// DayGroup groups tag aggregates by date for the Registro list.
type DayGroup struct {
	Label     string // "DD-mm-aaaa"
	TagGroups []TagGroup
}

// DayColumn holds entries for one date column in the Resumen calendar.
type DayColumn struct {
	Date    time.Time
	Label   string // "dd/mm/aaaa"
	Entries []controller.EntryTimeResponse
}

// SlotHeight is the pixel height of one 30-minute time slot in the calendar.
const SlotHeight = 32

// TotalCalendarHeight is the full pixel height of the 48-slot day grid.
const TotalCalendarHeight = 48 * SlotHeight

// FormatTagGroupDuration returns the summed duration of completed entries in the group.
// Returns "--:--:--" when there are no completed entries.
func FormatTagGroupDuration(tg TagGroup) string {
	if tg.Total == 0 {
		return "--:--:--"
	}
	h := int(tg.Total.Hours())
	m := int(tg.Total.Minutes()) % 60
	s := int(tg.Total.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}

func FormatDuration(e controller.EntryTimeResponse) string {
	if e.TimeEnd.IsZero() {
		return "--:--:--"
	}
	d := e.TimeEnd.Sub(e.TimeStart)
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
}

func FormatTime(t time.Time) string {
	return t.Format("15:04")
}

func FormatDate(t time.Time) string {
	return t.Format("02-01-2006")
}

func FormatEndLabel(e controller.EntryTimeResponse) string {
	if e.TimeEnd.IsZero() {
		return "en curso"
	}
	return FormatTime(e.TimeEnd) + " · " + FormatDate(e.TimeEnd)
}

// FormatEndShort returns only the end clock time, or "···" for an open entry.
func FormatEndShort(e controller.EntryTimeResponse) string {
	if e.TimeEnd.IsZero() {
		return "···"
	}
	return FormatTime(e.TimeEnd)
}

// Soft, modern pastel palette for tag blocks (readable with dark text).
var tagHexColors = []string{
	"#c7d2fe", // indigo-200
	"#ddd6fe", // violet-200
	"#bae6fd", // sky-200
	"#a7f3d0", // emerald-200
	"#fde68a", // amber-200
	"#fecdd3", // rose-200
	"#bfdbfe", // blue-200
	"#d9f99d", // lime-200
}

func TagHexColor(tag string) string {
	h := fnv.New32a()
	h.Write([]byte(tag))
	return tagHexColors[h.Sum32()%uint32(len(tagHexColors))]
}

func EntryTopPx(e controller.EntryTimeResponse) int {
	t := e.TimeStart
	slot := (t.Hour()*60 + t.Minute()) / 30
	return slot * SlotHeight
}

func EntryHeightPx(e controller.EntryTimeResponse) int {
	if e.TimeEnd.IsZero() {
		return SlotHeight
	}
	mins := int(e.TimeEnd.Sub(e.TimeStart).Minutes())
	if mins < 30 {
		mins = 30
	}
	slots := (mins + 29) / 30
	return slots * SlotHeight
}
