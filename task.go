package gtd

import (
	"context"
	"time"
)

type TaskStatus string

const (
	TaskStatusOpen    TaskStatus = "open"
	TaskStatusDone    TaskStatus = "done"
	TaskStatusDropped TaskStatus = "dropped"
)

type Task struct {
	ID          int64
	Title       string
	Description string
	Status      TaskStatus
	Assignee    *string
	ProjectID   *int64
	Due         *time.Time
	DeferUntil  *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
	// StatusChangedAt records when the task last entered its current status:
	// equal to CreatedAt on creation (the transition into open), then
	// overwritten by the supplied instant on every Complete/Drop/Reopen.
	StatusChangedAt time.Time
}

type TaskService interface {
	GetTask(ctx context.Context, id int64) (Task, error)
	ListTasks(ctx context.Context, filter TaskFilter) ([]Task, error)
	CreateTask(ctx context.Context, task Task) (Task, error)
	UpdateTask(ctx context.Context, task Task) (Task, error)
	CompleteTask(ctx context.Context, id int64, at time.Time) (Task, error)
	DropTask(ctx context.Context, id int64, at time.Time) (Task, error)
	ReopenTask(ctx context.Context, id int64, at time.Time) (Task, error)
	DeleteTask(ctx context.Context, id int64) error
	// MoveTaskUp / MoveTaskDown shift an open task one slot within the open
	// tasks that match filter. The filter scopes the swap to the user's current
	// view so the reorder is immediately visible; status is always forced to
	// open inside the move regardless of filter.Status.
	MoveTaskUp(ctx context.Context, id int64, filter TaskFilter) error
	MoveTaskDown(ctx context.Context, id int64, filter TaskFilter) error
	// MoveTaskFirst / MoveTaskLast move an open task ahead of / after every
	// open task matching filter, scoped to the user's current view exactly like
	// MoveTaskUp / MoveTaskDown. No-op when the task is already at that end.
	MoveTaskFirst(ctx context.Context, id int64, filter TaskFilter) error
	MoveTaskLast(ctx context.Context, id int64, filter TaskFilter) error
}

// DatePredicateKind discriminates how a DatePredicate constrains a date column.
type DatePredicateKind int

const (
	// OnOrBefore matches `column IS NOT NULL AND column <= Time` (used by Due).
	OnOrBefore DatePredicateKind = iota
	// AvailableAsOf matches `column IS NULL OR column <= Time` (used by Ready).
	AvailableAsOf
	// After matches `column > Time` (used by Defer).
	After
	// IsNull matches `column IS NULL`.
	IsNull
	// IsNotNull matches `column IS NOT NULL`.
	IsNotNull
)

// DatePredicate constrains a date column. Time-based kinds (OnOrBefore,
// AvailableAsOf, After) carry a resolved UTC time; IsNull/IsNotNull ignore Time.
type DatePredicate struct {
	Kind DatePredicateKind
	Time time.Time
}

type TaskFilter struct {
	Status    *TaskStatus
	Assignee  *string
	ProjectID *int64
	Due       *DatePredicate
	Ready     *DatePredicate
	Defer     *DatePredicate
	Search    []string
	TaskIDs   []int64
	// IncludeSomedayProjects keeps tasks whose project is parked (someday) in
	// the results. When false (default), someday-project tasks are excluded.
	IncludeSomedayProjects bool
}

func (f TaskFilter) WithStatus(s TaskStatus) TaskFilter {
	f.Status = &s
	return f
}

func (f TaskFilter) WithAssignee(a string) TaskFilter {
	f.Assignee = &a
	return f
}

func (f TaskFilter) WithProjectID(id int64) TaskFilter {
	f.ProjectID = &id
	return f
}

func (f TaskFilter) WithSearch(terms ...string) TaskFilter {
	f.Search = terms
	return f
}

func (f TaskFilter) WithTaskIDs(ids ...int64) TaskFilter {
	f.TaskIDs = ids
	return f
}
