// Package cmds holds small tea.Cmd helpers shared across the TUI.
package cmds

import tea "charm.land/bubbletea/v2"

// Emit wraps a value in a tea.Cmd so it can be batched alongside other cmds
// (e.g. with screen.Dismiss) without each caller defining a one-off closure.
func Emit[T any](v T) tea.Cmd {
	return func() tea.Msg { return v }
}
