package screen

import (
	tea "charm.land/bubbletea/v2"

	"github.com/qualidafial/gtd-tui/tui/internal/keymap"
)

// Screen represents a full-screen view in the application.
type Screen interface {
	Init() tea.Cmd
	Update(msg tea.Msg) (Screen, tea.Cmd)
	View() string

	// keymap.Map: Chords() returns the screen's keybindings already
	// aggregated (a composite concatenates its focused child's Chords ahead
	// of its own), the single source for both routing and resolved help.
	keymap.Map
}

// InputCapturer is an optional Screen capability. When a screen reports that it
// is capturing text input (e.g. a focused query bar), the app suppresses its
// global keybindings (tab, help toggle) so the keystrokes reach the screen.
type InputCapturer interface {
	CapturingInput() bool
}

// CapturingInput reports whether s is currently capturing text input.
func CapturingInput(s Screen) bool {
	c, ok := s.(InputCapturer)
	return ok && c.CapturingInput()
}

// Popper is satisfied by screens that can reveal a parent screen underneath.
type Popper interface {
	Pop() Screen
}

// PushMsg signals that a child screen should be pushed on top of the current view.
type PushMsg struct {
	Screen Screen
}

func Push(child Screen) tea.Cmd {
	return func() tea.Msg {
		return PushMsg{Screen: child}
	}
}

// DismissMsg signals that the current overlay screen should be dismissed.
type DismissMsg struct{}

// dismissed is a no-op Screen sentinel that overlays morph into when they
// emit Dismiss. It absorbs every subsequent message so the dismissing
// model's terminal-state branches don't re-fire on stray msgs.
type dismissed struct{}

func (dismissed) Init() tea.Cmd                    { return nil }
func (dismissed) Update(tea.Msg) (Screen, tea.Cmd) { return dismissed{}, nil }
func (dismissed) View() string                     { return "" }
func (dismissed) Chords() []keymap.Group           { return nil }

// Dismiss returns the no-op dismissed sentinel plus the cmd that signals the
// parent to pop this overlay. Optional extras are batched with the dispatch.
//
// Use directly: `return screen.Dismiss()` or `return screen.Dismiss(cmd)`.
//
// For tea.Sequence composition (where ordering matters), or for code that
// must remain as the active Screen after emitting the dispatch — e.g. the
// overlay wrapper itself, which has to stay so the parent can call Pop() —
// use DismissCmd to get just the cmd.
func Dismiss(extras ...tea.Cmd) (Screen, tea.Cmd) {
	if len(extras) == 0 {
		return dismissed{}, DismissCmd()
	}
	return dismissed{}, tea.Batch(append([]tea.Cmd{DismissCmd()}, extras...)...)
}

// DismissCmd is the freestanding dispatch cmd for callers that need finer
// composition (Sequence) or that cannot morph into the dismissed sentinel.
func DismissCmd() tea.Cmd {
	return func() tea.Msg {
		return DismissMsg{}
	}
}

// Replace returns next as the new model with a window-size request and its
// Init cmd batched. Use when one overlay's Update wants to swap itself for a
// different overlay in a single op, without going through pop-then-push
// (which would briefly expose the parent screen and add a stack level).
//
// The window-size request mirrors the Push handler in tui.app so the new
// inner gets sized exactly as if it had been pushed.
func Replace(next Screen) (Screen, tea.Cmd) {
	return next, tea.Batch(tea.RequestWindowSize, next.Init())
}

// InitMsg signals that the active screen should re-initialize.
type InitMsg struct{}

func InitCmd() tea.Cmd {
	return func() tea.Msg {
		return InitMsg{}
	}
}
