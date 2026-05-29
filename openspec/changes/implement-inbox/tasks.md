## 1. Domain Layer

- [ ] 1.1 Create `item.go` in root package with Item struct (ID, Title, Description, CreatedAt, UpdatedAt, ClarifiedIntoTaskID, ClarifiedIntoProjectID, Discarded). ClarifiedIntoReferenceID is added by implement-clarify.
- [ ] 1.2 Define InboxService interface in `item.go` with Create, List, Get methods

## 2. Database Migration

- [ ] 2.1 Create migration file `sqlite/migrations/0003_items.sql` with items table schema
- [ ] 2.2 Add CHECK constraint for non-empty title
- [ ] 2.3 Add CHECK constraint enforcing mutual exclusion of ClarifiedIntoTaskID/ClarifiedIntoProjectID/Discarded (implement-clarify extends this to include ClarifiedIntoReferenceID)

## 3. SQLite Implementation

- [ ] 3.1 Create `sqlite/item.go` implementing InboxService interface
- [ ] 3.2 Implement Create method using squirrel insert with UTC timestamps
- [ ] 3.3 Implement List method using squirrel select filtering unclarified/non-discarded items, ordered by created_at ASC
- [ ] 3.4 Implement Get method using squirrel select by ID
- [ ] 3.5 Add scanItem helper function following existing scanTask pattern

## 4. Testing

- [ ] 4.1 Create `sqlite/item_test.go` with table-driven tests using openTestDB
- [ ] 4.2 Test Create: valid item, empty title rejection
- [ ] 4.3 Test List: returns only unclarified items, FIFO ordering, empty result when all clarified
- [ ] 4.4 Test Get: valid ID, missing ID error

## 5. TUI

- [ ] 5.1 Create `tui/pages/inbox/` Screen implementing the inbox list
- [ ] 5.2 Render items returned by InboxService.List in FIFO order with empty-state message
- [ ] 5.3 Accept InboxService in the Screen constructor
- [ ] 5.4 Register an "Inbox" tab in `tui.New` ahead of the Tasks/Projects tabs
