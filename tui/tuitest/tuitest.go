package tuitest

import (
	"iter"

	tea "charm.land/bubbletea/v2"
)

type Updater[M Updater[M]] interface {
	Update(msg tea.Msg) (M, tea.Cmd)
}

type Initter[M Updater[M]] interface {
	Updater[M]

	Init() tea.Cmd
}

// Init runs Init and pumps all resulting commands through Update until the
// model settles, returning the final model.
func Init[M Initter[M]](m M) M {
	for s := range PumpInit(m) {
		m = s
	}
	return m
}

// PumpInit runs Init and pumps all resulting commands through Update until the
// model settles. Each (model, msg) pair is yielded so callers can inspect
// intermediate states or break early. The final yield is (model, nil) to
// deliver the settled model.
func PumpInit[M Initter[M]](m M) iter.Seq2[M, tea.Msg] {
	return func(yield func(M, tea.Msg) bool) {
		cmd := m.Init()
		if cmd == nil {
			yield(m, nil)
			return
		}
		pump(m, cmd, yield)
	}
}

func Send[M Updater[M]](m M, msg tea.Msg) M {
	for s := range PumpSend(m, msg) {
		m = s
	}
	return m
}

// PumpSend delivers msg to m.Update and pumps all resulting commands. Each
// (model, msg) pair produced during the pump is yielded. The final yield
// is (model, nil) to deliver the settled model.
func PumpSend[M Updater[M]](m M, msg tea.Msg) iter.Seq2[M, tea.Msg] {
	return func(yield func(M, tea.Msg) bool) {
		cmd := func() tea.Msg { return msg }
		pump(m, cmd, yield)
	}
}

func pump[M Updater[M]](m M, cmd tea.Cmd, yield func(M, tea.Msg) bool) {
	var cmds []tea.Cmd
	cmds = append(cmds, cmd)
	for i := 0; len(cmds) > 0 && i < 100; i++ {
		cmd = cmds[0]
		cmds = cmds[1:]
		msg := cmd()
		if msg == nil {
			continue
		}
		if batch, ok := msg.(tea.BatchMsg); ok {
			cmds = append(cmds, batch...)
			continue
		}
		m, cmd = m.Update(msg)
		if !yield(m, msg) {
			return
		}
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	yield(m, nil)
}
