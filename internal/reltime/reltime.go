// Package reltime renders timestamps as compact relative-time WHEN strings,
// providing a single shared vocabulary for the task list's chips and the task
// editor's status line.
package reltime

import (
	"fmt"
	"strings"
	"time"
)

// Format renders ref as a relative-time WHEN string against now. Day counts
// are calendar-day differences in the local timezone (both truncated to
// midnight), not 24-hour spans, so "tomorrow" stays "tomorrow" late at night.
//
// Future ladder: clock for a timed value still ahead today, "today" for a
// date-only today, "tomorrow", weekday names for 2–6 days, "Nd" up to 30 days,
// then an absolute YYYY-MM-DD date. Past ladder: clock for earlier today, "Nd"
// up to 30 days ago, then an absolute date (no tomorrow/weekday).
func Format(ref, now time.Time) string {
	refLocal := ref.Local()
	nowLocal := now.Local()

	refDay := truncateToDay(refLocal)
	nowDay := truncateToDay(nowLocal)
	// Round to absorb DST-shortened/lengthened days (23h or 25h).
	days := int((refDay.Sub(nowDay).Hours() + 12) / 24)
	if refDay.Before(nowDay) {
		days = -int((nowDay.Sub(refDay).Hours() + 12) / 24)
	}

	timed := !isLocalMidnight(refLocal)

	switch {
	case days == 0:
		if timed {
			return formatClock(refLocal)
		}
		return "today"
	case days > 0:
		switch {
		case days == 1:
			return "tomorrow"
		case days <= 6:
			return strings.ToLower(refLocal.Weekday().String())
		case days <= 30:
			return fmt.Sprintf("%dd", days)
		default:
			return refLocal.Format("2006-01-02")
		}
	default: // past
		n := -days
		if n <= 30 {
			return fmt.Sprintf("%dd", n)
		}
		return refLocal.Format("2006-01-02")
	}
}

// truncateToDay returns local midnight of t's calendar day.
func truncateToDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// isLocalMidnight reports whether t has no time-of-day component (mirrors the
// date-only check in date.formatDate).
func isLocalMidnight(t time.Time) bool {
	return t.Hour() == 0 && t.Minute() == 0 && t.Second() == 0
}

// formatClock renders a time-of-day like "3pm" or "3:30pm".
func formatClock(t time.Time) string {
	if t.Minute() == 0 {
		return strings.ToLower(t.Format("3pm"))
	}
	return strings.ToLower(t.Format("3:04pm"))
}
