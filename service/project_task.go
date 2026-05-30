package service

import (
	"context"
	"time"

	"github.com/qualidafial/gtd-tui"
)

type ProjectTaskService struct {
	inner     gtd.TaskService
	projectID int64
}

func NewProjectTaskService(inner gtd.TaskService, projectID int64) *ProjectTaskService {
	return &ProjectTaskService{inner: inner, projectID: projectID}
}

var _ gtd.TaskService = (*ProjectTaskService)(nil)

func (s *ProjectTaskService) GetTask(ctx context.Context, id int64) (gtd.Task, error) {
	return s.inner.GetTask(ctx, id)
}

func (s *ProjectTaskService) ListTasks(ctx context.Context, filter gtd.TaskFilter) ([]gtd.Task, error) {
	filter.ProjectID = &s.projectID
	return s.inner.ListTasks(ctx, filter)
}

func (s *ProjectTaskService) CreateTask(ctx context.Context, task gtd.Task) (gtd.Task, error) {
	task.ProjectID = &s.projectID
	return s.inner.CreateTask(ctx, task)
}

func (s *ProjectTaskService) UpdateTask(ctx context.Context, task gtd.Task) (gtd.Task, error) {
	return s.inner.UpdateTask(ctx, task)
}

func (s *ProjectTaskService) CompleteTask(ctx context.Context, id int64, at time.Time) (gtd.Task, error) {
	return s.inner.CompleteTask(ctx, id, at)
}

func (s *ProjectTaskService) DropTask(ctx context.Context, id int64, at time.Time) (gtd.Task, error) {
	return s.inner.DropTask(ctx, id, at)
}

func (s *ProjectTaskService) ReopenTask(ctx context.Context, id int64, at time.Time) (gtd.Task, error) {
	return s.inner.ReopenTask(ctx, id, at)
}

func (s *ProjectTaskService) DeleteTask(ctx context.Context, id int64) error {
	return s.inner.DeleteTask(ctx, id)
}

func (s *ProjectTaskService) MoveTaskUp(ctx context.Context, id int64, filter gtd.TaskFilter) error {
	filter.ProjectID = &s.projectID
	return s.inner.MoveTaskUp(ctx, id, filter)
}

func (s *ProjectTaskService) MoveTaskDown(ctx context.Context, id int64, filter gtd.TaskFilter) error {
	filter.ProjectID = &s.projectID
	return s.inner.MoveTaskDown(ctx, id, filter)
}