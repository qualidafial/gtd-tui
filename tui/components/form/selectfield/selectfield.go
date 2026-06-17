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
	"github.com/qualidafial/gtd-tui/tui/internal/keymap"
)

var (
	// moveKey claims up/down (with hidden vim aliases k/j) so the list, not
	// form navigation, receives them while focused.
	moveKey = key.NewBinding(
		key.WithKeys("up", "k", "down", "j"),
		key.WithHelp("↑/↓", "choose"),
	)
	// filterKey advertises and claims the list's filter trigger.
	filterKey = key.NewBinding(key.WithKeys("/"), key.WithHelp("/", "filter"))
	// submitKey is claimed only while the list is filtering, so Enter routes
	// to the list (accepting the filter) instead of submitting the form via
	// its last-field rule.
	submitKey = key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select"))
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
	hideWhenEmpty bool
	initialMatch  func(T) bool

	// Option configuration replayed by setOptions whenever the option set
	// changes, so dynamic updates preserve construction-time behavior.
	hasNone        bool
	noneDisplay    string
	explicitHeight bool

	err     error
	lastVal T
}

// FieldOption configures a [Model] at construction time. (Named
// `FieldOption` rather than `Option` so it doesn't shadow the user-facing
// [Option][T] item type.)
type FieldOption[T comparable] func(*Model[T])

// WithValidator installs a validator run by Validate. It receives the
// currently selected value of T.
func WithValidator[T comparable](fn func(T) error) FieldOption[T] {
	return func(m *Model[T]) { m.validator = fn }
}

// WithVisible installs a visibility predicate. The default visibility
// returns true.
func WithVisible[T comparable](p func(form.Values) bool) FieldOption[T] {
	return func(m *Model[T]) { m.visible = p }
}

// WithHideWhenEmpty hides the field while it has no real (non-WithNone)
// options. Combined with [Model.SetOptions], this lets a field be built up
// front and stay hidden until its options load — or remain hidden when the
// loaded set turns out to be empty. It composes with WithVisible: the field
// shows only when it has options AND the predicate passes.
func WithHideWhenEmpty[T comparable]() FieldOption[T] {
	return func(m *Model[T]) { m.hideWhenEmpty = true }
}

// WithHeight sets the rendered list height in rows. The default is the
// number of options, capped at 8. Setting an explicit height pins it, so
// later SetOptions calls no longer auto-size the list.
func WithHeight[T comparable](rows int) FieldOption[T] {
	return func(m *Model[T]) {
		m.list.SetHeight(rows)
		m.explicitHeight = true
	}
}

// WithInitialValue selects the option whose Value equals v at
// construction time. Applied after all other options (including
// WithNone) so the index is correct.
func WithInitialValue[T comparable](v T) FieldOption[T] {
	return func(m *Model[T]) {
		m.initialMatch = func(got T) bool { return got == v }
	}
}

// WithInitialFn selects the first option whose Value satisfies match at
// construction time. Use when T has no useful `==` (e.g. pointer types
// where equality should be by pointed-to value). Applied after all
// other options.
func WithInitialFn[T comparable](match func(T) bool) FieldOption[T] {
	return func(m *Model[T]) { m.initialMatch = match }
}

// WithNone prepends a synthetic option that, when selected, yields the
// zero value of T (for pointer-typed T, that is nil). Intended for
// "(none)"-style entries on optional pointer-typed selects. The synthetic
// option is re-prepended on every SetOptions call.
func WithNone[T comparable](display string) FieldOption[T] {
	return func(m *Model[T]) {
		m.hasNone = true
		m.noneDisplay = display
	}
}

// New creates a selectfield over options. key is required and non-empty.
// The first option in the resulting list is initially selected. Options
// may be nil and supplied later via [Model.SetOptions].
func New[T comparable](k, label string, options []Option[T], opts ...FieldOption[T]) Model[T] {
	if k == "" {
		panic("selectfield: key is required")
	}

	l := list.New(nil, delegate{}, 0, 1)
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
	m.setOptions(options)
	return m
}

// SetOptions replaces the field's options after construction, enabling
// option sets that are loaded or computed asynchronously. Any WithNone
// entry is re-prepended, the list height is recomputed (unless WithHeight
// pinned an explicit height), and the WithInitial* selection is re-applied
// against the new options. Returns the updated field.
func (m Model[T]) SetOptions(options []Option[T]) Model[T] {
	m.setOptions(options)
	return m
}

// setOptions rebuilds the underlying list items from options, applying the
// replayed WithNone/WithHeight/WithInitial* configuration. It mutates the
// receiver in place; callers using the value-type Field contract should go
// through New or SetOptions.
func (m *Model[T]) setOptions(options []Option[T]) {
	items := make([]list.Item, 0, len(options)+1)
	if m.hasNone {
		items = append(items, item[T]{opt: Option[T]{Display: m.noneDisplay}}) // Value is zero
	}
	for _, o := range options {
		items = append(items, item[T]{opt: o})
	}
	_ = m.list.SetItems(items)

	if !m.explicitHeight {
		m.list.SetHeight(defaultHeight(len(options)))
	}

	if m.initialMatch != nil {
		for i, it := range m.list.Items() {
			if got, ok := it.(item[T]); ok && m.initialMatch(got.opt.Value) {
				m.list.Select(i)
				break
			}
		}
	}
}

// defaultHeight derives the auto-sized list height from the number of
// (non-synthetic) options: the option count, capped at 8 and floored at 1.
func defaultHeight(options int) int {
	const heightCap = 8
	switch {
	case options > heightCap:
		return heightCap
	case options < 1:
		return 1
	default:
		return options
	}
}

// form.Field interface --------------------------------------------------------

func (m Model[T]) Key() string { return m.key }

// Focused reports whether the field has focus. The list widget does not
// have an internal focus flag, so the field tracks it via a sentinel: a
// selectfield is "focused" when its list has any items selected by index.
// In practice we track focus through a small companion flag.
func (m Model[T]) Focused() bool { return m.focused }

func (m Model[T]) Visible(v form.Values) bool {
	if m.hideWhenEmpty && m.realOptionCount() == 0 {
		return false
	}
	if m.visible == nil {
		return true
	}
	return m.visible(v)
}

// realOptionCount is the number of selectable options excluding any
// WithNone synthetic entry.
func (m Model[T]) realOptionCount() int {
	n := len(m.list.Items())
	if m.hasNone {
		n--
	}
	return n
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

// Keys claims up/down (so the list moves the selection rather than the
// form advancing fields) and the filter trigger, and advertises both. The
// list's broader internal keymap (q, ?, page nav) is intentionally not
// surfaced — it would shadow app/overlay bindings — but those keys still
// reach the list via Update when pressed while focused.
func (m Model[T]) Keys() []keymap.Group {
	g := keymap.Group{
		{Binding: moveKey, Show: []string{"up", "down"}, Vis: keymap.Short},
		{Binding: filterKey, Vis: keymap.Short},
	}
	if m.list.FilterState() == list.Filtering {
		// While typing a filter, claim Enter so the form forwards it to the
		// list (which accepts the filter) instead of treating it as
		// next-field. When not filtering, Enter is left unclaimed so the
		// form's last-field rule can submit a terminal selectfield.
		g = append(g, keymap.Binding{Binding: submitKey, Vis: keymap.Short})
	}
	return []keymap.Group{g}
}
