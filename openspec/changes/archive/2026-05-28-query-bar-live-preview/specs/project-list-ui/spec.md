## MODIFIED Requirements

### Requirement: Cancel project query edit on esc
Pressing esc while editing SHALL revert the query bar's text to the last applied query, blur the bar, and cause the project list to reload using that previously-applied query. The list SHALL end in the same state as immediately after the last commit, undoing any live-previewed filter introduced by debounced typing since.

#### Scenario: Cancel reverts text and snaps list back to last applied query
- **WHEN** the user edits the query and presses esc
- **THEN** the query bar reverts to the last applied query
- **AND** the project list reflects that previously-applied query
- **AND** the query bar is no longer focused
