package inbox

import (
	"context"
	"fmt"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/qualidafial/gtd-tui"
	"github.com/qualidafial/gtd-tui/tui/components/screen"
	"github.com/qualidafial/gtd-tui/tui/internal/keymap"
	"github.com/qualidafial/gtd-tui/tui/pages/inbox/clarify"
	"github.com/qualidafial/gtd-tui/tui/pages/inbox/itemcapture"
)

var emptyStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Italic(true)

// Model renders the inbox: unclarified, non-discarded items in FIFO order.
// `+`/`insert` opens the capture overlay; `enter` on a selected item opens
// the clarify wizard.
type Model struct {
	svc        gtd.InboxService
	taskSvc    gtd.TaskService
	projectSvc gtd.ProjectService
	items      []gtd.Item
	list    list.Model
	KeyMap  KeyMap
	width   int
	height  int
}

type itemsLoadedMsg struct{ items []gtd.Item }

func New(svc gtd.InboxService, taskSvc gtd.TaskService, projectSvc gtd.ProjectService) Model {
	keys := DefaultKeyMap()
	l := list.New(nil, newDelegate(keys), 0, 0)
	l.SetStatusBarItemName("item", "items")
	l.SetShowTitle(false)
	l.SetShowHelp(false)
	l.DisableQuitKeybindings()
	l.KeyMap.Filter.SetEnabled(false)

	m := Model{svc: svc, taskSvc: taskSvc, projectSvc: projectSvc, list: l, KeyMap: keys}
	m.updateKeybindings()
	return m
}

func (m Model) Init() tea.Cmd { return m.loadCmd() }

func (m Model) loadCmd() tea.Cmd {
	svc := m.svc
	return func() tea.Msg {
		items, err := svc.List(context.Background())
		if err != nil {
			return fmt.Errorf("load inbox: %w", err)
		}
		return itemsLoadedMsg{items: items}
	}
}

func (m Model) Update(msg tea.Msg) (screen.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width, msg.Height)
		return m, nil
	case itemsLoadedMsg:
		m.items = msg.items
		items := make([]list.Item, len(msg.items))
		for i, it := range msg.items {
			items[i] = Item{item: it}
		}
		cmd := m.list.SetItems(items)
		m.updateKeybindings()
		return m, cmd
	case screen.InitMsg:
		// Returning from a child (capture overlay or, later, clarify wizard) —
		// reload the inbox so newly-captured items appear and clarified items
		// drop off.
		return m, m.loadCmd()
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.KeyMap.New):
			return m, screen.Push(itemcapture.New(m.svc))
		case key.Matches(msg, m.KeyMap.Clarify):
			it, ok := m.list.SelectedItem().(Item)
			if !ok {
				return m, nil
			}
			return m, screen.Push(clarify.New(it.item, m.svc, m.taskSvc, m.projectSvc))
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	if len(m.items) == 0 {
		msg := emptyStyle.Render("Inbox is empty.")
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, msg)
	}
	var b strings.Builder
	b.WriteString(m.list.View())
	return b.String()
}

func (m *Model) updateKeybindings() {
	selected := len(m.items) > 0
	m.KeyMap.Clarify.SetEnabled(selected)
}

// Keys contributes the inbox action group plus a list-navigation group.
// CursorUp/Down show in both bars; GoToStart/End in full help only.
func (m Model) Keys() []keymap.Group {
	return append(m.KeyMap.Keys(), keymap.Group{
		{Binding: m.list.KeyMap.CursorUp, Vis: keymap.Short},
		{Binding: m.list.KeyMap.CursorDown, Vis: keymap.Short},
		{Binding: m.list.KeyMap.GoToStart, Vis: keymap.Full},
		{Binding: m.list.KeyMap.GoToEnd, Vis: keymap.Full},
	})
}
