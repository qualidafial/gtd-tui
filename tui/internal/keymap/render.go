package keymap

import "strings"

// Render maps a chord's displayed keys to a help label. It only ever
// receives keys that were in a chord's displayed set, so a hidden routing
// alias can never appear in a rendered label. Hosts may supply their own.
type Render func(keys []string) string

// glyphs maps key names to their display glyphs. Unmapped keys fall back
// to their raw string.
var glyphs = map[string]string{
	"up":        "↑",
	"down":      "↓",
	"left":      "←",
	"right":     "→",
	"shift+tab": "⇧tab",
	"enter":     "enter",
	"esc":       "esc",
	"space":     "space",
}

// DefaultRender joins per-key glyphs from the central table with "/",
// falling back to the raw key string for unmapped keys.
func DefaultRender(keys []string) string {
	parts := make([]string, len(keys))
	for i, k := range keys {
		if g, ok := glyphs[k]; ok {
			parts[i] = g
		} else {
			parts[i] = k
		}
	}
	return strings.Join(parts, "/")
}
