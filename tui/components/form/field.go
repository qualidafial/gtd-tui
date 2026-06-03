package form

import (
	tea "charm.land/bubbletea/v2"

	"github.com/qualidafial/gtd-tui/tui/internal/keymap"
)

// Field is a single form input. Implementations should be value types;
// mutating methods (Focus, Blur, Update, SetWidth) return a new Field
// value rather than mutating in place.
type Field interface {
	// Key is the field's unique identifier within a form. Required and
	// non-empty.
	Key() string
	Focus() (Field, tea.Cmd)
	Blur() Field
	Focused() bool
	// Visible reports whether the field should participate in rendering,
	// navigation, validation, submit, and help. The provided Values
	// contains only the values of visible preceding fields.
	Visible(Values) bool
	Update(tea.Msg) (Field, tea.Cmd)
	View() string
	Value() any
	// Validate runs the field's validator against its current value and
	// returns a new Field whose subsequent View reflects the result (e.g.
	// an inline error message), along with the error itself for the
	// form's decision logic. The caller-supplied validator function must
	// be pure; Field.Validate is allowed to cache the result internally —
	// the cache is what the returned Field carries.
	Validate() (Field, error)
	// SetWidth notifies the field of the column width available to it.
	// Fields that wrap or truncate should honor this; trivial fields may
	// return themselves unchanged.
	SetWidth(int) Field

	// Chords returns the field's complete consumed keybindings as the
	// single source for both routing (which keys the form must forward to
	// this field instead of treating as navigation) and help. A field does
	// not expose a separate curated routing list.
	keymap.Map
}
