package service

import (
	"context"
	"fmt"

	"github.com/qualidafial/gtd-tui"
	"github.com/qualidafial/gtd-tui/sqlite"
)

// InboxService orchestrates inbox capture (Create/List/Get) and the four
// clarify operations (Discard, Incubate, ClarifyAsTask, ClarifyAsProject). Each
// clarify operation creates its destination entity and stamps the originating
// Item in a single transaction.
type InboxService struct {
	db *sqlite.DB
}

func NewInboxService(db *sqlite.DB) *InboxService {
	return &InboxService{db: db}
}

// Create persists a fresh inbox capture.
func (s *InboxService) Create(ctx context.Context, item gtd.Item) (gtd.Item, error) {
	return s.db.CreateItem(ctx, item)
}

// List returns unclarified, non-discarded items in FIFO order (oldest first).
func (s *InboxService) List(ctx context.Context) ([]gtd.Item, error) {
	return s.db.ListItems(ctx)
}

// Get fetches a single item by ID.
func (s *InboxService) Get(ctx context.Context, id int64) (gtd.Item, error) {
	return s.db.GetItem(ctx, id)
}

// Discard marks the item as non-actionable. No destination entity is created.
func (s *InboxService) Discard(ctx context.Context, itemID int64) (gtd.Item, error) {
	var out gtd.Item
	err := s.db.RunTx(ctx, func(ctx context.Context, tx *sqlite.DB) error {
		item, err := tx.GetItem(ctx, itemID)
		if err != nil {
			return err
		}
		if err := checkUnclarified(item); err != nil {
			return err
		}
		updated, err := tx.UpdateItemDiscarded(ctx, itemID)
		if err != nil {
			return err
		}
		out = updated
		return nil
	})
	if err != nil {
		return gtd.Item{}, fmt.Errorf("discard item %d: %w", itemID, err)
	}
	return out, nil
}

// Incubate spawns a Project in Status=someday for ideas to revisit later. No
// tasks are created: someday projects are dormant by definition. The project
// title and description fall back to the item's when not provided.
func (s *InboxService) Incubate(ctx context.Context, itemID int64, project gtd.Project) (gtd.Project, gtd.Item, error) {
	project.Status = gtd.ProjectStatusSomeday
	var outProject gtd.Project
	var outItem gtd.Item
	err := s.db.RunTx(ctx, func(ctx context.Context, tx *sqlite.DB) error {
		item, err := tx.GetItem(ctx, itemID)
		if err != nil {
			return err
		}
		if err := checkUnclarified(item); err != nil {
			return err
		}
		if project.Title == "" {
			project.Title = item.Title
		}
		if project.Description == "" {
			project.Description = item.Description
		}
		created, err := tx.CreateProject(ctx, project)
		if err != nil {
			return err
		}
		updated, err := tx.UpdateItemClarifiedIntoProject(ctx, itemID, created.ID)
		if err != nil {
			return err
		}
		outProject = created
		outItem = updated
		return nil
	})
	if err != nil {
		return gtd.Project{}, gtd.Item{}, fmt.Errorf("incubate item %d: %w", itemID, err)
	}
	return outProject, outItem, nil
}

// ClarifyAsProject spawns a Project in Status=open together with its first
// task, all in a single transaction. The first task is always created in
// Status=open; the wizard layer handles do-it-now by completing the task in a
// follow-up call after the user confirms. Subsequent tasks for the same
// project (added during the wizard's per-task loop) use TaskService.CreateTask
// directly with the new project's ID. This checkpoint shape guarantees that
// once the project is committed the item is clarified and no work is lost on
// accidental dismissal mid-loop.
//
// The caller MUST NOT pre-set firstTask.ProjectID — the new project's ID is
// stamped by this operation.
func (s *InboxService) ClarifyAsProject(ctx context.Context, itemID int64, project gtd.Project, firstTask gtd.Task) (gtd.Project, gtd.Task, gtd.Item, error) {
	if firstTask.ProjectID != nil {
		return gtd.Project{}, gtd.Task{}, gtd.Item{}, fmt.Errorf("clarify as project item %d: firstTask.ProjectID is owned by this operation", itemID)
	}
	if firstTask.Status != "" && firstTask.Status != gtd.TaskStatusOpen {
		return gtd.Project{}, gtd.Task{}, gtd.Item{}, fmt.Errorf("clarify as project item %d: firstTask.Status must be open (use CompleteTask after confirmation)", itemID)
	}

	project.Status = gtd.ProjectStatusOpen
	firstTask.Status = gtd.TaskStatusOpen

	var outProject gtd.Project
	var outTask gtd.Task
	var outItem gtd.Item
	err := s.db.RunTx(ctx, func(ctx context.Context, tx *sqlite.DB) error {
		item, err := tx.GetItem(ctx, itemID)
		if err != nil {
			return err
		}
		if err := checkUnclarified(item); err != nil {
			return err
		}
		if project.Title == "" {
			project.Title = item.Title
		}
		if project.Description == "" {
			project.Description = item.Description
		}
		created, err := tx.CreateProject(ctx, project)
		if err != nil {
			return err
		}
		firstTask.ProjectID = &created.ID
		createdTask, err := tx.CreateTask(ctx, firstTask)
		if err != nil {
			return err
		}
		updated, err := tx.UpdateItemClarifiedIntoProject(ctx, itemID, created.ID)
		if err != nil {
			return err
		}
		outProject = created
		outTask = createdTask
		outItem = updated
		return nil
	})
	if err != nil {
		return gtd.Project{}, gtd.Task{}, gtd.Item{}, fmt.Errorf("clarify as project item %d: %w", itemID, err)
	}
	return outProject, outTask, outItem, nil
}

// ClarifyAsTask spawns a Task targeting the optional ProjectID. The task is
// always created in Status=open; do-it-now is handled at the wizard layer by
// completing the task in a follow-up call after the user confirms. Title and
// description fall back to the item's when not provided.
func (s *InboxService) ClarifyAsTask(ctx context.Context, itemID int64, task gtd.Task) (gtd.Task, gtd.Item, error) {
	if task.Status != "" && task.Status != gtd.TaskStatusOpen {
		return gtd.Task{}, gtd.Item{}, fmt.Errorf("clarify task from item %d: task.Status must be open (use CompleteTask after confirmation)", itemID)
	}
	var outTask gtd.Task
	var outItem gtd.Item
	err := s.db.RunTx(ctx, func(ctx context.Context, tx *sqlite.DB) error {
		item, err := tx.GetItem(ctx, itemID)
		if err != nil {
			return err
		}
		if err := checkUnclarified(item); err != nil {
			return err
		}
		if task.Title == "" {
			task.Title = item.Title
		}
		if task.Description == "" {
			task.Description = item.Description
		}
		task.Status = gtd.TaskStatusOpen
		if task.ProjectID != nil {
			if _, err := tx.GetProject(ctx, *task.ProjectID); err != nil {
				return fmt.Errorf("invalid project: %w", err)
			}
		}
		created, err := tx.CreateTask(ctx, task)
		if err != nil {
			return err
		}
		updated, err := tx.UpdateItemClarifiedIntoTask(ctx, itemID, created.ID)
		if err != nil {
			return err
		}
		outTask = created
		outItem = updated
		return nil
	})
	if err != nil {
		return gtd.Task{}, gtd.Item{}, fmt.Errorf("clarify task from item %d: %w", itemID, err)
	}
	return outTask, outItem, nil
}

// checkUnclarified rejects items that already have a terminal state. Each
// clarify operation calls this inside its transaction so concurrent races still
// terminate at the CHECK constraint as a safety net.
func checkUnclarified(item gtd.Item) error {
	if item.Discarded {
		return fmt.Errorf("item %d already discarded", item.ID)
	}
	if item.ClarifiedIntoTaskID != nil {
		return fmt.Errorf("item %d already clarified into task %d", item.ID, *item.ClarifiedIntoTaskID)
	}
	if item.ClarifiedIntoProjectID != nil {
		return fmt.Errorf("item %d already clarified into project %d", item.ID, *item.ClarifiedIntoProjectID)
	}
	return nil
}

var _ gtd.InboxService = (*InboxService)(nil)
