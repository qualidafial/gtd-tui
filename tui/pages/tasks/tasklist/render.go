package tasklist

import (
	"fmt"
	"io"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/qualidafial/gtd-tui"
	"github.com/qualidafial/gtd-tui/internal/reltime"
)

// truncateToDay returns local midnight of t's calendar day.
func truncateToDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// isLocalMidnight reports whether t has no time-of-day component (mirrors the
// date-only check in date.formatDate).
func isLocalMidnight(t time.Time) bool {
	return t.Hour() == 0 && t.Minute() == 0 && t.Second() == 0
}

// chip is a rendered data fragment with its urgency style.
type chip struct {
	text  string
	style lipgloss.Style
}

// chipColors holds the urgency palette, resolved against the terminal
// background. Dark theme is the only tuned target.
type chipColors struct {
	overdue  lipgloss.Style // red
	dueToday lipgloss.Style // orange
	dueSoon  lipgloss.Style // yellow (2–6 days)
	dueLater lipgloss.Style // dim/neutral (7+ days)
	deferred lipgloss.Style // dim blue
	ready    lipgloss.Style // teal
	assignee lipgloss.Style // magenta
	project  lipgloss.Style // green
}

// newChipColors returns the urgency palette. Dark theme is the only tuned
// target. The hasDarkBg parameter is retained for forward compatibility
// with a future light-mode tune.
func newChipColors(hasDarkBg bool) chipColors {
	_ = hasDarkBg
	return chipColors{
		overdue:  lipgloss.NewStyle().Foreground(lipgloss.Color("9")),   // red
		dueToday: lipgloss.NewStyle().Foreground(lipgloss.Color("208")), // orange
		dueSoon:  lipgloss.NewStyle().Foreground(lipgloss.Color("11")),  // yellow
		dueLater: lipgloss.NewStyle().Foreground(lipgloss.Color("245")), // dim
		deferred: lipgloss.NewStyle().Foreground(lipgloss.Color("67")),  // dim blue
		ready:    lipgloss.NewStyle().Foreground(lipgloss.Color("44")),  // teal
		assignee: lipgloss.NewStyle().Foreground(lipgloss.Color("13")),  // magenta
		project:  lipgloss.NewStyle().Foreground(lipgloss.Color("36")),  // green
	}
}

// taskChips builds the ordered chips for a task: due/overdue, then defer/ready,
// then assignee, then project. Due and defer chips are suppressed on done and
// dropped tasks; assignee and project chips survive on done tasks; dropped
// tasks show no chips. projectName is the resolved title to render in the
// `+<name>` chip; an empty string suppresses the chip.
func taskChips(t gtd.Task, now time.Time, c chipColors, projectName string) []chip {
	var chips []chip

	if t.Status == gtd.TaskStatusDropped {
		return nil
	}

	if t.Status == gtd.TaskStatusOpen {
		if ch, ok := dueChip(t, now, c); ok {
			chips = append(chips, ch)
		}
		if ch, ok := deferChip(t, now, c); ok {
			chips = append(chips, ch)
		}
	}

	if t.Assignee != nil {
		chips = append(chips, chip{text: "@" + *t.Assignee, style: c.assignee})
	}

	if projectName != "" {
		chips = append(chips, chip{text: "+" + projectName, style: c.project})
	}

	return chips
}

// dueChip builds the due/overdue chip. The word is decided by the reference
// instant (end-of-day for a date-only due, the exact instant for a timed due);
// the WHEN string reflects the raw due timestamp.
func dueChip(t gtd.Task, now time.Time, c chipColors) (chip, bool) {
	if t.Due == nil {
		return chip{}, false
	}
	due := t.Due.Local()
	ref := due
	if isLocalMidnight(due) {
		ref = endOfLocalDay(due)
	}
	when := reltime.Format(due, now)

	if ref.After(now) {
		return chip{text: "due:" + when, style: dueStyle(due, now, c)}, true
	}
	return chip{text: "overdue:" + when, style: c.overdue}, true
}

// deferChip builds the defer/ready chip. The word is decided by the reference
// instant (start-of-day for a date-only defer, the exact instant for a timed
// defer); a passed reference means the task has resurfaced (ready).
func deferChip(t gtd.Task, now time.Time, c chipColors) (chip, bool) {
	if t.DeferUntil == nil {
		return chip{}, false
	}
	deferUntil := t.DeferUntil.Local()
	ref := deferUntil
	if isLocalMidnight(deferUntil) {
		ref = truncateToDay(deferUntil)
	}
	when := reltime.Format(deferUntil, now)

	if ref.After(now) {
		return chip{text: "defer:" + when, style: c.deferred}, true
	}
	return chip{text: "ready:" + when, style: c.ready}, true
}

// dueStyle picks the urgency color for a not-yet-due chip by calendar-day
// distance: today orange, 1–6 days yellow, 7+ days dim.
func dueStyle(due, now time.Time, c chipColors) lipgloss.Style {
	days := calendarDaysBetween(now, due)
	switch {
	case days <= 0:
		return c.dueToday
	case days <= 6:
		return c.dueSoon
	default:
		return c.dueLater
	}
}

// calendarDaysBetween returns the signed number of calendar days from a to b in
// local time.
func calendarDaysBetween(a, b time.Time) int {
	aDay := truncateToDay(a.Local())
	bDay := truncateToDay(b.Local())
	if bDay.Before(aDay) {
		return -int((aDay.Sub(bDay).Hours() + 12) / 24)
	}
	return int((bDay.Sub(aDay).Hours() + 12) / 24)
}

// endOfLocalDay returns 23:59:59.999999999 of t's calendar day in its location.
func endOfLocalDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, int(time.Second-time.Nanosecond), t.Location())
}

// title styling per status.
var (
	doneTitleStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("65")).Faint(true)
	droppedTitleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Faint(true).Strikethrough(true)
	openTitleStyle    = lipgloss.NewStyle()
)

// statusMarker returns the leading marker for a task status.
func statusMarker(s gtd.TaskStatus) string {
	switch s {
	case gtd.TaskStatusDone:
		return "[x]"
	case gtd.TaskStatusDropped:
		return "[-]"
	default:
		return "[ ]"
	}
}

// titleStyle returns the per-status title style.
func titleStyle(s gtd.TaskStatus) lipgloss.Style {
	switch s {
	case gtd.TaskStatusDone:
		return doneTitleStyle
	case gtd.TaskStatusDropped:
		return droppedTitleStyle
	default:
		return openTitleStyle
	}
}

// delegate renders task rows: a status marker, the title (truncated first under
// width pressure and carrying the per-status style and selection highlight),
// and inline urgency-colored chips that keep their colors on the selected row.
// projectResolver returns the project name to render in the project chip for
// a task, or "" to suppress the chip.
type projectResolver func(gtd.Task) string

type delegate struct {
	keys      KeyMap
	hasDarkBg bool
	project   projectResolver
}

func newDelegate(keys KeyMap, project projectResolver) *delegate {
	if project == nil {
		project = func(gtd.Task) string { return "" }
	}
	return &delegate{keys: keys, hasDarkBg: true, project: project}
}

func (d *delegate) Height() int  { return 1 }
func (d *delegate) Spacing() int { return 0 }

func (d *delegate) Update(msg tea.Msg, _ *list.Model) tea.Cmd {
	if bg, ok := msg.(tea.BackgroundColorMsg); ok {
		d.hasDarkBg = bg.IsDark()
	}
	return nil
}

func (d *delegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	it, ok := item.(Item)
	if !ok {
		return
	}
	width := m.Width()
	if width <= 0 {
		return
	}

	selected := index == m.Index()
	cursor := "  "
	if selected {
		cursor = "> "
	}

	marker := statusMarker(it.task.Status) + " "
	prefix := cursor + marker

	colors := newChipColors(d.hasDarkBg)
	chips := taskChips(it.task, time.Now(), colors, d.project(it.task))
	var chipParts []string
	for _, ch := range chips {
		chipParts = append(chipParts, ch.style.Render(ch.text))
	}
	chipStr := strings.Join(chipParts, " ")
	chipWidth := lipgloss.Width(chipStr)
	if chipWidth > 0 {
		chipWidth += 2 // leading gap between title and chips
	}

	// Title truncates first; chips are short and kept intact.
	titleBudget := width - lipgloss.Width(prefix) - chipWidth
	if titleBudget < 1 {
		titleBudget = 1
	}
	title := ansi.Truncate(it.task.Title, titleBudget, "…")

	ts := titleStyle(it.task.Status)
	if selected {
		ts = ts.Bold(true)
	}
	title = ts.Render(title)

	var b strings.Builder
	b.WriteString(prefix)
	b.WriteString(title)
	if chipStr != "" {
		b.WriteString(" ")
		b.WriteString(chipStr)
	}
	fmt.Fprint(w, b.String())
}

// ShortHelp and FullHelp preserve the new/edit hints the default delegate
// advertised; the list's own help is disabled, so these are advisory only.
func (d *delegate) ShortHelp() []key.Binding {
	return []key.Binding{d.keys.New, d.keys.Edit}
}

func (d *delegate) FullHelp() [][]key.Binding {
	return [][]key.Binding{{d.keys.New, d.keys.Edit}}
}
