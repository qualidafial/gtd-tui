package taskconvert

import (
	"errors"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qualidafial/gtd-tui"
	"github.com/qualidafial/gtd-tui/service"
	"github.com/qualidafial/gtd-tui/sqlite"
	"github.com/qualidafial/gtd-tui/tui/components/form"
	"github.com/qualidafial/gtd-tui/tui/components/screen"
	"github.com/qualidafial/gtd-tui/tui/components/screen/screentest"
)

func setup(t *testing.T) (gtd.TaskService, gtd.ProjectService) {
	t.Helper()
	db, err := sqlite.Open(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return service.NewTaskService(db), service.NewProjectService(db)
}

func TestWizard_PrePopulatesFromTask(t *testing.T) {
	task := gtd.Task{ID: 1, Title: "Plan offsite", Description: "Q3 team", Status: gtd.TaskStatusOpen}
	m := New(task, nil)

	vals := m.form.FieldValues()
	assert.Equal(t, "Plan offsite", vals["project_title"], "project title seeds from the task")
	assert.Equal(t, "Q3 team", vals["project_description"])
	assert.Equal(t, "Plan offsite", vals["task_title"], "reframed task title seeds from the task")
	assert.Equal(t, "Q3 team", vals["task_description"])
	assert.Equal(t, "", vals["outcome"], "outcome starts empty")
}

func TestWizard_CommitConvertsTaskToProject(t *testing.T) {
	taskSvc, projSvc := setup(t)
	ctx := t.Context()

	task, err := taskSvc.CreateTask(ctx, gtd.Task{Title: "Plan offsite", Description: "Q3 team", Status: gtd.TaskStatusOpen})
	require.NoError(t, err)

	m := New(task, projSvc)
	var s screen.Screen = screentest.Init(t, m)

	// Submit commits via ConvertTaskToProject with the collected (pre-populated)
	// values. Sending SubmittedMsg directly exercises the commit path.
	_, cmd := s.Update(form.SubmittedMsg{})
	require.NotNil(t, cmd)
	cmd() // runs convertCmd, persisting the conversion

	// The task now belongs to a new open project seeded from the task.
	got, err := taskSvc.GetTask(ctx, task.ID)
	require.NoError(t, err)
	require.NotNil(t, got.ProjectID)

	project, err := projSvc.GetProject(ctx, *got.ProjectID)
	require.NoError(t, err)
	assert.Equal(t, "Plan offsite", project.Title)
	assert.Equal(t, gtd.ProjectStatusOpen, project.Status)
}

func TestWizard_AbandonLeavesTaskStandalone(t *testing.T) {
	taskSvc, projSvc := setup(t)
	ctx := t.Context()

	task, err := taskSvc.CreateTask(ctx, gtd.Task{Title: "Plan offsite", Status: gtd.TaskStatusOpen})
	require.NoError(t, err)

	var m screen.Screen = New(task, projSvc)
	m = screentest.Init(t, m)

	_, dismissed := screentest.RunUntilDismiss(t, m, tea.KeyPressMsg{Code: tea.KeyEscape})
	require.True(t, dismissed, "esc should dismiss")

	// No project was created and the task remains standalone.
	got, err := taskSvc.GetTask(ctx, task.ID)
	require.NoError(t, err)
	assert.Nil(t, got.ProjectID)

	projects, err := projSvc.ListProjects(ctx, gtd.ProjectFilter{})
	require.NoError(t, err)
	assert.Empty(t, projects)
}

func TestWizard_CommitErrorDisplayed(t *testing.T) {
	task := gtd.Task{ID: 1, Title: "t", Status: gtd.TaskStatusOpen}
	m := New(task, nil)
	m.saving = true

	next, cmd := m.Update(convertedMsg{err: errors.New("disk full")})
	require.NotNil(t, cmd, "expected error cmd on commit failure")
	msg := cmd()
	err, ok := msg.(error)
	require.True(t, ok, "expected error msg, got %T", msg)
	assert.Contains(t, err.Error(), "disk full")

	nm, ok := next.(Model)
	require.True(t, ok)
	require.Error(t, nm.err, "model retains the error for the standoff")
	assert.False(t, nm.CapturingInput(), "input is suppressed while the error is shown")
}
