## 1. Item Domain Layer

- [x] 1.1 Create `item.go` in root package with Item struct (ID, Title, Description, CreatedAt, UpdatedAt, ClarifiedIntoTaskID, ClarifiedIntoProjectID, Discarded). `ClarifiedIntoReferenceID` is added by `implement-references`.
- [x] 1.2 Define the capture-side `InboxService` interface in `item.go` (Create, List, Get)

## 2. Database Migration

- [x] 2.1 Create migration for the `items` table with non-empty-title CHECK
- [x] 2.2 Add CHECK constraint enforcing mutual exclusion of `clarified_into_task_id` / `clarified_into_project_id` / `discarded` (extended by `implement-references` to include `clarified_into_reference_id`)

## 3. SQLite Implementation

- [x] 3.1 Create `sqlite/item.go` implementing the capture-side InboxService
- [x] 3.2 Implement Create with squirrel insert and UTC timestamps
- [x] 3.3 Implement List filtering unclarified/non-discarded items, ordered by created_at ASC
- [x] 3.4 Implement Get with squirrel select by ID
- [x] 3.5 Add `scanItem` helper following the existing `scanTask` pattern

## 4. SQLite Tests

- [x] 4.1 Create `sqlite/item_test.go` covering Create (valid + empty-title rejection), List (FIFO, only unclarified), Get (valid + missing)
- [x] 4.2 Test the mutual-exclusion CHECK constraint with direct SQL violations

## 5. Service-layer InboxService Setup

- [x] 5.1 Create `service/inbox.go` with InboxService struct
- [x] 5.2 Constructor accepts ItemStore, TaskStore, ProjectStore
- [x] 5.3 Wire a transaction provider for atomic operations
- [x] 5.4 Add a helper that checks whether an Item is already clarified or discarded

## 6. Discard Operation

- [x] 6.1 Implement InboxService.Discard
- [x] 6.2 Reject non-existent items
- [x] 6.3 Reject already-clarified items
- [x] 6.4 Tests including error cases

## 7. Incubate Operation

- [x] 7.1 Implement InboxService.Incubate creating a Project with `Status=someday` and stamping `Item.ClarifiedIntoProjectID`
- [x] 7.2 Default the project title to the Item title when not provided
- [x] 7.3 Wrap in a single transaction
- [x] 7.4 Tests including rollback behavior and verifying ReopenProject works on the incubated project

## 8. ClarifyAsTask Operation

- [x] 8.1 Implement InboxService.ClarifyAsTask creating a Task and stamping `Item.ClarifiedIntoTaskID`
- [x] 8.2 Support kind parameter (next_action, delegated)
- [x] 8.3 Support optional ProjectID with validation against existing projects
- [x] 8.4 Default the task title to the Item title when not provided
- [x] 8.5 Support `Status=done` for the do-it-now case
- [x] 8.6 Wrap in a single transaction
- [x] 8.7 Tests covering all variants

## 9. ClarifyAsProject Operation

- [x] 9.1 Implement InboxService.ClarifyAsProject creating a Project with `Status=open` and stamping `Item.ClarifiedIntoProjectID`
- [x] 9.2 Default the project title to the Item title when not provided
- [x] 9.3 Wrap in a single transaction
- [x] 9.4 Tests including rollback behavior and the initial-status difference from Incubate

## 10. Integration

- [x] 10.1 Wire the SQLite InboxService and the service-layer InboxService into application initialization
- [x] 10.2 Add integration tests for end-to-end capture → clarify flows
- [x] 10.3 Verify all clarify operations reject double-clarification

## 11. TUI

- [x] 11.1 Create `tui/pages/inbox/` Screen rendering InboxService.List in FIFO order with empty-state messaging
- [x] 11.2 Accept the capture-side InboxService in the inbox Screen's constructor
- [x] 11.3 Register an "Inbox" tab in `tui.New`
- [x] 11.4 Verify someday projects surface in the existing projects tab via the `status:someday` query filter (no new page)

## 12. Service signature: ClarifyAsProject checkpoint (project + first open task)

- [x] 12.1 Update `service.InboxService.ClarifyAsProject` signature to `(ctx, itemID, project, firstTask Task) (Project, Task, Item, error)`
- [x] 12.2 Enforce firstTask.Status is open (or empty); reject caller-supplied firstTask.ProjectID
- [x] 12.3 Persist project + first task + item stamp in one transaction; first task auto-gets the new project's ID
- [x] 12.4 Update `ClarifyAsTask` to also enforce open-only status (do-it-now moves to the wizard layer)
- [x] 12.5 Tests covering: open first task; non-open rejected; caller-supplied ProjectID rejected; wizard checkpoint pattern (commit-then-complete-via-TaskService; add successor task via TaskService.CreateTask)
- [x] 12.6 Resolve the design.md open question (ClarifyAsProject with completed first action)

## 13. Item capture overlay

- [x] 13.1 Create `tui/pages/inbox/itemcapture/` overlay with Title and Description fields (same structure as `taskedit`/`projectedit`, but new-item only)
- [x] 13.2 Wire `+`/`insert` on the inbox screen to push the capture overlay
- [x] 13.3 Overlay submits via `InboxService.Create` and dismisses on success
- [x] 13.4 Overlay tests for capture (save populates the inbox; cancel does not)

## 14. Clarify wizard

- [x] 14.1 Create `tui/pages/inbox/clarify/` package with a custom progressive-disclosure wizard component (no huh)
- [x] 14.2 Render the item header (title + description) at the top of the wizard
- [x] 14.3 Implement the question-stack model: each answered question persists on screen; the next unanswered question appears inline beneath
- [x] 14.4 Wire the actionable root question
- [x] 14.5 Non-actionable branch: Trash (with confirm) → Discard; Someday → inline title + description form → Incubate
- [x] 14.6 Actionable branch root: "multi-step?" splits into single-task and project branches
- [x] 14.7 Per-task block: next-action title, `<2 min?`, doer-if-not-2min (assignee field), project-attach (single-task branch only)
- [x] 14.8 Single-task commit: route to `ClarifyAsTask` creating an open task; if `<2 min` then push the do-it-now prompt
- [x] 14.9 Project branch: inline Title + Outcome + Description form, then run first per-task block
- [x] 14.10 Project first-task commit: route to `ClarifyAsProject` (project + first open task); if `<2 min` push do-it-now prompt then loop on confirm
- [x] 14.11 Project subsequent tasks: route to `TaskService.CreateTask` with `ProjectID` of the new project; if `<2 min` push do-it-now prompt then loop on confirm
- [x] 14.12 Loop exit: first task answered NOT `<2 min` becomes the open next-action and the wizard exits
- [x] 14.13 Do-it-now prompt component: standalone overlay with a "press enter to confirm complete / esc to leave open" UX; on confirm calls `TaskService.CompleteTask`
- [x] 14.14 Back-navigation: changing a previous answer discards all wizard state beneath it and re-asks subsequent questions (only valid BEFORE any checkpoint commit; once committed, back-nav is disabled to avoid invalidating persisted state)
- [x] 14.15 Wire `enter` on the inbox screen to push the wizard for the selected item
- [x] 14.16 Wizard tests for each branch: discard-confirm, incubate, single-task (all axes combos incl. do-it-now), project (esc after checkpoint leaves task open, do-it-now confirm then loop, loop exit on not-do-it-now), back-nav state collapse pre-checkpoint

> **Status:** All wizard tasks complete. Deferred to a future change: the
> project-attach picker (single-task branch with attach=Yes still returns a
> "not yet implemented" error; the No path commits as standalone and the
> existing-project case is handled today by editing the task from the task
> list). The user can attach a task to an existing project via the existing
> tasklist + projectpicker workflow.
