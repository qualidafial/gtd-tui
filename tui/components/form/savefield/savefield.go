// Package savefield is a terminal "[ Save ]" button for use as the last
// field in a [form.Form]. Pressing Enter while focused triggers the same
// path as the form's Save key — Submit runs synchronously and, on
// success, the form emits [form.SubmittedMsg].
package savefield

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"github.com/qualidafial/gtd-tui/tui/components/form"
	"github.com/qualidafial/gtd-tui/tui/internal/keymap"
)

var enterKey = key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "save"))

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

// Update emits [form.SubmitRequestMsg] when Enter is pressed while
// focused. The form intercepts that message and runs its normal Submit
// path — same as ctrl+s, with full validation.
func (m Model) Update(msg tea.Msg) (form.Field, tea.Cmd) {
	if !m.focused {
		return m, nil
	}
	km, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return m, nil
	}
	if km.Code == tea.KeyEnter && km.Mod == 0 {
		return m, func() tea.Msg { return form.SubmitRequestMsg{} }
	}
	return m, nil
}

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

// Keys claims Enter so the form forwards it here (where Update emits
// SubmitRequestMsg) instead of treating it as next-field navigation, and
// advertises it as "enter save".
func (m Model) Keys() []keymap.Group {
	return []keymap.Group{{{Binding: enterKey, Vis: keymap.Short}}}
}
