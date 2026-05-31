package tuitest

import (
	"fmt"
	"iter"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
)

const (
	// cmdRunTimeout is how long the pump waits for a tea.Cmd to produce its
	// message AFTER the running goroutine has signaled it is about to call
	// cmd(). Synchronous app cmds (database writes against :memory:, dismiss
	// /push messages, save callbacks) finish in microseconds; cursor blink
	// (530ms), tickers, and other periodic cmds sleep for hundreds of ms and
	// would otherwise keep the pump alive indefinitely, since the bubbles
	// cursor re-emits a Blink cmd in response to every BlinkMsg.
	//
	// Goroutine-startup jitter (GC pauses, parallel-test scheduling) is moved
	// outside this window by the ready-channel handshake in runSync, so the
	// budget only measures the cmd's actual run time and can stay tight.
	cmdRunTimeout = 10 * time.Millisecond

	// maxTotalCmdLatency is how long much total latency from all commands a single
	// call to pump() will allow before failing the test.
	maxTotalCmdLatency = 2 * time.Second
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
func Init[M Initter[M]](t *testing.T, m M) M {
	t.Helper()
	for s := range PumpInit(t, m) {
		m = s
	}
	return m
}

// PumpInit runs Init and pumps all resulting commands through Update until the
// model settles. Each (model, msg) pair is yielded so callers can inspect
// intermediate states or break early. The final yield is (model, nil) to
// deliver the settled model.
func PumpInit[M Initter[M]](t *testing.T, m M) iter.Seq2[M, tea.Msg] {
	t.Helper()
	return func(yield func(M, tea.Msg) bool) {
		t.Helper()
		cmd := m.Init()
		if cmd == nil {
			yield(m, nil)
			return
		}
		pump(t, m, cmd, yield)
	}
}

func Send[M Updater[M]](t *testing.T, m M, msg tea.Msg) M {
	t.Helper()
	for s := range PumpSend(t, m, msg) {
		m = s
	}
	return m
}

// PumpSend delivers msg to m.Update and pumps all resulting commands. Each
// (model, msg) pair produced during the pump is yielded. The final yield
// is (model, nil) to deliver the settled model.
func PumpSend[M Updater[M]](t *testing.T, m M, msg tea.Msg) iter.Seq2[M, tea.Msg] {
	t.Helper()
	return func(yield func(M, tea.Msg) bool) {
		t.Helper()
		cmd := func() tea.Msg { return msg }
		pump(t, m, cmd, yield)
	}
}

// TypeText sends each rune in text to m as a tea.KeyPressMsg with Code and
// Text both set to that rune, mirroring how a terminal delivers typed input.
// All resulting commands are pumped through Update before the next rune is
// sent so the model settles between keystrokes the same way it would for a
// live user.
func TypeText[M Updater[M]](t *testing.T, m M, text string) M {
	t.Helper()
	for _, r := range text {
		m = Send(t, m, tea.KeyPressMsg{Code: r, Text: string(r)})
	}
	return m
}

func pump[M Updater[M]](t *testing.T, m M, cmd tea.Cmd, yield func(M, tea.Msg) bool) {
	t.Helper()
	var cmds []tea.Cmd
	cmds = append(cmds, cmd)
	var i int

	var totalCmdLatency time.Duration
	msgCounts := map[string]int{}
	msgLatencies := map[string]time.Duration{}

	for len(cmds) > 0 {
		i++

		cmd = cmds[0]
		cmds = cmds[1:]

		before := time.Now()
		msg, ok := runSync(cmd)

		latency := time.Since(before)
		totalCmdLatency += latency
		msgType := fmt.Sprintf("%T", msg)
		msgCounts[msgType]++
		msgLatencies[msgType] = latency

		if !ok {
			// Timer-driven cmd (cursor.Blink, ticker, etc.); skip without
			// blocking and without chasing its periodic re-emission.
			continue
		}

		if totalCmdLatency > maxTotalCmdLatency {
			var b strings.Builder
			_, _ = b.WriteString("tuitest: timeout waiting for model updates to settle. message histogram:")
			for msgType = range msgCounts {
				count := msgCounts[msgType]
				latency = msgLatencies[msgType]
				avgLatency := latency / time.Duration(count)
				_, _ = fmt.Fprintf(&b, "\n\t%s: %d (total %s, avg %s)", msgType, count, latency, avgLatency)
			}
			t.Fatal(b.String())

		}

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

// runSync executes cmd and returns its message, or (nil, false) if the cmd
// does not finish within cmdRunTimeout. A ready channel synchronizes with the
// running goroutine so the timeout window only measures cmd's own work, not
// goroutine-startup jitter — without this, a freshly-spawned goroutine that
// gets descheduled (GC, parallel-test contention) could push a microsecond
// sync cmd past the timeout and be falsely dropped as a "tick" cmd. The
// abandoned goroutine for a real slow cmd writes its eventual result into
// the buffered channel and exits cleanly.
func runSync(cmd tea.Cmd) (tea.Msg, bool) {
	if cmd == nil {
		return nil, true
	}
	ready := make(chan struct{})
	ch := make(chan tea.Msg, 1)
	go func() {
		close(ready)
		ch <- cmd()
	}()
	<-ready
	select {
	case msg := <-ch:
		return msg, true
	case <-time.After(cmdRunTimeout):
		return nil, false
	}
}
