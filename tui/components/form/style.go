package form

import "charm.land/lipgloss/v2"

// Shared styles used by all form fields. Field packages reference these
// directly so styling stays consistent across the toolkit. A consumer
// that wants a different look can override individual vars (lipgloss
// styles are values; mutating these shared vars is global, so do it once
// at startup if at all).
var (
	// LabelStyle renders a field label when its field is not focused.
	LabelStyle = lipgloss.NewStyle().Bold(true)

	// FocusedLabelStyle renders a field label when its field is focused.
	// The accent foreground is how a user sees "this field has focus".
	FocusedLabelStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("13"))

	// ErrorStyle renders a field's cached validation error beneath the
	// value.
	ErrorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))

	// AccentStyle is the "selected" / "focused affordance" pill — bold
	// white on bright magenta. Used by the focused savefield button and
	// the selected radiofield option of a focused field.
	AccentStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("13")).
			Bold(true)

	// MutedAccentStyle is the selected pill on an *unfocused* field —
	// bright white on dim magenta. Tells the user "this is the current
	// selection here, but this field isn't where you're typing." Used
	// for the selected radiofield option when the field is blurred.
	MutedAccentStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("15")).
				Background(lipgloss.Color("90"))

	// DimStyle is the "blurred" / "unselected" pill — bright white (not
	// bold) on gray. Used by the blurred savefield button and unselected
	// radiofield options.
	DimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")).
			Background(lipgloss.Color("8"))
)
