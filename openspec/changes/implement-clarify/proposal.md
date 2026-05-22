## Why

The GTD methodology requires processing inbox items into actionable outcomes. The foundation specs define the clarify workflow requirements, but implementation is needed to let users transform captured items into tasks, projects, someday ideas, or references. Without clarify, items accumulate in the inbox with no way to move them forward.

## What Changes

- Implement `Someday` entity for parked ideas with ReviewedAt timestamp for periodic review surfacing
- Implement `Reference` entity for standalone retrieval content (recipes, config snippets, links)
- Implement `InboxService` with the 5 clarify operations:
  - `Discard` - mark item as discarded (non-actionable, unwanted)
  - `Incubate` - create Someday, link via Item.ClarifiedIntoSomedayID
  - `FileAsReference` - create Reference, link via Item.ClarifiedIntoReferenceID
  - `ClarifyAsTask` - create Task with kind/project options, link via Item.ClarifiedIntoTaskID
  - `ClarifyAsProject` - create Project, link via Item.ClarifiedIntoProjectID
- Each clarify operation is transactional: destination entity creation + Item.ClarifiedInto stamp occur atomically
- Add SQLite migrations for someday and reference tables
- Add SQLite implementations for SomedayStore and ReferenceStore
- Implement InboxService in service/ package for cross-entity orchestration

## Capabilities

### New Capabilities
- `clarify-operations`: The 5 transactional clarify operations on InboxService that transform inbox items into destination entities
- `someday-entity`: Someday entity for parked ideas with ReviewedAt timestamp and promotion support
- `reference-entity`: Reference entity for standalone retrieval content

### Modified Capabilities


## Impact

- **Domain types**: Add `Someday`, `Reference` structs in root package with service interfaces
- **Item entity**: Requires Item.ClarifiedIntoSomedayID and Item.ClarifiedIntoReferenceID nullable FK fields (assumes implement-inbox adds the base Item entity and ClarifiedIntoTaskID/ProjectID)
- **SQLite layer**: New migrations for someday/reference tables, new store implementations
- **Service layer**: New InboxService in service/ package orchestrating clarify transactions
- **Dependencies**: Depends on implement-inbox for Item entity and base ClarifiedInto fields
