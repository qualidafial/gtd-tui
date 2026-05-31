## Why

The GTD TUI app needs both halves of the inbox loop in one slice: capturing thoughts and processing them into Tasks and Projects. An Item entity without the clarify operations is dead weight — items pile up with no way to move them forward — and the clarify operations have no input source without Item capture. This change ships both: the Item entity and capture API, plus the four clarify operations that drain the inbox into Tasks and Projects. Reference handling is a follow-up scope (see `implement-references`).

Someday items are not a separate entity. They are projects in `Status=someday`, per the finalized `project-entity` and `project-service` specs. Incubate reuses that machinery instead of introducing a parallel Someday type.

## What Changes

- Add `Item` domain type to the root package with title, description, timestamps, `ClarifiedIntoTaskID` / `ClarifiedIntoProjectID` nullable FKs, and `Discarded` flag
- Define the capture-side `InboxService` in the root package with Create / List / Get
- Implement an `InboxService` in `service/` with four clarify operations:
  - `Discard` — mark item as discarded; no destination entity
  - `Incubate` — create a Project with `Status=someday`, stamp `Item.ClarifiedIntoProjectID`
  - `ClarifyAsTask` — create a Task (kind + optional project), stamp `Item.ClarifiedIntoTaskID`
  - `ClarifyAsProject` — create a Project with `Status=open`, stamp `Item.ClarifiedIntoProjectID`
- Each clarify operation is transactional: destination entity creation and Item stamping occur atomically
- Add SQLite migration for the `items` table with CHECK constraints (non-empty title, mutual exclusion of `ClarifiedIntoTaskID` / `ClarifiedIntoProjectID` / `Discarded`). `ClarifiedIntoReferenceID` is added later by `implement-references`.
- Add an inbox screen at `tui/pages/inbox/` registered as a tab in the root tabContainer. Someday projects surface in the existing projects tab via the `status:someday` query filter — no separate page.

## Capabilities

### New Capabilities

- `inbox-service`: Item entity, capture-side InboxService interface, and SQLite implementation for Create / List / Get
- `clarify-operations`: The four transactional clarify operations on the service-layer InboxService (Discard, Incubate, ClarifyAsTask, ClarifyAsProject). FileAsReference is added later by `implement-references`.
- `inbox-page`: Inbox screen rendering the unclarified item list, registered as a tab

### Modified Capabilities

(none — this change introduces new functionality without modifying existing capabilities)

## Impact

- **Root package**: New `item.go` with Item struct and capture-side InboxService interface
- **sqlite/**: New `item.go` and migration adding `items` table
- **service/**: New `service/inbox.go` orchestrating clarify operations across ItemStore / TaskStore / ProjectStore with a shared transaction
- **Dependencies**: No new external dependencies; depends on the finalized `project-entity` / `project-service` specs for `ProjectStatusSomeday` semantics consumed by Incubate
- **TUI**: New `tui/pages/inbox/` screen wired into `tui.New` as a tab
- **Follow-up**: `implement-references` extends the items table with `ClarifiedIntoReferenceID`, the CHECK constraint, and adds the FileAsReference operation + references TUI tab
