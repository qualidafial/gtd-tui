package inbox

import "github.com/qualidafial/gtd-tui"

// Item wraps gtd.Item to satisfy list.Item.
type Item struct{ item gtd.Item }

func (i Item) FilterValue() string { return i.item.Title }