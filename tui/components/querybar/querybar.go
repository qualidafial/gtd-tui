// Package querybar provides a reusable single-line query bar component for
// TUI screens. It wraps a textinput with focus/blur management, debounced
// validation, parse-error state, and inline error highlighting.
package querybar

import (
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/qualidafial/gtd-tui/tui/cmds"
)

// ParseError reports an invalid token in a query string. Start and End are
// rune offsets into the original query marking the [Start, End) range of the
// offending token, enabling inline range highlighting.
type ParseError struct {
	Message string
	Start   int
	End     int
}

func (e *ParseError) Error() string { return e.Message }

// ApplyMsg carries the query string the parent should load. It is emitted on
// three paths: enter (commit), esc (revert to the previously-applied query),
// and a successful debounced parse while focused (live preview). The parent
// handles all three the same way: parse Query into a typed filter and reload.
type ApplyMsg struct{ Query string }

// debounceMsg fires after a configurable delay following the last keystroke.
type debounceMsg struct{ seq int }

var errorHighlight = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Underline(true)

// Model is the query bar component state.
type Model struct {
	KeyMap       KeyMap
	input        textinput.Model
	validate     func(string) *ParseError
	debounce     time.Duration
	parseErr     *ParseError
	appliedQuery string
	focused      bool
	debounceSeq  int
}

// New creates a Model with the given prompt, placeholder, debounce interval,
// and validation function. Call SetValue to seed an initial query.
func New(prompt, placeholder string, debounce time.Duration, validate func(string) *ParseError) Model {
	ti := textinput.New()
	ti.Prompt = prompt
	ti.Placeholder = placeholder
	return Model{
		KeyMap:   DefaultKeyMap(),
		input:    ti,
		validate: validate,
		debounce: debounce,
	}
}

// Value returns the current text value of the query bar.
func (m Model) Value() string { return m.input.Value() }

// SetValue sets the text value and records it as the applied query.
func (m *Model) SetValue(s string) {
	m.input.SetValue(s)
	m.appliedQuery = s
}

// SetWidth sets the width of the underlying textinput.
func (m *Model) SetWidth(w int) {
	m.input.SetWidth(w)
}

func (m Model) ShortHelp() []key.Binding {
	return m.KeyMap.ShortHelp()
}

func (m Model) FullHelp() [][]key.Binding {
	return m.KeyMap.FullHelp()
}

// CapturingInput reports whether the query bar is focused.
func (m Model) CapturingInput() bool { return m.focused }

// Focus activates the query bar. If the current value is non-empty, a trailing
// space is appended so that new typing starts a separate token, and the cursor
// is placed at the end.
func (m Model) Focus() (Model, tea.Cmd) {
	m.focused = true
	if m.input.Value() != "" && !strings.HasSuffix(m.input.Value(), " ") {
		m.input.SetValue(m.input.Value() + " ")
	}
	cmd := m.input.Focus()
	m.input.CursorEnd()
	return m, cmd
}

// Blur deactivates the query bar and trims whitespace from the value.
func (m Model) Blur() Model {
	m.focused = false
	m.input.SetValue(strings.TrimSpace(m.input.Value()))
	m.input.Blur()
	return m
}

// Update processes messages for the query bar. When focused, it handles enter
// (apply), esc (cancel), and forwards other messages to the underlying
// textinput, scheduling a debounce validation tick after each keystroke.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case debounceMsg:
		if !m.focused {
			return m, nil
		}
		if msg.seq != m.debounceSeq {
			return m, nil
		}
		trimmed := strings.TrimSpace(m.input.Value())
		if pe := m.validate(trimmed); pe != nil {
			m.parseErr = pe
			return m, cmds.Emit(pe)
		}
		m.parseErr = nil
		return m, func() tea.Msg { return ApplyMsg{Query: trimmed} }
	case tea.KeyPressMsg:
		if !m.focused {
			return m, nil
		}
		// clear error on any key
		m.parseErr = nil
		switch {
		case key.Matches(msg, m.KeyMap.Apply):
			trimmed := strings.TrimSpace(m.input.Value())
			m.input.SetValue(trimmed)
			if pe := m.validate(trimmed); pe != nil {
				m.parseErr = pe
				return m, cmds.Emit(pe)
			}
			m.parseErr = nil
			m.appliedQuery = trimmed
			m.focused = false
			m.input.Blur()
			query := trimmed
			return m, func() tea.Msg { return ApplyMsg{Query: query} }
		case key.Matches(msg, m.KeyMap.Cancel):
			m.input.SetValue(m.appliedQuery)
			m.parseErr = nil
			m.focused = false
			m.input.Blur()
			return m, func() tea.Msg { return ApplyMsg{Query: m.appliedQuery} }
		default:
			// Forward the keystroke to the textinput and schedule a debounce tick.
			// Only keypresses bump the seq — cursor-blink and other internal ticks
			// would otherwise reset the window faster than it can elapse, so the
			// validation tick would always land stale and never fire.
			var cmd tea.Cmd
			m.input, cmd = m.input.Update(msg)
			m.debounceSeq++
			seq := m.debounceSeq
			tick := tea.Tick(m.debounce, func(time.Time) tea.Msg {
				return debounceMsg{seq: seq}
			})
			return m, tea.Batch(cmd, tick)
		}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

// View renders the query bar as a single line. When a ParseError is present,
// the offending rune range is highlighted with red foreground and underline
// using ansi.Cut-based post-processing of the textinput output.
func (m Model) View() string {
	view := m.input.View()
	if m.parseErr == nil {
		return view
	}

	promptW := lipgloss.Width(m.input.Prompt)
	start := promptW + m.parseErr.Start
	end := promptW + m.parseErr.End
	if end <= start {
		end = start + 1
	}

	before := ansi.Cut(view, 0, start)
	offending := ansi.Cut(view, start, end)
	after := ansi.Cut(view, end, len(view)*4)

	return lipgloss.JoinHorizontal(lipgloss.Left,
		before,
		errorHighlight.Render(ansi.Strip(offending)),
		after,
	)
}
