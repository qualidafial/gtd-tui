## 1. Domain Layer

- [ ] 1.1 Add TimelineEntry domain type to root package (ID, EntityType, EntityID, EventType, Details, CreatedAt)
- [ ] 1.2 Add TimelineService interface to root package (Record, ListForEntity, ListGlobal)
- [ ] 1.3 Add EntityType and EventType constants/enums for type safety

## 2. SQLite Schema

- [ ] 2.1 Create migration for timeline_entries table with columns (id, entity_type, entity_id, event_type, details, created_at)
- [ ] 2.2 Add index on (entity_type, entity_id) for entity-scoped queries
- [ ] 2.3 Add index on created_at for global timeline ordering

## 3. SQLite TimelineService Implementation

- [ ] 3.1 Implement TimelineService.Record() - insert entry and return with ID/CreatedAt populated
- [ ] 3.2 Implement TimelineService.ListForEntity() - query by entity_type/entity_id ordered by created_at ASC
- [ ] 3.3 Implement TimelineService.ListGlobal() - query with limit and cursor, ordered by created_at DESC
- [ ] 3.4 Add tests for TimelineService CRUD operations

## 4. Integrate Timeline into Task Operations

- [ ] 4.1 Emit "created" entry in CreateTask
- [ ] 4.2 Emit "updated" entry in UpdateTask for non-status changes
- [ ] 4.3 Emit "status_changed" entry in CompleteTask with from/to details
- [ ] 4.4 Emit "status_changed" entry in DropTask with from/to details
- [ ] 4.5 Emit "status_changed" entry in ReopenTask with from/to details
- [ ] 4.6 Add tests verifying timeline entries for task operations

## 5. Integrate Timeline into Project Operations

- [ ] 5.1 Emit "created" entry in CreateProject
- [ ] 5.2 Emit "updated" entry in UpdateProject for non-status changes
- [ ] 5.3 Emit "status_changed" entry in CompleteProject with from/to details
- [ ] 5.4 Emit "status_changed" entry in DropProject with from/to details
- [ ] 5.5 Emit "status_changed" entry in ParkProject with from/to details
- [ ] 5.6 Emit "status_changed" entry in UnparkProject with from/to details
- [ ] 5.7 Add tests verifying timeline entries for project operations

## 6. Integrate Timeline into Inbox/Clarify Operations

- [ ] 6.1 Emit "created" entry when inbox Item is captured
- [ ] 6.2 Emit "clarified" entry on Item when ClarifyAsTask is called (with into_type/into_id)
- [ ] 6.3 Emit "created" entry on Task from ClarifyAsTask (with from_item reference)
- [ ] 6.4 Emit "clarified" entry on Item when ClarifyAsProject is called
- [ ] 6.5 Emit "created" entry on Project from ClarifyAsProject (with from_item reference)
- [ ] 6.6 Emit "clarified" entry on Item when Incubate is called (with into_type="someday")
- [ ] 6.7 Emit "clarified" entry on Item when FileAsReference is called
- [ ] 6.8 Emit "clarified" entry on Item when Discard is called (with discarded=true)
- [ ] 6.9 Add tests verifying timeline entries for clarify operations

## 7. Integrate Timeline into Meeting Operations

- [ ] 7.1 Emit "created" entry in CreateMeeting
- [ ] 7.2 Emit "updated" entry in UpdateMeeting
- [ ] 7.3 Add tests verifying timeline entries for meeting operations

## 8. Integrate Timeline into Comment Operations

- [ ] 8.1 Emit "commented" entry on parent Task/Project when Comment is created
- [ ] 8.2 Ensure edit-with-comment generates both "updated" and "commented" entries
- [ ] 8.3 Add tests verifying timeline entries for comment operations

## 9. Extended Timeline Queries

- [ ] 9.1 Implement project timeline aggregation (include entries for tasks in project)
- [ ] 9.2 Implement meeting timeline aggregation (include entries for linked entities)
- [ ] 9.3 Add tests for aggregated timeline queries

## 10. TUI Integration

- [ ] 10.1 Create timeline view component displaying entries in chronological order
- [ ] 10.2 Add entity-scoped timeline view accessible from task/project/meeting detail pages
- [ ] 10.3 Add global timeline view as a top-level page for Reflect workflow
- [ ] 10.4 Format timeline entries with human-readable descriptions
- [ ] 10.5 Add keyboard navigation for timeline views
