// Package radiofield is an inline single-select field for use in a
// [form.Form]. It renders all options on one line with a radio-button
// affordance — `( ) Task  (•) Project` — and moves between them with
// left/right arrows. Equivalent to `huh.NewSelect().Inline(true)`.
//
// Intended for binary or small-N choices (task vs. project, self vs.
// delegated, yes/no). There is no filtering and no scroll.
package radiofield

import (
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/qualidafial/gtd-tui/tui/components/form"
	"github.com/qualidafial/gtd-tui/tui/internal/keymap"
)

var (
	leftKey  = key.NewBinding(key.WithKeys("left"))
	rightKey = key.NewBinding(key.WithKeys("right"))
	// chooseKey is the displayed/claimed binding spanning both arrows;
	// leftKey/rightKey drive direction in Update.
	chooseKey = key.NewBinding(key.WithKeys("left", "right"), key.WithHelp("←/→", "choose"))
)

// Option carries a display string and a value of T to be returned when
// the option is selected.
type Option[T comparable] struct {
	Display string
	Value   T
}

// Model is an inline single-select field.
type Model[T comparable] struct {
	key       string
	label     string
	options   []Option[T]
	idx       int
	validator func(T) error
	visible   func(form.Values) bool
	focused   bool

	err     error
	lastVal T
}

// fieldOption configures a [Model] at construction time. Named
// `fieldOption` so it does not shadow the user-facing [Option][T].
type fieldOption[T comparable] func(*Model[T])

// WithValidator installs a validator run by Validate.
func WithValidator[T comparable](fn func(T) error) fieldOption[T] {
	return func(m *Model[T]) { m.validator = fn }
}

// WithVisible installs a visibility predicate.
func WithVisible[T comparable](p func(form.Values) bool) fieldOption[T] {
	return func(m *Model[T]) { m.visible = p }
}

// WithInitialValue selects the option whose Value equals v at
// construction time. If no option matches, the first option remains
// selected.
func WithInitialValue[T comparable](v T) fieldOption[T] {
	return func(m *Model[T]) {
		for i, o := range m.options {
			if o.Value == v {
				m.idx = i
				return
			}
		}
	}
}

// New creates a radiofield. key is required. options must be non-empty.
func New[T comparable](k, label string, options []Option[T], opts ...fieldOption[T]) Model[T] {
	if k == "" {
		panic("radiofield: key is required")
	}
	if len(options) == 0 {
		panic("radiofield: at least one option is required")
	}
	m := Model[T]{
		key:     k,
		label:   label,
		options: append([]Option[T](nil), options...),
		idx:     0,
	}
	for _, opt := range opts {
		opt(&m)
	}
	return m
}

// form.Field interface --------------------------------------------------------

func (m Model[T]) Key() string   { return m.key }
func (m Model[T]) Focused() bool { return m.focused }
func (m Model[T]) Visible(v form.Values) bool {
	if m.visible == nil {
		return true
	}
	return m.visible(v)
}

func (m Model[T]) Focus() (form.Field, tea.Cmd) { m.focused = true; return m, nil }
func (m Model[T]) Blur() form.Field             { m.focused = false; return m }

func (m Model[T]) Update(msg tea.Msg) (form.Field, tea.Cmd) {
	if !m.focused {
		return m, nil
	}
	km, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return m, nil
	}
	switch {
	case key.Matches(km, leftKey):
		if m.idx > 0 {
			m.idx--
		}
	case key.Matches(km, rightKey):
		if m.idx < len(m.options)-1 {
			m.idx++
		}
	default:
		return m, nil
	}
	if m.err != nil && m.options[m.idx].Value != m.lastVal {
		m.err = nil
	}
	return m, nil
}

func (m Model[T]) SetWidth(int) form.Field { return m }

func (m Model[T]) View() string {
	parts := make([]string, 0, len(m.options))
	selectedStyle := form.AccentStyle
	if !m.focused {
		selectedStyle = form.MutedAccentStyle
	}
	for i, o := range m.options {
		text := "  " + o.Display + "  "
		style := form.DimStyle
		if i == m.idx {
			style = selectedStyle
		}
		parts = append(parts, style.Render(text))
	}
	row := strings.Join(parts, "  ")

	var b []string
	if m.label != "" {
		style := form.LabelStyle
		if m.focused {
			style = form.FocusedLabelStyle
		}
		b = append(b, style.Render(m.label))
	}
	b = append(b, row)
	if m.err != nil {
		b = append(b, form.ErrorStyle.Render(m.err.Error()))
	}
	return lipgloss.JoinVertical(lipgloss.Left, b...)
}

func (m Model[T]) Value() any { return m.options[m.idx].Value }

// SelectedValue returns the currently selected T directly.
func (m Model[T]) SelectedValue() T { return m.options[m.idx].Value }

func (m Model[T]) Validate() (form.Field, error) {
	v := m.options[m.idx].Value
	m.lastVal = v
	if m.validator == nil {
		m.err = nil
		return m, nil
	}
	m.err = m.validator(v)
	return m, m.err
}

// Keys claims left/right so the form forwards them here for selection
// (rather than treating them as navigation) and advertises them as
// "←/→ choose".
func (m Model[T]) Keys() []keymap.Group {
	return []keymap.Group{{{Binding: chooseKey, Vis: keymap.Short}}}
}
