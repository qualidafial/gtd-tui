package screen

import (
	"testing"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"github.com/qualidafial/gtd-tui/tui/internal/keymap"
)

var escBinding = key.NewBinding(key.WithKeys("esc"))

// innerEscMsg marks that esc reached the inner screen rather than being
// intercepted by the overlay's fallback dismiss.
type innerEscMsg struct{}

// stubInner optionally claims esc in its Keys(). When it claims esc and is
// sent one it returns innerEscMsg, so a test can tell forwarding from the
// overlay's generic dismiss.
type stubInner struct{ bindsEsc bool }

func (s stubInner) Init() tea.Cmd { return nil }

func (s stubInner) Update(msg tea.Msg) (Screen, tea.Cmd) {
	if kp, ok := msg.(tea.KeyPressMsg); ok && key.Matches(kp, escBinding) {
		return s, func() tea.Msg { return innerEscMsg{} }
	}
	return s, nil
}

func (s stubInner) View() string { return "" }

func (s stubInner) Keys() []keymap.Group {
	if !s.bindsEsc {
		return nil
	}
	return []keymap.Group{{{Binding: escBinding, Vis: keymap.Short}}}
}

// When the inner subtree claims esc, the overlay forwards it instead of
// dismissing, so the inner can handle it (e.g. cancel a filter, record a
// modal's outcome).
func TestOverlay_EscForwardedWhenInnerClaimsIt(t *testing.T) {
	o := Overlay(stubInner{}, stubInner{bindsEsc: true})
	_, cmd := o.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	if cmd == nil {
		t.Fatal("expected the inner's cmd, got nil")
	}
	if _, ok := cmd().(innerEscMsg); !ok {
		t.Fatalf("expected esc forwarded to inner (innerEscMsg), got %T", cmd())
	}
}

// When the inner subtree does not claim esc, the overlay falls back to its
// generic dismiss.
func TestOverlay_EscDismissesWhenInnerIgnoresIt(t *testing.T) {
	o := Overlay(stubInner{}, stubInner{bindsEsc: false})
	_, cmd := o.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	if cmd == nil {
		t.Fatal("expected a dismiss cmd, got nil")
	}
	if _, ok := cmd().(DismissMsg); !ok {
		t.Fatalf("expected overlay fallback dismiss (DismissMsg), got %T", cmd())
	}
}
