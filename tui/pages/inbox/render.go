package inbox

import (
	"fmt"
	"io"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"github.com/qualidafial/gtd-tui/tui/theme"
)

var (
	titleStyle       = lipgloss.NewStyle()
	descriptionStyle = theme.Value
)

type delegate struct {
	keys KeyMap
}

func newDelegate(keys KeyMap) *delegate { return &delegate{keys: keys} }

func (d *delegate) Height() int                         { return 1 }
func (d *delegate) Spacing() int                        { return 0 }
func (d *delegate) Update(tea.Msg, *list.Model) tea.Cmd { return nil }

func (d *delegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	it, ok := item.(Item)
	if !ok {
		return
	}
	width := m.Width()
	if width <= 0 {
		return
	}

	cursor := "  "
	if index == m.Index() {
		cursor = "> "
	}

	title := flatten(it.item.Title)
	desc := flatten(it.item.Description)

	ts := titleStyle
	if index == m.Index() {
		ts = ts.Bold(true)
	}

	descPart := ""
	if desc != "" {
		descPart = "  " + descriptionStyle.Render(ansi.Truncate(desc, max(width/3, 8), "…"))
	}

	budget := width - lipgloss.Width(cursor) - lipgloss.Width(descPart)
	if budget < 1 {
		budget = 1
	}
	titleRendered := ts.Render(ansi.Truncate(title, budget, "…"))

	var b strings.Builder
	b.WriteString(cursor)
	b.WriteString(titleRendered)
	b.WriteString(descPart)
	fmt.Fprint(w, b.String())
}

// flatten collapses runs of whitespace (including newlines) into single
// spaces so a multiline title or description renders on a single row.
func flatten(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func (d *delegate) ShortHelp() []key.Binding  { return nil }
func (d *delegate) FullHelp() [][]key.Binding { return nil }
