package tui

import (
	"errors"
	"strings"
	"testing"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"

	"github.com/qualidafial/gtd-tui/tui/components/screen"
)

// stubScreen is a no-op Screen for testing tui.Model logic.
type stubScreen struct{}

func (s stubScreen) Init() tea.Cmd                            { return nil }
func (s stubScreen) Update(tea.Msg) (screen.Screen, tea.Cmd) { return s, nil }
func (s stubScreen) View() string                             { return "" }
func (s stubScreen) KeyMap() help.KeyMap                      { return emptyKeyMap{} }
func (s stubScreen) CapturingInput() bool                     { return false }

type emptyKeyMap struct{}

func (emptyKeyMap) ShortHelp() []key.Binding  { return nil }
func (emptyKeyMap) FullHelp() [][]key.Binding { return nil }

func newTestModel() Model {
	return Model{active: stubScreen{}}
}

func TestModel_Error_SetsErrField(t *testing.T) {
	m := newTestModel()
	result, _ := m.Update(errors.New("boom"))
	got := result.(Model).err
	if got == nil || got.Error() != "boom" {
		t.Fatalf("err = %v, want 'boom'", got)
	}
}

func TestModel_Error_RenderedInFooter(t *testing.T) {
	m := newTestModel()
	m.err = errors.New("something went wrong")
	footer := m.renderFooter()
	if !strings.Contains(footer, "something went wrong") {
		t.Fatalf("footer %q does not contain error message", footer)
	}
}

func TestModel_Error_ReplacedByNewer(t *testing.T) {
	m := newTestModel()
	m.err = errors.New("old error")
	result, _ := m.Update(errors.New("new error"))
	got := result.(Model).err
	if got == nil || got.Error() != "new error" {
		t.Fatalf("err = %v, want 'new error'", got)
	}
}

func TestModel_KeyPress_ClearsError(t *testing.T) {
	m := newTestModel()
	m.err = errors.New("boom")
	result, _ := m.Update(tea.KeyPressMsg{Code: 'a', Text: "a"})
	if result.(Model).err != nil {
		t.Fatal("keypress should clear error")
	}
}

func TestModel_EscKey_ClearsError(t *testing.T) {
	m := newTestModel()
	m.err = errors.New("boom")
	result, _ := m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	if result.(Model).err != nil {
		t.Fatal("esc should clear error")
	}
}
