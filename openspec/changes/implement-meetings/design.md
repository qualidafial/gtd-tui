## Context

The GTD TUI currently has Task, Project, Item (inbox), Someday, Reference, and Comment entities. Meeting records are identified in DESIGN.md as a key capture mechanism: action items captured during meetings should flow to the inbox automatically with a link back to the meeting.

This change introduces Meeting and MeetingLink entities to enable meeting-centric action item capture. The foundation specs in `openspec/changes/foundation/specs/domain-model/spec.md` already define the requirements - this design focuses on implementation approach.

## Goals / Non-Goals

**Goals:**
- Implement Meeting entity following existing patterns (value semantics, service interface in root package, SQLite implementation)
- Implement MeetingLink join table with CHECK constraint enforcing exactly one FK set
- Provide MeetingService.AddActionItem that atomically creates inbox Item, MeetingLink, and appends to Meeting body
- Ensure MeetingLinks follow Items through clarification (rewrite to point at resulting entity)
- Match existing code patterns for consistency (squirrel queries, transaction handling, test style)

**Non-Goals:**
- TUI views for meetings (separate change)
- Attendee contact management (deferred per DESIGN.md)
- Calendar integration or scheduling features
- Parsing action items from meeting prose (action items are captured via explicit API)

## Decisions

### Decision: Store attendees as JSON []string

Store attendees in a single JSON column rather than a separate attendees table.

**Rationale**: DESIGN.md explicitly states "JSON []string for now; promote to a contacts table if querying by attendee becomes useful." This keeps the schema simple and matches the documented intent.

**Alternatives considered**:
- Separate attendees table with foreign key: Rejected as premature complexity given no attendee query use cases.

### Decision: MeetingLink uses nullable FKs with CHECK constraint

MeetingLink has TaskID, ProjectID, and ItemID columns, all nullable, with a CHECK constraint ensuring exactly one is set.

**Rationale**: This matches the existing pattern for Comment (TaskID/ProjectID) and Item.ClarifiedInto* fields. The architecture spec requires "exactly one FK set" constraints on multi-nullable-FK tables.

**Alternatives considered**:
- Polymorphic type/id columns: Rejected because it breaks foreign key enforcement.
- Separate tables per link type: Rejected as unnecessary fragmentation.

### Decision: AddActionItem is a single transaction

AddActionItem creates the inbox Item, inserts the MeetingLink, and updates the Meeting body all in one transaction.

**Rationale**: The architecture spec requires "Write operations SHALL open a transaction at the service method level." This ensures atomicity - no orphaned links if any step fails.

### Decision: MeetingLink rewriting during clarification

When an Item with a MeetingLink is clarified, the InboxService clarify methods must update the MeetingLink to point at the resulting entity (Task or Project).

**Rationale**: DESIGN.md specifies "the link follows the Item through clarification (it is rewritten to point at the resulting entity) so the meeting's provenance trail is preserved permanently."

**Implementation**: Clarify methods (ClarifyAsTask, ClarifyAsProject) query for MeetingLinks referencing the Item and update them to point at the new entity. This happens within the existing clarify transaction.

### Decision: Meeting body append format

AddActionItem appends a uniform line to the Meeting body. Format: `- [ ] <action item title>` on its own line.

**Rationale**: Markdown checkbox format is familiar and readable in the meeting notes. The spawned inbox Item is the source of truth for workflow; the meeting body line is for human reference.

## Risks / Trade-offs

**[Risk] MeetingLink rewriting adds complexity to clarify flow** -> Mitigation: Query for MeetingLinks is a simple WHERE ItemID = ? lookup; update is a single UPDATE statement within existing transaction.

**[Risk] Attendees as JSON limits query flexibility** -> Mitigation: Acceptable per DESIGN.md. Migration path exists if needed later.

**[Trade-off] Meeting body becomes denormalized (contains action item text)** -> This is intentional per DESIGN.md: "the meeting body is the source of truth for the original phrasing of any action items captured during the meeting."
