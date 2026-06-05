// Package textfield is a multi-line text field for use in a [form.Form].
// It wraps [bubbles/v2/textarea] and implements [form.Field].
//
// Unlike the default textarea binding, plain Enter does NOT insert a
// newline and is not consumed at all — the form's navigation rules
// (`tab`/`shift+tab`/`ctrl+s`) take precedence. To insert a newline, use
// `alt+enter` or `ctrl+j`; those bindings appear in the form's help
// footer while the textfield is focused.
package textfield

import (
	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/qualidafial/gtd-tui/tui/components/form"
	"github.com/qualidafial/gtd-tui/tui/internal/keymap"
)

// Model is a multi-line text field.
type Model struct {
	key       string
	label     string
	area      textarea.Model
	validator func(string) error
	visible   func(form.Values) bool

	KeyMap KeyMap

	err     error
	lastVal string
}

// Option configures a [Model] at construction time.
type Option func(*Model)

// WithValidator installs a validator run by Validate. It receives the
// field's current string value.
func WithValidator(fn func(string) error) Option {
	return func(m *Model) { m.validator = fn }
}

// WithPlaceholder sets the textarea placeholder displayed when the value
// is empty.
func WithPlaceholder(s string) Option {
	return func(m *Model) { m.area.Placeholder = s }
}

// WithValue seeds the field's initial value.
func WithValue(s string) Option {
	return func(m *Model) { m.area.SetValue(s) }
}

// WithVisible installs a visibility predicate. The default visibility
// returns true.
func WithVisible(p func(form.Values) bool) Option {
	return func(m *Model) { m.visible = p }
}

// New creates a textfield. key is required and must be non-empty.
func New(k, label string, opts ...Option) Model {
	if k == "" {
		panic("textfield: key is required")
	}

	keyMap := DefaultKeyMap()

	ta := textarea.New()
	ta.ShowLineNumbers = false
	// Rebind newline so plain Enter does not insert one — the form needs
	// Enter to remain available (or, more precisely, the textarea should
	// not consume it).
	ta.KeyMap.InsertNewline = keyMap.InsertNewline

	m := Model{
		key:    k,
		label:  label,
		area:   ta,
		KeyMap: keyMap,
	}
	for _, opt := range opts {
		opt(&m)
	}
	return m
}

// SetValue replaces the current text value.
func (m Model) SetValue(s string) Model {
	m.area.SetValue(s)
	return m
}

// form.Field interface --------------------------------------------------------

func (m Model) Key() string   { return m.key }
func (m Model) Focused() bool { return m.area.Focused() }

func (m Model) Visible(v form.Values) bool {
	if m.visible == nil {
		return true
	}
	return m.visible(v)
}

func (m Model) Focus() (form.Field, tea.Cmd) {
	cmd := m.area.Focus()
	return m, cmd
}

func (m Model) Blur() form.Field {
	m.area.Blur()
	return m
}

func (m Model) Update(msg tea.Msg) (form.Field, tea.Cmd) {
	// Drop plain Enter before the textarea sees it: the textarea's default
	// InsertNewline is rebound to alt+enter/ctrl+j, but Enter would still
	// move the cursor in some bubbles versions. Returning early keeps the
	// keystroke available for the form (and prevents accidental input).
	if km, ok := msg.(tea.KeyPressMsg); ok {
		if km.Code == tea.KeyEnter && km.Mod == 0 {
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.area, cmd = m.area.Update(msg)
	if m.err != nil && m.area.Value() != m.lastVal {
		m.err = nil
	}
	return m, cmd
}

func (m Model) SetWidth(w int) form.Field {
	m.area.SetWidth(w)
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
	b = append(b, m.area.View())
	if m.err != nil {
		b = append(b, form.ErrorStyle.Render(m.err.Error()))
	}
	return lipgloss.JoinVertical(lipgloss.Left, b...)
}

func (m Model) Value() any { return m.area.Value() }

func (m Model) Validate() (form.Field, error) {
	if m.validator == nil {
		m.err = nil
		m.lastVal = m.area.Value()
		return m, nil
	}
	m.lastVal = m.area.Value()
	m.err = m.validator(m.lastVal)
	return m, m.err
}

func (m Model) Keys() []keymap.Group { return m.KeyMap.Keys() }
