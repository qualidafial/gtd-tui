## 1. Item Domain Layer

- [ ] 1.1 Create `item.go` in root package with Item struct (ID, Title, Description, CreatedAt, UpdatedAt, ClarifiedIntoTaskID, ClarifiedIntoProjectID, Discarded). `ClarifiedIntoReferenceID` is added by `implement-references`.
- [ ] 1.2 Define the capture-side `InboxService` interface in `item.go` (Create, List, Get)

## 2. Database Migration

- [ ] 2.1 Create migration for the `items` table with non-empty-title CHECK
- [ ] 2.2 Add CHECK constraint enforcing mutual exclusion of `clarified_into_task_id` / `clarified_into_project_id` / `discarded` (extended by `implement-references` to include `clarified_into_reference_id`)

## 3. SQLite Implementation

- [ ] 3.1 Create `sqlite/item.go` implementing the capture-side InboxService
- [ ] 3.2 Implement Create with squirrel insert and UTC timestamps
- [ ] 3.3 Implement List filtering unclarified/non-discarded items, ordered by created_at ASC
- [ ] 3.4 Implement Get with squirrel select by ID
- [ ] 3.5 Add `scanItem` helper following the existing `scanTask` pattern

## 4. SQLite Tests

- [ ] 4.1 Create `sqlite/item_test.go` covering Create (valid + empty-title rejection), List (FIFO, only unclarified), Get (valid + missing)
- [ ] 4.2 Test the mutual-exclusion CHECK constraint with direct SQL violations

## 5. Service-layer InboxService Setup

- [ ] 5.1 Create `service/inbox.go` with InboxService struct
- [ ] 5.2 Constructor accepts ItemStore, TaskStore, ProjectStore
- [ ] 5.3 Wire a transaction provider for atomic operations
- [ ] 5.4 Add a helper that checks whether an Item is already clarified or discarded

## 6. Discard Operation

- [ ] 6.1 Implement InboxService.Discard
- [ ] 6.2 Reject non-existent items
- [ ] 6.3 Reject already-clarified items
- [ ] 6.4 Tests including error cases

## 7. Incubate Operation

- [ ] 7.1 Implement InboxService.Incubate creating a Project with `Status=someday` and stamping `Item.ClarifiedIntoProjectID`
- [ ] 7.2 Default the project title to the Item title when not provided
- [ ] 7.3 Wrap in a single transaction
- [ ] 7.4 Tests including rollback behavior and verifying ReopenProject works on the incubated project

## 8. ClarifyAsTask Operation

- [ ] 8.1 Implement InboxService.ClarifyAsTask creating a Task and stamping `Item.ClarifiedIntoTaskID`
- [ ] 8.2 Support kind parameter (next_action, delegated)
- [ ] 8.3 Support optional ProjectID with validation against existing projects
- [ ] 8.4 Default the task title to the Item title when not provided
- [ ] 8.5 Support `Status=done` for the do-it-now case
- [ ] 8.6 Wrap in a single transaction
- [ ] 8.7 Tests covering all variants

## 9. ClarifyAsProject Operation

- [ ] 9.1 Implement InboxService.ClarifyAsProject creating a Project with `Status=open` and stamping `Item.ClarifiedIntoProjectID`
- [ ] 9.2 Default the project title to the Item title when not provided
- [ ] 9.3 Wrap in a single transaction
- [ ] 9.4 Tests including rollback behavior and the initial-status difference from Incubate

## 10. Integration

- [ ] 10.1 Wire the SQLite InboxService and the service-layer InboxService into application initialization
- [ ] 10.2 Add integration tests for end-to-end capture → clarify flows
- [ ] 10.3 Verify all clarify operations reject double-clarification

## 11. TUI

- [ ] 11.1 Create `tui/pages/inbox/` Screen rendering InboxService.List in FIFO order with empty-state messaging
- [ ] 11.2 Accept the capture-side InboxService in the inbox Screen's constructor
- [ ] 11.3 Register an "Inbox" tab in `tui.New`
- [ ] 11.4 Verify someday projects surface in the existing projects tab via the `status:someday` query filter (no new page)
