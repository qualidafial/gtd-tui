## 1. Reference Entity

- [ ] 1.1 Create Reference struct in reference.go with ID, Title, Body, CreatedAt, UpdatedAt fields
- [ ] 1.2 Create ReferenceStore interface with Create, Get, Update, Delete, List methods
- [ ] 1.3 Create ReferenceListOptions struct for list filtering/ordering
- [ ] 1.4 Create SQLite migration for references table with CHECK constraint on non-empty title
- [ ] 1.5 Implement sqlite.ReferenceStore with all CRUD operations
- [ ] 1.6 Write sqlite.ReferenceStore tests covering CRUD operations and CHECK constraint

## 2. Item Entity Extensions

- [ ] 2.1 Add ClarifiedIntoReferenceID nullable FK field to Item struct
- [ ] 2.2 Add Discarded bool field to Item struct
- [ ] 2.3 Update Item SQLite migration with new column and CHECK constraint for mutual exclusion of TaskID/ProjectID/ReferenceID/Discarded
- [ ] 2.4 Update sqlite.ItemStore to handle new ClarifiedInto field
- [ ] 2.5 Write tests for Item ClarifiedInto mutual exclusion constraint

## 3. InboxService Setup

- [ ] 3.1 Create service/inbox.go with InboxService struct
- [ ] 3.2 Add constructor taking store dependencies (ItemStore, TaskStore, ProjectStore, ReferenceStore)
- [ ] 3.3 Add transaction provider function for atomic operations
- [ ] 3.4 Add helper to check if Item is already clarified

## 4. Discard Operation

- [ ] 4.1 Implement InboxService.Discard method
- [ ] 4.2 Add validation for non-existent item
- [ ] 4.3 Add validation for already-clarified item
- [ ] 4.4 Write tests for Discard including error cases

## 5. Incubate Operation

- [ ] 5.1 Implement InboxService.Incubate creating a Project with Status=someday and stamping Item.ClarifiedIntoProjectID
- [ ] 5.2 Add default title copying from Item
- [ ] 5.3 Add transaction wrapping for atomicity
- [ ] 5.4 Write tests for Incubate including rollback behavior and verifying ReopenProject path from the incubated project

## 6. FileAsReference Operation

- [ ] 6.1 Implement InboxService.FileAsReference method with Reference creation
- [ ] 6.2 Add default title and body copying from Item
- [ ] 6.3 Add transaction wrapping for atomicity
- [ ] 6.4 Write tests for FileAsReference including rollback behavior

## 7. ClarifyAsTask Operation

- [ ] 7.1 Implement InboxService.ClarifyAsTask method with Task creation
- [ ] 7.2 Support kind parameter (next_action, delegated)
- [ ] 7.3 Support optional ProjectID with validation
- [ ] 7.4 Add default title copying from Item
- [ ] 7.5 Support Status=done for do-it-now case
- [ ] 7.6 Add transaction wrapping for atomicity
- [ ] 7.7 Write tests for ClarifyAsTask including all variants

## 8. ClarifyAsProject Operation

- [ ] 8.1 Implement InboxService.ClarifyAsProject creating a Project with Status=open and stamping Item.ClarifiedIntoProjectID
- [ ] 8.2 Add default title copying from Item
- [ ] 8.3 Add transaction wrapping for atomicity
- [ ] 8.4 Write tests for ClarifyAsProject including rollback behavior and distinguishing initial status from Incubate

## 9. Integration

- [ ] 9.1 Wire InboxService into application initialization
- [ ] 9.2 Add integration tests for complete clarify workflows
- [ ] 9.3 Verify all clarify operations prevent double-clarification

## 10. TUI

- [ ] 10.1 Create `tui/pages/references/` Screen rendering ReferenceStore.List with querybar-driven title filter
- [ ] 10.2 Accept ReferenceStore in the Screen constructor
- [ ] 10.3 Register a "References" tab in `tui.New`
- [ ] 10.4 Verify someday projects surface in the existing projects tab via the `status:someday` query filter (no new page)
