// Package theme is the single source of truth for the TUI's visual
// styling: the color palette and the reusable lipgloss styles built on
// top of it. Every other package references these instead of constructing
// its own styles, so a tweak here changes the whole app.
//
// Styles are lipgloss values; method calls return copies. A consumer that
// needs a variation derives one in place (e.g. theme.Label.Width(10)) and
// the shared definition stays untouched.
package theme

import "charm.land/lipgloss/v2"

// Palette. Colors whose role matches huh's Charm theme use that theme's
// exact (dark-background) values so the app reads as a Charm app: fuchsia
// for focus/selection, indigo for the logo, green for ready/needs-action,
// red for errors, cream for text on fills, and huh's 243 neutral text
// tone. App-specific hues with no huh analog (due-date orange/yellow,
// deferred blue) keep their tuned ANSI codes. A single
// value can serve more than one semantic role (e.g. Danger is both errors
// and overdue); names lean on the role that drove the choice and the
// comments note the rest. Only a dark terminal background is tuned today.
var (
	// Status / urgency hues.
	Danger   = lipgloss.Color("#ED567A") // huh red: errors, overdue
	DueToday = lipgloss.Color("208")     // orange
	Warning  = lipgloss.Color("136")     // dark amber: due soon, banners
	Ready    = lipgloss.Color("#02BF87") // huh green: ready / needs-action
	Deferred = lipgloss.Color("67")      // dim blue
	Project  = lipgloss.Color("#7571F9") // huh indigo: project association
	Done     = lipgloss.Color("65")      // muted green: completed items
	Accent   = lipgloss.Color("#F780E2") // huh fuchsia: focus affordance, assignee
	Active   = lipgloss.Color("#F780E2") // huh fuchsia: active tab / bullet

	// Neutral text tones, dimmest to brightest.
	Subtle = lipgloss.Color("240")     // labels, hints, inactive tabs, dropped
	Muted  = lipgloss.Color("243")     // huh description: secondary values, descriptions
	White  = lipgloss.Color("#FFFDF5") // huh cream: text on accent fills

	// Surfaces / fills.
	LogoBg        = lipgloss.Color("#7571F9") // huh indigo: logo background
	LogoFg        = lipgloss.Color("#FFFDF5") // huh cream: logo text
	MutedAccentBg = lipgloss.Color("#A24FA0") // dimmed fuchsia: selected pill on a blurred field
	PillBg        = lipgloss.Color("237")     // huh blurred-button bg: unselected pill
)

// Reusable styles. Heads, labels, values and item-title variants recur
// across pages; build them once here.
var (
	// Title is a bold heading.
	Title = lipgloss.NewStyle().Bold(true)

	// Subtitle is faint secondary heading or description text.
	Subtitle = lipgloss.NewStyle().Faint(true)

	// Label is a de-emphasized field label. Callers add .Width(n) to
	// align a label column.
	Label = lipgloss.NewStyle().Foreground(Subtle)

	// Value is secondary value/description text.
	Value = lipgloss.NewStyle().Foreground(Muted)

	// Error renders error text.
	Error = lipgloss.NewStyle().Foreground(Danger)

	// DoneTitle renders the title of a completed item.
	DoneTitle = lipgloss.NewStyle().Foreground(Done).Faint(true)

	// DroppedTitle renders the title of a dropped item.
	DroppedTitle = lipgloss.NewStyle().Foreground(Subtle).Faint(true).Strikethrough(true)
)
