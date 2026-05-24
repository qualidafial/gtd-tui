package tasklist

import "github.com/qualidafial/gtd-tui"

// Item wraps gtd.Task to satisfy list.Item. Rendering lives in the custom
// delegate (see render.go), which reads the wrapped task directly; only
// FilterValue is required by the list.
type Item struct{ task gtd.Task }

func (t Item) FilterValue() string { return t.task.Title }
