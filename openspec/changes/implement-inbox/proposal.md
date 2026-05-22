## Why

The GTD TUI app needs its foundation: the Item entity and inbox capture workflow. Without a working inbox, users cannot capture thoughts, and without capture there is no GTD system. This is the first concrete functionality to implement after the architectural foundation specs are in place.

## What Changes

- Add `Item` domain type to the root package with title, description, timestamps, and `ClarifiedInto` pointer fields
- Define `InboxService` interface in the root package with Create, List, and Get operations
- Implement SQLite storage for items with migration, squirrel queries, and constraint enforcement
- Make inbox the default entry point when unprocessed items exist

## Capabilities

### New Capabilities

- `inbox-service`: Item entity definition, InboxService interface, and SQLite implementation for inbox capture operations (Create, List, Get)

### Modified Capabilities

(none - this change introduces new functionality without modifying existing capabilities)

## Impact

- **Root package**: New `item.go` file with Item struct and InboxService interface
- **sqlite/**: New `item.go` implementation file and migration adding `items` table
- **Dependencies**: No new dependencies (uses existing modernc.org/sqlite and squirrel)
- **TUI**: Future changes will wire inbox to UI; this change establishes the service layer
