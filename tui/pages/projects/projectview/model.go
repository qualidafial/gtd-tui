package projectview

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/qualidafial/gtd-tui"
	"github.com/qualidafial/gtd-tui/service"
	"github.com/qualidafial/gtd-tui/tui/cmds"
	"github.com/qualidafial/gtd-tui/tui/components/screen"
	"github.com/qualidafial/gtd-tui/tui/internal/keymap"
	"github.com/qualidafial/gtd-tui/tui/pages/projects/projectconvert"
	"github.com/qualidafial/gtd-tui/tui/pages/projects/projectedit"
	"github.com/qualidafial/gtd-tui/tui/pages/projects/taskpicker"
	"github.com/qualidafial/gtd-tui/tui/pages/tasks/tasklist"
	"github.com/qualidafial/gtd-tui/tui/theme"
)

var (
	titleStyle  = theme.Title
	labelStyle  = theme.Label.Width(10)
	valueStyle  = theme.Value
	statusStyle = theme.Value
)

type Model struct {
	project       gtd.Project
	projectSvc    gtd.ProjectService
	taskSvc       gtd.TaskService
	pickerFn      tasklist.PickerFactory
	tasks         screen.Screen
	KeyMap        KeyMap
	taskCount     int // tasks of any status under this project; -1 until loaded
	hasCandidates bool
	width         int
	height        int
}

func New(
	project gtd.Project,
	taskSvc gtd.TaskService,
	projectSvc gtd.ProjectService,
	pickerFn tasklist.PickerFactory,
	taskViewFn tasklist.ViewFactory,
) Model {
	wrapped := service.NewProjectTaskService(taskSvc, project.ID)
	projectNameFn := func(_ int64) string { return project.Title }
	// enter pushes the task view; e edits. The in-project list is itself a
	// detail view, but enter on a task should still drill into that task.
	tasks := tasklist.New(wrapped, "", pickerFn, nil, projectNameFn, false, taskViewFn)

	m := Model{
		project:    project,
		projectSvc: projectSvc,
		taskSvc:    taskSvc,
		pickerFn:   pickerFn,
		tasks:      tasks,
		KeyMap:     DefaultKeyMap(),
		taskCount:  -1, // unknown until statsCmd loads it; keeps guards off
	}
	m.updateKeybindings()
	return m
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.reloadCmd(), m.tasks.Init(), m.statsCmd())
}

// statsCmd loads the project's task count (any status) and whether any
// standalone open task exists, so the Convert-to-Task and Link-Task action
// guards can be reconciled. It is fired on Init and after any embedded task
// list reload (which signals the project's tasks may have changed).
func (m Model) statsCmd() tea.Cmd {
	svc := m.taskSvc
	id := m.project.ID
	return func() tea.Msg {
		all, err := svc.ListTasks(context.Background(), gtd.TaskFilter{ProjectID: &id, IncludeSomedayProjects: true})
		if err != nil {
			return fmt.Errorf("load project stats: %w", err)
		}
		cand, err := svc.ListTasks(context.Background(), gtd.TaskFilter{Status: new(gtd.TaskStatusOpen)})
		if err != nil {
			return fmt.Errorf("load project stats: %w", err)
		}
		has := slices.ContainsFunc(cand, gtd.IsStandalone)
		return statsLoadedMsg{taskCount: len(all), hasCandidates: has}
	}
}

type statsLoadedMsg struct {
	taskCount     int
	hasCandidates bool
}

type linkedMsg struct{ err error }

type convertedMsg struct{ err error }

func (m Model) reloadCmd() tea.Cmd {
	if m.projectSvc == nil || m.project.ID == 0 {
		return nil
	}
	svc := m.projectSvc
	id := m.project.ID
	return func() tea.Msg {
		p, err := svc.GetProject(context.Background(), id)
		if err != nil {
			return fmt.Errorf("reload project: %w", err)
		}
		return projectReloadedMsg{project: p}
	}
}

type projectReloadedMsg struct {
	project gtd.Project
}

func (m Model) Update(msg tea.Msg) (screen.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case projectReloadedMsg:
		m.project = msg.project
		m.updateKeybindings()
		return m, tea.RequestWindowSize

	case statsLoadedMsg:
		m.taskCount = msg.taskCount
		m.hasCandidates = msg.hasCandidates
		m.updateKeybindings()
		return m, nil

	case tasklist.TasksLoadedMsg:
		// The embedded list reloaded; its contents may have changed, so refresh
		// the action guards. Still forward the message to the list.
		var cmd tea.Cmd
		m.tasks, cmd = m.tasks.Update(msg)
		return m, tea.Batch(cmd, m.statsCmd())

	case taskpicker.SelectedMsg:
		return m, m.linkCmd(msg.Task.ID)

	case linkedMsg:
		if msg.err != nil {
			err := msg.err
			return m, cmds.Emit(fmt.Errorf("link task: %w", err))
		}
		// Reload the embedded task list; its TasksLoadedMsg refreshes stats.
		return m, m.tasks.Init()

	case projectconvert.ConfirmedMsg:
		return m, m.convertCmd()

	case convertedMsg:
		if msg.err != nil {
			err := msg.err
			return m, cmds.Emit(fmt.Errorf("convert to task: %w", err))
		}
		// The project is gone; pop back to the project list, which reloads.
		return screen.Dismiss()

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		headerHeight := lipgloss.Height(m.renderHeader())
		taskMsg := tea.WindowSizeMsg{
			Width:  msg.Width,
			Height: msg.Height - headerHeight,
		}
		var cmd tea.Cmd
		m.tasks, cmd = m.tasks.Update(taskMsg)
		return m, cmd

	case tea.KeyPressMsg:
		if !screen.CapturingInput(m.tasks) {
			switch {
			case key.Matches(msg, m.KeyMap.Edit):
				return m, screen.Push(projectedit.New(m.project, m.projectSvc, nil))
			case key.Matches(msg, m.KeyMap.LinkTask):
				return m, screen.Push(taskpicker.New(m.taskSvc))
			case key.Matches(msg, m.KeyMap.ConvertToTask):
				return m, screen.Push(projectconvert.New(m.project))
			}
		}
	}

	var cmd tea.Cmd
	m.tasks, cmd = m.tasks.Update(msg)
	return m, cmd
}

// linkCmd links the chosen standalone task into this project.
func (m Model) linkCmd(taskID int64) tea.Cmd {
	svc := m.projectSvc
	pid := m.project.ID
	return func() tea.Msg {
		if _, err := svc.LinkTaskToProject(context.Background(), taskID, pid); err != nil {
			return linkedMsg{err: err}
		}
		return linkedMsg{}
	}
}

// convertCmd collapses this project into a standalone task.
func (m Model) convertCmd() tea.Cmd {
	svc := m.projectSvc
	id := m.project.ID
	return func() tea.Msg {
		if _, err := svc.ConvertProjectToTask(context.Background(), id); err != nil {
			return convertedMsg{err: err}
		}
		return convertedMsg{}
	}
}

// updateKeybindings reconciles the action guards: Convert-to-Task is available
// only for an open, empty project; Link-Task only when a standalone open task
// exists to link.
func (m *Model) updateKeybindings() {
	m.KeyMap.ConvertToTask.SetEnabled(gtd.CanConvertProjectToTask(m.project, m.taskCount))
	m.KeyMap.LinkTask.SetEnabled(m.hasCandidates)
}

func (m Model) View() string {
	return lipgloss.JoinVertical(lipgloss.Left,
		m.renderHeader(),
		m.tasks.View(),
	)
}

func (m Model) renderHeader() string {
	var lines []string

	lines = append(lines, titleStyle.Render(m.project.Title))

	lines = append(lines, labelStyle.Render("Status:")+" "+statusStyle.Render(statusLabel(m.project.Status)))

	if m.project.Outcome != "" {
		lines = append(lines, labelStyle.Render("Outcome:")+" "+valueStyle.Render(m.project.Outcome))
	}

	if m.project.Due != nil {
		lines = append(lines, labelStyle.Render("Due:")+" "+valueStyle.Render(m.project.Due.Local().Format(time.DateOnly)))
	}

	lines = append(lines, "")
	return strings.Join(lines, "\n")
}

func statusLabel(s gtd.ProjectStatus) string {
	switch s {
	case gtd.ProjectStatusOpen:
		return "Open"
	case gtd.ProjectStatusSomeday:
		return "Someday"
	case gtd.ProjectStatusDone:
		return "Done"
	case gtd.ProjectStatusDropped:
		return "Dropped"
	default:
		return fmt.Sprintf("%s", s)
	}
}

func (m Model) CapturingInput() bool {
	return screen.CapturingInput(m.tasks)
}

func (m Model) Keys() []keymap.Group {
	return slices.Concat(m.KeyMap.Keys(), m.tasks.Keys())
}
