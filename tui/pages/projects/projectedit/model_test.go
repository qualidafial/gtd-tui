package projectedit

import (
	"errors"
	"testing"
	"time"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qualidafial/gtd-tui"
	"github.com/qualidafial/gtd-tui/service"
	"github.com/qualidafial/gtd-tui/sqlite"
	"github.com/qualidafial/gtd-tui/tui/components/screen"
	"github.com/qualidafial/gtd-tui/tui/components/screen/screentest"
)

func openTestDB(t *testing.T) *sqlite.DB {
	t.Helper()
	db, err := sqlite.Open(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return db
}

// stubScreen is a no-op Screen used as the result of a test view factory.
type stubScreen struct{}

func (stubScreen) Init() tea.Cmd                           { return nil }
func (stubScreen) Update(tea.Msg) (screen.Screen, tea.Cmd) { return stubScreen{}, nil }
func (stubScreen) View() string                            { return "" }
func (stubScreen) ShortHelp() []key.Binding                { return nil }
func (stubScreen) FullHelp() [][]key.Binding               { return nil }

func TestView_HeaderShown_ExistingProject(t *testing.T) {
	p := gtd.Project{
		ID:              1,
		Title:           "Build shed",
		Status:          gtd.ProjectStatusOpen,
		StatusChangedAt: time.Now().AddDate(0, 0, -3),
		CreatedAt:       time.Now().AddDate(0, 0, -7),
		UpdatedAt:       time.Now().AddDate(0, 0, -1),
	}
	m := New(p, nil, nil)
	view := ansi.Strip(m.View())

	assert.Contains(t, view, "Project ID:")
	assert.Contains(t, view, "1")
	assert.Contains(t, view, "Status:")
	assert.Contains(t, view, "Open (3d)")
	assert.Contains(t, view, "Created:")
	assert.Contains(t, view, "Updated:")
}

func TestView_NoHeader_NewProject(t *testing.T) {
	m := New(gtd.Project{}, nil, nil)
	view := ansi.Strip(m.View())

	assert.NotContains(t, view, "Project ID:")
	assert.NotContains(t, view, "Status:")
	assert.NotContains(t, view, "Created:")
}

func TestSaveError_ReturnsErrorCmd(t *testing.T) {
	m := New(gtd.Project{ID: 1, Title: "Existing"}, nil, nil)

	updated, cmd := m.Update(projectSavedMsg{err: errors.New("disk full")})
	_ = updated
	require.NotNil(t, cmd, "expected error cmd on save failure")
	msg := cmd()
	err, ok := msg.(error)
	require.True(t, ok, "expected error msg, got %T", msg)
	assert.Contains(t, err.Error(), "disk full")
}

func TestSaveError_EscClearsAndResumesForm(t *testing.T) {
	m := New(gtd.Project{ID: 1, Title: "Existing"}, nil, nil)

	withErr, _ := m.Update(projectSavedMsg{err: errors.New("disk full")})
	assert.NotNil(t, withErr.(Model).err)

	cleared, cmd := withErr.(Model).Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	assert.Nil(t, cmd)
	assert.Nil(t, cleared.(Model).err)
}

func TestSaveError_OtherKeysSwallowed(t *testing.T) {
	m := New(gtd.Project{ID: 1, Title: "Existing"}, nil, nil)

	withErr, _ := m.Update(projectSavedMsg{err: errors.New("disk full")})

	_, cmd := withErr.(Model).Update(tea.KeyPressMsg{Code: 'x', Text: "x"})
	assert.Nil(t, cmd)
}

func TestUpdate_CreateSuccess_ReplacesWithView(t *testing.T) {
	viewCalled := false
	factory := func(p gtd.Project) screen.Screen {
		viewCalled = true
		return stubScreen{}
	}
	m := New(gtd.Project{}, nil, factory)

	created := gtd.Project{ID: 42, Title: "New"}
	updated, cmd := m.Update(projectSavedMsg{project: created, created: true})

	assert.NotNil(t, cmd, "expected a cmd batching window-size + init")
	assert.True(t, viewCalled, "view factory should run")
	_, ok := updated.(stubScreen)
	assert.True(t, ok, "expected the model to be replaced by the view; got %T", updated)
}

func TestUpdate_UpdateSuccess_DismissOnly(t *testing.T) {
	m := New(gtd.Project{ID: 1, Title: "Existing"}, nil, nil)

	_, cmd := m.Update(projectSavedMsg{project: gtd.Project{ID: 1}, created: false})
	assert.NotNil(t, cmd)

	msg := cmd()
	_, ok := msg.(screen.DismissMsg)
	assert.True(t, ok, "expected DismissMsg, got %T", msg)
}

func TestCtrlEnter_SavesExistingProject(t *testing.T) {
	db := openTestDB(t)
	svc := service.NewProjectService(db)
	created, err := svc.CreateProject(t.Context(), gtd.Project{Title: "Build shed", Outcome: "Shed built", Status: gtd.ProjectStatusOpen})
	require.NoError(t, err)

	// Tab advances Title → Outcome in the new form toolkit; type a suffix
	// onto the prefilled outcome and submit via ctrl+s.
	var s screen.Screen = New(created, svc, nil)
	s = screentest.Init(t, s)
	s = screentest.Send(t, s, tea.KeyPressMsg{Code: tea.KeyTab})
	s = screentest.TypeText(t, s, " v2")

	_, dismissed := screentest.RunUntilDismiss(t, s, tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
	require.True(t, dismissed)

	got, err := svc.GetProject(t.Context(), created.ID)
	require.NoError(t, err)
	assert.Equal(t, "Shed built v2", got.Outcome)
}

func TestCtrlEnter_NewProject_TriggersDismissThenPush(t *testing.T) {
	db := openTestDB(t)
	svc := service.NewProjectService(db)

	viewCalled := false
	factory := func(p gtd.Project) screen.Screen {
		viewCalled = true
		return stubScreen{}
	}

	var s screen.Screen = New(gtd.Project{}, svc, factory)
	s = screentest.Init(t, s)

	s = screentest.TypeText(t, s, "New project")
	s = screentest.Send(t, s, tea.KeyPressMsg{Code: tea.KeyTab})
	s = screentest.TypeText(t, s, "ship it")

	for st := range screentest.PumpSend(t, s, tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl}) {
		s = st
	}

	assert.True(t, viewCalled, "expected the view factory to fire after create-on-ctrl+s")

	projects, err := svc.ListProjects(t.Context(), gtd.ProjectFilter{})
	require.NoError(t, err)
	require.Len(t, projects, 1)
	assert.Equal(t, "New project", projects[0].Title)
}

func TestCtrlEnter_ValidationFails_NoSave(t *testing.T) {
	db := openTestDB(t)
	svc := service.NewProjectService(db)

	// Empty Title and Outcome — the title validator fails first.
	var s screen.Screen = New(gtd.Project{}, svc, nil)
	s = screentest.Init(t, s)

	_, dismissed := screentest.RunUntilDismiss(t, s, tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
	require.False(t, dismissed, "overlay must not dismiss with invalid inputs")

	projects, err := svc.ListProjects(t.Context(), gtd.ProjectFilter{})
	require.NoError(t, err)
	assert.Empty(t, projects)
}

func TestEsc_DismissesWithoutSaving(t *testing.T) {
	m := New(gtd.Project{}, nil, nil)

	_, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	require.NotNil(t, cmd)
	msg := cmd()
	_, ok := msg.(screen.DismissMsg)
	assert.True(t, ok, "expected DismissMsg from esc, got %T", msg)
}
