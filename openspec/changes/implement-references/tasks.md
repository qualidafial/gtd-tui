## 1. Reference Domain Layer

- [ ] 1.1 Create `reference.go` in root package with Reference struct (ID, Title, Body, CreatedAt, UpdatedAt)
- [ ] 1.2 Define `ReferenceStore` interface with Create, Get, Update, Delete, List
- [ ] 1.3 Define `ReferenceListOptions` struct for list filtering/ordering

## 2. References Table Migration

- [ ] 2.1 Create migration for the `references` table with CHECK constraint on non-empty title
- [ ] 2.2 Add UTC timestamp columns and a title index suitable for filter queries

## 3. Items Table Rebuild

- [ ] 3.1 In the same migration (or a sequential one), rebuild the `items` table to add `clarified_into_reference_id INTEGER REFERENCES references(id)`
- [ ] 3.2 Extend the mutual-exclusion CHECK to include the new column
- [ ] 3.3 Disable foreign keys for the rebuild, copy rows from old to new, drop old, rename new, re-enable foreign keys
- [ ] 3.4 Add a migration round-trip test that seeds rows under the old shape and asserts the new shape and CHECK behavior

## 4. SQLite Reference Implementation

- [ ] 4.1 Create `sqlite/reference.go` implementing ReferenceStore with squirrel queries and UTC timestamps
- [ ] 4.2 Add `scanReference` helper following the existing `scanTask` pattern
- [ ] 4.3 Tests covering CRUD and the non-empty-title CHECK

## 5. Item Struct + Store Updates

- [ ] 5.1 Add `ClarifiedIntoReferenceID *int64` to the Item struct
- [ ] 5.2 Update `sqlite.ItemStore` queries to read/write the new column
- [ ] 5.3 Tests covering the extended mutual-exclusion CHECK

## 6. FileAsReference Operation

- [ ] 6.1 Extend the service-layer `InboxService` constructor to accept ReferenceStore
- [ ] 6.2 Implement `InboxService.FileAsReference` creating a Reference and stamping `Item.ClarifiedIntoReferenceID`
- [ ] 6.3 Default title from Item title and body from Item description
- [ ] 6.4 Wrap in a single transaction
- [ ] 6.5 Tests including rollback behavior and double-clarification rejection

## 7. Integration

- [ ] 7.1 Wire ReferenceStore into application initialization and the InboxService constructor at the call site
- [ ] 7.2 Add an integration test for capture → FileAsReference

## 8. TUI

- [ ] 8.1 Create `tui/pages/references/` Screen rendering ReferenceStore.List with querybar-driven title filter
- [ ] 8.2 Accept ReferenceStore in the Screen constructor
- [ ] 8.3 Register a "References" tab in `tui.New`
