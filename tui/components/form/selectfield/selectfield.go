// Package selectfield is a single-select field for use in a [form.Form].
// It wraps [bubbles/v2/list] and implements [form.Field]. Filtering (`/`)
// and arrow-key navigation are inherited from the list widget.
package selectfield

import (
	"fmt"
	"io"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/qualidafial/gtd-tui/tui/components/form"
)


// Option carries a display string and a value of T to be returned when
// the option is selected.
type Option[T comparable] struct {
	Display string
	Value   T
}

// item adapts an Option[T] to the [list.Item] interface.
type item[T comparable] struct {
	opt Option[T]
}

func (i item[T]) FilterValue() string { return i.opt.Display }

// delegate is a one-line item renderer with a focus marker. It is
// independent of T so it can be shared across all selectfield
// instantiations.
type delegate struct{}

func (delegate) Height() int                         { return 1 }
func (delegate) Spacing() int                        { return 0 }
func (delegate) Update(tea.Msg, *list.Model) tea.Cmd { return nil }
func (delegate) Render(w io.Writer, m list.Model, index int, it list.Item) {
	s := it.FilterValue()
	if index == m.Index() {
		_, _ = fmt.Fprint(w, form.FocusedLabelStyle.Render("> "+s))
		return
	}
	_, _ = fmt.Fprint(w, "  "+s)
}

// Model is a single-select field.
type Model[T comparable] struct {
	key           string
	label         string
	list          list.Model
	validator     func(T) error
	visible       func(form.Values) bool
	focused       bool
	submitOnEnter bool
	initialMatch  func(T) bool

	err     error
	lastVal T
}

// fieldOption configures a [Model] at construction time. (Named
// `fieldOption` internally so it doesn't shadow the user-facing
// [Option][T] item type.)
type fieldOption[T comparable] func(*Model[T])

// WithValidator installs a validator run by Validate. It receives the
// currently selected value of T.
func WithValidator[T comparable](fn func(T) error) fieldOption[T] {
	return func(m *Model[T]) { m.validator = fn }
}

// WithVisible installs a visibility predicate. The default visibility
// returns true.
func WithVisible[T comparable](p func(form.Values) bool) fieldOption[T] {
	return func(m *Model[T]) { m.visible = p }
}

// WithHeight sets the rendered list height in rows. The default is the
// number of options, capped at 8.
func WithHeight[T comparable](rows int) fieldOption[T] {
	return func(m *Model[T]) { m.list.SetHeight(rows) }
}

// WithSubmitOnEnter makes Enter on a non-filtering list commit the
// selection by emitting [form.SubmitRequestMsg]. Useful when the
// selectfield is the only field of a single-pick overlay where there is
// no trailing savefield. While the list is in the Filtering state Enter
// is left to the list (it accepts the filter); once the filter is
// applied or unfiltered, Enter submits.
func WithSubmitOnEnter[T comparable]() fieldOption[T] {
	return func(m *Model[T]) { m.submitOnEnter = true }
}

// WithInitialValue selects the option whose Value equals v at
// construction time. Applied after all other options (including
// WithNone) so the index is correct.
func WithInitialValue[T comparable](v T) fieldOption[T] {
	return func(m *Model[T]) {
		m.initialMatch = func(got T) bool { return got == v }
	}
}

// WithInitialFn selects the first option whose Value satisfies match at
// construction time. Use when T has no useful `==` (e.g. pointer types
// where equality should be by pointed-to value). Applied after all
// other options.
func WithInitialFn[T comparable](match func(T) bool) fieldOption[T] {
	return func(m *Model[T]) { m.initialMatch = match }
}

// WithNone prepends a synthetic option that, when selected, yields the
// zero value of T (for pointer-typed T, that is nil). Intended for
// "(none)"-style entries on optional pointer-typed selects.
func WithNone[T comparable](display string) fieldOption[T] {
	return func(m *Model[T]) {
		items := m.list.Items()
		none := item[T]{opt: Option[T]{Display: display}} // Value is zero
		newItems := append([]list.Item{none}, items...)
		_ = m.list.SetItems(newItems)
	}
}

// New creates a selectfield over options. key is required and non-empty.
// The first option in the resulting list is initially selected.
func New[T comparable](k, label string, options []Option[T], opts ...fieldOption[T]) Model[T] {
	if k == "" {
		panic("selectfield: key is required")
	}

	items := make([]list.Item, len(options))
	for i, o := range options {
		items[i] = item[T]{opt: o}
	}

	defaultHeight := len(options)
	const heightCap = 8
	if defaultHeight > heightCap {
		defaultHeight = heightCap
	}
	if defaultHeight < 1 {
		defaultHeight = 1
	}

	l := list.New(items, delegate{}, 0, defaultHeight)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetShowPagination(false)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(true)

	m := Model[T]{
		key:   k,
		label: label,
		list:  l,
	}
	for _, opt := range opts {
		opt(&m)
	}
	if m.initialMatch != nil {
		for i, it := range m.list.Items() {
			if got, ok := it.(item[T]); ok && m.initialMatch(got.opt.Value) {
				m.list.Select(i)
				break
			}
		}
	}
	return m
}

// form.Field interface --------------------------------------------------------

func (m Model[T]) Key() string { return m.key }

// Focused reports whether the field has focus. The list widget does not
// have an internal focus flag, so the field tracks it via a sentinel: a
// selectfield is "focused" when its list has any items selected by index.
// In practice we track focus through a small companion flag.
func (m Model[T]) Focused() bool { return m.focused }

func (m Model[T]) Visible(v form.Values) bool {
	if m.visible == nil {
		return true
	}
	return m.visible(v)
}

func (m Model[T]) Focus() (form.Field, tea.Cmd) {
	m.focused = true
	return m, nil
}

func (m Model[T]) Blur() form.Field {
	m.focused = false
	return m
}

func (m Model[T]) Update(msg tea.Msg) (form.Field, tea.Cmd) {
	if !m.focused {
		return m, nil
	}
	if m.submitOnEnter {
		if km, ok := msg.(tea.KeyPressMsg); ok &&
			km.Code == tea.KeyEnter && km.Mod == 0 &&
			m.list.FilterState() != list.Filtering {
			return m, func() tea.Msg { return form.SubmitRequestMsg{} }
		}
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	if m.err != nil && m.currentValue() != m.lastVal {
		m.err = nil
	}
	return m, cmd
}

func (m Model[T]) SetWidth(w int) form.Field {
	m.list.SetWidth(w)
	return m
}

func (m Model[T]) View() string {
	var b []string
	if m.label != "" {
		style := form.LabelStyle
		if m.focused {
			style = form.FocusedLabelStyle
		}
		b = append(b, style.Render(m.label))
	}
	b = append(b, m.list.View())
	if m.err != nil {
		b = append(b, form.ErrorStyle.Render(m.err.Error()))
	}
	return lipgloss.JoinVertical(lipgloss.Left, b...)
}

func (m Model[T]) Value() any { return m.currentValue() }

// SelectedValue returns the currently selected T. If the list has no
// items it returns the zero value of T.
func (m Model[T]) SelectedValue() T { return m.currentValue() }


func (m Model[T]) currentValue() T {
	if it, ok := m.list.SelectedItem().(item[T]); ok {
		return it.opt.Value
	}
	var zero T
	return zero
}

func (m Model[T]) Validate() (form.Field, error) {
	v := m.currentValue()
	m.lastVal = v
	if m.validator == nil {
		m.err = nil
		return m, nil
	}
	m.err = m.validator(v)
	return m, m.err
}

func (m Model[T]) ShortHelp() []key.Binding {
	return m.list.ShortHelp()
}

func (m Model[T]) FullHelp() [][]key.Binding {
	return m.list.FullHelp()
}

