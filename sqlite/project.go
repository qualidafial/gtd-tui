package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/qualidafial/gtd-tui"
	"github.com/qualidafial/gtd-tui/internal/orderkey"
)

var projectColumns = []string{
	"id", "title", "outcome", "description", "due", "status",
	"created_at", "updated_at", "status_changed_at",
}

func (d *DB) GetProject(ctx context.Context, id int64) (gtd.Project, error) {
	query, args, err := sq.Select(projectColumns...).From("projects").Where(sq.Eq{"id": id}).ToSql()
	if err != nil {
		return gtd.Project{}, err
	}
	project, err := scanProject(d.db.QueryRowContext(ctx, query, args...))
	if errors.Is(err, sql.ErrNoRows) {
		return gtd.Project{}, fmt.Errorf("project %d: not found", id)
	}
	if err != nil {
		return gtd.Project{}, fmt.Errorf("project %d: %w", id, err)
	}
	return project, nil
}

func (d *DB) ListProjects(ctx context.Context, filter gtd.ProjectFilter) ([]gtd.Project, error) {
	q := applyProjectListFilter(sq.Select(projectColumns...).From("projects"), filter)
	q = q.OrderBy(
		"CASE status WHEN 'open' THEN 0 WHEN 'someday' THEN 1 ELSE 2 END ASC",
		"order_key ASC",
		"status_changed_at DESC",
		"id ASC",
	)

	query, args, err := q.ToSql()
	if err != nil {
		return nil, err
	}
	rows, err := d.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("projects: %w", err)
	}
	defer rows.Close()

	var projects []gtd.Project
	for rows.Next() {
		project, err := scanProject(rows)
		if err != nil {
			return nil, fmt.Errorf("projects: %w", err)
		}
		projects = append(projects, project)
	}
	return projects, rows.Err()
}

// applyProjectListFilter translates a gtd.ProjectFilter into SQL WHERE clauses
// on q. Used by ListProjects and by shiftProject to compute filtered candidate
// orderings.
func applyProjectListFilter(q sq.SelectBuilder, filter gtd.ProjectFilter) sq.SelectBuilder {
	if filter.Status != nil {
		q = q.Where(sq.Eq{"status": string(*filter.Status)})
	}
	for _, term := range filter.Search {
		pattern := likeContains(term)
		q = q.Where("(lower(title) LIKE ? OR lower(outcome) LIKE ? OR lower(description) LIKE ?)",
			pattern, pattern, pattern)
	}
	return q
}

func isOrderedProjectStatus(s gtd.ProjectStatus) bool {
	return s == gtd.ProjectStatusOpen || s == gtd.ProjectStatusSomeday
}

func (d *DB) CreateProject(ctx context.Context, project gtd.Project) (gtd.Project, error) {
	now := time.Now().UTC()
	project.CreatedAt = now
	project.UpdatedAt = now
	project.StatusChangedAt = now
	if project.Status == "" {
		project.Status = gtd.ProjectStatusOpen
	}

	var key any
	if isOrderedProjectStatus(project.Status) {
		k, err := d.nextProjectOrderKey(ctx, project.Status)
		if err != nil {
			return gtd.Project{}, fmt.Errorf("create project: %w", err)
		}
		key = k
	}

	query, args, err := sq.Insert("projects").
		Columns("title", "outcome", "description", "due", "status", "order_key", "created_at", "updated_at", "status_changed_at").
		Values(project.Title, project.Outcome, project.Description, nullTime(project.Due),
			string(project.Status), key, project.CreatedAt, project.UpdatedAt, project.StatusChangedAt).
		ToSql()
	if err != nil {
		return gtd.Project{}, err
	}
	res, err := d.db.ExecContext(ctx, query, args...)
	if err != nil {
		return gtd.Project{}, fmt.Errorf("create project: %w", err)
	}
	project.ID, err = res.LastInsertId()
	if err != nil {
		return gtd.Project{}, fmt.Errorf("create project: %w", err)
	}
	return project, nil
}

func (d *DB) UpdateProject(ctx context.Context, project gtd.Project) (gtd.Project, error) {
	project.UpdatedAt = time.Now().UTC()

	err := d.RunTx(ctx, func(ctx context.Context, tx *DB) error {
		current, err := tx.GetProject(ctx, project.ID)
		if err != nil {
			return err
		}
		if project.Status != current.Status {
			return fmt.Errorf("UpdateProject cannot change status; use CompleteProject/DropProject/ParkProject/ReopenProject")
		}
		project.StatusChangedAt = current.StatusChangedAt

		query, args, err := sq.Update("projects").
			Set("title", project.Title).
			Set("outcome", project.Outcome).
			Set("description", project.Description).
			Set("due", nullTime(project.Due)).
			Set("updated_at", project.UpdatedAt).
			Where(sq.Eq{"id": project.ID}).
			ToSql()
		if err != nil {
			return err
		}
		_, err = tx.db.ExecContext(ctx, query, args...)
		return err
	})
	if err != nil {
		return gtd.Project{}, fmt.Errorf("update project %d: %w", project.ID, err)
	}
	return project, nil
}

// DeleteProject removes a project row. Callers are responsible for ensuring no
// tasks reference it; ConvertProjectToTask guards this by requiring the project
// to be empty.
func (d *DB) DeleteProject(ctx context.Context, id int64) error {
	query, args, err := sq.Delete("projects").Where(sq.Eq{"id": id}).ToSql()
	if err != nil {
		return err
	}
	if _, err := d.db.ExecContext(ctx, query, args...); err != nil {
		return fmt.Errorf("delete project %d: %w", id, err)
	}
	return nil
}

func (d *DB) CompleteProject(ctx context.Context, id int64, cascade bool, at time.Time) (gtd.Project, error) {
	return d.transitionProject(ctx, id, gtd.ProjectStatusDone, at, cascade, gtd.TaskStatusDone)
}

func (d *DB) DropProject(ctx context.Context, id int64, cascade bool, at time.Time) (gtd.Project, error) {
	return d.transitionProject(ctx, id, gtd.ProjectStatusDropped, at, cascade, gtd.TaskStatusDropped)
}

func (d *DB) ParkProject(ctx context.Context, id int64, at time.Time) (gtd.Project, error) {
	return d.transitionProject(ctx, id, gtd.ProjectStatusSomeday, at, false, "")
}

func (d *DB) ReopenProject(ctx context.Context, id int64, at time.Time) (gtd.Project, error) {
	return d.transitionProject(ctx, id, gtd.ProjectStatusOpen, at, false, "")
}

// transitionProject atomically sets the project's status and stamps
// status_changed_at with at. For terminal transitions (done/dropped) it either
// cascades the supplied task status onto pending tasks or detaches them,
// preserving the invariant that no pending tasks remain under a closed project.
// Assigns a fresh order_key when entering open or someday; clears it for
// done/dropped.
func (d *DB) transitionProject(ctx context.Context, id int64, newStatus gtd.ProjectStatus, at time.Time, cascade bool, taskStatus gtd.TaskStatus) (gtd.Project, error) {
	var project gtd.Project
	err := d.RunTx(ctx, func(ctx context.Context, tx *DB) error {
		current, err := tx.GetProject(ctx, id)
		if err != nil {
			return err
		}

		now := time.Now().UTC()
		statusChangedAt := at.UTC()

		update := sq.Update("projects").
			Set("status", string(newStatus)).
			Set("updated_at", now).
			Set("status_changed_at", statusChangedAt).
			Where(sq.Eq{"id": id})

		if isOrderedProjectStatus(newStatus) {
			key, err := tx.nextProjectOrderKey(ctx, newStatus)
			if err != nil {
				return err
			}
			update = update.Set("order_key", key)
		} else {
			update = update.Set("order_key", nil)
		}

		query, args, err := update.ToSql()
		if err != nil {
			return err
		}
		if _, err := tx.db.ExecContext(ctx, query, args...); err != nil {
			return err
		}

		if taskStatus != "" {
			if cascade {
				if err := tx.cascadeTaskStatus(ctx, id, taskStatus, statusChangedAt); err != nil {
					return err
				}
			} else {
				if err := tx.detachPendingTasks(ctx, id); err != nil {
					return err
				}
			}
		}

		current.Status = newStatus
		current.UpdatedAt = now
		current.StatusChangedAt = statusChangedAt
		project = current
		return nil
	})
	if err != nil {
		return gtd.Project{}, fmt.Errorf("transition project %d: %w", id, err)
	}
	return project, nil
}

// cascadeTaskStatus marks all pending tasks under the project with the terminal
// status, clearing their order_key and stamping status_changed_at = at.
func (d *DB) cascadeTaskStatus(ctx context.Context, projectID int64, status gtd.TaskStatus, at time.Time) error {
	query, args, err := sq.Update("tasks").
		Set("status", string(status)).
		Set("order_key", nil).
		Set("updated_at", time.Now().UTC()).
		Set("status_changed_at", at).
		Where(sq.Eq{"project_id": projectID, "status": string(gtd.TaskStatusOpen)}).
		ToSql()
	if err != nil {
		return err
	}
	_, err = d.db.ExecContext(ctx, query, args...)
	return err
}

// detachPendingTasks sets project_id = NULL for all pending tasks under the
// project, making them standalone. Done/dropped tasks stay attached.
func (d *DB) detachPendingTasks(ctx context.Context, projectID int64) error {
	query, args, err := sq.Update("tasks").
		Set("project_id", nil).
		Set("updated_at", time.Now().UTC()).
		Where(sq.Eq{"project_id": projectID, "status": string(gtd.TaskStatusOpen)}).
		ToSql()
	if err != nil {
		return err
	}
	_, err = d.db.ExecContext(ctx, query, args...)
	return err
}

// MoveProjectUp shifts the project one slot earlier within projects of the
// same status that match filter. No-op when already at the top of the filtered
// set.
func (d *DB) MoveProjectUp(ctx context.Context, id int64, filter gtd.ProjectFilter) error {
	return d.RunTx(ctx, func(ctx context.Context, tx *DB) error {
		return tx.shiftProject(ctx, id, -1, filter)
	})
}

// MoveProjectDown shifts the project one slot later within projects of the
// same status that match filter. No-op when already at the bottom of the
// filtered set.
func (d *DB) MoveProjectDown(ctx context.Context, id int64, filter gtd.ProjectFilter) error {
	return d.RunTx(ctx, func(ctx context.Context, tx *DB) error {
		return tx.shiftProject(ctx, id, +1, filter)
	})
}

// MoveProjectFirst moves the project ahead of every same-status project
// matching filter. No-op when already first in the filtered set; rejected for
// done or dropped projects.
func (d *DB) MoveProjectFirst(ctx context.Context, id int64, filter gtd.ProjectFilter) error {
	return d.RunTx(ctx, func(ctx context.Context, tx *DB) error {
		status, others, pos, err := tx.projectReorderState(ctx, id, filter)
		if err != nil {
			return err
		}
		if pos == 0 {
			return nil
		}
		return tx.moveProjectTo(ctx, id, 0, status, others)
	})
}

// MoveProjectLast moves the project after every same-status project
// matching filter. No-op when already last in the filtered set; rejected for
// done or dropped projects.
func (d *DB) MoveProjectLast(ctx context.Context, id int64, filter gtd.ProjectFilter) error {
	return d.RunTx(ctx, func(ctx context.Context, tx *DB) error {
		status, others, pos, err := tx.projectReorderState(ctx, id, filter)
		if err != nil {
			return err
		}
		if pos == len(others) {
			return nil
		}
		return tx.moveProjectTo(ctx, id, len(others), status, others)
	})
}

// shiftProject moves id by delta slots within the same-status projects
// matching filter.
func (d *DB) shiftProject(ctx context.Context, id int64, delta int, filter gtd.ProjectFilter) error {
	status, others, pos, err := d.projectReorderState(ctx, id, filter)
	if err != nil {
		return err
	}
	return d.moveProjectTo(ctx, id, pos+delta, status, others)
}

// projectReorderState loads the reorder context for an orderable project: its
// status, the same-status projects matching filter (excluding id, ordered by
// key), and id's current insertion index among them. Returns an error for
// done or dropped projects.
func (d *DB) projectReorderState(ctx context.Context, id int64, filter gtd.ProjectFilter) (gtd.ProjectStatus, []statusEntry, int, error) {
	project, err := d.GetProject(ctx, id)
	if err != nil {
		return "", nil, 0, err
	}
	if !isOrderedProjectStatus(project.Status) {
		return "", nil, 0, fmt.Errorf("project %d: cannot reorder %s projects", id, project.Status)
	}

	key, err := d.projectOrderKey(ctx, id)
	if err != nil {
		return "", nil, 0, err
	}
	others, err := d.projectOrderByStatus(ctx, project.Status, filter, id)
	if err != nil {
		return "", nil, 0, err
	}

	pos := 0
	for _, o := range others {
		if o.key < key {
			pos++
		}
	}
	return project.Status, others, pos, nil
}

// moveProjectTo slots id at insertion index newPos among others (the filtered
// same-status neighbors, excluding id). The fast path slots a new key via
// orderkey.Between against the filtered neighbors; on exhaustion the entire
// same-status group is renumbered, preserving the relative order of every
// non-moving project. A newPos out of range is a no-op.
func (d *DB) moveProjectTo(ctx context.Context, id int64, newPos int, status gtd.ProjectStatus, others []statusEntry) error {
	if newPos < 0 || newPos > len(others) {
		return nil
	}

	var prevKey, nextKey string
	if newPos > 0 {
		prevKey = others[newPos-1].key
	}
	if newPos < len(others) {
		nextKey = others[newPos].key
	}

	if newKey, ok := orderkey.Between(prevKey, nextKey); ok {
		return d.setProjectOrderKey(ctx, id, newKey)
	}

	// Exhaustion fallback: renumber the entire same-status group. Splice the
	// moving id into the full ordering between its filtered prev/next neighbors
	// so the visible move is preserved and every other project keeps its
	// relative position.
	full, err := d.projectOrderByStatus(ctx, status, gtd.ProjectFilter{}, id)
	if err != nil {
		return err
	}
	insertAt := 0
	if newPos > 0 {
		prevID := others[newPos-1].id
		for i, e := range full {
			if e.id == prevID {
				insertAt = i + 1
				break
			}
		}
	}
	keys := orderkey.Renumber(len(full) + 1)
	for i, p := range full[:insertAt] {
		if err := d.setProjectOrderKey(ctx, p.id, keys[i]); err != nil {
			return err
		}
	}
	if err := d.setProjectOrderKey(ctx, id, keys[insertAt]); err != nil {
		return err
	}
	for i, p := range full[insertAt:] {
		if err := d.setProjectOrderKey(ctx, p.id, keys[insertAt+1+i]); err != nil {
			return err
		}
	}
	return nil
}

// projectOrderByStatus returns id/order_key pairs for the projects with the
// given status that match filter, ordered by (order_key, id) and excluding
// excludeID. Status in filter is always overridden to the supplied status.
// Pass an empty ProjectFilter to load the entire same-status group.
func (d *DB) projectOrderByStatus(ctx context.Context, status gtd.ProjectStatus, filter gtd.ProjectFilter, excludeID int64) ([]statusEntry, error) {
	filter.Status = &status
	q := applyProjectListFilter(sq.Select("id", "order_key").From("projects"), filter)
	if excludeID != 0 {
		q = q.Where(sq.NotEq{"id": excludeID})
	}
	q = q.OrderBy("order_key ASC", "id ASC")
	return scanStatusEntries(ctx, d, q)
}

func (d *DB) nextProjectOrderKey(ctx context.Context, status gtd.ProjectStatus) (string, error) {
	var maxKey sql.NullString
	query, args, err := sq.Select("MAX(order_key)").From("projects").
		Where(sq.Eq{"status": string(status)}).ToSql()
	if err != nil {
		return "", err
	}
	if err := d.db.QueryRowContext(ctx, query, args...).Scan(&maxKey); err != nil {
		return "", err
	}
	key, ok := orderkey.Between(maxKey.String, "")
	if !ok {
		return "", fmt.Errorf("order keys exhausted")
	}
	return key, nil
}

func (d *DB) projectOrderKey(ctx context.Context, id int64) (string, error) {
	query, args, err := sq.Select("order_key").From("projects").Where(sq.Eq{"id": id}).ToSql()
	if err != nil {
		return "", err
	}
	var key sql.NullString
	err = d.db.QueryRowContext(ctx, query, args...).Scan(&key)
	if errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("project %d: not found", id)
	}
	if err != nil {
		return "", err
	}
	return key.String, nil
}

func (d *DB) setProjectOrderKey(ctx context.Context, id int64, key string) error {
	query, args, err := sq.Update("projects").
		Set("order_key", key).
		Where(sq.Eq{"id": id}).
		ToSql()
	if err != nil {
		return err
	}
	res, err := d.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("project %d: not found", id)
	}
	return nil
}

func (d *DB) CountTasksByProjects(ctx context.Context, projectIDs []int64) (map[int64]gtd.ProjectTaskCounts, error) {
	if len(projectIDs) == 0 {
		return nil, nil
	}
	query, args, err := sq.Select(
		"project_id",
		"SUM(CASE WHEN status = 'done' THEN 1 ELSE 0 END) AS complete",
		"COUNT(*) AS total",
	).
		From("tasks").
		Where(sq.And{
			sq.Eq{"project_id": projectIDs},
			sq.NotEq{"status": string(gtd.TaskStatusDropped)},
		}).
		GroupBy("project_id").
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := d.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("count tasks by projects: %w", err)
	}
	defer rows.Close()

	counts := make(map[int64]gtd.ProjectTaskCounts, len(projectIDs))
	for rows.Next() {
		var projectID int64
		var c gtd.ProjectTaskCounts
		if err := rows.Scan(&projectID, &c.Complete, &c.Total); err != nil {
			return nil, fmt.Errorf("count tasks by projects: %w", err)
		}
		counts[projectID] = c
	}
	return counts, rows.Err()
}

func scanProject(s scanner) (gtd.Project, error) {
	var project gtd.Project
	var due sql.NullTime
	err := s.Scan(
		&project.ID, &project.Title, &project.Outcome, &project.Description,
		&due, &project.Status, &project.CreatedAt, &project.UpdatedAt, &project.StatusChangedAt,
	)
	if err != nil {
		return gtd.Project{}, err
	}
	if due.Valid {
		project.Due = &due.Time
	}
	return project, nil
}