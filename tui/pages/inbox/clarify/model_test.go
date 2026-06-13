package clarify_test

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qualidafial/gtd-tui"
	"github.com/qualidafial/gtd-tui/service"
	"github.com/qualidafial/gtd-tui/sqlite"
	"github.com/qualidafial/gtd-tui/tui/components/screen"
	"github.com/qualidafial/gtd-tui/tui/components/screen/screentest"
	"github.com/qualidafial/gtd-tui/tui/pages/inbox/clarify"
	"github.com/qualidafial/gtd-tui/tui/pages/inbox/clarify/doitnow"
)

type env struct {
	inboxSvc gtd.InboxService
	taskSvc  gtd.TaskService
	projSvc  gtd.ProjectService
}

func setup(t *testing.T) env {
	t.Helper()
	db, err := sqlite.Open(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return env{
		inboxSvc: service.NewInboxService(db),
		taskSvc:  service.NewTaskService(db),
		projSvc:  service.NewProjectService(db),
	}
}

func sendKey(t *testing.T, s screen.Screen, code rune) screen.Screen {
	t.Helper()
	return screentest.Send(t, s, tea.KeyPressMsg{Code: code, Text: string(code)})
}

func sendCode(t *testing.T, s screen.Screen, code rune) screen.Screen {
	t.Helper()
	return screentest.Send(t, s, tea.KeyPressMsg{Code: code})
}

func tab(t *testing.T, s screen.Screen) screen.Screen {
	t.Helper()
	return sendCode(t, s, tea.KeyTab)
}

func ctrlS() tea.KeyPressMsg { return tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl} }

// TestSingleTask_NotDoItNow_CommitsAndDismisses drives the happy path with
// every radio at its default ("Yes" actionable, "Single task", "<2min No",
// "Me" doer); the prefilled title flows through and ClarifyAsTask runs.
func TestSingleTask_NotDoItNow_CommitsAndDismisses(t *testing.T) {
	e := setup(t)
	ctx := t.Context()

	item, err := e.inboxSvc.Create(ctx, gtd.Item{Title: "Call dentist", Description: "before friday"})
	require.NoError(t, err)

	var s screen.Screen = clarify.New(item, e.inboxSvc, e.taskSvc, e.projSvc)
	s = screentest.Init(t, s)

	_, dismissed := screentest.RunUntilDismiss(t, s, ctrlS())
	require.True(t, dismissed)

	items, err := e.inboxSvc.List(ctx)
	require.NoError(t, err)
	assert.Empty(t, items)

	tasks, err := e.taskSvc.ListTasks(ctx, gtd.TaskFilter{})
	require.NoError(t, err)
	require.Len(t, tasks, 1)
	assert.Equal(t, "Call dentist", tasks[0].Title)
	assert.Equal(t, "before friday", tasks[0].Description)
	assert.Equal(t, gtd.TaskStatusOpen, tasks[0].Status)
	assert.Nil(t, tasks[0].Assignee)
	assert.Nil(t, tasks[0].ProjectID)
}

// TestSingleTask_DoItNow_ReplacesWithDoitnow: flip <2min radio to Yes, then
// ctrl+s. Single-task do-it-now is terminal, so the wizard commits the task
// and replaces itself with the doitnow overlay (dismissing it exits straight
// to the inbox rather than returning to the spent form).
func TestSingleTask_DoItNow_ReplacesWithDoitnow(t *testing.T) {
	e := setup(t)
	ctx := t.Context()

	item, err := e.inboxSvc.Create(ctx, gtd.Item{Title: "Email Sam"})
	require.NoError(t, err)

	var s screen.Screen = clarify.New(item, e.inboxSvc, e.taskSvc, e.projSvc)
	s = screentest.Init(t, s)

	// Walk from the actionable radio to the under2Min radio. Order in the
	// initial form: actionable → multiStep → taskTitle → taskDesc →
	// under2Min. Tab three times to reach under2Min, then press Right to
	// flip it to Yes.
	s = tab(t, s) // → multiStep
	s = tab(t, s) // → taskTitle
	s = tab(t, s) // → taskDesc
	s = tab(t, s) // → under2Min
	s = sendCode(t, s, tea.KeyRight)

	s = screentest.Send(t, s, ctrlS())
	assert.IsType(t, doitnow.Model{}, s, "single-task do-it-now should replace the wizard with the doitnow overlay")

	tasks, err := e.taskSvc.ListTasks(ctx, gtd.TaskFilter{})
	require.NoError(t, err)
	require.Len(t, tasks, 1)
	assert.Equal(t, gtd.TaskStatusOpen, tasks[0].Status)
}

// TestNonActionable_Trash_Discards: actionable=No, nonAct=Trash (default),
// ctrl+s commits via Discard.
func TestNonActionable_Trash_Discards(t *testing.T) {
	e := setup(t)
	ctx := t.Context()

	item, err := e.inboxSvc.Create(ctx, gtd.Item{Title: "trash this"})
	require.NoError(t, err)

	var s screen.Screen = clarify.New(item, e.inboxSvc, e.taskSvc, e.projSvc)
	s = screentest.Init(t, s)

	// Flip actionable to No.
	s = sendCode(t, s, tea.KeyRight)

	_, dismissed := screentest.RunUntilDismiss(t, s, ctrlS())
	require.True(t, dismissed)

	got, err := e.inboxSvc.Get(ctx, item.ID)
	require.NoError(t, err)
	assert.True(t, got.Discarded)
}

// TestNonActionable_Someday_Incubates: actionable=No, nonAct=Someday,
// somedayTitle prefilled, ctrl+s incubates.
func TestNonActionable_Someday_Incubates(t *testing.T) {
	e := setup(t)
	ctx := t.Context()

	item, err := e.inboxSvc.Create(ctx, gtd.Item{Title: "learn pottery", Description: "later this year"})
	require.NoError(t, err)

	var s screen.Screen = clarify.New(item, e.inboxSvc, e.taskSvc, e.projSvc)
	s = screentest.Init(t, s)

	// Flip actionable to No, then tab to nonAct and flip to Someday.
	s = sendCode(t, s, tea.KeyRight) // actionable → No
	s = tab(t, s)                    // → nonAct
	s = sendCode(t, s, tea.KeyRight) // nonAct → Someday

	_, dismissed := screentest.RunUntilDismiss(t, s, ctrlS())
	require.True(t, dismissed)

	projects, err := e.projSvc.ListProjects(ctx, gtd.ProjectFilter{}.WithStatus(gtd.ProjectStatusSomeday))
	require.NoError(t, err)
	require.Len(t, projects, 1)
	assert.Equal(t, "learn pottery", projects[0].Title)
	assert.Equal(t, "later this year", projects[0].Description)
}

// TestProject_FirstTask_EntersProjectLoop: multiStep=Project triggers
// ClarifyAsProject; after success the wizard transitions to the per-task
// loop form rather than dismissing.
func TestProject_FirstTask_EntersProjectLoop(t *testing.T) {
	e := setup(t)
	ctx := t.Context()

	item, err := e.inboxSvc.Create(ctx, gtd.Item{Title: "Refactor auth", Description: "split modules"})
	require.NoError(t, err)

	var s screen.Screen = clarify.New(item, e.inboxSvc, e.taskSvc, e.projSvc)
	s = screentest.Init(t, s)

	// actionable=Yes (default). Tab to multiStep, flip to Project.
	s = tab(t, s)                    // → multiStep
	s = sendCode(t, s, tea.KeyRight) // → Project
	// projectTitle is now visible and prefilled. Tab to projectOutcome
	// and type one.
	s = tab(t, s) // → projectTitle
	s = tab(t, s) // → projectOutcome
	s = screentest.TypeText(t, s, "auth is modular")

	// Type a first-task title (overwrite the item-title prefill) by
	// tabbing forward to taskTitle.
	s = tab(t, s) // → projectDesc
	s = tab(t, s) // → taskTitle
	// Clear the prefilled title.
	for range 16 {
		s = sendCode(t, s, tea.KeyBackspace)
	}
	s = screentest.TypeText(t, s, "Sketch design")

	// Run ctrl+s; do NOT exit on dismiss — the wizard transitions to the
	// project loop and stays on screen.
	for st := range screentest.PumpSend(t, s, ctrlS()) {
		s = st
	}

	projects, err := e.projSvc.ListProjects(ctx, gtd.ProjectFilter{}.WithStatus(gtd.ProjectStatusOpen))
	require.NoError(t, err)
	require.Len(t, projects, 1)
	assert.Equal(t, "Refactor auth", projects[0].Title)
	assert.Equal(t, "auth is modular", projects[0].Outcome)

	tasks, err := e.taskSvc.ListTasks(ctx, gtd.TaskFilter{})
	require.NoError(t, err)
	require.Len(t, tasks, 1)
	assert.Equal(t, "Sketch design", tasks[0].Title)
	require.NotNil(t, tasks[0].ProjectID)
	assert.Equal(t, projects[0].ID, *tasks[0].ProjectID)
}

// TestProject_DoItNow_LoopsAfterResult: the project do-it-now branch commits
// the first task, pushes the doitnow overlay, and — once doitnow returns its
// ResultMsg — rebuilds the per-task loop form. This guards the regression
// where the wizard stayed in its saving state and dropped the ResultMsg,
// leaving the form frozen until Esc.
// driveProjectToDoItNow clarifies item as a project whose first task is flagged
// do-it-now and commits, leaving the wizard on the pushed doitnow overlay. It
// returns the settled wizard and the committed tasks.
func driveProjectToDoItNow(t *testing.T, e env, item gtd.Item) (screen.Screen, []gtd.Task) {
	t.Helper()

	var s screen.Screen = clarify.New(item, e.inboxSvc, e.taskSvc, e.projSvc)
	s = screentest.Init(t, s)

	// actionable=Yes (default). Walk to Project, set an outcome, flip the
	// first task's <2min radio to Yes (do it now).
	s = tab(t, s)                    // → multiStep
	s = sendCode(t, s, tea.KeyRight) // → Project
	s = tab(t, s)                    // → projectTitle (prefilled)
	s = tab(t, s)                    // → projectOutcome
	s = screentest.TypeText(t, s, "auth is modular")
	s = tab(t, s)                    // → projectDesc
	s = tab(t, s)                    // → taskTitle (prefilled)
	s = tab(t, s)                    // → taskDesc
	s = tab(t, s)                    // → under2Min
	s = sendCode(t, s, tea.KeyRight) // → Yes (do it now)

	// Commit. The wizard pushes doitnow (no app stack here, so the push is
	// just a PushMsg in the stream); afterwards it must be out of its saving
	// state and ready for the ResultMsg.
	var sawPush bool
	for st, msg := range screentest.PumpSend(t, s, ctrlS()) {
		s = st
		if _, ok := msg.(screen.PushMsg); ok {
			sawPush = true
		}
	}
	require.True(t, sawPush, "project do-it-now should push the doitnow overlay")

	tasks, err := e.taskSvc.ListTasks(t.Context(), gtd.TaskFilter{})
	require.NoError(t, err)
	return s, tasks
}

// TestProject_DoItNow_CompletedLoops: completing the task at the do-it-now
// prompt (enter) loops back to a fresh per-task form.
func TestProject_DoItNow_CompletedLoops(t *testing.T) {
	e := setup(t)
	item, err := e.inboxSvc.Create(t.Context(), gtd.Item{Title: "Refactor auth"})
	require.NoError(t, err)

	s, tasks := driveProjectToDoItNow(t, e, item)
	require.Len(t, tasks, 1)

	// doitnow resolves with the task completed.
	s = screentest.Send(t, s, doitnow.ResultMsg{TaskID: tasks[0].ID, Completed: true})
	// The rebuilt loop form needs a size before it renders its fields.
	s = screentest.Send(t, s, tea.WindowSizeMsg{Width: 80, Height: 24})

	view := s.View()
	assert.Contains(t, view, "Next task title", "wizard should rebuild the loop form after doitnow")
	assert.Contains(t, view, "(done)", "completed first task should render as done")
}

// TestProject_DoItNow_LeftOpenDismisses: leaving the task open at the do-it-now
// prompt (esc) exits the whole wizard back to the inbox instead of looping.
func TestProject_DoItNow_LeftOpenDismisses(t *testing.T) {
	e := setup(t)
	item, err := e.inboxSvc.Create(t.Context(), gtd.Item{Title: "Refactor auth"})
	require.NoError(t, err)

	s, tasks := driveProjectToDoItNow(t, e, item)
	require.Len(t, tasks, 1)

	// doitnow resolves with the task left open: the wizard dismisses.
	var dismissed bool
	for st, msg := range screentest.PumpSend(t, s, doitnow.ResultMsg{TaskID: tasks[0].ID, Completed: false}) {
		s = st
		if _, ok := msg.(screen.DismissMsg); ok {
			dismissed = true
		}
	}
	assert.True(t, dismissed, "leaving the task open should dismiss the wizard")
}

// TestEsc_AtAnyTime_Dismisses: Esc is the wizard-level cancel; it dismisses
// regardless of form position.
func TestEsc_AtAnyTime_Dismisses(t *testing.T) {
	e := setup(t)
	ctx := t.Context()

	item, err := e.inboxSvc.Create(ctx, gtd.Item{Title: "x"})
	require.NoError(t, err)

	var s screen.Screen = clarify.New(item, e.inboxSvc, e.taskSvc, e.projSvc)
	s = screentest.Init(t, s)

	_, dismissed := screentest.RunUntilDismiss(t, s, tea.KeyPressMsg{Code: tea.KeyEscape})
	require.True(t, dismissed)
}

// TestSingleTask_AttachExistingProject: in the single-task branch the wizard
// offers a Project select of open projects; choosing one sets the created
// task's ProjectID.
func TestSingleTask_AttachExistingProject(t *testing.T) {
	e := setup(t)
	ctx := t.Context()

	proj, err := e.projSvc.CreateProject(ctx, gtd.Project{Title: "Website", Status: gtd.ProjectStatusOpen})
	require.NoError(t, err)

	item, err := e.inboxSvc.Create(ctx, gtd.Item{Title: "Buy domain"})
	require.NoError(t, err)

	var s screen.Screen = clarify.New(item, e.inboxSvc, e.taskSvc, e.projSvc)
	s = screentest.Init(t, s)
	s = screentest.Send(t, s, tea.WindowSizeMsg{Width: 80, Height: 100})

	// The Project select is shown because an open project exists; its default
	// row is "(none)". (Counterpart to TestSingleTask_NoProjects_FieldHidden.)
	assert.Contains(t, s.View(), "(none)", "Project select should be shown when an open project exists")

	// Single-task defaults. Tab to the Project select:
	// actionable → multiStep → taskTitle → taskDesc → under2Min → doer → project.
	for range 6 {
		s = tab(t, s)
	}
	// Move from "(none)" to the first real project.
	s = sendCode(t, s, tea.KeyDown)

	_, dismissed := screentest.RunUntilDismiss(t, s, ctrlS())
	require.True(t, dismissed)

	tasks, err := e.taskSvc.ListTasks(ctx, gtd.TaskFilter{})
	require.NoError(t, err)
	require.Len(t, tasks, 1)
	require.NotNil(t, tasks[0].ProjectID)
	assert.Equal(t, proj.ID, *tasks[0].ProjectID)
}

// TestSingleTask_NoneYieldsStandalone: leaving the Project select on its
// "(none)" default produces a standalone task even when open projects exist.
func TestSingleTask_NoneYieldsStandalone(t *testing.T) {
	e := setup(t)
	ctx := t.Context()

	_, err := e.projSvc.CreateProject(ctx, gtd.Project{Title: "Website", Status: gtd.ProjectStatusOpen})
	require.NoError(t, err)

	item, err := e.inboxSvc.Create(ctx, gtd.Item{Title: "Standalone errand"})
	require.NoError(t, err)

	var s screen.Screen = clarify.New(item, e.inboxSvc, e.taskSvc, e.projSvc)
	s = screentest.Init(t, s)

	// Commit without touching the Project select.
	_, dismissed := screentest.RunUntilDismiss(t, s, ctrlS())
	require.True(t, dismissed)

	tasks, err := e.taskSvc.ListTasks(ctx, gtd.TaskFilter{})
	require.NoError(t, err)
	require.Len(t, tasks, 1)
	assert.Nil(t, tasks[0].ProjectID)
}

// TestSingleTask_DoItNow_CarriesProject: a sub-2-minute do-it-now single task
// still carries the chosen ProjectID into the committed (open) task before the
// doitnow overlay takes over.
func TestSingleTask_DoItNow_CarriesProject(t *testing.T) {
	e := setup(t)
	ctx := t.Context()

	proj, err := e.projSvc.CreateProject(ctx, gtd.Project{Title: "Website", Status: gtd.ProjectStatusOpen})
	require.NoError(t, err)

	item, err := e.inboxSvc.Create(ctx, gtd.Item{Title: "Ping host"})
	require.NoError(t, err)

	var s screen.Screen = clarify.New(item, e.inboxSvc, e.taskSvc, e.projSvc)
	s = screentest.Init(t, s)

	s = tab(t, s)                    // → multiStep
	s = tab(t, s)                    // → taskTitle
	s = tab(t, s)                    // → taskDesc
	s = tab(t, s)                    // → under2Min
	s = sendCode(t, s, tea.KeyRight) // → Yes (do it now); doer hides
	s = tab(t, s)                    // → project
	s = sendCode(t, s, tea.KeyDown)  // select the project

	s = screentest.Send(t, s, ctrlS())
	assert.IsType(t, doitnow.Model{}, s)

	tasks, err := e.taskSvc.ListTasks(ctx, gtd.TaskFilter{})
	require.NoError(t, err)
	require.Len(t, tasks, 1)
	assert.Equal(t, gtd.TaskStatusOpen, tasks[0].Status)
	require.NotNil(t, tasks[0].ProjectID)
	assert.Equal(t, proj.ID, *tasks[0].ProjectID)
}

// TestSingleTask_NoProjects_FieldHidden: with no open projects the Project
// select is not shown at all, and the committed task is standalone.
func TestSingleTask_NoProjects_FieldHidden(t *testing.T) {
	e := setup(t)
	ctx := t.Context()

	item, err := e.inboxSvc.Create(ctx, gtd.Item{Title: "Solo task"})
	require.NoError(t, err)

	var s screen.Screen = clarify.New(item, e.inboxSvc, e.taskSvc, e.projSvc)
	s = screentest.Init(t, s)

	// Render the full single-task form with a tall window so nothing clips.
	// The Project select's only row would be "(none)"; its absence proves the
	// field is not shown at all when there are no open projects.
	s = screentest.Send(t, s, tea.WindowSizeMsg{Width: 80, Height: 100})
	assert.NotContains(t, s.View(), "(none)", "Project select should be hidden when there are no open projects")

	_, dismissed := screentest.RunUntilDismiss(t, s, ctrlS())
	require.True(t, dismissed)

	tasks, err := e.taskSvc.ListTasks(ctx, gtd.TaskFilter{})
	require.NoError(t, err)
	require.Len(t, tasks, 1)
	assert.Nil(t, tasks[0].ProjectID)
}

// TestProjectLoop_SecondTaskAttachesToNewProject: loop tasks attach to the
// freshly-created project, never to a stray select value — the loop form has
// no Project field. A pre-existing open project must not capture them.
func TestProjectLoop_SecondTaskAttachesToNewProject(t *testing.T) {
	e := setup(t)
	ctx := t.Context()

	other, err := e.projSvc.CreateProject(ctx, gtd.Project{Title: "Unrelated", Status: gtd.ProjectStatusOpen})
	require.NoError(t, err)

	item, err := e.inboxSvc.Create(ctx, gtd.Item{Title: "Refactor auth"})
	require.NoError(t, err)

	var s screen.Screen = clarify.New(item, e.inboxSvc, e.taskSvc, e.projSvc)
	s = screentest.Init(t, s)

	// Project branch with a first (not do-it-now) task → enters the loop.
	s = tab(t, s)                    // → multiStep
	s = sendCode(t, s, tea.KeyRight) // → Project
	s = tab(t, s)                    // → projectTitle (prefilled)
	s = tab(t, s)                    // → projectOutcome
	s = screentest.TypeText(t, s, "auth modular")
	s = tab(t, s) // → projectDesc
	s = tab(t, s) // → taskTitle (prefilled)
	for st := range screentest.PumpSend(t, s, ctrlS()) {
		s = st
	}

	// In the loop form: capture and commit a second task.
	s = screentest.Send(t, s, tea.WindowSizeMsg{Width: 80, Height: 24})
	s = screentest.TypeText(t, s, "Write tests")
	for st := range screentest.PumpSend(t, s, ctrlS()) {
		s = st
	}

	open, err := e.projSvc.ListProjects(ctx, gtd.ProjectFilter{}.WithStatus(gtd.ProjectStatusOpen))
	require.NoError(t, err)
	require.Len(t, open, 2) // "Unrelated" + new "Refactor auth"

	tasks, err := e.taskSvc.ListTasks(ctx, gtd.TaskFilter{})
	require.NoError(t, err)
	require.Len(t, tasks, 2)
	for _, task := range tasks {
		require.NotNil(t, task.ProjectID)
		assert.NotEqual(t, other.ID, *task.ProjectID)
	}
}

// Compile-time guard: sendKey is exported only for the table tests below;
// we silence the unused-warning if they are removed.
var _ = sendKey
