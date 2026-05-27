package service

import (
	"context"
	"time"

	"github.com/qualidafial/gtd-tui"
	"github.com/qualidafial/gtd-tui/sqlite"
)

type ProjectService struct {
	db *sqlite.DB
}

func NewProjectService(db *sqlite.DB) *ProjectService {
	return &ProjectService{db: db}
}

func (s *ProjectService) GetProject(ctx context.Context, id int64) (gtd.Project, error) {
	return s.db.GetProject(ctx, id)
}

func (s *ProjectService) ListProjects(ctx context.Context, filter gtd.ProjectFilter) ([]gtd.Project, error) {
	return s.db.ListProjects(ctx, filter)
}

func (s *ProjectService) CreateProject(ctx context.Context, project gtd.Project) (gtd.Project, error) {
	return s.db.CreateProject(ctx, project)
}

func (s *ProjectService) UpdateProject(ctx context.Context, project gtd.Project) (gtd.Project, error) {
	return s.db.UpdateProject(ctx, project)
}

func (s *ProjectService) CompleteProject(ctx context.Context, id int64, cascade bool, at time.Time) (gtd.Project, error) {
	return s.db.CompleteProject(ctx, id, cascade, at)
}

func (s *ProjectService) DropProject(ctx context.Context, id int64, cascade bool, at time.Time) (gtd.Project, error) {
	return s.db.DropProject(ctx, id, cascade, at)
}

func (s *ProjectService) ParkProject(ctx context.Context, id int64, at time.Time) (gtd.Project, error) {
	return s.db.ParkProject(ctx, id, at)
}

func (s *ProjectService) ReopenProject(ctx context.Context, id int64, at time.Time) (gtd.Project, error) {
	return s.db.ReopenProject(ctx, id, at)
}

func (s *ProjectService) MoveProjectUp(ctx context.Context, id int64) error {
	return s.db.MoveProjectUp(ctx, id)
}

func (s *ProjectService) MoveProjectDown(ctx context.Context, id int64) error {
	return s.db.MoveProjectDown(ctx, id)
}

func (s *ProjectService) CountTasksByProjects(ctx context.Context, projectIDs []int64) (map[int64]gtd.ProjectTaskCounts, error) {
	return s.db.CountTasksByProjects(ctx, projectIDs)
}

var _ gtd.ProjectService = (*ProjectService)(nil)