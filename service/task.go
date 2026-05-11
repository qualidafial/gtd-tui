package service

import (
	"context"

	gtd "github.com/qualidafial/gtd-tui"
	"github.com/qualidafial/gtd-tui/sqlite"
)

type TaskService struct {
	db *sqlite.DB
}

func NewTaskService(db *sqlite.DB) *TaskService {
	return &TaskService{db: db}
}

func (s *TaskService) Task(ctx context.Context, id int64) (gtd.Task, error) {
	return s.db.Task(ctx, id)
}

func (s *TaskService) Tasks(ctx context.Context, filter gtd.TaskFilter) ([]gtd.Task, error) {
	return s.db.Tasks(ctx, filter)
}

func (s *TaskService) CreateTask(ctx context.Context, task gtd.Task) (gtd.Task, error) {
	return s.db.CreateTask(ctx, task)
}

func (s *TaskService) UpdateTask(ctx context.Context, task gtd.Task) (gtd.Task, error) {
	return s.db.UpdateTask(ctx, task)
}

func (s *TaskService) DeleteTask(ctx context.Context, id int64) error {
	return s.db.DeleteTask(ctx, id)
}

var _ gtd.TaskService = (*TaskService)(nil)
