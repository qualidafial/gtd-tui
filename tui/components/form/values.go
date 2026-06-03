package form

// Values is an immutable snapshot of the visible prior fields' values
// supplied to a field's Visible predicate.
type Values interface {
	// Get returns the Value() of the visible preceding field whose Key()
	// matches key, or nil if no such field is in the snapshot. A field's
	// own Value is never present in the snapshot supplied to its own
	// Visible call; hidden fields are excluded from snapshots supplied to
	// later fields.
	Get(key string) any
}

type valuesMap map[string]any

func (v valuesMap) Get(k string) any { return v[k] }
