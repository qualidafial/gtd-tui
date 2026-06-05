// Package datefield is a single-line date/time field for use in a
// [form.Form]. It wraps [bubbles/v2/textinput] with a natural-language
// date parser ([naturaltime]) and implements [form.Field].
//
// The Value type is *time.Time so empty input parses to nil. Accepted
// input formats:
//
//   - YYYY-MM-DD
//   - YYYY-MM-DD HH:MM
//   - YYYY-MM-DD HH:MM:SS
//   - natural-language expressions parsed by naturaltime (e.g.
//     "tomorrow", "in 3 hours", "next monday at 9am")
//
// On Blur, if the text parses to a non-nil time, the displayed text is
// replaced with its absolute representation (e.g. "tomorrow at 3pm" →
// "2026-06-02 15:00") so the user can see exactly what will be saved.
package datefield

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/sho0pi/naturaltime"

	"github.com/qualidafial/gtd-tui/tui/components/form"
	"github.com/qualidafial/gtd-tui/tui/internal/keymap"
)

// Model is a date/time field bound to a *time.Time. Empty input means nil.
type Model struct {
	key       string
	label     string
	input     textinput.Model
	validator func(*time.Time) error
	visible   func(form.Values) bool

	err     error
	lastVal string
}

// Option configures a [Model] at construction time.
type Option func(*Model)

// WithValidator installs a validator run by Validate. It receives the
// parsed *time.Time (nil if the input is empty).
func WithValidator(fn func(*time.Time) error) Option {
	return func(m *Model) { m.validator = fn }
}

// WithPlaceholder sets the textinput placeholder displayed when empty.
func WithPlaceholder(s string) Option {
	return func(m *Model) { m.input.Placeholder = s }
}

// WithValue seeds the field's initial value. Nil renders as empty.
func WithValue(t *time.Time) Option {
	return func(m *Model) { m.input.SetValue(formatDate(t)) }
}

// WithVisible installs a visibility predicate. Default returns true.
func WithVisible(p func(form.Values) bool) Option {
	return func(m *Model) { m.visible = p }
}

// New creates a datefield. key is required and non-empty.
func New(k, label string, opts ...Option) Model {
	if k == "" {
		panic("datefield: key is required")
	}
	ti := textinput.New()
	ti.Placeholder = "YYYY-MM-DD [HH:MM[:SS]] or e.g. \"tomorrow\""

	m := Model{
		key:   k,
		label: label,
		input: ti,
	}
	for _, opt := range opts {
		opt(&m)
	}
	return m
}

// form.Field interface --------------------------------------------------------

func (m Model) Key() string   { return m.key }
func (m Model) Focused() bool { return m.input.Focused() }

func (m Model) Visible(v form.Values) bool {
	if m.visible == nil {
		return true
	}
	return m.visible(v)
}

func (m Model) Focus() (form.Field, tea.Cmd) {
	cmd := m.input.Focus()
	return m, cmd
}

// Blur snaps the displayed text to the absolute representation of the
// parsed value (e.g. "tomorrow at 3pm" → "2026-06-02 15:00") so the user
// can see what will be saved. If parsing fails the text is left as-is
// and the cached error is set.
func (m Model) Blur() form.Field {
	m.input.Blur()
	t, err := parseDate(m.input.Value())
	if err != nil {
		m.err = err
		return m
	}
	m.input.SetValue(formatDate(t))
	m.lastVal = m.input.Value()
	m.err = nil
	return m
}

func (m Model) Update(msg tea.Msg) (form.Field, tea.Cmd) {
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	if m.err != nil && m.input.Value() != m.lastVal {
		m.err = nil
	}
	return m, cmd
}

func (m Model) SetWidth(w int) form.Field {
	m.input.SetWidth(w)
	return m
}

func (m Model) View() string {
	var b []string
	if m.label != "" {
		style := form.LabelStyle
		if m.Focused() {
			style = form.FocusedLabelStyle
		}
		b = append(b, style.Render(m.label))
	}
	b = append(b, m.input.View())
	if m.err != nil {
		b = append(b, form.ErrorStyle.Render(m.err.Error()))
	}
	return lipgloss.JoinVertical(lipgloss.Left, b...)
}

// Value returns the parsed *time.Time as an `any`. Empty input is nil.
// Parse errors return nil and leave the cached error populated for View.
func (m Model) Value() any {
	t, err := parseDate(m.input.Value())
	if err != nil {
		return (*time.Time)(nil)
	}
	return t
}

// SelectedTime returns the parsed time directly. nil means empty or
// unparseable input.
func (m Model) SelectedTime() *time.Time {
	t, err := parseDate(m.input.Value())
	if err != nil {
		return nil
	}
	return t
}

func (m Model) Validate() (form.Field, error) {
	m.lastVal = m.input.Value()
	t, err := parseDate(m.lastVal)
	if err != nil {
		m.err = err
		return m, err
	}
	if m.validator != nil {
		if vErr := m.validator(t); vErr != nil {
			m.err = vErr
			return m, vErr
		}
	}
	m.err = nil
	return m, nil
}

// Keys returns no bindings: a datefield is a free-text input whose runes
// are consumed by the textinput while focused; it claims and advertises
// nothing that collides with form navigation.
func (m Model) Keys() []keymap.Group { return nil }

// Parsing / formatting --------------------------------------------------------

var dateLayouts = []string{
	"2006-01-02 15:04:05",
	"2006-01-02 15:04",
	"2006-01-02",
}

// naturalParser is the lazily-initialized naturaltime parser. The first
// call compiles an embedded JS program in a goja runtime; once is enough.
var naturalParser = sync.OnceValues(naturaltime.New)

// parseDate returns nil for empty input, or a parsed local time.
func parseDate(s string) (*time.Time, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil, nil
	}
	for _, layout := range dateLayouts {
		if t, err := time.ParseInLocation(layout, s, time.Local); err == nil {
			return &t, nil
		}
	}

	parser, err := naturalParser()
	if err != nil {
		return nil, fmt.Errorf("init natural-language date parser: %w", err)
	}
	t, err := parser.ParseDate(s, time.Now())
	if err != nil {
		return nil, fmt.Errorf("invalid date %q: %w", s, err)
	}
	if t == nil {
		return nil, fmt.Errorf("invalid date %q", s)
	}
	return t, nil
}

// formatDate renders nil as "" and chooses date-only vs date+time based
// on whether the time component is midnight local.
func formatDate(t *time.Time) string {
	if t == nil {
		return ""
	}
	local := t.Local()
	if local.Hour() == 0 && local.Minute() == 0 && local.Second() == 0 {
		return local.Format("2006-01-02")
	}
	return local.Format("2006-01-02 15:04")
}
