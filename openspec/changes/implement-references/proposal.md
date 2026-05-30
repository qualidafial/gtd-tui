## Why

The capture + clarify loop landed in `implement-inbox` covers Tasks and Projects but not references — standalone retrieval content (recipes, snippets, links) that should never become work to do. This change adds the Reference entity, extends the clarify loop with `FileAsReference`, and surfaces references as their own tab.

## What Changes

- Add `Reference` domain type and `ReferenceStore` interface in the root package
- Add SQLite migration for the `references` table with CHECK constraint on non-empty title; add `sqlite/reference.go` implementing ReferenceStore
- Add `ClarifiedIntoReferenceID` (*int64) to the `Item` struct; ALTER `items` to add the column and extend the existing mutual-exclusion CHECK constraint to include it
- Add `FileAsReference` operation to the service-layer InboxService; extend the InboxService constructor to accept a `ReferenceStore`
- Add a references screen at `tui/pages/references/` with querybar-driven title filtering, registered as a tab in the root tabContainer

## Capabilities

### New Capabilities

- `reference-entity`: Reference entity, ReferenceStore interface, and SQLite implementation
- `references-page`: References screen with title filtering via the shared querybar, registered as a tab

### Modified Capabilities

- `inbox-service`: Item gains `ClarifiedIntoReferenceID`; items table gains the `clarified_into_reference_id` column and its CHECK arm
- `clarify-operations`: InboxService constructor gains ReferenceStore; FileAsReference operation added; mutual-exclusion requirement extends to ClarifiedIntoReferenceID

## Impact

- **Root package**: New `reference.go`; `item.go` gains `ClarifiedIntoReferenceID` field
- **sqlite/**: New `reference.go`; new migration creating `references` and altering `items` to add the FK column + extended CHECK
- **service/**: `service/inbox.go` constructor signature extended; new `FileAsReference` method
- **Dependencies**: Depends on `implement-inbox` for the base Item entity, items table, and service-layer InboxService
- **TUI**: New `tui/pages/references/` screen wired into `tui.New` as a tab
