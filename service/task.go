package service

import (
	"context"

	"github.com/qualidafial/gtd-tui"
	"github.com/qualidafial/gtd-tui/sqlite"
)

type TaskService struct {
	db *sqlite.DB
}

func NewTaskService(db *sqlite.DB) *TaskService {
	return &TaskService{db: db}
}

func (s *TaskService) GetTask(ctx context.Context, id int64) (gtd.Task, error) {
	return s.db.GetTask(ctx, id)
}

func (s *TaskService) ListTasks(ctx context.Context, filter gtd.TaskFilter) ([]gtd.Task, error) {
	return s.db.ListTasks(ctx, filter)
}

func (s *TaskService) CreateTask(ctx context.Context, task gtd.Task) (gtd.Task, error) {
	return s.db.CreateTask(ctx, task)
}

func (s *TaskService) UpdateTask(ctx context.Context, task gtd.Task) (gtd.Task, error) {
	return s.db.UpdateTask(ctx, task)
}

func (s *TaskService) CompleteTask(ctx context.Context, id int64) (gtd.Task, error) {
	return s.db.CompleteTask(ctx, id)
}

func (s *TaskService) DropTask(ctx context.Context, id int64) (gtd.Task, error) {
	return s.db.DropTask(ctx, id)
}

func (s *TaskService) ReopenTask(ctx context.Context, id int64) (gtd.Task, error) {
	return s.db.ReopenTask(ctx, id)
}

func (s *TaskService) DeleteTask(ctx context.Context, id int64) error {
	return s.db.DeleteTask(ctx, id)
}

func (s *TaskService) MoveUp(ctx context.Context, id int64) error {
	return s.db.MoveUp(ctx, id)
}

func (s *TaskService) MoveDown(ctx context.Context, id int64) error {
	return s.db.MoveDown(ctx, id)
}

var _ gtd.TaskService = (*TaskService)(nil)
