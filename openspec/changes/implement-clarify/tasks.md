## 1. Someday Entity

- [ ] 1.1 Create Someday struct in someday.go with ID, Title, Description, ReviewedAt, CreatedAt, UpdatedAt fields
- [ ] 1.2 Create SomedayStore interface with Create, Get, Update, Delete, List methods
- [ ] 1.3 Create SomedayListOptions struct for list filtering/ordering
- [ ] 1.4 Create SQLite migration for someday table with CHECK constraint on non-empty title
- [ ] 1.5 Implement sqlite.SomedayStore with all CRUD operations
- [ ] 1.6 Write sqlite.SomedayStore tests covering CRUD operations and CHECK constraint

## 2. Reference Entity

- [ ] 2.1 Create Reference struct in reference.go with ID, Title, Body, CreatedAt, UpdatedAt fields
- [ ] 2.2 Create ReferenceStore interface with Create, Get, Update, Delete, List methods
- [ ] 2.3 Create ReferenceListOptions struct for list filtering/ordering
- [ ] 2.4 Create SQLite migration for reference table with CHECK constraint on non-empty title
- [ ] 2.5 Implement sqlite.ReferenceStore with all CRUD operations
- [ ] 2.6 Write sqlite.ReferenceStore tests covering CRUD operations and CHECK constraint

## 3. Item Entity Extensions

- [ ] 3.1 Add ClarifiedIntoSomedayID and ClarifiedIntoReferenceID nullable FK fields to Item struct
- [ ] 3.2 Add Discarded bool field to Item struct
- [ ] 3.3 Update Item SQLite migration with new columns and CHECK constraint for mutual exclusion
- [ ] 3.4 Update sqlite.ItemStore to handle new ClarifiedInto fields
- [ ] 3.5 Write tests for Item ClarifiedInto mutual exclusion constraint

## 4. InboxService Setup

- [ ] 4.1 Create service/inbox.go with InboxService struct
- [ ] 4.2 Add constructor taking store dependencies (ItemStore, TaskStore, ProjectStore, SomedayStore, ReferenceStore)
- [ ] 4.3 Add transaction provider function for atomic operations
- [ ] 4.4 Add helper to check if Item is already clarified

## 5. Discard Operation

- [ ] 5.1 Implement InboxService.Discard method
- [ ] 5.2 Add validation for non-existent item
- [ ] 5.3 Add validation for already-clarified item
- [ ] 5.4 Write tests for Discard including error cases

## 6. Incubate Operation

- [ ] 6.1 Implement InboxService.Incubate method with Someday creation
- [ ] 6.2 Add default title copying from Item
- [ ] 6.3 Add transaction wrapping for atomicity
- [ ] 6.4 Write tests for Incubate including rollback behavior

## 7. FileAsReference Operation

- [ ] 7.1 Implement InboxService.FileAsReference method with Reference creation
- [ ] 7.2 Add default title and body copying from Item
- [ ] 7.3 Add transaction wrapping for atomicity
- [ ] 7.4 Write tests for FileAsReference including rollback behavior

## 8. ClarifyAsTask Operation

- [ ] 8.1 Implement InboxService.ClarifyAsTask method with Task creation
- [ ] 8.2 Support kind parameter (next_action, delegated)
- [ ] 8.3 Support optional ProjectID with validation
- [ ] 8.4 Add default title copying from Item
- [ ] 8.5 Support Status=done for do-it-now case
- [ ] 8.6 Add transaction wrapping for atomicity
- [ ] 8.7 Write tests for ClarifyAsTask including all variants

## 9. ClarifyAsProject Operation

- [ ] 9.1 Implement InboxService.ClarifyAsProject method with Project creation
- [ ] 9.2 Add default title copying from Item
- [ ] 9.3 Add transaction wrapping for atomicity
- [ ] 9.4 Write tests for ClarifyAsProject including rollback behavior

## 10. Integration

- [ ] 10.1 Wire InboxService into application initialization
- [ ] 10.2 Add integration tests for complete clarify workflows
- [ ] 10.3 Verify all clarify operations prevent double-clarification
