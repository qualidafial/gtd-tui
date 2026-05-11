package gtd

import (
	"context"
	"time"
)

type TaskStatus string

const (
	TaskStatusInbox    TaskStatus = "inbox"
	TaskStatusActive   TaskStatus = "active"
	TaskStatusWaiting  TaskStatus = "waiting"
	TaskStatusDeferred TaskStatus = "deferred"
	TaskStatusDone     TaskStatus = "done"
	TaskStatusDropped  TaskStatus = "dropped"
)

type Task struct {
	ID          int64
	Title       string
	Description string
	Status      TaskStatus
	Due         *time.Time
	DeferUntil  *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type TaskService interface {
	Task(ctx context.Context, id int64) (Task, error)
	Tasks(ctx context.Context, filter TaskFilter) ([]Task, error)
	CreateTask(ctx context.Context, task Task) (Task, error)
	UpdateTask(ctx context.Context, task Task) (Task, error)
	DeleteTask(ctx context.Context, id int64) error
}

type TaskFilter struct {
	Status     *TaskStatus
	ProjectIDs []int64
	TaskIDs    []int64
	Query      string
}
