package projectlist

import "github.com/qualidafial/gtd-tui"

// Item wraps a project and its task counts to satisfy list.Item.
type Item struct {
	project gtd.Project
	counts  gtd.ProjectTaskCounts
}

func (it Item) FilterValue() string { return it.project.Title }
