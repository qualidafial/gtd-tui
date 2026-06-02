package itemcapture_test

import (
	"context"
	"errors"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/qualidafial/gtd-tui"
	"github.com/qualidafial/gtd-tui/service"
	"github.com/qualidafial/gtd-tui/sqlite"
	"github.com/qualidafial/gtd-tui/tui/components/screen"
	"github.com/qualidafial/gtd-tui/tui/components/screen/screentest"
	"github.com/qualidafial/gtd-tui/tui/pages/inbox/itemcapture"
)

func openTestDB(t *testing.T) *sqlite.DB {
	t.Helper()
	db, err := sqlite.Open(":memory:")
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return db
}

func TestCapture_TypingAndSubmitting_CreatesItem(t *testing.T) {
	db := openTestDB(t)
	svc := service.NewInboxService(db)

	var s screen.Screen = itemcapture.New(svc)
	s = screentest.Init(t, s)

	s = screentest.TypeText(t, s, "Call dentist")
	// Tab advances Title → Description in the new form toolkit.
	s = screentest.Send(t, s, tea.KeyPressMsg{Code: tea.KeyTab})
	s = screentest.TypeText(t, s, "before friday")

	// Tab to the trailing Save button, then Enter to submit.
	s = screentest.Send(t, s, tea.KeyPressMsg{Code: tea.KeyTab})
	_, dismissed := screentest.RunUntilDismiss(t, s, tea.KeyPressMsg{Code: tea.KeyEnter})
	require.True(t, dismissed, "expected the overlay to dismiss after save")

	items, err := svc.List(t.Context())
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "Call dentist", items[0].Title)
	assert.Equal(t, "before friday", items[0].Description)
}

func TestCapture_TitleOnly_StillCreates(t *testing.T) {
	db := openTestDB(t)
	svc := service.NewInboxService(db)

	var s screen.Screen = itemcapture.New(svc)
	s = screentest.Init(t, s)

	s = screentest.TypeText(t, s, "Quick capture")

	// Ctrl+s submits from anywhere; an empty Description is allowed.
	_, dismissed := screentest.RunUntilDismiss(t, s, tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
	require.True(t, dismissed)

	items, err := svc.List(t.Context())
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "Quick capture", items[0].Title)
	assert.Empty(t, items[0].Description)
}

func TestCapture_EmptyTitle_DoesNotCreate(t *testing.T) {
	db := openTestDB(t)
	svc := service.NewInboxService(db)

	var s screen.Screen = itemcapture.New(svc)
	s = screentest.Init(t, s)

	// Try to submit with empty Title — Submit validates in declaration
	// order, Title's validator rejects empty input, the form does not
	// emit SubmittedMsg, and the overlay stays open.
	for _, msg := range screentest.PumpSend(t, s, tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl}) {
		if _, ok := msg.(screen.DismissMsg); ok {
			t.Fatalf("overlay should not dismiss with an empty title")
		}
	}

	items, err := svc.List(t.Context())
	require.NoError(t, err)
	assert.Empty(t, items)
}

func TestCapture_CtrlEnter_SavesFromTitle(t *testing.T) {
	db := openTestDB(t)
	svc := service.NewInboxService(db)

	var s screen.Screen = itemcapture.New(svc)
	s = screentest.Init(t, s)

	s = screentest.TypeText(t, s, "Quick capture via ctrl+enter")

	_, dismissed := screentest.RunUntilDismiss(t, s, tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
	require.True(t, dismissed)

	items, err := svc.List(t.Context())
	require.NoError(t, err)
	require.Len(t, items, 1)
	assert.Equal(t, "Quick capture via ctrl+enter", items[0].Title)
}

func TestCapture_CtrlEnter_EmptyTitle_DoesNotCreate(t *testing.T) {
	db := openTestDB(t)
	svc := service.NewInboxService(db)

	var s screen.Screen = itemcapture.New(svc)
	s = screentest.Init(t, s)

	_, dismissed := screentest.RunUntilDismiss(t, s, tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl})
	require.False(t, dismissed, "overlay must not dismiss with empty title")

	items, err := svc.List(t.Context())
	require.NoError(t, err)
	assert.Empty(t, items)
}

func TestCapture_Cancel_DoesNotCreate(t *testing.T) {
	db := openTestDB(t)
	svc := service.NewInboxService(db)

	var s screen.Screen = itemcapture.New(svc)
	s = screentest.Init(t, s)

	s = screentest.TypeText(t, s, "Will be cancelled")
	s = screentest.Send(t, s, tea.KeyPressMsg{Code: tea.KeyEscape})

	items, err := svc.List(t.Context())
	require.NoError(t, err)
	assert.Empty(t, items)
}

// errSvc wraps a real InboxService and forces Create to fail. Used only to
// drive the save-error path, which the real :memory: stack cannot reach for a
// valid item.
type errSvc struct {
	gtd.InboxService
	err error
}

func (e *errSvc) Create(context.Context, gtd.Item) (gtd.Item, error) {
	return gtd.Item{}, e.err
}

func TestCapture_SaveError_EmitsErrorMessageAndStaysOpen(t *testing.T) {
	db := openTestDB(t)
	svc := &errSvc{InboxService: service.NewInboxService(db), err: errors.New("disk full")}

	var s screen.Screen = itemcapture.New(svc)
	s = screentest.Init(t, s)

	s = screentest.TypeText(t, s, "doomed")

	var sawError, dismissed bool
	for st, msg := range screentest.PumpSend(t, s, tea.KeyPressMsg{Code: 's', Mod: tea.ModCtrl}) {
		s = st
		if err, ok := msg.(error); ok && err != nil {
			sawError = true
		}
		if _, ok := msg.(screen.DismissMsg); ok {
			dismissed = true
		}
	}
	require.True(t, sawError, "expected an error message to be emitted on save failure")
	assert.False(t, dismissed, "overlay must not dismiss when save fails")
}
