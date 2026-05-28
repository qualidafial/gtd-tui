package taskstatus

import (
	"context"
	"fmt"
	"time"

	"github.com/qualidafial/gtd-tui"
)

// Transition identifies a task status change initiated from the task list.
type Transition int

const (
	Complete Transition = iota
	Drop
	Reopen
)

// spec describes how a transition is presented and applied.
type spec struct {
	title       string
	description func(title string) string
	affirmative string
	apply       func(svc gtd.TaskService, ctx context.Context, id int64, at time.Time) (gtd.Task, error)
}

var specs = map[Transition]spec{
	Complete: {
		title:       "Complete task?",
		description: func(t string) string { return fmt.Sprintf("%q will be marked done.", t) },
		affirmative: "Complete",
		apply: func(svc gtd.TaskService, ctx context.Context, id int64, at time.Time) (gtd.Task, error) {
			return svc.CompleteTask(ctx, id, at)
		},
	},
	Drop: {
		title:       "Drop task?",
		description: func(t string) string { return fmt.Sprintf("%q will be moved to Dropped.", t) },
		affirmative: "Drop",
		apply: func(svc gtd.TaskService, ctx context.Context, id int64, at time.Time) (gtd.Task, error) {
			return svc.DropTask(ctx, id, at)
		},
	},
	Reopen: {
		title:       "Reopen task?",
		description: func(t string) string { return fmt.Sprintf("%q will be moved back to open.", t) },
		affirmative: "Reopen",
		apply: func(svc gtd.TaskService, ctx context.Context, id int64, at time.Time) (gtd.Task, error) {
			return svc.ReopenTask(ctx, id, at)
		},
	},
}
