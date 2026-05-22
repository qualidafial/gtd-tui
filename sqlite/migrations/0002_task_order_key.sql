ALTER TABLE tasks ADD COLUMN order_key TEXT;

UPDATE tasks
SET order_key = printf('%016d', (
    SELECT COUNT(*) FROM tasks AS t2
    WHERE t2.status = tasks.status
      AND (
          t2.created_at < tasks.created_at
          OR (t2.created_at = tasks.created_at AND t2.id < tasks.id)
      )
) + 1)
WHERE status NOT IN ('done', 'dropped');

CREATE INDEX idx_tasks_status_order_key ON tasks(status, order_key);
