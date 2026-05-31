// Package clarify walks the GTD clarify decision tree for a single inbox item.
// The wizard renders all answered questions as a persistent column and reveals
// the next unanswered question inline. Leaf decisions commit via the
// appropriate InboxService operation; tasks marked do-it-now push the doitnow
// overlay so the user can confirm completion, and project clarifications loop
// to define successive tasks until one is committed as the open next-action.
//
// Branches:
//
//   - Actionable=Yes / Multi-step=No → single-task: per-task block → ClarifyAsTask
//   - Actionable=Yes / Multi-step=Yes → project: project form → loop per-task
//     blocks (ClarifyAsProject for the first, TaskService.CreateTask after)
//   - Actionable=No / Trash → confirm → Discard
//   - Actionable=No / Someday → title + description → Incubate
//
// Back-navigation (Esc) reverts to the previous step and clears its answer.
// Once any commit lands (single-task save, ClarifyAsProject checkpoint, any
// loop task), back-nav is disabled — Esc dismisses instead.
//
// Deferred to a future iteration: the project-attach picker (selecting an
// existing project for a single task). When the user answers "Attach to a
// project? Yes", the wizard currently shows a "not yet implemented" message;
// the No path commits as standalone.
package clarify

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textarea"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/qualidafial/gtd-tui"
	"github.com/qualidafial/gtd-tui/tui/components/screen"
	"github.com/qualidafial/gtd-tui/tui/pages/inbox/clarify/doitnow"
)

var (
	headerStyle   = lipgloss.NewStyle().Bold(true)
	itemDescStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("250"))
	questionStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	answerStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("36"))
	doneStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("65")).Faint(true)
	activeStyle   = lipgloss.NewStyle().Bold(true)
	errStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	hintStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Italic(true)
)

const questionColWidth = 22

type step int

const (
	stepActionable step = iota
	stepNonActionable
	stepDiscardConfirm
	stepSomedayTitle
	stepSomedayDescription
	stepMultiStep
	stepProjectTitle
	stepProjectOutcome
	stepProjectDescription
	stepTaskTitle
	stepTaskDescription
	stepUnder2Min
	stepDoer
	stepAssignee
	stepAttachProject
	stepSaving
	stepDoItNow
	stepDone
)

type doer int

const (
	doerMe doer = iota
	doerSomeoneElse
)

type nonActChoice int

const (
	choiceTrash nonActChoice = iota
	choiceSomeday
)

// answers accumulates the user's decisions. Pointer fields denote "answered"
// vs "not yet asked"; string fields are empty until set.
type answers struct {
	actionable    *bool
	nonActChoice  *nonActChoice
	somedayTitle  string
	somedayDesc   string
	multiStep     *bool
	projectTitle  string
	projectOut    string
	projectDesc   string
	taskTitle     string
	taskDesc      string
	under2Min     *bool
	doer          *doer
	assignee      string
	attachProject *bool
}

type Model struct {
	item    gtd.Item
	svc     gtd.InboxService
	taskSvc gtd.TaskService

	KeyMap KeyMap

	step     step
	ans      answers
	input    textinput.Model // single-line: titles, assignee
	textarea textarea.Model  // multi-line: descriptions
	width    int

	// history is the stack of steps to pop back through for Esc. Cleared once
	// the wizard reaches its first commit; back-nav is disabled afterward
	// because clearing an answer would invalidate persisted state.
	history []step

	// project-branch loop state. committedProject is set after the first
	// ClarifyAsProject; committedTasks accumulates every task that's been
	// persisted (open or done).
	committedProject *gtd.Project
	committedTasks   []gtd.Task

	err error
}

func New(item gtd.Item, svc gtd.InboxService, taskSvc gtd.TaskService) Model {
	ti := textinput.New()
	ti.CharLimit = 200

	ta := textarea.New()
	ta.CharLimit = 2000
	ta.ShowLineNumbers = false
	taKeys := textarea.DefaultKeyMap()
	taKeys.InsertNewline = key.NewBinding(key.WithKeys("alt+enter"), key.WithHelp("alt+enter", "newline"))
	ta.KeyMap = taKeys

	return Model{
		item:     item,
		svc:      svc,
		taskSvc:  taskSvc,
		KeyMap:   defaultKeyMap(),
		step:     stepActionable,
		input:    ti,
		textarea: ta,
	}
}

func (m Model) Init() tea.Cmd { return nil }

// --- top-level Update ------------------------------------------------------

func (m Model) Update(msg tea.Msg) (screen.Screen, tea.Cmd) {
	if m.err != nil {
		if kp, ok := msg.(tea.KeyPressMsg); ok && key.Matches(kp, m.KeyMap.Back) {
			m.err = nil
		}
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		inputWidth := max(msg.Width-questionColWidth-2, 10)
		m.input.SetWidth(inputWidth)
		m.textarea.SetWidth(inputWidth)
		m.textarea.SetHeight(4)
		return m, nil

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
	case tea.KeyPressMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

// --- key dispatch ----------------------------------------------------------

func (m Model) handleKey(msg tea.KeyPressMsg) (screen.Screen, tea.Cmd) {
	if key.Matches(msg, m.KeyMap.Back) {
		return m.handleBack()
	}

	switch m.step {
	case stepActionable:
		return m.handleActionable(msg)
	case stepNonActionable:
		return m.handleNonActionable(msg)
	case stepDiscardConfirm:
		return m.handleDiscardConfirm(msg)
	case stepSomedayTitle:
		return m.handleTextStep(msg, func(a *answers, v string) { a.somedayTitle = v }, stepSomedayDescription, m.item.Description, true)
	case stepSomedayDescription:
		return m.handleTextStep(msg, func(a *answers, v string) { a.somedayDesc = v }, stepSaving, "", true, m.incubateCmd())
	case stepMultiStep:
		return m.handleMultiStep(msg)
	case stepProjectTitle:
		return m.handleTextStep(msg, func(a *answers, v string) { a.projectTitle = v }, stepProjectOutcome, "", false)
	case stepProjectOutcome:
		return m.handleTextStep(msg, func(a *answers, v string) { a.projectOut = v }, stepProjectDescription, m.item.Description, true)
	case stepProjectDescription:
		return m.handleTextStep(msg, func(a *answers, v string) { a.projectDesc = v }, stepTaskTitle, "", true)
	case stepTaskTitle:
		return m.handleTaskTitle(msg)
	case stepTaskDescription:
		return m.handleTaskDescription(msg)
	case stepUnder2Min:
		return m.handleUnder2Min(msg)
	case stepDoer:
		return m.handleDoer(msg)
	case stepAssignee:
		return m.handleAssignee(msg)
	case stepAttachProject:
		return m.handleAttachProject(msg)
	}
	return m, nil
}

// --- step handlers (triage / non-actionable) -------------------------------

func (m Model) handleActionable(msg tea.KeyPressMsg) (screen.Screen, tea.Cmd) {
	switch {
	case key.Matches(msg, m.KeyMap.Yes):
		t := true
		m.ans.actionable = &t
		m.advance(stepMultiStep)
	case key.Matches(msg, m.KeyMap.No):
		f := false
		m.ans.actionable = &f
		m.advance(stepNonActionable)
	}
	return m, nil
}

func (m Model) handleNonActionable(msg tea.KeyPressMsg) (screen.Screen, tea.Cmd) {
	switch {
	case key.Matches(msg, m.KeyMap.Trash):
		c := choiceTrash
		m.ans.nonActChoice = &c
		m.advance(stepDiscardConfirm)
	case key.Matches(msg, m.KeyMap.Someday):
		c := choiceSomeday
		m.ans.nonActChoice = &c
		m.advance(stepSomedayTitle)
		m.input.SetValue(m.item.Title)
		m.input.CursorEnd()
		return m, m.input.Focus()
	}
	return m, nil
}

func (m Model) handleDiscardConfirm(msg tea.KeyPressMsg) (screen.Screen, tea.Cmd) {
	switch {
	case key.Matches(msg, m.KeyMap.Yes):
		m.history = nil // commit-bound
		m.step = stepSaving
		return m, m.discardCmd()
	case key.Matches(msg, m.KeyMap.No):
		return m.handleBack()
	}
	return m, nil
}

// --- step handlers (actionable / project / per-task) -----------------------

func (m Model) handleMultiStep(msg tea.KeyPressMsg) (screen.Screen, tea.Cmd) {
	switch {
	case key.Matches(msg, m.KeyMap.Yes):
		t := true
		m.ans.multiStep = &t
		m.advance(stepProjectTitle)
		m.input.SetValue(m.item.Title)
		m.input.CursorEnd()
		return m, m.input.Focus()
	case key.Matches(msg, m.KeyMap.No):
		f := false
		m.ans.multiStep = &f
		m.advance(stepTaskTitle)
		m.input.SetValue(m.item.Title)
		m.input.CursorEnd()
		return m, m.input.Focus()
	}
	return m, nil
}

func (m Model) handleTaskTitle(msg tea.KeyPressMsg) (screen.Screen, tea.Cmd) {
	if key.Matches(msg, m.KeyMap.Confirm) {
		title := strings.TrimSpace(m.input.Value())
		if title == "" {
			return m, nil
		}
		m.ans.taskTitle = title
		m.input.Blur()
		m.input.Reset()
		m.advance(stepTaskDescription)
		// Pre-populate description with item desc only for the FIRST task in a
		// project loop (or single-task branch). Subsequent loop tasks start
		// blank — they're new actions, not the original captured item.
		if !m.inLoopAfterFirstTask() {
			m.textarea.SetValue(m.item.Description)
		}
		return m, m.textarea.Focus()
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m Model) handleTaskDescription(msg tea.KeyPressMsg) (screen.Screen, tea.Cmd) {
	if key.Matches(msg, m.KeyMap.Confirm) {
		m.ans.taskDesc = strings.TrimSpace(m.textarea.Value())
		m.textarea.Blur()
		m.textarea.Reset()
		m.advance(stepUnder2Min)
		return m, nil
	}
	var cmd tea.Cmd
	m.textarea, cmd = m.textarea.Update(msg)
	return m, cmd
}

func (m Model) handleUnder2Min(msg tea.KeyPressMsg) (screen.Screen, tea.Cmd) {
	switch {
	case key.Matches(msg, m.KeyMap.Yes):
		t := true
		m.ans.under2Min = &t
		// Do-it-now: self is the doer, no delegate question.
		d := doerMe
		m.ans.doer = &d
		return m.afterPerTask()
	case key.Matches(msg, m.KeyMap.No):
		f := false
		m.ans.under2Min = &f
		m.advance(stepDoer)
	}
	return m, nil
}

func (m Model) handleDoer(msg tea.KeyPressMsg) (screen.Screen, tea.Cmd) {
	switch {
	case key.Matches(msg, m.KeyMap.Me):
		d := doerMe
		m.ans.doer = &d
		return m.afterPerTask()
	case key.Matches(msg, m.KeyMap.Someone):
		d := doerSomeoneElse
		m.ans.doer = &d
		m.advance(stepAssignee)
		m.input.Reset()
		return m, m.input.Focus()
	}
	return m, nil
}

func (m Model) handleAssignee(msg tea.KeyPressMsg) (screen.Screen, tea.Cmd) {
	if key.Matches(msg, m.KeyMap.Confirm) {
		value := strings.TrimSpace(m.input.Value())
		if value == "" {
			return m, nil
		}
		m.ans.assignee = value
		m.input.Blur()
		m.input.Reset()
		return m.afterPerTask()
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m Model) handleAttachProject(msg tea.KeyPressMsg) (screen.Screen, tea.Cmd) {
	switch {
	case key.Matches(msg, m.KeyMap.Yes):
		t := true
		m.ans.attachProject = &t
		m.err = errors.New("project-attach picker is not yet implemented")
	case key.Matches(msg, m.KeyMap.No):
		f := false
		m.ans.attachProject = &f
		m.history = nil
		m.step = stepSaving
		return m, m.clarifyAsTaskCmd()
	}
	return m, nil
}

// afterPerTask is the per-task block's exit hook. The single-task branch goes
// to the project-attach question; the project branch commits the task
// immediately (first via ClarifyAsProject, subsequent via TaskService). Must
// return the model so the caller's value-receiver gets the advanced step.
func (m Model) afterPerTask() (screen.Screen, tea.Cmd) {
	if m.ans.multiStep != nil && *m.ans.multiStep {
		m.history = nil
		m.step = stepSaving
		if m.committedProject == nil {
			return m, m.clarifyAsProjectCmd()
		}
		return m, m.createProjectTaskCmd()
	}
	m.advance(stepAttachProject)
	return m, nil
}

// inLoopAfterFirstTask reports whether the project loop has already committed
// at least one task. Used to suppress item-description prepopulation on
// subsequent task definitions.
func (m Model) inLoopAfterFirstTask() bool {
	return m.committedProject != nil && len(m.committedTasks) >= 1
}

// --- generic text-step helper ----------------------------------------------

// handleTextStep handles a text-input step that advances to `nextStep` on
// Enter, writing the trimmed value into m.ans via the supplied setField
// closure. The closure pattern is required because handleTextStep has a value
// receiver — taking a pointer to a caller-owned field would mutate the
// caller's local copy (which then gets discarded when handleTextStep returns
// its own local m). Passing a setter lets the helper write to its own m.ans
// before returning, so the mutation actually sticks.
//
// allowEmpty controls whether an empty trimmed value blocks advancement. If a
// commit cmd is supplied, advancement skips the next-step focus and instead
// emits the commit cmd (transitioning to stepSaving). If commit is nil, the
// helper focuses the next step's widget per focusForStep.
func (m Model) handleTextStep(msg tea.KeyPressMsg, setField func(*answers, string), nextStep step, preNext string, allowEmpty bool, commit ...tea.Cmd) (screen.Screen, tea.Cmd) {
	useTextarea := m.step == stepSomedayDescription || m.step == stepProjectDescription
	if key.Matches(msg, m.KeyMap.Confirm) {
		var value string
		if useTextarea {
			value = strings.TrimSpace(m.textarea.Value())
		} else {
			value = strings.TrimSpace(m.input.Value())
		}
		if value == "" && !allowEmpty {
			return m, nil
		}
		setField(&m.ans, value)
		if useTextarea {
			m.textarea.Blur()
			m.textarea.Reset()
		} else {
			m.input.Blur()
			m.input.Reset()
		}
		if len(commit) > 0 && commit[0] != nil {
			m.history = nil
			m.step = stepSaving
			return m, commit[0]
		}
		m.advance(nextStep)
		return m, m.focusForStep(nextStep, preNext)
	}
	if useTextarea {
		var cmd tea.Cmd
		m.textarea, cmd = m.textarea.Update(msg)
		return m, cmd
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m *Model) focusForStep(s step, prefill string) tea.Cmd {
	switch s {
	case stepSomedayDescription, stepProjectDescription:
		m.textarea.SetValue(prefill)
		return m.textarea.Focus()
	case stepSomedayTitle, stepProjectTitle, stepProjectOutcome,
		stepTaskTitle, stepAssignee:
		m.input.SetValue(prefill)
		m.input.CursorEnd()
		return m.input.Focus()
	}
	return nil
}

// --- navigation ------------------------------------------------------------

func (m *Model) advance(to step) {
	m.history = append(m.history, m.step)
	m.step = to
}

func (m Model) handleBack() (screen.Screen, tea.Cmd) {
	if len(m.history) == 0 {
		return screen.Dismiss()
	}
	// Pop history.
	prev := m.history[len(m.history)-1]
	m.history = m.history[:len(m.history)-1]
	// Clear the answer that was set when we ADVANCED FROM prev (i.e. the
	// answer associated with the step we're returning TO).
	m.clearAnswerFor(prev)
	// Reset any in-progress input widget for the current step before switching
	// back. The new step may set its own widget state below.
	m.input.Blur()
	m.input.Reset()
	m.textarea.Blur()
	m.textarea.Reset()
	m.step = prev
	// Re-focus / re-populate widgets for text steps.
	switch prev {
	case stepSomedayTitle:
		m.input.SetValue(m.item.Title)
		m.input.CursorEnd()
		return m, m.input.Focus()
	case stepSomedayDescription:
		m.textarea.SetValue(m.item.Description)
		return m, m.textarea.Focus()
	case stepProjectTitle:
		m.input.SetValue(m.item.Title)
		m.input.CursorEnd()
		return m, m.input.Focus()
	case stepProjectOutcome:
		m.input.SetValue("")
		return m, m.input.Focus()
	case stepProjectDescription:
		m.textarea.SetValue(m.item.Description)
		return m, m.textarea.Focus()
	case stepTaskTitle:
		// Use the original item title only when this is the first task; for
		// subsequent tasks the user enters a fresh title.
		if !m.inLoopAfterFirstTask() {
			m.input.SetValue(m.item.Title)
		}
		m.input.CursorEnd()
		return m, m.input.Focus()
	case stepTaskDescription:
		m.textarea.SetValue(m.ans.taskDesc) // restore whatever they had
		return m, m.textarea.Focus()
	case stepAssignee:
		m.input.SetValue(m.ans.assignee)
		m.input.CursorEnd()
		return m, m.input.Focus()
	}
	return m, nil
}

// clearAnswerFor zeros the answer field associated with the named step (the
// step we are returning TO on a back action). This is the "answer that gets
// re-asked" per the back-nav spec.
func (m *Model) clearAnswerFor(s step) {
	switch s {
	case stepActionable:
		m.ans.actionable = nil
	case stepNonActionable:
		m.ans.nonActChoice = nil
	case stepSomedayTitle:
		m.ans.somedayTitle = ""
	case stepSomedayDescription:
		m.ans.somedayDesc = ""
	case stepMultiStep:
		m.ans.multiStep = nil
	case stepProjectTitle:
		m.ans.projectTitle = ""
	case stepProjectOutcome:
		m.ans.projectOut = ""
	case stepProjectDescription:
		m.ans.projectDesc = ""
	case stepTaskTitle:
		m.ans.taskTitle = ""
	case stepTaskDescription:
		m.ans.taskDesc = ""
	case stepUnder2Min:
		m.ans.under2Min = nil
	case stepDoer:
		m.ans.doer = nil
	case stepAssignee:
		m.ans.assignee = ""
	case stepAttachProject:
		m.ans.attachProject = nil
	}
}

// --- save cmds -------------------------------------------------------------

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
	task gtd.Task
	err  error
}

type clarifiedAsProjectMsg struct {
	project gtd.Project
	task    gtd.Task
	err     error
}

type projectTaskCreatedMsg struct {
	task gtd.Task
	err  error
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

func (m Model) incubateCmd() tea.Cmd {
	svc := m.svc
	id := m.item.ID
	project := gtd.Project{
		Title:       m.ans.somedayTitle,
		Description: m.ans.somedayDesc,
	}
	return func() tea.Msg {
		p, i, err := svc.Incubate(context.Background(), id, project)
		if err != nil {
			slog.Error("incubate item: " + err.Error())
		}
		return incubatedMsg{project: p, item: i, err: err}
	}
}

func (m Model) clarifyAsTaskCmd() tea.Cmd {
	svc := m.svc
	id := m.item.ID
	task := m.taskFromAnswers()
	return func() tea.Msg {
		t, _, err := svc.ClarifyAsTask(context.Background(), id, task)
		if err != nil {
			slog.Error("clarify task: " + err.Error())
		}
		return clarifiedAsTaskMsg{task: t, err: err}
	}
}

func (m Model) clarifyAsProjectCmd() tea.Cmd {
	svc := m.svc
	id := m.item.ID
	project := gtd.Project{
		Title:       m.ans.projectTitle,
		Outcome:     m.ans.projectOut,
		Description: m.ans.projectDesc,
	}
	firstTask := m.taskFromAnswers()
	return func() tea.Msg {
		p, t, _, err := svc.ClarifyAsProject(context.Background(), id, project, firstTask)
		if err != nil {
			slog.Error("clarify as project: " + err.Error())
		}
		return clarifiedAsProjectMsg{project: p, task: t, err: err}
	}
}

func (m Model) createProjectTaskCmd() tea.Cmd {
	taskSvc := m.taskSvc
	task := m.taskFromAnswers()
	task.ProjectID = &m.committedProject.ID
	return func() tea.Msg {
		t, err := taskSvc.CreateTask(context.Background(), task)
		if err != nil {
			slog.Error("create project task: " + err.Error())
		}
		return projectTaskCreatedMsg{task: t, err: err}
	}
}

func (m Model) taskFromAnswers() gtd.Task {
	task := gtd.Task{
		Title:       m.ans.taskTitle,
		Description: m.ans.taskDesc,
		Status:      gtd.TaskStatusOpen,
	}
	if m.ans.assignee != "" {
		a := m.ans.assignee
		task.Assignee = &a
	}
	return task
}

// --- save msg handlers -----------------------------------------------------

func (m Model) handleDiscarded(msg discardedMsg) (screen.Screen, tea.Cmd) {
	if msg.err != nil {
		m.err = msg.err
		err := msg.err
		return m, func() tea.Msg { return fmt.Errorf("discard: %w", err) }
	}
	m.step = stepDone
	return screen.Dismiss()
}

func (m Model) handleIncubated(msg incubatedMsg) (screen.Screen, tea.Cmd) {
	if msg.err != nil {
		m.err = msg.err
		err := msg.err
		return m, func() tea.Msg { return fmt.Errorf("incubate: %w", err) }
	}
	m.step = stepDone
	return screen.Dismiss()
}

func (m Model) handleClarifiedAsTask(msg clarifiedAsTaskMsg) (screen.Screen, tea.Cmd) {
	if msg.err != nil {
		m.err = msg.err
		err := msg.err
		return m, func() tea.Msg { return fmt.Errorf("clarify task: %w", err) }
	}
	if m.ans.under2Min != nil && *m.ans.under2Min {
		m.step = stepDoItNow
		return m, screen.Push(doitnow.New(msg.task, m.taskSvc))
	}
	m.step = stepDone
	return screen.Dismiss()
}

func (m Model) handleClarifiedAsProject(msg clarifiedAsProjectMsg) (screen.Screen, tea.Cmd) {
	if msg.err != nil {
		m.err = msg.err
		err := msg.err
		return m, func() tea.Msg { return fmt.Errorf("clarify as project: %w", err) }
	}
	p := msg.project
	m.committedProject = &p
	m.committedTasks = append(m.committedTasks, msg.task)
	if m.ans.under2Min != nil && *m.ans.under2Min {
		m.step = stepDoItNow
		return m, screen.Push(doitnow.New(msg.task, m.taskSvc))
	}
	m.step = stepDone
	return screen.Dismiss()
}

func (m Model) handleProjectTaskCreated(msg projectTaskCreatedMsg) (screen.Screen, tea.Cmd) {
	if msg.err != nil {
		m.err = msg.err
		err := msg.err
		return m, func() tea.Msg { return fmt.Errorf("create project task: %w", err) }
	}
	m.committedTasks = append(m.committedTasks, msg.task)
	if m.ans.under2Min != nil && *m.ans.under2Min {
		m.step = stepDoItNow
		return m, screen.Push(doitnow.New(msg.task, m.taskSvc))
	}
	m.step = stepDone
	return screen.Dismiss()
}

func (m Model) handleDoItNowResult(msg doitnow.ResultMsg) (screen.Screen, tea.Cmd) {
	// Update the last committed task's status to reflect what happened.
	if n := len(m.committedTasks); n > 0 && msg.Completed {
		m.committedTasks[n-1].Status = gtd.TaskStatusDone
	}
	// In the project branch, a confirmed do-it-now loops to the next task.
	// Anywhere else (single-task, or unconfirmed do-it-now in project) we
	// exit the wizard.
	inProjectLoop := m.committedProject != nil
	if inProjectLoop && msg.Completed {
		m.resetForNextLoopTask()
		return m, m.input.Focus()
	}
	m.step = stepDone
	return screen.Dismiss()
}

// resetForNextLoopTask clears per-task answer state and returns the wizard
// to stepTaskTitle for the next iteration of the project loop. The committed
// project + task list are preserved for the persistent column.
func (m *Model) resetForNextLoopTask() {
	m.ans.taskTitle = ""
	m.ans.taskDesc = ""
	m.ans.under2Min = nil
	m.ans.doer = nil
	m.ans.assignee = ""
	m.ans.attachProject = nil
	m.step = stepTaskTitle
	m.input.Reset()
	m.textarea.Reset()
}

// --- rendering -------------------------------------------------------------

func (m Model) View() string {
	var b strings.Builder

	b.WriteString(headerStyle.Render(m.item.Title))
	b.WriteByte('\n')
	if m.item.Description != "" {
		b.WriteString(m.wrappedDescription(m.item.Description))
		b.WriteByte('\n')
	}
	b.WriteByte('\n')

	for _, line := range m.answeredLines() {
		b.WriteString(line)
		b.WriteByte('\n')
	}

	b.WriteString(m.activeView())

	if m.err != nil {
		b.WriteString("\n\n")
		b.WriteString(errStyle.Render(m.err.Error()))
		b.WriteString("\n")
		b.WriteString(hintStyle.Render("press esc to dismiss"))
	}
	return b.String()
}

func (m Model) wrappedDescription(s string) string {
	style := itemDescStyle
	if m.width > 0 {
		style = style.Width(m.width)
	}
	return style.Render(s)
}

func (m Model) answeredLines() []string {
	var lines []string
	if m.ans.actionable != nil {
		lines = append(lines, m.qa("Actionable", yn(*m.ans.actionable)))
	}
	if m.ans.nonActChoice != nil {
		label := "Trash"
		if *m.ans.nonActChoice == choiceSomeday {
			label = "Someday"
		}
		lines = append(lines, m.qa("Trash or Someday", label))
	}
	if m.step > stepSomedayTitle && m.ans.somedayTitle != "" {
		lines = append(lines, m.qa("Someday title", m.ans.somedayTitle))
	}
	if m.step > stepSomedayDescription && m.ans.somedayDesc != "" {
		lines = append(lines, m.qa("Someday description", m.ans.somedayDesc))
	}
	if m.ans.multiStep != nil {
		lines = append(lines, m.qa("Multi-step", yn(*m.ans.multiStep)))
	}
	if m.step > stepProjectTitle && m.ans.projectTitle != "" {
		lines = append(lines, m.qa("Project title", m.ans.projectTitle))
	}
	if m.step > stepProjectOutcome && m.ans.projectOut != "" {
		lines = append(lines, m.qa("Outcome", m.ans.projectOut))
	}
	if m.step > stepProjectDescription && m.ans.projectDesc != "" {
		lines = append(lines, m.qa("Project description", m.ans.projectDesc))
	}
	// Committed tasks (project loop): render each with status.
	for i, t := range m.committedTasks {
		label := fmt.Sprintf("Task #%d", i+1)
		status := "open"
		if t.Status == gtd.TaskStatusDone {
			status = "done"
		}
		valueStyle := answerStyle
		if t.Status == gtd.TaskStatusDone {
			valueStyle = doneStyle
		}
		lines = append(lines, m.joinRow(questionStyle.Render(label), valueStyle.Render(fmt.Sprintf("%s [%s]", t.Title, status))))
	}
	if m.step > stepTaskTitle && m.ans.taskTitle != "" {
		lines = append(lines, m.qa("Next action", m.ans.taskTitle))
	}
	if m.step > stepTaskDescription && m.ans.taskDesc != "" {
		lines = append(lines, m.qa("Description", m.ans.taskDesc))
	}
	if m.ans.under2Min != nil {
		lines = append(lines, m.qa("< 2 minutes", yn(*m.ans.under2Min)))
	}
	if m.ans.doer != nil {
		label := "Me"
		if *m.ans.doer == doerSomeoneElse {
			label = "Someone else"
		}
		lines = append(lines, m.qa("Who's doing it", label))
	}
	if m.step > stepAssignee && m.ans.assignee != "" {
		lines = append(lines, m.qa("Assignee", m.ans.assignee))
	}
	if m.ans.attachProject != nil {
		lines = append(lines, m.qa("Attach to a project", yn(*m.ans.attachProject)))
	}
	return lines
}

func (m Model) activeView() string {
	switch m.step {
	case stepActionable:
		return m.activePrompt("Actionable", "(y/n)")
	case stepNonActionable:
		return m.activePrompt("Trash or Someday", "(t/s)")
	case stepDiscardConfirm:
		return m.activePrompt("Really discard?", "(y/n)")
	case stepSomedayTitle:
		return m.activeInput("Someday title", m.input.View())
	case stepSomedayDescription:
		return m.activeInput("Someday description", m.textarea.View())
	case stepMultiStep:
		return m.activePrompt("Multi-step", "(y/n)")
	case stepProjectTitle:
		return m.activeInput("Project title", m.input.View())
	case stepProjectOutcome:
		return m.activeInput("Outcome", m.input.View())
	case stepProjectDescription:
		return m.activeInput("Project description", m.textarea.View())
	case stepTaskTitle:
		return m.activeInput("Next action", m.input.View())
	case stepTaskDescription:
		return m.activeInput("Description", m.textarea.View())
	case stepUnder2Min:
		return m.activePrompt("< 2 minutes", "(y/n)")
	case stepDoer:
		return m.activePrompt("Who's doing it", "(m=me, s=someone else)")
	case stepAssignee:
		return m.activeInput("Assignee", m.input.View())
	case stepAttachProject:
		return m.activePrompt("Attach to a project", "(y/n)")
	case stepSaving:
		return hintStyle.Render("Saving…")
	case stepDoItNow:
		return hintStyle.Render("Waiting for do-it-now confirmation…")
	}
	return ""
}

func (m Model) labelWidth() int { return questionColWidth + 1 }

func (m Model) answerWidth() int {
	if m.width <= 0 {
		return 0
	}
	return max(m.width-m.labelWidth(), 10)
}

func (m Model) qa(question, answer string) string {
	return m.joinRow(questionStyle.Render(question), answerStyle.Render(answer))
}

func (m Model) activePrompt(question, hint string) string {
	return m.joinRow(activeStyle.Render(question), hintStyle.Render(hint))
}

func (m Model) activeInput(question, input string) string {
	return m.joinRow(activeStyle.Render(question), input)
}

func (m Model) joinRow(label, content string) string {
	labelCell := lipgloss.NewStyle().Width(m.labelWidth()).Render(label + ":")
	contentCell := content
	if w := m.answerWidth(); w > 0 {
		contentCell = lipgloss.NewStyle().Width(w).Render(content)
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, labelCell, contentCell)
}

func yn(v bool) string {
	if v {
		return "Yes"
	}
	return "No"
}

// --- Screen interface -----------------------------------------------------

func (m Model) CapturingInput() bool {
	switch m.step {
	case stepSomedayTitle, stepSomedayDescription,
		stepProjectTitle, stepProjectOutcome, stepProjectDescription,
		stepTaskTitle, stepTaskDescription, stepAssignee:
		return true
	}
	return false
}

func (m Model) ShortHelp() []key.Binding {
	switch m.step {
	case stepSomedayTitle, stepSomedayDescription,
		stepProjectTitle, stepProjectOutcome, stepProjectDescription,
		stepTaskTitle, stepTaskDescription, stepAssignee:
		return []key.Binding{m.KeyMap.Confirm, m.KeyMap.Back}
	case stepDoer:
		return []key.Binding{m.KeyMap.Me, m.KeyMap.Someone, m.KeyMap.Back}
	case stepNonActionable:
		return []key.Binding{m.KeyMap.Trash, m.KeyMap.Someday, m.KeyMap.Back}
	default:
		return []key.Binding{m.KeyMap.Yes, m.KeyMap.No, m.KeyMap.Back}
	}
}

func (m Model) FullHelp() [][]key.Binding {
	return [][]key.Binding{m.ShortHelp()}
}