package reltime

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// now is a stable reference: 2026-05-24 (Sunday) at 12:00 local.
var now = time.Date(2026, 5, 24, 12, 0, 0, 0, time.Local)

func dateOnly(y int, m time.Month, d int) time.Time {
	return time.Date(y, m, d, 0, 0, 0, 0, time.Local)
}

func TestFormat(t *testing.T) {
	tests := []struct {
		name string
		ref  time.Time
		want string
	}{
		{"date-only today", dateOnly(2026, 5, 24), "today"},
		{"tomorrow", dateOnly(2026, 5, 25), "tomorrow"},
		{"weekday band low (2d)", dateOnly(2026, 5, 26), "tuesday"},
		{"weekday band high (6d)", dateOnly(2026, 5, 30), "saturday"},
		{"seven days", dateOnly(2026, 5, 31), "7d"},
		{"thirty days", dateOnly(2026, 6, 23), "30d"},
		{"thirty-one days absolute", dateOnly(2026, 6, 24), "2026-06-24"},
		{"timed today still ahead", time.Date(2026, 5, 24, 15, 0, 0, 0, time.Local), "3pm"},
		{"timed earlier today", time.Date(2026, 5, 24, 9, 30, 0, 0, time.Local), "9:30am"},
		{"past three days", dateOnly(2026, 5, 21), "3d"},
		{"past thirty days", dateOnly(2026, 4, 24), "30d"},
		{"past thirty-one days absolute", dateOnly(2026, 4, 23), "2026-04-23"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, Format(tt.ref, now))
		})
	}
}
