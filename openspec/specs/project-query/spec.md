# project-query Specification

## Purpose
Defines the `projectquery` parser that translates a compact query string into a `gtd.ProjectFilter`, including supported keys, free-text search semantics, and parse-error reporting with rune ranges.

## Requirements

### Requirement: Project query parser
The `projectquery` package SHALL parse a compact query string into a `gtd.ProjectFilter`. The grammar is whitespace-tokenized: a token of the form `key:value` whose key is recognized sets a structured filter field; any other token is a free-text search term added to `ProjectFilter.Search`.

#### Scenario: Empty query yields zero filter
- **WHEN** the query string is empty
- **THEN** the parser SHALL return a zero-value `ProjectFilter` and no error

#### Scenario: Free-text search tokens
- **WHEN** the query is `shed plans`
- **THEN** `ProjectFilter.Search` SHALL be `["shed", "plans"]`

#### Scenario: Unrecognized key treated as free text
- **WHEN** the query contains `foo:bar`
- **THEN** `foo:bar` SHALL be added to `Search` as a free-text token

### Requirement: Status filter key
The parser SHALL recognize `status:` with values `open`, `someday`, `done`, and `dropped`. An invalid status value SHALL return a `*querybar.ParseError` with the offending token's rune range.

#### Scenario: Valid status filter
- **WHEN** the query is `status:someday`
- **THEN** `ProjectFilter.Status` SHALL point to `ProjectStatusSomeday`

#### Scenario: Invalid status value
- **WHEN** the query is `status:bogus`
- **THEN** the parser SHALL return a `*querybar.ParseError` with `Start` and `End` marking the `status:bogus` token

### Requirement: ProjectFilter.Search field
`ProjectFilter` SHALL include a `Search []string` field. When non-empty, `ListProjects` SHALL filter to projects whose title, outcome, or description contains every search term (case-insensitive substring match). Each term must appear in at least one field; all terms must match for the project to be included.

#### Scenario: Search matches title
- **WHEN** `Search` is `["shed"]` and a project has title "Build shed"
- **THEN** the project SHALL be included in results

#### Scenario: Search matches outcome
- **WHEN** `Search` is `["functional"]` and a project has outcome "A functional shed"
- **THEN** the project SHALL be included in results

#### Scenario: Search requires all terms
- **WHEN** `Search` is `["shed", "garage"]` and a project has title "Build shed" but no field contains "garage"
- **THEN** the project SHALL NOT be included in results

### Requirement: ParseError with range
Parse errors SHALL be returned as `*querybar.ParseError` with `Message`, `Start`, and `End` fields. `Start` and `End` are rune offsets into the original query marking the `[Start, End)` range of the offending token.

#### Scenario: Error range marks the bad token
- **WHEN** the query is `shed status:bogus plans`
- **THEN** `Start` SHALL be 5 and `End` SHALL be 17 (the rune range of `status:bogus`)
