// Package savefield is a terminal "[ Save ]" button for use as the last
// field in a [form.Form]. It is a valueless focus placeholder: it carries
// no value of its own and always validates. It does not claim Enter —
// submitting when it holds focus comes from the form's "Enter on the last
// visible field submits" rule, not from the field itself.
package savefield

import (
	tea "charm.land/bubbletea/v2"

	"github.com/qualidafial/gtd-tui/tui/components/form"
	"github.com/qualidafial/gtd-tui/tui/internal/keymap"
)

// Model is a single focusable "Save" button.
type Model struct {
	key     string
	label   string
	visible func(form.Values) bool
	focused bool
}

// Option configures a [Model] at construction time.
type Option func(*Model)

// WithLabel overrides the default "Save" label (e.g. "Update", "Create").
func WithLabel(s string) Option {
	return func(m *Model) { m.label = s }
}

// WithVisible installs a visibility predicate. Default returns true.
func WithVisible(p func(form.Values) bool) Option {
	return func(m *Model) { m.visible = p }
}

// New creates a savefield. key is required and non-empty.
func New(k string, opts ...Option) Model {
	if k == "" {
		panic("savefield: key is required")
	}
	m := Model{
		key:   k,
		label: "Save",
	}
	for _, opt := range opts {
		opt(&m)
	}
	return m
}

// form.Field interface --------------------------------------------------------

func (m Model) Key() string   { return m.key }
func (m Model) Focused() bool { return m.focused }

func (m Model) Visible(v form.Values) bool {
	if m.visible == nil {
		return true
	}
	return m.visible(v)
}

func (m Model) Focus() (form.Field, tea.Cmd) { m.focused = true; return m, nil }
func (m Model) Blur() form.Field             { m.focused = false; return m }

// Update is a no-op: the savefield consumes no keys. When it is the last
// visible field, the form submits on Enter via its own last-field rule.
func (m Model) Update(tea.Msg) (form.Field, tea.Cmd) { return m, nil }

func (m Model) SetWidth(int) form.Field { return m }

func (m Model) View() string {
	text := "  " + m.label + "  "
	if m.focused {
		return form.AccentStyle.Render(text)
	}
	return form.DimStyle.Render(text)
}

func (m Model) Value() any                    { return nil }
func (m Model) Validate() (form.Field, error) { return m, nil }

// Keys returns no bindings: the savefield is a valueless focus placeholder
// and does not claim Enter, so the form's last-field rule receives it and
// submits.
func (m Model) Keys() []keymap.Group { return nil }
