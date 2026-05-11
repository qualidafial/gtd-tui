package gtd

import (
	"context"
	"time"
)

type ProjectTask struct {
	ProjectID int64
	TaskID    int64
	CreatedAt time.Time
}

type ProjectTaskFilter struct {
	ProjectIDs []int64
	TaskIDs    []int64
}

type ProjectTaskService interface {
	ProjectTasks(ctx context.Context, filter ProjectTaskFilter) ([]ProjectTask, error)
	AddTaskToProject(ctx context.Context, taskID, projectID int64) error
	RemoveTaskFromProject(ctx context.Context, taskID, projectID int64) error
}
