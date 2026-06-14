## MODIFIED Requirements

### Requirement: Project reordering
ProjectService SHALL provide `MoveProjectUp(ctx context.Context, id int64, filter ProjectFilter) error` and `MoveProjectDown(ctx context.Context, id int64, filter ProjectFilter) error` to shift a project one position within projects of the same status that also match `filter`, and `MoveProjectFirst(ctx context.Context, id int64, filter ProjectFilter) error` and `MoveProjectLast(ctx context.Context, id int64, filter ProjectFilter) error` to move a project ahead of / after every same-status project that matches `filter`. The moving project's status group is always the universe — open projects are reordered among open projects; someday projects are reordered among someday projects. The supplied `filter` (Search, and any Status that matches the moving project's status) SHALL narrow the candidate neighbors further. Reordering uses fractional-indexed order keys (same `orderkey` package as tasks) with a renumber fallback when keys are exhausted; on exhaustion the renumber SHALL act on the entire same-status group (not just the filtered subset), preserving every non-moving project's relative position. All four moves SHALL be rejected for done/dropped projects. The new position SHALL be computed against the *filtered* set so a move is relative to the visible list; same-status projects outside the filter MAY interleave with filtered projects as a result, and on key exhaustion the moving project may visibly jump several positions in unfiltered views. `MoveProjectFirst` on the first filtered project and `MoveProjectLast` on the last filtered project SHALL be no-ops.

#### Scenario: Move open project up
- **WHEN** MoveProjectUp is called on an open project that is not first among open projects matching the filter
- **THEN** the project moves one position earlier among the filtered open projects

#### Scenario: Move someday project down
- **WHEN** MoveProjectDown is called on a someday project that is not last among someday projects matching the filter
- **THEN** the project moves one position later among the filtered someday projects

#### Scenario: Move open project first
- **WHEN** MoveProjectFirst is called on an open project that is not already first among open projects matching the filter
- **THEN** the project SHALL receive an order_key earlier than every other filtered open project
- **AND** open projects that do not match the filter SHALL retain their existing order_keys

#### Scenario: Move someday project last
- **WHEN** MoveProjectLast is called on a someday project that is not already last among someday projects matching the filter
- **THEN** the project SHALL receive an order_key later than every other filtered someday project
- **AND** projects that do not match the filter SHALL retain their existing order_keys

#### Scenario: Move to boundary is a no-op
- **WHEN** MoveProjectFirst is called on the first filtered project of its status group, or MoveProjectLast on the last
- **THEN** no order_keys SHALL change

#### Scenario: Reorder rejected for done/dropped project
- **WHEN** MoveProjectUp, MoveProjectDown, MoveProjectFirst, or MoveProjectLast is called on a done or dropped project
- **THEN** an error is returned

#### Scenario: Move down within a search filter
- **WHEN** MoveProjectDown is called on an open project with a Search filter that matches a subset of open projects
- **THEN** the project SHALL receive a new order_key between the next filtered project and the one after it
- **AND** open projects that do not match the filter SHALL retain their existing order_keys

#### Scenario: Key exhaustion renumbers the entire same-status group
- **WHEN** any project move is called and `orderkey.Between` cannot produce a key strictly between the filtered prev/next neighbors
- **THEN** every project in the moving project's status group SHALL be assigned a fresh evenly-spaced order_key in its current order, with the moving project slotted at its target position relative to its filtered neighbors
- **AND** the relative order of every non-moving project in that status group SHALL be preserved
