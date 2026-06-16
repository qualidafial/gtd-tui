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

// Palette. A single ANSI code can serve more than one semantic role
// (e.g. Danger is both errors and overdue); names lean on the role that
// drove the choice and the comments note the rest. Only a dark terminal
// background is tuned today.
var (
	// Status / urgency hues.
	Danger   = lipgloss.Color("9")   // errors, overdue
	DueToday = lipgloss.Color("208") // orange
	Warning  = lipgloss.Color("11")  // yellow: due soon, banners
	Ready    = lipgloss.Color("44")  // teal: ready / needs-action
	Deferred = lipgloss.Color("67")  // dim blue
	Project  = lipgloss.Color("36")  // green: project association
	Done     = lipgloss.Color("65")  // muted green: completed items
	Accent   = lipgloss.Color("13")  // magenta: focus affordance, assignee
	Active   = lipgloss.Color("212") // pink: active tab / bullet

	// Neutral text tones, dimmest to brightest.
	Subtle = lipgloss.Color("240") // labels, hints, inactive tabs, dropped
	Muted  = lipgloss.Color("245") // secondary values, descriptions
	Soft   = lipgloss.Color("250") // slightly brighter descriptions
	White  = lipgloss.Color("15")  // text on accent fills

	// Surfaces / fills.
	LogoBg        = lipgloss.Color("62")  // logo background
	LogoFg        = lipgloss.Color("230") // logo text
	MutedAccentBg = lipgloss.Color("90")  // selected pill on a blurred field
	PillBg        = lipgloss.Color("8")   // unselected pill background
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
