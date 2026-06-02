// Package form provides a small, synchronous form toolkit for TUI overlays.
//
// A [Model] owns an ordered list of [Field]s with a single focus index.
// Navigation, submit, and cancel keys live on the form's [KeyMap]; fields
// only consume keys that mean something locally (typing, cursor movement,
// selection toggles). Submit walks visible fields' validators in order and
// reports the first failure synchronously — no message-loop round trips.
//
// A field's [Field.Visible] predicate receives a [Values] snapshot that
// contains only the values of visible fields *preceding* it in declaration
// order. A field can therefore reveal/hide later fields, but cannot see
// itself or any later sibling, and hidden fields contribute nothing to the
// snapshot any later field sees.
//
// Rendering goes through a [viewport.Model] so tall forms scroll
// automatically; the form ensures the focused field stays in view.
package form

import (
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// SubmittedMsg is emitted by [Model.Update] after a successful Submit so
// the surrounding overlay can perform its save side effect without
// inspecting Model internals.
type SubmittedMsg struct{}

// SubmitRequestMsg asks the form to invoke Submit, exactly as if the
// user had pressed the Save key. Fields emit this from their Update cmd
// to request submission — savefield uses it to wire Enter-to-submit.
// External code generally should not emit this; press the Save key
// instead.
type SubmitRequestMsg struct{}

// Values is an immutable snapshot of the visible prior fields' values
// supplied to a field's Visible predicate.
type Values interface {
	// Get returns the Value() of the visible preceding field whose Key()
	// matches key, or nil if no such field is in the snapshot. A field's
	// own Value is never present in the snapshot supplied to its own
	// Visible call; hidden fields are excluded from snapshots supplied to
	// later fields.
	Get(key string) any
}

type valuesMap map[string]any

func (v valuesMap) Get(k string) any { return v[k] }

// Field is a single form input. Implementations should be value types;
// mutating methods (Focus, Blur, Update, SetWidth) return a new Field
// value rather than mutating in place.
type Field interface {
	// Key is the field's unique identifier within a form. Required and
	// non-empty.
	Key() string
	Focus() (Field, tea.Cmd)
	Blur() Field
	Focused() bool
	// Visible reports whether the field should participate in rendering,
	// navigation, validation, submit, and help. The provided Values
	// contains only the values of visible preceding fields.
	Visible(Values) bool
	Update(tea.Msg) (Field, tea.Cmd)
	View() string
	Value() any
	// Validate runs the field's validator against its current value and
	// returns a new Field whose subsequent View reflects the result (e.g.
	// an inline error message), along with the error itself for the
	// form's decision logic. The caller-supplied validator function must
	// be pure; Field.Validate is allowed to cache the result internally —
	// the cache is what the returned Field carries.
	Validate() (Field, error)
	// SetWidth notifies the field of the column width available to it.
	// Fields that wrap or truncate should honor this; trivial fields may
	// return themselves unchanged.
	SetWidth(int) Field
	ShortHelp() []key.Binding
	FullHelp() [][]key.Binding
}

// KeyMap is the navigation/submit keymap shared by every form. Cancel
// (Esc) is intentionally not owned by the form — overlays decide what
// cancellation means (dismiss, clear error, back out a wizard step) and
// surface their own Esc binding in the overlay's help.
type KeyMap struct {
	Next key.Binding
	Prev key.Binding
	Save key.Binding
}

// DefaultKeyMap returns the gtd-tui form keymap.
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Next: key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next")),
		Prev: key.NewBinding(key.WithKeys("shift+tab"), key.WithHelp("shift+tab", "prev")),
		Save: key.NewBinding(key.WithKeys("ctrl+s"), key.WithHelp("ctrl+s", "save")),
	}
}

// Model is a form holding an ordered list of fields with a single focus.
type Model struct {
	KeyMap KeyMap

	fields   []Field
	focus    int     // -1 if no field is focused
	initCmd  tea.Cmd // applied on Init (stashed at New so Init can stay return-only)
	viewport viewport.Model
	width    int
	height   int
}

// New constructs a form. It panics if any field has an empty Key or if two
// fields share a Key — uniqueness is a construction-time invariant. The
// first visible field is focused (using the progressive snapshot rule).
func New(fields ...Field) Model {
	seen := make(map[string]struct{}, len(fields))
	for _, f := range fields {
		k := f.Key()
		if k == "" {
			panic("form: field has empty key")
		}
		if _, dup := seen[k]; dup {
			panic("form: duplicate field key: " + k)
		}
		seen[k] = struct{}{}
	}

	fs := append([]Field(nil), fields...)
	m := Model{
		KeyMap:   DefaultKeyMap(),
		fields:   fs,
		focus:    -1,
		viewport: viewport.New(),
	}

	vis := visibility(fs)
	for i, v := range vis {
		if !v {
			continue
		}
		nf, cmd := fs[i].Focus()
		fs[i] = nf
		m.focus = i
		m.initCmd = cmd
		break
	}
	m.fields = fs
	return m
}

// Init returns the focus cmd for the initially focused field batched with a
// window-size request so the form is ready to lay itself out on first View.
func (m Model) Init() tea.Cmd {
	return tea.Batch(m.initCmd, tea.RequestWindowSize)
}

// Update routes form-level keys (Save/Next/Prev) to the form's own
// handlers, threads window-size changes to every field, and otherwise
// dispatches the message to the focused field. On a successful Submit via
// the Save key, Update emits [SubmittedMsg].
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case SubmitRequestMsg:
		m, cmd := m.handleSubmit()
		return m.syncViewport(), cmd

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.SetWidth(msg.Width)
		for i, f := range m.fields {
			m.fields[i] = f.SetWidth(msg.Width)
		}
		return m.syncViewport(), nil

	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.KeyMap.Save):
			m, cmd := m.handleSubmit()
			return m.syncViewport(), cmd
		case key.Matches(msg, m.KeyMap.Next):
			m, cmd := m.Next()
			return m.syncViewport(), cmd
		case key.Matches(msg, m.KeyMap.Prev):
			m, cmd := m.Prev()
			return m.syncViewport(), cmd
		}
	}

	if m.focus < 0 {
		return m.syncViewport(), nil
	}

	nf, cmd := m.fields[m.focus].Update(msg)
	m.fields[m.focus] = nf

	m, rcmd := m.refocusIfHidden()
	return m.syncViewport(), tea.Batch(cmd, rcmd)
}

// View returns the viewport's rendered content. Width and height come from
// the most recent [tea.WindowSizeMsg]; until one arrives the viewport
// renders an empty frame.
func (m Model) View() string { return m.viewport.View() }

// FieldValues returns a map of every currently-visible field's Value
// keyed by its Key. Hidden fields are excluded under the same
// progressive-snapshot rule used everywhere else — a hidden field
// contributes nothing to what the host should see at save time. Hosts
// use this to read field values when building the side effect that
// follows SubmittedMsg.
func (m Model) FieldValues() map[string]any {
	vis := visibility(m.fields)
	out := make(map[string]any, len(m.fields))
	for i, f := range m.fields {
		if !vis[i] {
			continue
		}
		out[f.Key()] = f.Value()
	}
	return out
}

// Focused returns the currently focused Field, or nil if no field is
// focused (which is the state immediately after a successful Submit, or
// when every field is hidden).
func (m Model) Focused() Field {
	if m.focus < 0 || m.focus >= len(m.fields) {
		return nil
	}
	return m.fields[m.focus]
}

// Next runs the currently focused field's validator and, on success,
// moves focus to the next visible field. If the validator fails, focus
// stays put — the field's own View will now reflect the error — so the
// user can fix the input. It is a no-op if no later visible field exists.
func (m Model) Next() (Model, tea.Cmd) {
	if m.focus >= 0 && m.focus < len(m.fields) {
		nf, err := m.fields[m.focus].Validate()
		m.fields[m.focus] = nf
		if err != nil {
			return m, nil
		}
	}
	vis := visibility(m.fields)
	for i := m.focus + 1; i < len(m.fields); i++ {
		if vis[i] {
			return m.focusIndex(i)
		}
	}
	return m, nil
}

// Prev moves focus to the previous visible field. It is a no-op if no
// earlier visible field exists.
func (m Model) Prev() (Model, tea.Cmd) {
	vis := visibility(m.fields)
	for i := m.focus - 1; i >= 0; i-- {
		if vis[i] {
			return m.focusIndex(i)
		}
	}
	return m, nil
}

// Submit validates every visible field in declaration order. On the first
// validator that returns a non-nil error, Submit focuses that field and
// returns (model, false, cmd). If every visible field validates, it
// returns (model, true, nil). Hidden fields are never validated and never
// fail the form.
func (m Model) Submit() (Model, bool, tea.Cmd) {
	vis := visibility(m.fields)
	for i, f := range m.fields {
		if !vis[i] {
			continue
		}
		nf, err := f.Validate()
		m.fields[i] = nf
		if err != nil {
			nm, cmd := m.focusIndex(i)
			return nm, false, cmd
		}
	}
	return m, true, nil
}

// ShortHelp returns the form-level bindings composed with the focused
// (visible) field's bindings. Screens can delegate help with
// `return f.ShortHelp()`.
func (m Model) ShortHelp() []key.Binding {
	bs := []key.Binding{m.KeyMap.Next, m.KeyMap.Prev, m.KeyMap.Save}
	if f := m.Focused(); f != nil {
		bs = append(bs, f.ShortHelp()...)
	}
	return bs
}

// FullHelp returns the form-level bindings on the first row and the
// focused field's full help on subsequent rows.
func (m Model) FullHelp() [][]key.Binding {
	rows := [][]key.Binding{{m.KeyMap.Next, m.KeyMap.Prev, m.KeyMap.Save}}
	if f := m.Focused(); f != nil {
		rows = append(rows, f.FullHelp()...)
	}
	return rows
}

// handleSubmit calls Submit and batches the form-emitted SubmittedMsg when
// validation passes.
func (m Model) handleSubmit() (Model, tea.Cmd) {
	m, ok, cmd := m.Submit()
	if !ok {
		return m, cmd
	}
	return m, tea.Batch(cmd, submittedCmd)
}

func submittedCmd() tea.Msg { return SubmittedMsg{} }

// focusIndex blurs the current focus (if any) and focuses field i. If i
// already holds focus, it is a no-op.
func (m Model) focusIndex(i int) (Model, tea.Cmd) {
	if i == m.focus {
		return m, nil
	}
	if m.focus >= 0 && m.focus < len(m.fields) {
		m.fields[m.focus] = m.fields[m.focus].Blur()
	}
	nf, cmd := m.fields[i].Focus()
	m.fields[i] = nf
	m.focus = i
	return m, cmd
}

// refocusIfHidden moves focus off a now-hidden field, searching forward
// first and then backward for the next visible field. If no visible field
// exists, focus is cleared.
//
// Under the progressive-snapshot rule a focused field cannot hide itself
// via its own Update (its visibility predicate depends only on prior
// fields, which the focused field cannot mutate). This routine is
// therefore defensive — it guards against future field types that mutate
// shared state through means outside the Update path.
func (m Model) refocusIfHidden() (Model, tea.Cmd) {
	if m.focus < 0 {
		return m, nil
	}
	vis := visibility(m.fields)
	if vis[m.focus] {
		return m, nil
	}

	for i := m.focus + 1; i < len(m.fields); i++ {
		if vis[i] {
			return m.focusIndex(i)
		}
	}
	for i := m.focus - 1; i >= 0; i-- {
		if vis[i] {
			return m.focusIndex(i)
		}
	}

	m.fields[m.focus] = m.fields[m.focus].Blur()
	m.focus = -1
	return m, nil
}

// syncViewport rebuilds the viewport's content from the current visible
// fields and scrolls so that the focused field is in view. Visible
// fields are separated by a single blank line so labels breathe.
func (m Model) syncViewport() Model {
	vis := visibility(m.fields)
	parts := make([]string, 0, 2*len(m.fields))
	focusedStart, focusedEnd := -1, -1
	lineCount := 0
	first := true
	for i, f := range m.fields {
		if !vis[i] {
			continue
		}
		if !first {
			parts = append(parts, "")
			lineCount++
		}
		first = false

		view := f.View()
		h := lipgloss.Height(view)
		if i == m.focus {
			focusedStart = lineCount
			focusedEnd = lineCount + h - 1
		}
		parts = append(parts, view)
		lineCount += h
	}
	content := lipgloss.JoinVertical(lipgloss.Left, parts...)
	m.viewport.SetContent(content)

	// Size the viewport to content when content fits within the allotted
	// height; otherwise cap at the allotted height so tall forms scroll.
	// Without this, a short form would pad its viewport to the full
	// window height and push the help footer off-screen.
	contentHeight := lipgloss.Height(content)
	viewHeight := contentHeight
	if m.height > 0 && m.height < viewHeight {
		viewHeight = m.height
	}
	m.viewport.SetHeight(viewHeight)

	if focusedStart >= 0 {
		// Ensure both ends of the focused field are visible; calling end
		// first then start biases the scroll toward showing the label.
		m.viewport.EnsureVisible(focusedEnd, 0, 0)
		m.viewport.EnsureVisible(focusedStart, 0, 0)
	}
	return m
}

// visibility computes the per-field visibility mask using the progressive
// snapshot rule: field i's predicate sees a Values containing only the
// values of visible fields with index < i.
func visibility(fs []Field) []bool {
	vis := make([]bool, len(fs))
	vm := valuesMap{}
	for i, f := range fs {
		if f.Visible(vm) {
			vis[i] = true
			vm[f.Key()] = f.Value()
		}
	}
	return vis
}
