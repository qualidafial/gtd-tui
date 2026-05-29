package projectedit

import (
	"errors"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qualidafial/gtd-tui"
	"github.com/qualidafial/gtd-tui/tui/components/screen"
)

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
	assert.Equal(t, huh.StateNormal, cleared.(Model).form.State)
}

func TestSaveError_OtherKeysSwallowed(t *testing.T) {
	m := New(gtd.Project{ID: 1, Title: "Existing"}, nil, nil)

	withErr, _ := m.Update(projectSavedMsg{err: errors.New("disk full")})

	_, cmd := withErr.(Model).Update(tea.KeyPressMsg{Code: 'x', Text: "x"})
	assert.Nil(t, cmd)
}

func TestUpdate_CreateSuccess_DismissAndPushView(t *testing.T) {
	viewCalled := false
	factory := func(p gtd.Project) screen.Screen {
		viewCalled = true
		return nil
	}
	m := New(gtd.Project{}, nil, factory)

	created := gtd.Project{ID: 42, Title: "New"}
	updated, cmd := m.Update(projectSavedMsg{project: created, created: true})
	_ = updated

	assert.NotNil(t, cmd)
	assert.True(t, viewCalled)
}

func TestUpdate_UpdateSuccess_DismissOnly(t *testing.T) {
	m := New(gtd.Project{ID: 1, Title: "Existing"}, nil, nil)

	_, cmd := m.Update(projectSavedMsg{project: gtd.Project{ID: 1}, created: false})
	assert.NotNil(t, cmd)

	msg := cmd()
	_, ok := msg.(screen.DismissMsg)
	assert.True(t, ok, "expected DismissMsg, got %T", msg)
}

func TestEsc_DismissesWithoutSaving(t *testing.T) {
	m := New(gtd.Project{}, nil, nil)

	// Init the form so it's in normal state
	m.form.State = huh.StateAborted

	updated, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	_ = updated
	// The form should produce a dismiss when aborted
	assert.NotNil(t, cmd)
}
