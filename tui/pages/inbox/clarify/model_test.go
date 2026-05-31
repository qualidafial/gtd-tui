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

// pressKey is a small helper: send a single keypress and discard the
// resulting model assignment (so the test reads naturally as a script).
func pressKey(t *testing.T, s screen.Screen, code rune) screen.Screen {
	t.Helper()
	return screentest.Send(t, s, tea.KeyPressMsg{Code: code, Text: string(code)})
}

func pressEnter(t *testing.T, s screen.Screen) screen.Screen {
	t.Helper()
	return screentest.Send(t, s, tea.KeyPressMsg{Code: tea.KeyEnter})
}

func pressEsc(t *testing.T, s screen.Screen) screen.Screen {
	t.Helper()
	return screentest.Send(t, s, tea.KeyPressMsg{Code: tea.KeyEscape})
}


// TestSingleTask_NotDoItNow_CommitsAndDismisses drives the simplest happy path:
// actionable=Yes → multi-step=No → accept inherited title and description →
// <2min=No → me → no project → ClarifyAsTask commits → wizard dismisses.
func TestSingleTask_NotDoItNow_CommitsAndDismisses(t *testing.T) {
	e := setup(t)
	ctx := t.Context()

	item, err := e.inboxSvc.Create(ctx, gtd.Item{Title: "Call dentist", Description: "before friday"})
	require.NoError(t, err)

	var s screen.Screen = clarify.New(item, e.inboxSvc, e.taskSvc)
	s = screentest.Init(t, s)

	s = pressKey(t, s, 'y') // Actionable? Yes
	s = pressKey(t, s, 'n') // Multi-step? No
	s = pressEnter(t, s)    // accept pre-populated title
	s = pressEnter(t, s)    // accept pre-populated description
	s = pressKey(t, s, 'n') // < 2min? No
	s = pressKey(t, s, 'm') // doer = me

	_, dismissed := screentest.RunUntilDismiss(t, s, tea.KeyPressMsg{Code: 'n', Text: "n"})
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

// TestSingleTask_DoItNow_PushesDoitnowAndCompletes covers the do-it-now path:
// <2min=Yes → ClarifyAsTask commits open → doitnow overlay pushes → confirm
// completes the task. The wizard dismisses after the doitnow result.
func TestSingleTask_DoItNow_PushesDoitnowAndCompletes(t *testing.T) {
	e := setup(t)
	ctx := t.Context()

	item, err := e.inboxSvc.Create(ctx, gtd.Item{Title: "Email Sam"})
	require.NoError(t, err)

	var s screen.Screen = clarify.New(item, e.inboxSvc, e.taskSvc)
	s = screentest.Init(t, s)

	s = pressKey(t, s, 'y') // actionable
	s = pressKey(t, s, 'n') // single task
	s = pressEnter(t, s)    // accept title
	s = pressEnter(t, s)    // accept (empty) description
	s = pressKey(t, s, 'y') // <2min Yes (auto-sets doer=me; advances to attach prompt)

	// 'n' for attach=No commits the task and pushes the doitnow overlay.
	var sawPush bool
	for st, msg := range screentest.PumpSend(t, s, tea.KeyPressMsg{Code: 'n', Text: "n"}) {
		s = st
		if _, ok := msg.(screen.PushMsg); ok {
			sawPush = true
			break
		}
	}
	require.True(t, sawPush, "do-it-now path should push the doitnow overlay")

	// Task exists, still open.
	tasks, err := e.taskSvc.ListTasks(ctx, gtd.TaskFilter{})
	require.NoError(t, err)
	require.Len(t, tasks, 1)
	assert.Equal(t, gtd.TaskStatusOpen, tasks[0].Status)
}

// TestNonActionable_TrashConfirm_Discards covers actionable=No → trash → confirm
// → svc.Discard is called and the wizard dismisses.
func TestNonActionable_TrashConfirm_Discards(t *testing.T) {
	e := setup(t)
	ctx := t.Context()

	item, err := e.inboxSvc.Create(ctx, gtd.Item{Title: "trash this"})
	require.NoError(t, err)

	var s screen.Screen = clarify.New(item, e.inboxSvc, e.taskSvc)
	s = screentest.Init(t, s)

	s = pressKey(t, s, 'n') // not actionable
	s = pressKey(t, s, 't') // trash

	_, dismissed := screentest.RunUntilDismiss(t, s, tea.KeyPressMsg{Code: 'y', Text: "y"}) // confirm
	require.True(t, dismissed)

	got, err := e.inboxSvc.Get(ctx, item.ID)
	require.NoError(t, err)
	assert.True(t, got.Discarded, "discard should have stamped the item")

	items, err := e.inboxSvc.List(ctx)
	require.NoError(t, err)
	assert.Empty(t, items)
}

// TestNonActionable_TrashConfirm_NoReturnsToChoice covers the "no" path of the
// discard confirmation: the wizard backs out to the trash/someday question
// without persisting anything.
func TestNonActionable_TrashConfirm_NoReturnsToChoice(t *testing.T) {
	e := setup(t)
	ctx := t.Context()

	item, err := e.inboxSvc.Create(ctx, gtd.Item{Title: "maybe trash"})
	require.NoError(t, err)

	var s screen.Screen = clarify.New(item, e.inboxSvc, e.taskSvc)
	s = screentest.Init(t, s)

	s = pressKey(t, s, 'n')   // not actionable
	s = pressKey(t, s, 't')   // trash
	s = pressKey(t, s, 'n')   // confirm? no — backs out
	s = pressKey(t, s, 's')   // pick someday instead
	s = pressEnter(t, s)      // accept default someday title

	_, dismissed := screentest.RunUntilDismiss(t, s, tea.KeyPressMsg{Code: tea.KeyEnter}) // accept empty desc
	require.True(t, dismissed)

	// Item not discarded; clarified into a someday project instead.
	got, err := e.inboxSvc.Get(ctx, item.ID)
	require.NoError(t, err)
	assert.False(t, got.Discarded)
	require.NotNil(t, got.ClarifiedIntoProjectID)

	p, err := e.projSvc.GetProject(ctx, *got.ClarifiedIntoProjectID)
	require.NoError(t, err)
	assert.Equal(t, gtd.ProjectStatusSomeday, p.Status)
	assert.Equal(t, "maybe trash", p.Title)
}

// TestNonActionable_Someday_Incubates covers actionable=No → someday → fill
// title/desc → svc.Incubate creates a someday project.
func TestNonActionable_Someday_Incubates(t *testing.T) {
	e := setup(t)
	ctx := t.Context()

	item, err := e.inboxSvc.Create(ctx, gtd.Item{Title: "learn pottery", Description: "later this year"})
	require.NoError(t, err)

	var s screen.Screen = clarify.New(item, e.inboxSvc, e.taskSvc)
	s = screentest.Init(t, s)

	s = pressKey(t, s, 'n') // not actionable
	s = pressKey(t, s, 's') // someday
	s = pressEnter(t, s)    // accept default title

	_, dismissed := screentest.RunUntilDismiss(t, s, tea.KeyPressMsg{Code: tea.KeyEnter}) // accept default desc
	require.True(t, dismissed)

	projects, err := e.projSvc.ListProjects(ctx, gtd.ProjectFilter{}.WithStatus(gtd.ProjectStatusSomeday))
	require.NoError(t, err)
	require.Len(t, projects, 1)
	assert.Equal(t, "learn pottery", projects[0].Title)
	assert.Equal(t, "later this year", projects[0].Description)
}

// TestProject_FirstTaskNotDoItNow_CommitsAndDismisses covers the project
// branch: multi-step=Yes → project form → first task per-task block → <2min=No
// → ClarifyAsProject commits project + first task open → wizard dismisses.
func TestProject_FirstTaskNotDoItNow_CommitsAndDismisses(t *testing.T) {
	e := setup(t)
	ctx := t.Context()

	item, err := e.inboxSvc.Create(ctx, gtd.Item{Title: "Refactor auth", Description: "split modules"})
	require.NoError(t, err)

	var s screen.Screen = clarify.New(item, e.inboxSvc, e.taskSvc)
	s = screentest.Init(t, s)

	s = pressKey(t, s, 'y') // actionable
	s = pressKey(t, s, 'y') // multi-step
	s = pressEnter(t, s)    // accept project title
	s = screentest.TypeText(t, s, "auth is modular")
	s = pressEnter(t, s)    // outcome
	s = pressEnter(t, s)    // accept project description
	s = screentest.TypeText(t, s, "Sketch design")
	s = pressEnter(t, s)    // task title
	s = pressEnter(t, s)    // accept (blank) task description
	s = pressKey(t, s, 'n') // <2min? No

	_, dismissed := screentest.RunUntilDismiss(t, s, tea.KeyPressMsg{Code: 'm', Text: "m"}) // me
	require.True(t, dismissed)

	projects, err := e.projSvc.ListProjects(ctx, gtd.ProjectFilter{}.WithStatus(gtd.ProjectStatusOpen))
	require.NoError(t, err)
	require.Len(t, projects, 1)
	assert.Equal(t, "Refactor auth", projects[0].Title)
	assert.Equal(t, "auth is modular", projects[0].Outcome)

	tasks, err := e.taskSvc.ListTasks(ctx, gtd.TaskFilter{})
	require.NoError(t, err)
	require.Len(t, tasks, 1)
	assert.Equal(t, "Sketch design", tasks[0].Title)
	assert.Equal(t, gtd.TaskStatusOpen, tasks[0].Status)
	require.NotNil(t, tasks[0].ProjectID)
	assert.Equal(t, projects[0].ID, *tasks[0].ProjectID)
}

// TestBackNav_FromMultiStep_ClearsAnswer ensures Esc on a non-root step pops
// back, clearing the previous answer for re-asking.
func TestBackNav_FromMultiStep_ClearsAnswer(t *testing.T) {
	e := setup(t)
	ctx := t.Context()

	item, err := e.inboxSvc.Create(ctx, gtd.Item{Title: "x"})
	require.NoError(t, err)

	var s screen.Screen = clarify.New(item, e.inboxSvc, e.taskSvc)
	s = screentest.Init(t, s)

	s = pressKey(t, s, 'y') // actionable Yes → stepMultiStep
	s = pressEsc(t, s)      // back to stepActionable; clears actionable answer

	view := s.View()
	assert.Contains(t, view, "Actionable", "should be back at the actionable prompt")
	// answered column shouldn't show the cleared answer any more
	assert.NotContains(t, view, "Yes\n", "actionable=Yes should have been cleared")
}

// TestBackNav_AtRoot_Dismisses ensures Esc on the root question dismisses
// instead of looping forever.
func TestBackNav_AtRoot_Dismisses(t *testing.T) {
	e := setup(t)
	ctx := t.Context()

	item, err := e.inboxSvc.Create(ctx, gtd.Item{Title: "x"})
	require.NoError(t, err)

	var s screen.Screen = clarify.New(item, e.inboxSvc, e.taskSvc)
	s = screentest.Init(t, s)

	_, dismissed := screentest.RunUntilDismiss(t, s, tea.KeyPressMsg{Code: tea.KeyEscape})
	require.True(t, dismissed)
}
