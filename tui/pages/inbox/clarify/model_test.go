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

	var s screen.Screen = clarify.New(item, e.inboxSvc, e.taskSvc)
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

// TestSingleTask_DoItNow_PushesDoitnow: flip <2min radio to Yes, then
// ctrl+s. The clarify path commits the task and pushes the doitnow overlay.
func TestSingleTask_DoItNow_PushesDoitnow(t *testing.T) {
	e := setup(t)
	ctx := t.Context()

	item, err := e.inboxSvc.Create(ctx, gtd.Item{Title: "Email Sam"})
	require.NoError(t, err)

	var s screen.Screen = clarify.New(item, e.inboxSvc, e.taskSvc)
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

	var sawPush bool
	for st, msg := range screentest.PumpSend(t, s, ctrlS()) {
		s = st
		if _, ok := msg.(screen.PushMsg); ok {
			sawPush = true
			break
		}
	}
	require.True(t, sawPush, "do-it-now path should push the doitnow overlay")

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

	var s screen.Screen = clarify.New(item, e.inboxSvc, e.taskSvc)
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

	var s screen.Screen = clarify.New(item, e.inboxSvc, e.taskSvc)
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

	var s screen.Screen = clarify.New(item, e.inboxSvc, e.taskSvc)
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

// TestEsc_AtAnyTime_Dismisses: Esc is the wizard-level cancel; it dismisses
// regardless of form position.
func TestEsc_AtAnyTime_Dismisses(t *testing.T) {
	e := setup(t)
	ctx := t.Context()

	item, err := e.inboxSvc.Create(ctx, gtd.Item{Title: "x"})
	require.NoError(t, err)

	var s screen.Screen = clarify.New(item, e.inboxSvc, e.taskSvc)
	s = screentest.Init(t, s)

	_, dismissed := screentest.RunUntilDismiss(t, s, tea.KeyPressMsg{Code: tea.KeyEscape})
	require.True(t, dismissed)
}

// Compile-time guard: sendKey is exported only for the table tests below;
// we silence the unused-warning if they are removed.
var _ = sendKey
