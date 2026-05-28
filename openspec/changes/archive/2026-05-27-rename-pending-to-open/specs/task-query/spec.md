# task-query Delta Spec

## MODIFIED Requirements

### Requirement: Recognized keys
The parser SHALL recognize the keys `status`, `assignee`, `due`, `defer`, and `ready`. `status` accepts open/done/dropped; `assignee` accepts any string; `due`, `defer`, and `ready` accept date-predicate values (`ready` accepts only threshold values, not `none`/`any`). The `kind` key SHALL NOT be recognized.

#### Scenario: status key
- **WHEN** parsing `status:dropped`
- **THEN** TaskFilter.Status = TaskStatusDropped

#### Scenario: kind key is unrecognized
- **WHEN** parsing `kind:delegated`
- **THEN** the token is treated as a free-text search term