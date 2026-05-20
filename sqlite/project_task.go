package sqlite

// import (
// 	"context"
// 	"fmt"

// 	sq "github.com/Masterminds/squirrel"
// 	"github.com/qualidafial/gtd-tui"
// )

// var projectTaskColumns = []string{
// 	"project_id", "task_id", "created_at",
// }

// func (d *DB) AddTaskToProject(ctx context.Context, taskID, projectID int64) error {
// 	query, args, err := sq.Insert("project_tasks").
// 		Columns("task_id", "project_id").
// 		Values(taskID, projectID).
// 		ToSql()
// 	if err != nil {
// 		return err
// 	}
// 	if _, err := d.db.ExecContext(ctx, query, args...); err != nil {
// 		return fmt.Errorf("add task %d to project %d: %w", taskID, projectID, err)
// 	}
// 	return nil
// }

// func (d *DB) RemoveTaskFromProject(ctx context.Context, taskID, projectID int64) error {
// 	query, args, err := sq.Delete("project_tasks").
// 		Where(sq.Eq{
// 			"task_id":    taskID,
// 			"project_id": projectID,
// 		}).
// 		ToSql()
// 	if err != nil {
// 		return err
// 	}
// 	if _, err := d.db.ExecContext(ctx, query, args...); err != nil {
// 		return fmt.Errorf("remove task %d from project %d: %w", taskID, projectID, err)
// 	}
// 	return nil
// }

// func (d *DB) ProjectTasks(ctx context.Context, filter gtd.ProjectTaskFilter) (_ []gtd.ProjectTask, err error) {
// 	defer func() {
// 		if err != nil {
// 			err = fmt.Errorf("project_tasks: %w", err)
// 		}
// 	}()

// 	q := sq.Select(projectTaskColumns...)
// 	if len(filter.ProjectIDs) > 0 {
// 		q = q.Where(sq.Eq{"project_id": filter.ProjectIDs}).OrderBy("project_id")
// 	}
// 	if len(filter.TaskIDs) > 0 {
// 		q = q.Where(sq.Eq{"task_id": filter.TaskIDs}).OrderBy("task_id")
// 	}

// 	query, args, err := q.ToSql()
// 	if err != nil {
// 		return nil, err
// 	}

// 	rows, err := d.db.QueryContext(ctx, query, args...)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer rows.Close()

// 	var pts []gtd.ProjectTask
// 	for rows.Next() {
// 		var pt gtd.ProjectTask
// 		err = rows.Scan(&pt.ProjectID, &pt.TaskID, &pt.CreatedAt)
// 		if err != nil {
// 			return nil, err
// 		}
// 		pts = append(pts, pt)
// 	}
// 	return pts, nil
// }
