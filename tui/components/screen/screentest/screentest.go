package screentest

import (
	"iter"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/qualidafial/gtd-tui/tui/components/screen"
	"github.com/qualidafial/gtd-tui/tui/tuitest"
)

func Init(t *testing.T, s screen.Screen) screen.Screen {
	t.Helper()
	for m := range PumpInit(t, s) {
		s = m
	}
	return s
}

func PumpInit(t *testing.T, s screen.Screen) iter.Seq2[screen.Screen, tea.Msg] {
	t.Helper()
	return tuitest.PumpInit(t, s)
}

func Send(t *testing.T, s screen.Screen, msg tea.Msg) screen.Screen {
	t.Helper()
	for m := range PumpSend(t, s, msg) {
		s = m
	}
	return s
}

func PumpSend(t *testing.T, s screen.Screen, msg tea.Msg) iter.Seq2[screen.Screen, tea.Msg] {
	t.Helper()
	return tuitest.PumpSend(t, s, msg)
}

// TypeText is the Screen-specialized form of tuitest.TypeText.
func TypeText(t *testing.T, s screen.Screen, text string) screen.Screen {
	t.Helper()
	return tuitest.TypeText(t, s, text)
}

// RunUntilDismiss pumps msg through s and returns (final-screen, true) the
// first time a screen.DismissMsg appears in the resulting message stream, or
// (final-screen, false) if the pump settles without one. This is the standard
// "drive the final keystroke and assert the overlay closes" idiom for
// overlay-screen tests.
func RunUntilDismiss(t *testing.T, s screen.Screen, msg tea.Msg) (screen.Screen, bool) {
	t.Helper()
	var dismissed bool
	for st, m := range PumpSend(t, s, msg) {
		s = st
		if _, ok := m.(screen.DismissMsg); ok {
			dismissed = true
			break
		}
	}
	return s, dismissed
}
