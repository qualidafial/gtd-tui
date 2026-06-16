package projectlist

import (
	"slices"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/qualidafial/gtd-tui"
	"github.com/qualidafial/gtd-tui/service"
	"github.com/qualidafial/gtd-tui/sqlite"
	"github.com/qualidafial/gtd-tui/tui/components/screen"
	"github.com/qualidafial/gtd-tui/tui/pages/projects/projectedit"
	"github.com/qualidafial/gtd-tui/tui/pages/projects/projectstatus"
	"github.com/qualidafial/gtd-tui/tui/pages/projects/projectview"
)

func openTestSvc(t *testing.T) gtd.ProjectService {
	t.Helper()
	db, err := sqlite.Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return service.NewProjectService(db)
}

func seedProject(t *testing.T, svc gtd.ProjectService, p gtd.Project) gtd.Project {
	t.Helper()
	created, err := svc.CreateProject(t.Context(), p)
	if err != nil {
		t.Fatalf("seed project: %v", err)
	}
	return created
}

func loadProjects(m Model, projects []gtd.Project) Model {
	updated, _ := m.Update(projectsLoadedMsg{projects: projects})
	return updated.(Model)
}

func sendKey(m Model, msg tea.KeyPressMsg) (Model, tea.Cmd) {
	updated, cmd := m.Update(msg)
	return updated.(Model), cmd
}

func pushScreen(t *testing.T, cmd tea.Cmd) screen.Screen {
	t.Helper()
	if cmd == nil {
		t.Fatal("expected a cmd, got nil")
	}
	msg := cmd()
	push, ok := msg.(screen.PushMsg)
	if !ok {
		t.Fatalf("expected PushMsg, got %T", msg)
	}
	return push.Screen
}

func TestModel_Load_AppliesItems(t *testing.T) {
	svc := openTestSvc(t)
	seedProject(t, svc, gtd.Project{Title: "Alpha", Status: gtd.ProjectStatusOpen})
	seedProject(t, svc, gtd.Project{Title: "Beta", Status: gtd.ProjectStatusSomeday})

	projects, _ := svc.ListProjects(t.Context(), gtd.ProjectFilter{})
	m := loadProjects(New(svc, nil, nil, nil), projects)

	if got := len(m.list.Items()); got != 2 {
		t.Fatalf("expected 2 items; got %d", got)
	}
}

func TestModel_PlusKey_PushesCreateOverlay(t *testing.T) {
	m := New(openTestSvc(t), nil, nil, nil)
	_, cmd := sendKey(m, tea.KeyPressMsg{Code: '+', Text: "+"})
	s := pushScreen(t, cmd)
	if _, ok := s.(projectedit.Model); !ok {
		t.Fatalf("expected projectedit.Model, got %T", s)
	}
}

func TestModel_Space_CompletePushesConfirmation(t *testing.T) {
	svc := openTestSvc(t)
	p := seedProject(t, svc, gtd.Project{Title: "P", Status: gtd.ProjectStatusOpen})

	m := loadProjects(New(svc, nil, nil, nil), []gtd.Project{p})
	_, cmd := sendKey(m, tea.KeyPressMsg{Code: ' ', Text: " "})

	s := pushScreen(t, cmd)
	ps, ok := s.(projectstatus.Model)
	if !ok {
		t.Fatalf("expected projectstatus.Model, got %T", s)
	}
	if ps.Transition() != projectstatus.Complete {
		t.Fatalf("transition = %v, want Complete", ps.Transition())
	}
}

func TestModel_Space_ReopenIsImmediate(t *testing.T) {
	svc := openTestSvc(t)
	p := seedProject(t, svc, gtd.Project{Title: "P", Status: gtd.ProjectStatusSomeday})

	m := loadProjects(New(svc, nil, nil, nil), []gtd.Project{p})
	_, cmd := sendKey(m, tea.KeyPressMsg{Code: ' ', Text: " "})

	if cmd == nil {
		t.Fatal("expected a reload cmd after reopen")
	}
	if msg := cmd(); msg != nil {
		if _, ok := msg.(screen.PushMsg); ok {
			t.Fatal("reopen should not push a confirmation overlay")
		}
	}
}

func TestModel_Delete_DropPushesConfirmation(t *testing.T) {
	svc := openTestSvc(t)
	p := seedProject(t, svc, gtd.Project{Title: "P", Status: gtd.ProjectStatusOpen})

	m := loadProjects(New(svc, nil, nil, nil), []gtd.Project{p})
	_, cmd := sendKey(m, tea.KeyPressMsg{Code: tea.KeyDelete})

	s := pushScreen(t, cmd)
	ps, ok := s.(projectstatus.Model)
	if !ok {
		t.Fatalf("expected projectstatus.Model, got %T", s)
	}
	if ps.Transition() != projectstatus.Drop {
		t.Fatalf("transition = %v, want Drop", ps.Transition())
	}
}

func TestModel_Delete_DisabledForDone(t *testing.T) {
	svc := openTestSvc(t)
	p := seedProject(t, svc, gtd.Project{Title: "P", Status: gtd.ProjectStatusDone})

	m := loadProjects(New(svc, nil, nil, nil), []gtd.Project{p})
	_, cmd := sendKey(m, tea.KeyPressMsg{Code: tea.KeyDelete})

	if cmd != nil {
		if msg := cmd(); msg != nil {
			if _, ok := msg.(screen.PushMsg); ok {
				t.Fatal("delete on done project should not push an overlay")
			}
		}
	}
}

func TestModel_S_ParkIsImmediate(t *testing.T) {
	svc := openTestSvc(t)
	p := seedProject(t, svc, gtd.Project{Title: "P", Status: gtd.ProjectStatusOpen})

	m := loadProjects(New(svc, nil, nil, nil), []gtd.Project{p})
	_, cmd := sendKey(m, tea.KeyPressMsg{Code: 's', Text: "s"})

	if cmd == nil {
		t.Fatal("expected a park cmd")
	}
	if _, ok := cmd().(screen.PushMsg); ok {
		t.Fatal("park should not push an overlay")
	}
}

func TestModel_MoveBindings_Boundaries(t *testing.T) {
	svc := openTestSvc(t)
	p1 := seedProject(t, svc, gtd.Project{Title: "A", Status: gtd.ProjectStatusOpen})
	p2 := seedProject(t, svc, gtd.Project{Title: "B", Status: gtd.ProjectStatusOpen})
	p3 := seedProject(t, svc, gtd.Project{Title: "C", Status: gtd.ProjectStatusDone})

	m := loadProjects(New(svc, nil, nil, nil), []gtd.Project{p1, p2, p3})

	down := func(m Model) Model {
		u, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyDown})
		return u.(Model)
	}

	tests := []struct {
		name           string
		model          Model
		wantUp, wantDn bool
	}{
		{"first open", m, false, true},
		{"last open", down(m), true, false},
		{"done item", down(down(m)), false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.model.KeyMap.MoveUp.Enabled(); got != tt.wantUp {
				t.Errorf("MoveUp enabled = %v, want %v", got, tt.wantUp)
			}
			if got := tt.model.KeyMap.MoveDown.Enabled(); got != tt.wantDn {
				t.Errorf("MoveDown enabled = %v, want %v", got, tt.wantDn)
			}
			// Move-to-top tracks move-up; move-to-bottom tracks move-down.
			if got := tt.model.KeyMap.MoveFirst.Enabled(); got != tt.wantUp {
				t.Errorf("MoveFirst enabled = %v, want %v", got, tt.wantUp)
			}
			if got := tt.model.KeyMap.MoveLast.Enabled(); got != tt.wantDn {
				t.Errorf("MoveLast enabled = %v, want %v", got, tt.wantDn)
			}
		})
	}
}

func TestModel_MoveLast_ReordersAndKeepsCursor(t *testing.T) {
	svc := openTestSvc(t)
	first := seedProject(t, svc, gtd.Project{Title: "A", Status: gtd.ProjectStatusOpen})
	seedProject(t, svc, gtd.Project{Title: "B", Status: gtd.ProjectStatusOpen})
	seedProject(t, svc, gtd.Project{Title: "C", Status: gtd.ProjectStatusOpen})

	projects, err := svc.ListProjects(t.Context(), gtd.ProjectFilter{})
	if err != nil {
		t.Fatalf("list projects: %v", err)
	}

	m := loadProjects(New(svc, nil, nil, nil), projects)

	// Cursor starts on the first project (A). shift+end moves it to the bottom.
	_, cmd := sendKey(m, tea.KeyPressMsg{Code: tea.KeyEnd, Mod: tea.ModShift})
	if cmd == nil {
		t.Fatal("expected a reorder cmd from shift+end")
	}
	msg, ok := cmd().(projectsReorderedMsg)
	if !ok {
		t.Fatalf("expected projectsReorderedMsg, got %T", cmd())
	}
	if msg.selectID != first.ID {
		t.Fatalf("selectID = %d, want %d (the moved project)", msg.selectID, first.ID)
	}
	gotOrder := make([]string, len(msg.projects))
	for i, p := range msg.projects {
		gotOrder[i] = p.Title
	}
	if want := []string{"B", "C", "A"}; !slices.Equal(gotOrder, want) {
		t.Fatalf("order after move = %v, want %v", gotOrder, want)
	}

	// Applying the msg keeps the cursor on the moved project.
	applied, _ := m.Update(msg)
	if sel, ok := applied.(Model).list.SelectedItem().(Item); !ok || sel.project.ID != first.ID {
		t.Fatal("cursor not on moved project after reorder")
	}
}

func TestModel_Enter_PushesProjectView(t *testing.T) {
	svc := openTestSvc(t)
	p := seedProject(t, svc, gtd.Project{Title: "P", Status: gtd.ProjectStatusOpen})

	m := loadProjects(New(svc, nil, nil, nil), []gtd.Project{p})
	_, cmd := sendKey(m, tea.KeyPressMsg{Code: tea.KeyEnter})
	s := pushScreen(t, cmd)
	if _, ok := s.(projectview.Model); !ok {
		t.Fatalf("expected projectview.Model, got %T", s)
	}
}

func TestModel_QueryBar_DefaultQuery(t *testing.T) {
	m := New(openTestSvc(t), nil, nil, nil)
	if got := m.query.Value(); got != defaultProjectQuery {
		t.Fatalf("default query = %q, want %q", got, defaultProjectQuery)
	}
}

func TestModel_QueryBar_FocusOnSlash(t *testing.T) {
	m := New(openTestSvc(t), nil, nil, nil)
	m2, _ := sendKey(m, tea.KeyPressMsg{Code: '/', Text: "/"})
	if !m2.query.CapturingInput() {
		t.Fatal("'/' should focus the query bar")
	}
	if !m2.CapturingInput() {
		t.Fatal("model CapturingInput() should be true when query bar is focused")
	}
}

func TestModel_QueryBar_ResetRevertsToDefaultAndReloads(t *testing.T) {
	m := New(openTestSvc(t), nil, nil, nil)
	m.query.SetValue("status:someday")
	m.updateKeybindings()

	m2, cmd := sendKey(m, tea.KeyPressMsg{Code: '\\', Text: "\\"})
	if got := m2.query.Value(); got != defaultProjectQuery {
		t.Fatalf("query after revert = %q, want %q", got, defaultProjectQuery)
	}
	if cmd == nil {
		t.Fatal("expected a reload cmd from revert")
	}
	if _, ok := cmd().(projectsLoadedMsg); !ok {
		t.Fatal("revert cmd should reload projects")
	}
}

func TestModel_QueryBar_ResetDisabledAtDefault(t *testing.T) {
	m := New(openTestSvc(t), nil, nil, nil)
	if m.KeyMap.ResetQuery.Enabled() {
		t.Fatal("ResetQuery should be disabled when query equals default")
	}
	m2, cmd := sendKey(m, tea.KeyPressMsg{Code: '\\', Text: "\\"})
	if cmd != nil {
		t.Fatal("revert at default should not issue a cmd")
	}
	if got := m2.query.Value(); got != defaultProjectQuery {
		t.Fatalf("value changed at default: %q", got)
	}
}

func TestModel_QueryBar_CancelReverts(t *testing.T) {
	m := New(openTestSvc(t), nil, nil, nil)
	// focus
	m2, _ := sendKey(m, tea.KeyPressMsg{Code: '/', Text: "/"})
	// esc to cancel
	m3, cmd := sendKey(m2, tea.KeyPressMsg{Code: tea.KeyEscape})
	if cmd == nil {
		t.Fatal("expected cmd from cancel")
	}
	if m3.query.CapturingInput() {
		t.Fatal("query bar should be blurred after cancel")
	}
	// value should revert to default
	if got := m3.query.Value(); got != defaultProjectQuery {
		t.Fatalf("after cancel, query = %q, want %q", got, defaultProjectQuery)
	}
}
