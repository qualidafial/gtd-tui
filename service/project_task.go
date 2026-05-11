package service

import (
	"context"

	gtd "github.com/qualidafial/gtd-tui"
	"github.com/qualidafial/gtd-tui/sqlite"
)

type ProjectTaskService struct {
	db *sqlite.DB
}

func NewProjectTaskService(db *sqlite.DB) *ProjectTaskService {
	return &ProjectTaskService{db: db}
}

func (s *ProjectTaskService) ProjectTasks(ctx context.Context, filter gtd.ProjectTaskFilter) ([]gtd.ProjectTask, error) {
	return s.db.ProjectTasks(ctx, filter)
}

func (s *ProjectTaskService) AddTaskToProject(ctx context.Context, taskID, projectID int64) error {
	return s.db.AddTaskToProject(ctx, taskID, projectID)
}

func (s *ProjectTaskService) RemoveTaskFromProject(ctx context.Context, taskID, projectID int64) error {
	return s.db.RemoveTaskFromProject(ctx, taskID, projectID)
}

var _ gtd.ProjectTaskService = (*ProjectTaskService)(nil)
