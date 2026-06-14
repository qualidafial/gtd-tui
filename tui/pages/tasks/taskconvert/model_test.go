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
	"github.com/qualidafial/gtd-tui/tui/internal/keymap"
)

func setup(t *testing.T) (gtd.TaskService, gtd.ProjectService) {
	t.Helper()
	db, err := sqlite.Open(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return service.NewTaskService(db), service.NewProjectService(db)
}

// stubScreen stands in for the project view the wizard lands on after a
// successful convert, carrying the project it was constructed with so tests
// can assert the factory received the freshly created project.
type stubScreen struct{ project gtd.Project }

func (s stubScreen) Init() tea.Cmd                           { return nil }
func (s stubScreen) Update(tea.Msg) (screen.Screen, tea.Cmd) { return s, nil }
func (s stubScreen) View() string                            { return "" }
func (s stubScreen) Keys() []keymap.Group                    { return nil }

func TestWizard_PrePopulatesFromTask(t *testing.T) {
	task := gtd.Task{ID: 1, Title: "Plan offsite", Description: "Q3 team", Status: gtd.TaskStatusOpen}
	m := New(task, nil, nil)

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

	var landed gtd.Project
	viewFn := func(p gtd.Project) screen.Screen {
		landed = p
		return stubScreen{project: p}
	}
	m := New(task, projSvc, viewFn)
	var s screen.Screen = screentest.Init(t, m)

	// Submit commits via ConvertTaskToProject with the collected (pre-populated)
	// values. Sending SubmittedMsg directly exercises the commit path.
	s, cmd := s.Update(form.SubmittedMsg{})
	require.NotNil(t, cmd)
	converted := cmd() // runs convertCmd, persisting the conversion

	// The task now belongs to a new open project seeded from the task.
	got, err := taskSvc.GetTask(ctx, task.ID)
	require.NoError(t, err)
	require.NotNil(t, got.ProjectID)

	project, err := projSvc.GetProject(ctx, *got.ProjectID)
	require.NoError(t, err)
	assert.Equal(t, "Plan offsite", project.Title)
	assert.Equal(t, gtd.ProjectStatusOpen, project.Status)

	// On success the wizard lands on the new project's view (via the factory),
	// not a dismiss.
	next, _ := s.Update(converted)
	stub, ok := next.(stubScreen)
	require.True(t, ok, "expected to land on the project view stub, got %T", next)
	assert.Equal(t, *got.ProjectID, stub.project.ID, "factory built the view for the created project")
	assert.Equal(t, *got.ProjectID, landed.ID)
}

func TestWizard_AbandonLeavesTaskStandalone(t *testing.T) {
	taskSvc, projSvc := setup(t)
	ctx := t.Context()

	task, err := taskSvc.CreateTask(ctx, gtd.Task{Title: "Plan offsite", Status: gtd.TaskStatusOpen})
	require.NoError(t, err)

	var m screen.Screen = New(task, projSvc, nil)
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
	m := New(task, nil, nil)
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
