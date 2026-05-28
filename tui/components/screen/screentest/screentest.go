package screentest

import (
	"iter"

	tea "charm.land/bubbletea/v2"
	"github.com/qualidafial/gtd-tui/tui/components/screen"
	"github.com/qualidafial/gtd-tui/tui/tuitest"
)

func Init(s screen.Screen) screen.Screen {
	for m := range PumpInit(s) {
		s = m
	}
	return s
}

func PumpInit(s screen.Screen) iter.Seq2[screen.Screen, tea.Msg] {
	return tuitest.PumpInit(s)
}

func Send(s screen.Screen, msg tea.Msg) screen.Screen {
	for m := range PumpSend(s, msg) {
		s = m
	}
	return s
}

func PumpSend(s screen.Screen, msg tea.Msg) iter.Seq2[screen.Screen, tea.Msg] {
	return tuitest.PumpSend(s, msg)
}
