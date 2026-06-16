package form

import (
	"charm.land/lipgloss/v2"

	"github.com/qualidafial/gtd-tui/tui/theme"
)

// Shared styles used by all form fields. Field packages reference these
// directly so styling stays consistent across the toolkit. Colors come
// from the theme palette; a consumer that wants a different look can
// override individual vars (lipgloss styles are values; mutating these
// shared vars is global, so do it once at startup if at all).
var (
	// LabelStyle renders a field label when its field is not focused.
	LabelStyle = theme.Title

	// FocusedLabelStyle renders a field label when its field is focused.
	// The accent foreground is how a user sees "this field has focus".
	FocusedLabelStyle = theme.Title.Foreground(theme.Accent)

	// ErrorStyle renders a field's cached validation error beneath the
	// value.
	ErrorStyle = theme.Error

	// AccentStyle is the "selected" / "focused affordance" pill — bold
	// white on bright magenta. Used by the focused savefield button and
	// the selected radiofield option of a focused field.
	AccentStyle = lipgloss.NewStyle().
			Foreground(theme.White).
			Background(theme.Accent).
			Bold(true)

	// MutedAccentStyle is the selected pill on an *unfocused* field —
	// bright white on dim magenta. Tells the user "this is the current
	// selection here, but this field isn't where you're typing." Used
	// for the selected radiofield option when the field is blurred.
	MutedAccentStyle = lipgloss.NewStyle().
				Foreground(theme.White).
				Background(theme.MutedAccentBg)

	// DimStyle is the "blurred" / "unselected" pill — bright white (not
	// bold) on gray. Used by the blurred savefield button and unselected
	// radiofield options.
	DimStyle = lipgloss.NewStyle().
			Foreground(theme.White).
			Background(theme.PillBg)
)
