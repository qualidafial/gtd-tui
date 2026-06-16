// Package clarify walks the GTD clarify decision tree for a single inbox
// item. The wizard is built on the form toolkit:
//
//   - An initial form covers every branch of the tree (actionable/non-
//     actionable, single-task/project, do-it-now, delegate). Visibility
//     predicates reveal the questions relevant to each path. The trailing
//     savefield commits via the appropriate InboxService call.
//   - For the project branch, the first task is committed with
//     ClarifyAsProject; the wizard then replaces its form with a fresh
//     per-task form so the user can capture additional next actions in
//     the same project, looping until they press Esc.
//   - For tasks the user said they'd do in under two minutes the wizard
//     pushes the doitnow overlay; on completion + project context the
//     wizard rebuilds the per-task form for the next iteration.
package clarify

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/qualidafial/gtd-tui"
	"github.com/qualidafial/gtd-tui/tui/cmds"
	"github.com/qualidafial/gtd-tui/tui/components/form"
	"github.com/qualidafial/gtd-tui/tui/components/form/inputfield"
	"github.com/qualidafial/gtd-tui/tui/components/form/radiofield"
	"github.com/qualidafial/gtd-tui/tui/components/form/savefield"
	"github.com/qualidafial/gtd-tui/tui/components/form/selectfield"
	"github.com/qualidafial/gtd-tui/tui/components/form/textfield"
	"github.com/qualidafial/gtd-tui/tui/components/screen"
	"github.com/qualidafial/gtd-tui/tui/internal/keymap"
	"github.com/qualidafial/gtd-tui/tui/pages/inbox/clarify/doitnow"
	"github.com/qualidafial/gtd-tui/tui/theme"
)

var (
	keyBack = key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel"))

	headerStyle   = theme.Title
	itemDescStyle = lipgloss.NewStyle().Foreground(theme.Muted)
	doneStyle     = theme.DoneTitle
	errStyle      = theme.Error
	hintStyle     = theme.Label.Italic(true)
)

// phase tracks which form the wizard is currently showing.
type phase int

const (
	phaseInitial     phase = iota // root tree → leaf decision
	phaseProjectLoop              // per-task add loop after ClarifyAsProject
)

type Model struct {
	item       gtd.Item
	svc        gtd.InboxService
	taskSvc    gtd.TaskService
	projectSvc gtd.ProjectService

	phase phase
	form  form.Model

	committedProject *gtd.Project
	committedTasks   []gtd.Task

	saving bool
	err    error
}

func New(item gtd.Item, svc gtd.InboxService, taskSvc gtd.TaskService, projectSvc gtd.ProjectService) Model {
	// The initial form is built immediately with an empty Project select; the
	// open-project list loads asynchronously (Init → loadProjectsCmd →
	// projectsLoadedMsg) and is populated in place via form.UpdateField. The
	// Project select uses WithHideWhenEmpty so it stays hidden until (and
	// unless) options arrive.
	return Model{
		item:       item,
		svc:        svc,
		taskSvc:    taskSvc,
		projectSvc: projectSvc,
		phase:      phaseInitial,
		form:       buildInitialForm(item),
	}
}

func (m Model) Init() tea.Cmd { return tea.Batch(m.form.Init(), m.loadProjectsCmd()) }

// -- form construction ------------------------------------------------------

// Predicate keys we read out of FieldValues. Centralizing the strings here
// keeps the buildXxxForm closures and the Submit handler in sync.
const (
	kActionable   = "actionable"
	kNonAct       = "nonAct"
	kSomedayTitle = "somedayTitle"
	kSomedayDesc  = "somedayDesc"
	kMultiStep    = "multiStep"
	kProjectTitle = "projectTitle"
	kProjectOut   = "projectOutcome"
	kProjectDesc  = "projectDesc"
	kTaskTitle    = "taskTitle"
	kTaskDesc     = "taskDesc"
	kProject      = "project"
	kUnder2Min    = "under2Min"
	kDoer         = "doer"
	kAssignee     = "assignee"
	kSaveDiscard  = "saveDiscard"
	kSaveSomeday  = "saveSomeday"
	kSaveTask     = "saveTask"
	kSaveProject  = "saveProject"
	kSaveLoop     = "saveLoop"
)

// buildInitialForm assembles the full clarify form with progressive
// visibility predicates so each branch's questions reveal as the user
// answers the parent radio. The single-task branch's optional Project
// select starts empty and is populated later via [form.Model.UpdateField]
// once the open-project list loads.
func buildInitialForm(item gtd.Item) form.Model {
	requireNonEmpty := func(label string) func(string) error {
		return func(s string) error {
			if strings.TrimSpace(s) == "" {
				return errors.New(label + " is required")
			}
			return nil
		}
	}

	actionable := radiofield.New(kActionable, "Actionable?", []radiofield.Option[bool]{
		{Display: "Yes", Value: true},
		{Display: "No", Value: false},
	})

	nonAct := radiofield.New(kNonAct, "Non-actionable choice", []radiofield.Option[string]{
		{Display: "Trash", Value: "trash"},
		{Display: "Someday", Value: "someday"},
	}, radiofield.WithVisible[string](func(v form.Values) bool {
		got, _ := v.Get(kActionable).(bool)
		return !got
	}))

	somedayTitle := inputfield.New(kSomedayTitle, "Someday title",
		inputfield.WithValue(item.Title),
		inputfield.WithValidator(requireNonEmpty("title")),
		inputfield.WithVisible(func(v form.Values) bool {
			return v.Get(kNonAct) == "someday"
		}),
	)
	somedayDesc := textfield.New(kSomedayDesc, "Someday description",
		textfield.WithValue(item.Description),
		textfield.WithVisible(func(v form.Values) bool {
			return v.Get(kNonAct) == "someday"
		}),
	)
	saveSomeday := savefield.New(kSaveSomeday, savefield.WithLabel("Park"),
		savefield.WithVisible(func(v form.Values) bool {
			return v.Get(kNonAct) == "someday"
		}),
	)
	saveDiscard := savefield.New(kSaveDiscard, savefield.WithLabel("Discard"),
		savefield.WithVisible(func(v form.Values) bool {
			return v.Get(kNonAct) == "trash"
		}),
	)

	multiStep := radiofield.New(kMultiStep, "Multi-step?", []radiofield.Option[bool]{
		{Display: "Single task", Value: false},
		{Display: "Project", Value: true},
	}, radiofield.WithVisible[bool](func(v form.Values) bool {
		got, _ := v.Get(kActionable).(bool)
		return got
	}))

	projectTitle := inputfield.New(kProjectTitle, "Project title",
		inputfield.WithValue(item.Title),
		inputfield.WithValidator(requireNonEmpty("project title")),
		inputfield.WithVisible(func(v form.Values) bool {
			got, _ := v.Get(kMultiStep).(bool)
			return got
		}),
	)
	projectOutcome := inputfield.New(kProjectOut, "Outcome",
		inputfield.WithValidator(requireNonEmpty("outcome")),
		inputfield.WithVisible(func(v form.Values) bool {
			got, _ := v.Get(kMultiStep).(bool)
			return got
		}),
	)
	projectDesc := textfield.New(kProjectDesc, "Project description",
		textfield.WithValue(item.Description),
		textfield.WithVisible(func(v form.Values) bool {
			got, _ := v.Get(kMultiStep).(bool)
			return got
		}),
	)

	taskTitle := inputfield.New(kTaskTitle, "Task title",
		inputfield.WithValue(item.Title),
		inputfield.WithValidator(requireNonEmpty("task title")),
		inputfield.WithVisible(taskFieldsVisible),
	)
	taskDesc := textfield.New(kTaskDesc, "Task description",
		textfield.WithValue(item.Description),
		textfield.WithVisible(taskFieldsVisible),
	)
	under2Min := radiofield.New(kUnder2Min, "< 2 min?", []radiofield.Option[bool]{
		{Display: "No", Value: false},
		{Display: "Yes (do it now)", Value: true},
	}, radiofield.WithVisible[bool](taskFieldsVisible))
	doer := radiofield.New(kDoer, "Who?", []radiofield.Option[string]{
		{Display: "Me", Value: "me"},
		{Display: "Someone else", Value: "someone"},
	}, radiofield.WithVisible[string](func(v form.Values) bool {
		if !taskFieldsVisible(v) {
			return false
		}
		under, _ := v.Get(kUnder2Min).(bool)
		return !under
	}))
	assignee := inputfield.New(kAssignee, "Assignee",
		inputfield.WithValidator(requireNonEmpty("assignee")),
		inputfield.WithVisible(func(v form.Values) bool {
			return v.Get(kDoer) == "someone"
		}),
	)

	// Single-task branch only: an optional select to attach the task to an
	// existing open project. Defaults to "(none)" (standalone). Shown for
	// every single task regardless of the <2 min / doer answers, and never
	// in the project branch (the first task auto-attaches to the new project)
	// nor when there are no open projects to attach to. Options load
	// asynchronously; WithHideWhenEmpty keeps it hidden until they arrive (and
	// when the loaded set is empty).
	project := selectfield.New[int64](kProject, "Project", nil,
		selectfield.WithNone[int64]("(none)"),
		selectfield.WithHideWhenEmpty[int64](),
		selectfield.WithVisible[int64](func(v form.Values) bool {
			actYes, _ := v.Get(kActionable).(bool)
			multi, _ := v.Get(kMultiStep).(bool)
			return actYes && !multi
		}),
	)

	saveTask := savefield.New(kSaveTask, savefield.WithLabel("Create task"),
		savefield.WithVisible(func(v form.Values) bool {
			got, _ := v.Get(kMultiStep).(bool)
			actYes, _ := v.Get(kActionable).(bool)
			return actYes && !got
		}),
	)
	saveProject := savefield.New(kSaveProject, savefield.WithLabel("Create project"),
		savefield.WithVisible(func(v form.Values) bool {
			got, _ := v.Get(kMultiStep).(bool)
			actYes, _ := v.Get(kActionable).(bool)
			return actYes && got
		}),
	)

	return form.New(
		actionable,
		nonAct, somedayTitle, somedayDesc, saveDiscard, saveSomeday,
		multiStep,
		projectTitle, projectOutcome, projectDesc,
		taskTitle, taskDesc, under2Min, doer, assignee, project,
		saveTask, saveProject,
	)
}

// taskFieldsVisible is the shared predicate for the per-task block of the
// initial form (visible whenever actionable=true).
func taskFieldsVisible(v form.Values) bool {
	got, _ := v.Get(kActionable).(bool)
	return got
}

// buildLoopForm assembles a fresh per-task form for the project loop. The
// fields are the same as the initial form's per-task block, but no value
// is prefilled from the original item — loop tasks are new actions.
func buildLoopForm() form.Model {
	requireNonEmpty := func(label string) func(string) error {
		return func(s string) error {
			if strings.TrimSpace(s) == "" {
				return errors.New(label + " is required")
			}
			return nil
		}
	}

	taskTitle := inputfield.New(kTaskTitle, "Next task title",
		inputfield.WithValidator(requireNonEmpty("task title")),
	)
	taskDesc := textfield.New(kTaskDesc, "Next task description")
	under2Min := radiofield.New(kUnder2Min, "< 2 min?", []radiofield.Option[bool]{
		{Display: "No", Value: false},
		{Display: "Yes (do it now)", Value: true},
	})
	doer := radiofield.New(kDoer, "Who?", []radiofield.Option[string]{
		{Display: "Me", Value: "me"},
		{Display: "Someone else", Value: "someone"},
	}, radiofield.WithVisible[string](func(v form.Values) bool {
		under, _ := v.Get(kUnder2Min).(bool)
		return !under
	}))
	assignee := inputfield.New(kAssignee, "Assignee",
		inputfield.WithValidator(requireNonEmpty("assignee")),
		inputfield.WithVisible(func(v form.Values) bool {
			return v.Get(kDoer) == "someone"
		}),
	)
	save := savefield.New(kSaveLoop, savefield.WithLabel("Save task"))

	return form.New(taskTitle, taskDesc, under2Min, doer, assignee, save)
}

// -- Update -----------------------------------------------------------------

func (m Model) Update(msg tea.Msg) (screen.Screen, tea.Cmd) {
	if m.err != nil {
		if kp, ok := msg.(tea.KeyPressMsg); ok && key.Matches(kp, keyBack) {
			m.err = nil
		}
		return m, nil
	}

	if m.saving {
		return m.handleSavingMsg(msg)
	}

	switch msg := msg.(type) {
	case projectsLoadedMsg:
		return m.handleProjectsLoaded(msg)
	case discardedMsg:
		return m.handleDiscarded(msg)
	case incubatedMsg:
		return m.handleIncubated(msg)
	case clarifiedAsTaskMsg:
		return m.handleClarifiedAsTask(msg)
	case clarifiedAsProjectMsg:
		return m.handleClarifiedAsProject(msg)
	case projectTaskCreatedMsg:
		return m.handleProjectTaskCreated(msg)
	case doitnow.ResultMsg:
		return m.handleDoItNowResult(msg)
	case form.SubmittedMsg:
		_ = msg
		return m.handleSubmitted()
	case tea.KeyPressMsg:
		if key.Matches(msg, keyBack) {
			return screen.Dismiss()
		}
	}

	var cmd tea.Cmd
	m.form, cmd = m.form.Update(msg)
	return m, cmd
}

// handleSavingMsg routes only the result messages that can move the wizard
// forward while a save is in flight; other messages are dropped to prevent
// re-firing the form pipeline mid-commit.
func (m Model) handleSavingMsg(msg tea.Msg) (screen.Screen, tea.Cmd) {
	switch msg := msg.(type) {
	case discardedMsg:
		return m.handleDiscarded(msg)
	case incubatedMsg:
		return m.handleIncubated(msg)
	case clarifiedAsTaskMsg:
		return m.handleClarifiedAsTask(msg)
	case clarifiedAsProjectMsg:
		return m.handleClarifiedAsProject(msg)
	case projectTaskCreatedMsg:
		return m.handleProjectTaskCreated(msg)
	}
	return m, nil
}

// handleSubmitted inspects the current form's values and dispatches the
// matching service call.
func (m Model) handleSubmitted() (screen.Screen, tea.Cmd) {
	vals := m.form.FieldValues()

	if m.phase == phaseProjectLoop {
		m.saving = true
		return m, m.createProjectTaskCmd(vals)
	}

	actionable, _ := vals[kActionable].(bool)
	if !actionable {
		switch vals[kNonAct] {
		case "trash":
			m.saving = true
			return m, m.discardCmd()
		case "someday":
			m.saving = true
			return m, m.incubateCmd(vals)
		}
		return m, nil
	}

	multiStep, _ := vals[kMultiStep].(bool)
	if multiStep {
		m.saving = true
		return m, m.clarifyAsProjectCmd(vals)
	}
	m.saving = true
	return m, m.clarifyAsTaskCmd(vals)
}

// -- result handlers --------------------------------------------------------

// handleProjectsLoaded populates the Project select in place once the
// open-project list is available. A load failure surfaces via the error
// path; the wizard remains usable (without project attachment) once the
// error is dismissed. The select belongs only to the initial form, so the
// update is skipped if the wizard has already moved into the project loop.
func (m Model) handleProjectsLoaded(msg projectsLoadedMsg) (screen.Screen, tea.Cmd) {
	if msg.err != nil {
		m.err = msg.err
		return m, cmds.Emit(fmt.Errorf("load projects: %w", msg.err))
	}
	if m.phase != phaseInitial {
		return m, nil
	}
	opts := make([]selectfield.Option[int64], 0, len(msg.projects))
	for _, p := range msg.projects {
		opts = append(opts, selectfield.Option[int64]{Display: p.Title, Value: p.ID})
	}
	m.form = m.form.UpdateField(kProject, func(f form.Field) form.Field {
		return f.(selectfield.Model[int64]).SetOptions(opts)
	})
	return m, nil
}

func (m Model) handleDiscarded(msg discardedMsg) (screen.Screen, tea.Cmd) {
	if msg.err != nil {
		m.err = msg.err
		m.saving = false
		err := msg.err
		return m, cmds.Emit(fmt.Errorf("discard: %w", err))
	}
	return screen.Dismiss()
}

func (m Model) handleIncubated(msg incubatedMsg) (screen.Screen, tea.Cmd) {
	if msg.err != nil {
		m.err = msg.err
		m.saving = false
		err := msg.err
		return m, cmds.Emit(fmt.Errorf("incubate: %w", err))
	}
	return screen.Dismiss()
}

func (m Model) handleClarifiedAsTask(msg clarifiedAsTaskMsg) (screen.Screen, tea.Cmd) {
	if msg.err != nil {
		m.err = msg.err
		m.saving = false
		err := msg.err
		return m, cmds.Emit(fmt.Errorf("clarify task: %w", err))
	}
	m.saving = false
	if msg.under2Min {
		// Single-task do-it-now is terminal. Replace the wizard with the
		// doitnow overlay so dismissing it exits straight to the inbox
		// rather than dropping back onto the spent clarify form.
		return screen.Replace(doitnow.New(msg.task, m.taskSvc))
	}
	return screen.Dismiss()
}

func (m Model) handleClarifiedAsProject(msg clarifiedAsProjectMsg) (screen.Screen, tea.Cmd) {
	if msg.err != nil {
		m.err = msg.err
		m.saving = false
		err := msg.err
		return m, cmds.Emit(fmt.Errorf("clarify as project: %w", err))
	}
	p := msg.project
	m.committedProject = &p
	m.committedTasks = append(m.committedTasks, msg.task)
	m.saving = false
	if msg.under2Min {
		// Project do-it-now loops: keep the wizard underneath so its
		// ResultMsg can rebuild the loop form once doitnow resolves.
		return m, screen.Push(doitnow.New(msg.task, m.taskSvc))
	}
	return m.enterProjectLoop()
}

func (m Model) handleProjectTaskCreated(msg projectTaskCreatedMsg) (screen.Screen, tea.Cmd) {
	if msg.err != nil {
		m.err = msg.err
		m.saving = false
		err := msg.err
		return m, cmds.Emit(fmt.Errorf("create project task: %w", err))
	}
	m.committedTasks = append(m.committedTasks, msg.task)
	m.saving = false
	if msg.under2Min {
		return m, screen.Push(doitnow.New(msg.task, m.taskSvc))
	}
	return m.enterProjectLoop()
}

func (m Model) handleDoItNowResult(msg doitnow.ResultMsg) (screen.Screen, tea.Cmd) {
	if !msg.Completed {
		// The user left the task open from the do-it-now prompt (esc): exit
		// the whole wizard back to the inbox rather than continuing the loop.
		return screen.Dismiss()
	}
	if n := len(m.committedTasks); n > 0 {
		m.committedTasks[n-1].Status = gtd.TaskStatusDone
	}
	if m.committedProject != nil {
		// Project context: the completed task loops back to a fresh per-task
		// form so the user can capture the next next-action.
		return m.enterProjectLoop()
	}
	return screen.Dismiss()
}

// enterProjectLoop swaps the form for a fresh per-task form so the user
// can capture the next next-action in the project.
func (m Model) enterProjectLoop() (screen.Screen, tea.Cmd) {
	m.phase = phaseProjectLoop
	m.saving = false
	m.form = buildLoopForm()
	return m, m.form.Init()
}

// -- save cmds --------------------------------------------------------------

type projectsLoadedMsg struct {
	projects []gtd.Project
	err      error
}

func (m Model) loadProjectsCmd() tea.Cmd {
	svc := m.projectSvc
	return func() tea.Msg {
		projects, err := svc.ListProjects(context.Background(), gtd.ProjectFilter{}.WithStatus(gtd.ProjectStatusOpen))
		if err != nil {
			slog.Error("load projects: " + err.Error())
		}
		return projectsLoadedMsg{projects: projects, err: err}
	}
}

type discardedMsg struct {
	item gtd.Item
	err  error
}

type incubatedMsg struct {
	project gtd.Project
	item    gtd.Item
	err     error
}

type clarifiedAsTaskMsg struct {
	task      gtd.Task
	under2Min bool
	err       error
}

type clarifiedAsProjectMsg struct {
	project   gtd.Project
	task      gtd.Task
	under2Min bool
	err       error
}

type projectTaskCreatedMsg struct {
	task      gtd.Task
	under2Min bool
	err       error
}

func (m Model) discardCmd() tea.Cmd {
	svc := m.svc
	id := m.item.ID
	return func() tea.Msg {
		item, err := svc.Discard(context.Background(), id)
		if err != nil {
			slog.Error("discard item: " + err.Error())
		}
		return discardedMsg{item: item, err: err}
	}
}

func (m Model) incubateCmd(vals map[string]any) tea.Cmd {
	title, _ := vals[kSomedayTitle].(string)
	desc, _ := vals[kSomedayDesc].(string)
	svc := m.svc
	id := m.item.ID
	project := gtd.Project{Title: title, Description: desc}
	return func() tea.Msg {
		p, i, err := svc.Incubate(context.Background(), id, project)
		if err != nil {
			slog.Error("incubate item: " + err.Error())
		}
		return incubatedMsg{project: p, item: i, err: err}
	}
}

func (m Model) clarifyAsTaskCmd(vals map[string]any) tea.Cmd {
	svc := m.svc
	id := m.item.ID
	task := taskFromVals(vals)
	under, _ := vals[kUnder2Min].(bool)
	return func() tea.Msg {
		t, _, err := svc.ClarifyAsTask(context.Background(), id, task)
		if err != nil {
			slog.Error("clarify task: " + err.Error())
		}
		return clarifiedAsTaskMsg{task: t, under2Min: under, err: err}
	}
}

func (m Model) clarifyAsProjectCmd(vals map[string]any) tea.Cmd {
	svc := m.svc
	id := m.item.ID
	title, _ := vals[kProjectTitle].(string)
	out, _ := vals[kProjectOut].(string)
	desc, _ := vals[kProjectDesc].(string)
	project := gtd.Project{Title: title, Outcome: out, Description: desc}
	firstTask := taskFromVals(vals)
	under, _ := vals[kUnder2Min].(bool)
	return func() tea.Msg {
		p, t, _, err := svc.ClarifyAsProject(context.Background(), id, project, firstTask)
		if err != nil {
			slog.Error("clarify as project: " + err.Error())
		}
		return clarifiedAsProjectMsg{project: p, task: t, under2Min: under, err: err}
	}
}

func (m Model) createProjectTaskCmd(vals map[string]any) tea.Cmd {
	taskSvc := m.taskSvc
	task := taskFromVals(vals)
	task.ProjectID = &m.committedProject.ID
	under, _ := vals[kUnder2Min].(bool)
	return func() tea.Msg {
		t, err := taskSvc.CreateTask(context.Background(), task)
		if err != nil {
			slog.Error("create project task: " + err.Error())
		}
		return projectTaskCreatedMsg{task: t, under2Min: under, err: err}
	}
}

func taskFromVals(vals map[string]any) gtd.Task {
	title, _ := vals[kTaskTitle].(string)
	desc, _ := vals[kTaskDesc].(string)
	task := gtd.Task{
		Title:       title,
		Description: desc,
		Status:      gtd.TaskStatusOpen,
	}
	if doer, _ := vals[kDoer].(string); doer == "someone" {
		if asg, _ := vals[kAssignee].(string); asg != "" {
			a := asg
			task.Assignee = &a
		}
	}
	// The project select is present only in the single-task branch of the
	// initial form. When absent (loop form) or left on "(none)" the value is
	// the zero int64, leaving the task standalone.
	if pid, _ := vals[kProject].(int64); pid != 0 {
		task.ProjectID = new(pid)
	}
	return task
}

// -- rendering --------------------------------------------------------------

func (m Model) View() string {
	var sections []string

	sections = append(sections, headerStyle.Render(m.item.Title))
	if m.item.Description != "" {
		sections = append(sections, itemDescStyle.Render(m.item.Description))
	}
	sections = append(sections, "")

	if m.committedProject != nil {
		sections = append(sections,
			doneStyle.Render(fmt.Sprintf("Project: %s", m.committedProject.Title)),
		)
		for _, t := range m.committedTasks {
			line := fmt.Sprintf("  · %s", t.Title)
			if t.Status == gtd.TaskStatusDone {
				line += " (done)"
			}
			sections = append(sections, doneStyle.Render(line))
		}
		sections = append(sections, "")
	}

	sections = append(sections, m.form.View())

	if m.err != nil {
		sections = append(sections,
			"",
			errStyle.Render(m.err.Error()),
			hintStyle.Render("press esc to dismiss"),
		)
	}

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m Model) CapturingInput() bool { return m.err == nil && !m.saving }

// Keys aggregates the form's resolved bindings and appends this screen's
// own esc binding as a trailing group; Resolve subtracts the overlay's
// duplicate esc.
func (m Model) Keys() []keymap.Group {
	return append(m.form.Keys(), keymap.Group{{Binding: keyBack, Vis: keymap.Short}})
}
