package taskview

import (
	"context"
	"fmt"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/qualidafial/gtd-tui"
	"github.com/qualidafial/gtd-tui/tui/components/screen"
	"github.com/qualidafial/gtd-tui/tui/internal/keymap"
	"github.com/qualidafial/gtd-tui/tui/pages/tasks/taskedit"
	"github.com/qualidafial/gtd-tui/tui/pages/tasks/tasklist"
	"github.com/qualidafial/gtd-tui/tui/pages/tasks/taskstatus"
)

var (
	titleStyle = lipgloss.NewStyle().Bold(true)
	labelStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Width(12)
	valueStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
)

// ProjectViewFactory builds the linked project's view screen, used by the
// go-to-project action. It is injected (rather than importing projectview
// directly) so the task package does not depend on the projects package,
// avoiding an import cycle. When nil, go-to-project is disabled.
type ProjectViewFactory func(project gtd.Project) screen.Screen

type Model struct {
	task          gtd.Task
	taskSvc       gtd.TaskService
	projectNameFn tasklist.ProjectNameFunc
	pickerFn      tasklist.PickerFactory
	convertFn     tasklist.ConvertFactory
	projectViewFn ProjectViewFactory
	KeyMap        KeyMap
	width         int
	height        int
}

func New(
	task gtd.Task,
	taskSvc gtd.TaskService,
	projectNameFn tasklist.ProjectNameFunc,
	pickerFn tasklist.PickerFactory,
	convertFn tasklist.ConvertFactory,
	projectViewFn ProjectViewFactory,
) Model {
	m := Model{
		task:          task,
		taskSvc:       taskSvc,
		projectNameFn: projectNameFn,
		pickerFn:      pickerFn,
		convertFn:     convertFn,
		projectViewFn: projectViewFn,
		KeyMap:        DefaultKeyMap(),
	}
	m.updateKeybindings()
	return m
}

func (m Model) Init() tea.Cmd {
	return m.reloadCmd()
}

// reloadCmd re-fetches the task by ID so the header reflects changes made by an
// action overlay. The application re-runs Init on the revealed screen after an
// overlay dismisses, so this refreshes the view with no per-action plumbing.
func (m Model) reloadCmd() tea.Cmd {
	if m.taskSvc == nil || m.task.ID == 0 {
		return nil
	}
	svc := m.taskSvc
	id := m.task.ID
	return func() tea.Msg {
		t, err := svc.GetTask(context.Background(), id)
		if err != nil {
			return fmt.Errorf("reload task: %w", err)
		}
		return taskReloadedMsg{task: t}
	}
}

type taskReloadedMsg struct {
	task gtd.Task
}

func (m Model) Update(msg tea.Msg) (screen.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case taskReloadedMsg:
		m.task = msg.task
		m.updateKeybindings()
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, m.KeyMap.Edit):
			return m, screen.Push(taskedit.New(m.task, m.taskSvc, m.resolveProjectName(), nil))

		case key.Matches(msg, m.KeyMap.ToggleComplete):
			transition := taskstatus.Complete
			if m.task.Status != gtd.TaskStatusOpen {
				transition = taskstatus.Reopen
			}
			return m, screen.Push(taskstatus.New(m.task, m.taskSvc, transition))

		case key.Matches(msg, m.KeyMap.Drop):
			return m, screen.Push(taskstatus.New(m.task, m.taskSvc, taskstatus.Drop))

		case key.Matches(msg, m.KeyMap.AssignToProject):
			if m.pickerFn != nil {
				return m, screen.Push(m.pickerFn(m.task))
			}

		case key.Matches(msg, m.KeyMap.ConvertToProject):
			if m.convertFn != nil {
				return m, screen.Push(m.convertFn(m.task))
			}

		case key.Matches(msg, m.KeyMap.GoToProject):
			if m.task.ProjectID != nil && m.projectViewFn != nil {
				// projectview reloads the full project from its ID on Init.
				return screen.Replace(m.projectViewFn(gtd.Project{ID: *m.task.ProjectID}))
			}
		}
	}

	return m, nil
}

func (m Model) resolveProjectName() string {
	if m.task.ProjectID == nil || m.projectNameFn == nil {
		return ""
	}
	return m.projectNameFn(*m.task.ProjectID)
}

// updateKeybindings reconciles per-task action availability: drop only for an
// open task, convert-to-project only for a standalone task, go-to-project only
// for a task with a project, and the toggle label tracks the current status.
func (m *Model) updateKeybindings() {
	open := m.task.Status == gtd.TaskStatusOpen

	label := "complete"
	if !open {
		label = "reopen"
	}
	m.KeyMap.ToggleComplete.SetHelp("space", label)
	m.KeyMap.Drop.SetEnabled(open)
	m.KeyMap.AssignToProject.SetEnabled(m.pickerFn != nil)
	m.KeyMap.ConvertToProject.SetEnabled(gtd.IsStandalone(m.task) && m.convertFn != nil)
	m.KeyMap.GoToProject.SetEnabled(m.task.ProjectID != nil && m.projectViewFn != nil)
}

func (m Model) View() string {
	return m.renderHeader()
}

func (m Model) renderHeader() string {
	var lines []string

	lines = append(lines, titleStyle.Render(m.task.Title))
	lines = append(lines, m.field("Status:", statusLabel(m.task.Status)))

	if name := m.resolveProjectName(); name != "" {
		lines = append(lines, m.field("Project:", "+"+name))
	}
	if m.task.Assignee != nil && *m.task.Assignee != "" {
		lines = append(lines, m.field("Assignee:", *m.task.Assignee))
	}
	if m.task.Due != nil {
		lines = append(lines, m.field("Due:", m.task.Due.Local().Format(time.DateOnly)))
	}
	if m.task.Description != "" {
		lines = append(lines, m.field("Description:", m.task.Description))
	}

	return strings.Join(lines, "\n")
}

func (m Model) field(label, value string) string {
	return labelStyle.Render(label) + " " + valueStyle.Render(value)
}

func statusLabel(s gtd.TaskStatus) string {
	switch s {
	case gtd.TaskStatusOpen:
		return "Open"
	case gtd.TaskStatusDone:
		return "Done"
	case gtd.TaskStatusDropped:
		return "Dropped"
	default:
		return string(s)
	}
}

func (m Model) Keys() []keymap.Group {
	return m.KeyMap.Keys()
}
