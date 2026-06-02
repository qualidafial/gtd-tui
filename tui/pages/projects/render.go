package projects

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

var (
	openTitleStyle    = lipgloss.NewStyle()
	somedayTitleStyle = lipgloss.NewStyle().Faint(true)
	doneTitleStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("65")).Faint(true)
	droppedTitleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Faint(true).Strikethrough(true)
)

func projectStatusMarker(s gtd.ProjectStatus) string {
	switch s {
	case gtd.ProjectStatusOpen:
		return "[ ]"
	case gtd.ProjectStatusSomeday:
		return "[?]"
	case gtd.ProjectStatusDone:
		return "[x]"
	case gtd.ProjectStatusDropped:
		return "[-]"
	default:
		return "[ ]"
	}
}

func projectTitleStyle(s gtd.ProjectStatus) lipgloss.Style {
	switch s {
	case gtd.ProjectStatusOpen:
		return openTitleStyle
	case gtd.ProjectStatusSomeday:
		return somedayTitleStyle
	case gtd.ProjectStatusDone:
		return doneTitleStyle
	case gtd.ProjectStatusDropped:
		return droppedTitleStyle
	default:
		return openTitleStyle
	}
}

type chip struct {
	text  string
	style lipgloss.Style
}

type chipColors struct {
	overdue  lipgloss.Style
	dueToday lipgloss.Style
	dueSoon  lipgloss.Style
	dueLater lipgloss.Style
	warning  lipgloss.Style
}

func newChipColors(hasDarkBg bool) chipColors {
	_ = hasDarkBg
	return chipColors{
		overdue:  lipgloss.NewStyle().Foreground(lipgloss.Color("9")), // red
		dueToday: lipgloss.NewStyle().Foreground(lipgloss.Color("208")),
		dueSoon:  lipgloss.NewStyle().Foreground(lipgloss.Color("11")),
		dueLater: lipgloss.NewStyle().Foreground(lipgloss.Color("245")),
		warning:  lipgloss.NewStyle().Foreground(lipgloss.Color("44")), // teal
	}
}

func projectChips(p gtd.Project, counts gtd.ProjectTaskCounts, now time.Time, c chipColors) []chip {
	var chips []chip

	// task progress chip
	if counts.Total > 0 {
		chips = append(chips, chip{
			text:  fmt.Sprintf("%d/%d tasks", counts.Complete, counts.Total),
			style: lipgloss.NewStyle().Foreground(lipgloss.Color("245")),
		})
	}

	// "needs action" warning for open projects with no pending tasks
	// pending = Total - Complete; fires when all non-dropped tasks are done (or no tasks)
	if p.Status == gtd.ProjectStatusOpen && counts.Complete == counts.Total {
		chips = append(chips, chip{text: "needs action", style: c.warning})
	}

	// due chip — only meaningful for open projects
	if p.Due != nil && p.Status == gtd.ProjectStatusOpen {
		chips = append(chips, dueChip(p.Due, now, c))
	}

	return chips
}

func dueChip(due *time.Time, now time.Time, c chipColors) chip {
	d := due.Local()
	ref := d
	if isLocalMidnight(d) {
		ref = endOfLocalDay(d)
	}
	when := reltime.Format(d, now)
	if ref.Before(now) {
		return chip{text: "overdue:" + when, style: c.overdue}
	}
	return chip{text: "due:" + when, style: dueStyle(d, now, c)}
}

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

func isLocalMidnight(t time.Time) bool {
	return t.Hour() == 0 && t.Minute() == 0 && t.Second() == 0
}

func endOfLocalDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, int(time.Second-time.Nanosecond), t.Location())
}

func calendarDaysBetween(a, b time.Time) int {
	aDay := truncateToDay(a.Local())
	bDay := truncateToDay(b.Local())
	if bDay.Before(aDay) {
		return -int((aDay.Sub(bDay).Hours() + 12) / 24)
	}
	return int((bDay.Sub(aDay).Hours() + 12) / 24)
}

func truncateToDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

type delegate struct {
	keys      KeyMap
	hasDarkBg bool
}

func newDelegate(keys KeyMap) *delegate {
	return &delegate{keys: keys, hasDarkBg: true}
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

	marker := projectStatusMarker(it.project.Status) + " "
	prefix := cursor + marker

	colors := newChipColors(d.hasDarkBg)
	chips := projectChips(it.project, it.counts, time.Now(), colors)
	var chipParts []string
	for _, ch := range chips {
		chipParts = append(chipParts, ch.style.Render(ch.text))
	}
	chipStr := strings.Join(chipParts, " ")
	chipWidth := lipgloss.Width(chipStr)
	if chipWidth > 0 {
		chipWidth += 2 // leading gap
	}

	titleBudget := width - lipgloss.Width(prefix) - chipWidth
	if titleBudget < 1 {
		titleBudget = 1
	}
	title := ansi.Truncate(it.project.Title, titleBudget, "…")

	ts := projectTitleStyle(it.project.Status)
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

func (d *delegate) ShortHelp() []key.Binding {
	return []key.Binding{d.keys.New}
}

func (d *delegate) FullHelp() [][]key.Binding {
	return [][]key.Binding{{d.keys.New}}
}
