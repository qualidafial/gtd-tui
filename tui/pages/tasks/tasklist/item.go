package tasklist

import "github.com/qualidafial/gtd-tui"

// Item wraps gtd.Task to satisfy list.DefaultItem.
type Item struct{ task gtd.Task }

func (t Item) FilterValue() string { return t.task.Title }
func (t Item) Title() string       { return t.task.Title }
func (t Item) Description() string { return "" }
