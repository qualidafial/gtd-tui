## 1. Domain Types

- [ ] 1.1 Add Meeting struct to meeting.go in root package (ID, Title, Body, StartTime, EndTime, Attendees []string, CreatedAt, UpdatedAt)
- [ ] 1.2 Add MeetingLink struct to meeting.go (ID, MeetingID, TaskID, ProjectID, ItemID as nullable *int64, CreatedAt)
- [ ] 1.3 Add MeetingFilter struct with MeetingIDs field
- [ ] 1.4 Add MeetingService interface with Meeting, Meetings, CreateMeeting, UpdateMeeting, DeleteMeeting, MeetingLinks, AddActionItem methods

## 2. Database Schema

- [ ] 2.1 Create migration 0003_meetings.sql with meetings table (id, title, body, start_time, end_time, attendees JSON, created_at, updated_at)
- [ ] 2.2 Add CHECK constraint on meetings for non-empty title
- [ ] 2.3 Add meeting_links table (id, meeting_id, task_id, project_id, item_id, created_at)
- [ ] 2.4 Add CHECK constraint on meeting_links ensuring exactly one of task_id/project_id/item_id is set
- [ ] 2.5 Add foreign keys with ON DELETE CASCADE for meeting_links

## 3. SQLite Meeting Implementation

- [ ] 3.1 Create sqlite/meeting.go with MeetingService struct embedding *DB
- [ ] 3.2 Implement Meeting(ctx, id) - fetch single meeting by ID
- [ ] 3.3 Implement Meetings(ctx, filter) - list meetings with optional ID filter
- [ ] 3.4 Implement CreateMeeting(ctx, meeting) - insert meeting, return with server-assigned fields
- [ ] 3.5 Implement UpdateMeeting(ctx, meeting) - update meeting, refresh UpdatedAt
- [ ] 3.6 Implement DeleteMeeting(ctx, id) - delete meeting (links cascade)
- [ ] 3.7 Implement MeetingLinks(ctx, meetingID) - fetch links for a meeting

## 4. AddActionItem Implementation

- [ ] 4.1 Implement AddActionItem(ctx, meetingID, title, description) in sqlite/meeting.go
- [ ] 4.2 Create inbox Item within transaction
- [ ] 4.3 Create MeetingLink with ItemID set within transaction
- [ ] 4.4 Append "- [ ] <title>" line to Meeting body within transaction
- [ ] 4.5 Handle newline formatting (add \n before if body doesn't end with newline)
- [ ] 4.6 Return created Item and updated Meeting

## 5. MeetingLink Clarification

- [ ] 5.1 Add helper function to query MeetingLinks by ItemID
- [ ] 5.2 Add helper function to rewrite MeetingLinks (ItemID -> TaskID or ProjectID)
- [ ] 5.3 Update ClarifyAsTask to rewrite MeetingLinks within transaction
- [ ] 5.4 Update ClarifyAsProject to rewrite MeetingLinks within transaction

## 6. Meeting CRUD Tests

- [ ] 6.1 Create sqlite/meeting_test.go with openTestDB setup
- [ ] 6.2 Test CreateMeeting with valid data, verify server-assigned fields
- [ ] 6.3 Test CreateMeeting rejects empty title (CHECK constraint)
- [ ] 6.4 Test Meeting(id) returns meeting, Meeting(invalid) returns error
- [ ] 6.5 Test Meetings with empty filter returns all, with IDs filter returns subset
- [ ] 6.6 Test UpdateMeeting updates fields, refreshes UpdatedAt
- [ ] 6.7 Test DeleteMeeting removes meeting and cascades to links

## 7. MeetingLink Tests

- [ ] 7.1 Test MeetingLink with TaskID only - accepted
- [ ] 7.2 Test MeetingLink with ProjectID only - accepted
- [ ] 7.3 Test MeetingLink with ItemID only - accepted
- [ ] 7.4 Test MeetingLink with multiple FKs - CHECK violation
- [ ] 7.5 Test MeetingLink with no FKs - CHECK violation
- [ ] 7.6 Test MeetingLinks(meetingID) returns correct links

## 8. AddActionItem Tests

- [ ] 8.1 Test AddActionItem creates inbox Item with correct title/description
- [ ] 8.2 Test AddActionItem creates MeetingLink with ItemID
- [ ] 8.3 Test AddActionItem appends line to empty body
- [ ] 8.4 Test AddActionItem appends line to body ending with newline
- [ ] 8.5 Test AddActionItem appends line to body not ending with newline
- [ ] 8.6 Test AddActionItem returns Item with populated ID/timestamps
- [ ] 8.7 Test AddActionItem returns Meeting with updated body and UpdatedAt
- [ ] 8.8 Test AddActionItem rejects empty title
- [ ] 8.9 Test AddActionItem rejects invalid meeting ID

## 9. MeetingLink Clarification Tests

- [ ] 9.1 Test ClarifyAsTask rewrites MeetingLink from ItemID to TaskID
- [ ] 9.2 Test ClarifyAsProject rewrites MeetingLink from ItemID to ProjectID
- [ ] 9.3 Test Discard preserves MeetingLink pointing at Item
- [ ] 9.4 Test Item with multiple MeetingLinks - all rewritten on clarify
- [ ] 9.5 Test clarification rollback reverts MeetingLink rewrite
