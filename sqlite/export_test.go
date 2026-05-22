package sqlite

import "context"

// SetOrderKeyForTest forces a task's order_key. Used by tests to provoke
// the renumber path in MoveToTop / MoveBetween.
func (d *DB) SetOrderKeyForTest(ctx context.Context, id int64, key string) error {
	return d.setOrderKey(ctx, id, key)
}
