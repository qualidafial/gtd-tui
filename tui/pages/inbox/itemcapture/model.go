package itemcapture

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/qualidafial/gtd-tui/tui/cmds"

	"github.com/qualidafial/gtd-tui"
	"github.com/qualidafial/gtd-tui/tui/components/form"
	"github.com/qualidafial/gtd-tui/tui/components/form/inputfield"
	"github.com/qualidafial/gtd-tui/tui/components/form/savefield"
	"github.com/qualidafial/gtd-tui/tui/components/form/textfield"
	"github.com/qualidafial/gtd-tui/tui/components/screen"
	"github.com/qualidafial/gtd-tui/tui/internal/keymap"
)

var keyBack = key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back"))

// Model is the new-inbox-item capture overlay. Items are write-once: this
// overlay is the only way to set Title and Description on an Item — all
// later refinement happens via the clarify wizard.
type Model struct {
	svc    gtd.InboxService
	form   form.Model
	saving bool
	err    error
}

func New(svc gtd.InboxService) Model {
	title := inputfield.New("title", "Title",
		inputfield.WithValidator(func(s string) error {
			if len(s) == 0 {
				return errors.New("title is required")
			}
			return nil
		}),
	)
	desc := textfield.New("description", "Description")
	save := savefield.New("save")

	return Model{
		svc:  svc,
		form: form.New(title, desc, save),
	}
}

func (m Model) Init() tea.Cmd { return m.form.Init() }

func (m Model) Update(msg tea.Msg) (screen.Screen, tea.Cmd) {
	// Save-error standoff: an earlier save failed; require Esc to clear
	// the error before the form can be re-driven.
	if m.err != nil {
		if kp, ok := msg.(tea.KeyPressMsg); ok && key.Matches(kp, keyBack) {
			m.err = nil
		}
		return m, nil
	}

	// Save in flight: only the saved-msg can move us forward.
	if m.saving {
		if savedMsg, ok := msg.(itemSavedMsg); ok {
			return m.handleSaved(savedMsg)
		}
		return m, nil
	}

	// Esc cancels the overlay before any save is in flight.
	if kp, ok := msg.(tea.KeyPressMsg); ok && key.Matches(kp, keyBack) {
		return screen.Dismiss()
	}

	switch msg := msg.(type) {
	case form.SubmittedMsg:
		m.saving = true
		return m, m.saveCmd()
	case itemSavedMsg:
		return m.handleSaved(msg)
	}

	var cmd tea.Cmd
	m.form, cmd = m.form.Update(msg)
	return m, cmd
}

func (m Model) handleSaved(msg itemSavedMsg) (screen.Screen, tea.Cmd) {
	if msg.err != nil {
		m.err = msg.err
		m.saving = false
		err := msg.err
		return m, cmds.Emit(fmt.Errorf("save failed: %w", err))
	}
	return screen.Dismiss()
}

func (m Model) saveCmd() tea.Cmd {
	values := m.form.FieldValues()
	title, _ := values["title"].(string)
	desc, _ := values["description"].(string)
	item := gtd.Item{Title: title, Description: desc}
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

func (m Model) CapturingInput() bool { return m.err == nil && !m.saving }

// Keys aggregates the form's resolved bindings and appends this screen's
// own esc binding as a trailing group; Resolve subtracts the overlay's
// duplicate esc.
func (m Model) Keys() []keymap.Group {
	return append(m.form.Keys(), keymap.Group{{Binding: keyBack, Vis: keymap.Short}})
}

type itemSavedMsg struct {
	item gtd.Item
	err  error
}
