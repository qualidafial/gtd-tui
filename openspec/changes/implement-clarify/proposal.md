## Why

The GTD methodology requires processing inbox items into actionable outcomes. The foundation specs define the clarify workflow requirements, but implementation is needed to let users transform captured items into tasks, projects, or references. Without clarify, items accumulate in the inbox with no way to move them forward.

Someday items are not a distinct entity — they are projects in `someday` status. The Project entity already carries `ProjectStatusSomeday` and `ParkProject`/`ReopenProject` transitions (per the finalized `project-entity` and `project-service` specs), so `Incubate` reuses that machinery instead of introducing a parallel Someday type.

## What Changes

- Implement `Reference` entity for standalone retrieval content (recipes, config snippets, links)
- Implement `InboxService` with the 5 clarify operations:
  - `Discard` - mark item as discarded (non-actionable, unwanted)
  - `Incubate` - create a Project with `Status=someday`, link via `Item.ClarifiedIntoProjectID`
  - `FileAsReference` - create a Reference, link via `Item.ClarifiedIntoReferenceID`
  - `ClarifyAsTask` - create a Task with kind/project options, link via `Item.ClarifiedIntoTaskID`
  - `ClarifyAsProject` - create a Project with `Status=open`, link via `Item.ClarifiedIntoProjectID`
- Each clarify operation is transactional: destination entity creation + Item.ClarifiedInto stamp occur atomically
- Add SQLite migration and implementation for the references table / ReferenceStore
- Implement InboxService in service/ package for cross-entity orchestration
- Add a references screen at `tui/pages/references/` and register it as a tab. Someday projects surface in the existing projects tab via the `status:someday` query filter — no separate page

## Capabilities

### New Capabilities
- `clarify-operations`: The 5 transactional clarify operations on InboxService that transform inbox items into destination entities
- `reference-entity`: Reference entity for standalone retrieval content
- `references-page`: References screen with title filtering via the shared querybar and registered as a tab

### Modified Capabilities


## Impact

- **Domain types**: Add `Reference` struct + service interface in the root package; reuse existing `Project` for incubated items
- **Item entity**: Requires `Item.ClarifiedIntoReferenceID` nullable FK field (assumes implement-inbox adds the base Item entity and `ClarifiedIntoTaskID`/`ClarifiedIntoProjectID`). `ClarifiedIntoSomedayID` is NOT needed — incubated items use `ClarifiedIntoProjectID`
- **SQLite layer**: New migration for references table, new ReferenceStore implementation
- **Service layer**: New InboxService in service/ package orchestrating clarify transactions
- **Dependencies**: Depends on implement-inbox for Item entity and base ClarifiedInto fields; depends on the existing `project-service` spec for `ProjectStatusSomeday` semantics
- **TUI**: New `tui/pages/references/` screen wired into `tui.New` as a tab
