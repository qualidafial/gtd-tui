package projectview

import (
	"strings"
	"testing"
	"time"

	"github.com/qualidafial/gtd-tui/tui/components/screen/screentest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	tea "charm.land/bubbletea/v2"

	"github.com/qualidafial/gtd-tui"
	"github.com/qualidafial/gtd-tui/service"
	"github.com/qualidafial/gtd-tui/sqlite"
)

type env struct {
	taskSvc    gtd.TaskService
	projectSvc gtd.ProjectService
}

func setup(t *testing.T) env {
	t.Helper()
	db, err := sqlite.Open(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return env{
		taskSvc:    service.NewTaskService(db),
		projectSvc: service.NewProjectService(db),
	}
}

func TestHeader_AllFields(t *testing.T) {
	due := time.Date(2026, 6, 15, 0, 0, 0, 0, time.Local)
	p := gtd.Project{
		ID:      1,
		Title:   "Build shed",
		Status:  gtd.ProjectStatusOpen,
		Outcome: "A functional shed",
		Due:     &due,
	}
	m := New(p, nil, nil, nil)
	header := m.renderHeader()

	assert.Contains(t, header, "Build shed")
	assert.Contains(t, header, "Open")
	assert.Contains(t, header, "A functional shed")
	assert.Contains(t, header, "2026-06-15")
}

func TestHeader_OmitsEmpty(t *testing.T) {
	p := gtd.Project{
		ID:     1,
		Title:  "Minimal",
		Status: gtd.ProjectStatusOpen,
	}
	m := New(p, nil, nil, nil)
	header := m.renderHeader()

	assert.Contains(t, header, "Minimal")
	assert.Contains(t, header, "Open")
	assert.NotContains(t, header, "Outcome")
	assert.NotContains(t, header, "Due")
}

func TestTaskScoping(t *testing.T) {
	e := setup(t)
	ctx := t.Context()

	p, err := e.projectSvc.CreateProject(ctx, gtd.Project{Title: "P1", Status: gtd.ProjectStatusOpen})
	require.NoError(t, err)

	_, err = e.taskSvc.CreateTask(ctx, gtd.Task{Title: "In project", Status: gtd.TaskStatusOpen, ProjectID: &p.ID})
	require.NoError(t, err)
	_, err = e.taskSvc.CreateTask(ctx, gtd.Task{Title: "Standalone", Status: gtd.TaskStatusOpen})
	require.NoError(t, err)

	m := New(p, e.taskSvc, e.projectSvc, nil)
	m = screentest.Init(m).(Model)

	s, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = s.(Model)

	view := m.View()
	assert.Contains(t, view, "In project")
	assert.NotContains(t, view, "Standalone")
}

func TestCreateInheritsProject(t *testing.T) {
	e := setup(t)
	ctx := t.Context()

	p, err := e.projectSvc.CreateProject(ctx, gtd.Project{Title: "P1", Status: gtd.ProjectStatusOpen})
	require.NoError(t, err)

	wrapped := service.NewProjectTaskService(e.taskSvc, p.ID)
	created, err := wrapped.CreateTask(ctx, gtd.Task{Title: "New task", Status: gtd.TaskStatusOpen})
	require.NoError(t, err)
	require.NotNil(t, created.ProjectID)
	assert.Equal(t, p.ID, *created.ProjectID)
}

func TestHeader_StatusLabels(t *testing.T) {
	for _, tt := range []struct {
		status gtd.ProjectStatus
		label  string
	}{
		{gtd.ProjectStatusOpen, "Open"},
		{gtd.ProjectStatusSomeday, "Someday"},
		{gtd.ProjectStatusDone, "Done"},
		{gtd.ProjectStatusDropped, "Dropped"},
	} {
		t.Run(string(tt.status), func(t *testing.T) {
			p := gtd.Project{ID: 1, Title: "T", Status: tt.status}
			m := New(p, nil, nil, nil)
			header := m.renderHeader()
			found := false
			for _, line := range strings.Split(header, "\n") {
				if strings.Contains(line, tt.label) {
					found = true
					break
				}
			}
			assert.True(t, found, "header should contain %q", tt.label)
		})
	}
}
