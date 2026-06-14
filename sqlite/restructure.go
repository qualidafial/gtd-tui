package sqlite

import (
	"context"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/qualidafial/gtd-tui"
)

// ConvertTaskToProject promotes a standalone task into a new open project,
// keeping the task as the project's first action. The new project's Title and
// Description default from the task when left empty; project.Status is forced
// to open. The task is re-parented under the new project and its Title and
// Description are replaced with the reframed values (an empty reframed Title
// keeps the original, since title is required); all other task fields are
// preserved. The whole conversion is one transaction.
func (d *DB) ConvertTaskToProject(ctx context.Context, taskID int64, project gtd.Project, reframed gtd.Task) (gtd.Project, gtd.Task, error) {
	var outProject gtd.Project
	var outTask gtd.Task
	err := d.RunTx(ctx, func(ctx context.Context, tx *DB) error {
		task, err := tx.GetTask(ctx, taskID)
		if err != nil {
			return err
		}
		if task.ProjectID != nil {
			return fmt.Errorf("task %d already belongs to project %d", taskID, *task.ProjectID)
		}

		if project.Title == "" {
			project.Title = task.Title
		}
		if project.Description == "" {
			project.Description = task.Description
		}
		project.Status = gtd.ProjectStatusOpen
		created, err := tx.CreateProject(ctx, project)
		if err != nil {
			return err
		}

		if reframed.Title != "" {
			task.Title = reframed.Title
		}
		task.Description = reframed.Description
		task.ProjectID = &created.ID
		task.UpdatedAt = time.Now().UTC()
		query, args, err := sq.Update("tasks").
			Set("title", task.Title).
			Set("description", task.Description).
			Set("project_id", task.ProjectID).
			Set("updated_at", task.UpdatedAt).
			Where(sq.Eq{"id": taskID}).
			ToSql()
		if err != nil {
			return err
		}
		if _, err := tx.db.ExecContext(ctx, query, args...); err != nil {
			return err
		}

		outProject = created
		outTask = task
		return nil
	})
	if err != nil {
		return gtd.Project{}, gtd.Task{}, fmt.Errorf("convert task %d to project: %w", taskID, err)
	}
	return outProject, outTask, nil
}

// ConvertProjectToTask collapses an empty open project into a standalone task.
// It is guarded to projects with Status=open that have zero tasks of any
// status — the only lossless case. The new task inherits Title, Description and
// Due; a non-empty project Outcome is folded into the task Description so no
// content is lost. The project row is deleted. One transaction.
func (d *DB) ConvertProjectToTask(ctx context.Context, projectID int64) (gtd.Task, error) {
	var outTask gtd.Task
	err := d.RunTx(ctx, func(ctx context.Context, tx *DB) error {
		project, err := tx.GetProject(ctx, projectID)
		if err != nil {
			return err
		}
		if project.Status != gtd.ProjectStatusOpen {
			return fmt.Errorf("project %d: only open projects can convert to a task (status is %s)", projectID, project.Status)
		}
		n, err := tx.countProjectTasks(ctx, projectID)
		if err != nil {
			return err
		}
		if n > 0 {
			return fmt.Errorf("project %d: only empty projects can convert to a task (%d task(s) attached)", projectID, n)
		}

		task := gtd.Task{
			Title:       project.Title,
			Description: foldOutcome(project.Description, project.Outcome),
			Status:      gtd.TaskStatusOpen,
			Due:         project.Due,
		}
		created, err := tx.CreateTask(ctx, task)
		if err != nil {
			return err
		}
		if err := tx.DeleteProject(ctx, projectID); err != nil {
			return err
		}

		outTask = created
		return nil
	})
	if err != nil {
		return gtd.Task{}, fmt.Errorf("convert project %d to task: %w", projectID, err)
	}
	return outTask, nil
}

// LinkTaskToProject re-parents an existing standalone task into a project,
// placing it at the bottom of the project's task ordering. It rejects tasks
// that already belong to a project and non-existent projects. An open task is
// given a fresh order key (the global bottom, which is also the bottom of the
// project's filtered view); a closed task keeps its NULL order key. One
// transaction.
func (d *DB) LinkTaskToProject(ctx context.Context, taskID, projectID int64) (gtd.Task, error) {
	var outTask gtd.Task
	err := d.RunTx(ctx, func(ctx context.Context, tx *DB) error {
		task, err := tx.GetTask(ctx, taskID)
		if err != nil {
			return err
		}
		if task.ProjectID != nil {
			return fmt.Errorf("task %d already belongs to project %d", taskID, *task.ProjectID)
		}
		if _, err := tx.GetProject(ctx, projectID); err != nil {
			return err
		}

		task.ProjectID = &projectID
		task.UpdatedAt = time.Now().UTC()
		update := sq.Update("tasks").
			Set("project_id", projectID).
			Set("updated_at", task.UpdatedAt).
			Where(sq.Eq{"id": taskID})
		if !isClosedStatus(task.Status) {
			key, err := tx.nextTaskOrderKey(ctx)
			if err != nil {
				return err
			}
			update = update.Set("order_key", key)
		}
		query, args, err := update.ToSql()
		if err != nil {
			return err
		}
		if _, err := tx.db.ExecContext(ctx, query, args...); err != nil {
			return err
		}

		outTask = task
		return nil
	})
	if err != nil {
		return gtd.Task{}, fmt.Errorf("link task %d to project %d: %w", taskID, projectID, err)
	}
	return outTask, nil
}

// countProjectTasks returns the number of tasks attached to the project,
// counting every status (open, done, dropped).
func (d *DB) countProjectTasks(ctx context.Context, projectID int64) (int, error) {
	query, args, err := sq.Select("COUNT(*)").From("tasks").
		Where(sq.Eq{"project_id": projectID}).ToSql()
	if err != nil {
		return 0, err
	}
	var n int
	if err := d.db.QueryRowContext(ctx, query, args...).Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}

// foldOutcome appends a non-empty project outcome to the task description so
// the outcome is preserved when a project collapses to a task.
func foldOutcome(description, outcome string) string {
	if outcome == "" {
		return description
	}
	if description == "" {
		return outcome
	}
	return description + "\n\n" + outcome
}
