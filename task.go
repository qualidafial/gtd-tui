package gtd

import (
	"context"
	"time"
)

type TaskKind string

const (
	TaskKindNextAction TaskKind = "next_action"
	TaskKindDelegated  TaskKind = "delegated"
)

type TaskStatus string

const (
	TaskStatusPending TaskStatus = "pending"
	TaskStatusDone    TaskStatus = "done"
	TaskStatusDropped TaskStatus = "dropped"
)

type Task struct {
	ID          int64
	Title       string
	Description string
	Kind        TaskKind
	Status      TaskStatus
	Assignee    string
	Due         *time.Time
	DeferUntil  *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type TaskService interface {
	GetTask(ctx context.Context, id int64) (Task, error)
	ListTasks(ctx context.Context, filter TaskFilter) ([]Task, error)
	CreateTask(ctx context.Context, task Task) (Task, error)
	UpdateTask(ctx context.Context, task Task) (Task, error)
	CompleteTask(ctx context.Context, id int64) (Task, error)
	DropTask(ctx context.Context, id int64) (Task, error)
	ReopenTask(ctx context.Context, id int64) (Task, error)
	DeleteTask(ctx context.Context, id int64) error
	MoveUp(ctx context.Context, id int64) error
	MoveDown(ctx context.Context, id int64) error
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
	Status   *TaskStatus
	Kind     *TaskKind
	Assignee *string
	Due      *DatePredicate
	Ready    *DatePredicate
	Defer    *DatePredicate
	Search   []string
	TaskIDs  []int64
}

func (f TaskFilter) WithStatus(s TaskStatus) TaskFilter {
	f.Status = &s
	return f
}

func (f TaskFilter) WithKind(k TaskKind) TaskFilter {
	f.Kind = &k
	return f
}

func (f TaskFilter) WithAssignee(a string) TaskFilter {
	f.Assignee = &a
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
