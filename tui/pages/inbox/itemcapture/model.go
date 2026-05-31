package itemcapture

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"

	"github.com/qualidafial/gtd-tui"
	"github.com/qualidafial/gtd-tui/tui/components/screen"
)

var keyBack = key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back"))

// Model is the new-inbox-item capture overlay. Items are write-once: this
// overlay is the only way to set Title and Description on an Item — all later
// refinement happens via the clarify wizard, which inherits both fields and
// lets the user override when committing to a destination entity.
type Model struct {
	item   *gtd.Item
	svc    gtd.InboxService
	form   *huh.Form
	saving bool
	err    error
}

func New(svc gtd.InboxService) Model {
	item := &gtd.Item{}
	m := Model{item: item, svc: svc}

	fields := []huh.Field{
		huh.NewInput().
			Title("Title").
			Value(&item.Title).
			Validate(func(s string) error {
				if len(s) == 0 {
					return errors.New("title is required")
				}
				return nil
			}),
		huh.NewText().
			Title("Description").
			Value(&item.Description),
	}

	keymap := huh.NewDefaultKeyMap()
	keymap.Quit = keyBack

	m.form = huh.NewForm(huh.NewGroup(fields...)).
		WithShowErrors(true).
		WithShowHelp(false).
		WithKeyMap(keymap)
	return m
}

func (m Model) Init() tea.Cmd { return m.form.Init() }

func (m Model) Update(msg tea.Msg) (screen.Screen, tea.Cmd) {
	// Save-error standoff: the form has completed but the save failed, so the
	// user must press Esc to dismiss the error before the form can be re-driven.
	// Swallow everything else to avoid re-firing save or any form transition.
	if m.err != nil {
		if kp, ok := msg.(tea.KeyPressMsg); ok && key.Matches(kp, keyBack) {
			m.err = nil
			m.form.State = huh.StateNormal
		}
		return m, nil
	}
	if savedMsg, ok := msg.(itemSavedMsg); ok {
		if savedMsg.err != nil {
			m.err = savedMsg.err
			m.saving = false
			err := savedMsg.err
			return m, func() tea.Msg { return fmt.Errorf("save failed: %w", err) }
		}
		return screen.Dismiss()
	}

	if m.saving {
		return m, nil
	}

	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
	}

	switch m.form.State {
	case huh.StateAborted:
		return screen.Dismiss(cmd)
	case huh.StateCompleted:
		m.saving = true
		return m, tea.Batch(cmd, m.saveCmd())
	}
	return m, cmd
}

func (m Model) saveCmd() tea.Cmd {
	item := *m.item
	svc := m.svc
	return func() tea.Msg {
		saved, err := svc.Create(context.Background(), item)
		if err != nil {
			slog.Error("saving item: " + err.Error())
		}
		return itemSavedMsg{item: saved, err: err}
	}
}

func (m Model) View() string { return m.form.View() }

func (m Model) CapturingInput() bool { return m.form.State == huh.StateNormal }

func (m Model) ShortHelp() []key.Binding  { return m.form.KeyBinds() }
func (m Model) FullHelp() [][]key.Binding { return [][]key.Binding{m.form.KeyBinds()} }

type itemSavedMsg struct {
	item gtd.Item
	err  error
}
