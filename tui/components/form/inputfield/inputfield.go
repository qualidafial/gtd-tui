// Package inputfield is a single-line text field for use in a [form.Form].
// It wraps [bubbles/v2/textinput] and implements [form.Field].
package inputfield

import (
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/qualidafial/gtd-tui/tui/components/form"
	"github.com/qualidafial/gtd-tui/tui/internal/keymap"
)

// Model is a single-line text input field.
type Model struct {
	key       string
	label     string
	input     textinput.Model
	validator func(string) error
	visible   func(form.Values) bool

	err     error  // cached result of last Validate; cleared on value change
	lastVal string // value at last Validate, used to detect change
}

// Option configures a [Model] at construction time.
type Option func(*Model)

// WithValidator installs a validator run by Validate. It receives the
// field's current string value.
func WithValidator(fn func(string) error) Option {
	return func(m *Model) { m.validator = fn }
}

// WithPlaceholder sets the textinput placeholder displayed when the value
// is empty.
func WithPlaceholder(s string) Option {
	return func(m *Model) { m.input.Placeholder = s }
}

// WithValue seeds the field's initial value.
func WithValue(s string) Option {
	return func(m *Model) { m.input.SetValue(s) }
}

// WithVisible installs a visibility predicate. The default visibility
// returns true.
func WithVisible(p func(form.Values) bool) Option {
	return func(m *Model) { m.visible = p }
}

// New creates an inputfield. key is required and must be non-empty.
func New(k, label string, opts ...Option) Model {
	if k == "" {
		panic("inputfield: key is required")
	}
	m := Model{
		key:   k,
		label: label,
		input: textinput.New(),
	}
	for _, opt := range opts {
		opt(&m)
	}
	return m
}

// SetValue replaces the current text value.
func (m Model) SetValue(s string) Model {
	m.input.SetValue(s)
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

func (m Model) Blur() form.Field {
	m.input.Blur()
	return m
}

func (m Model) Update(msg tea.Msg) (form.Field, tea.Cmd) {
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	if m.err != nil && m.input.Value() != m.lastVal {
		m.err = nil // stop showing a stale error once the user edits the value
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

func (m Model) Value() any { return m.input.Value() }

func (m Model) Validate() (form.Field, error) {
	if m.validator == nil {
		m.err = nil
		m.lastVal = m.input.Value()
		return m, nil
	}
	m.lastVal = m.input.Value()
	m.err = m.validator(m.lastVal)
	return m, m.err
}

// Chords returns no bindings: a single-line text field consumes only
// free-text runes (handled by the textinput while focused) and the cursor
// keys, none of which it advertises or needs to claim from form
// navigation. Left/right cursor movement does not collide with the form's
// tab/up/down navigation.
func (m Model) Chords() []keymap.Group { return nil }
