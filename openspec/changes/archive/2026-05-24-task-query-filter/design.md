## Context

The task list (`tui/pages/tasks/tasklist`) loads tasks through `TaskService.ListTasks(ctx, TaskFilter)`. Today the filter is hard-coded to `status:pending` in `tui/app.go`, and `TaskFilter` supports `Status`, `Kind`, `IncludeDeferred`, and `TaskIDs`. The service and sqlite layers (capabilities `task-service`, `task-sqlite`) were finalized by the archived `implement-tasks` change. This change adds a user-facing query language over that surface, extends the filter with free-text and date predicates, and **removes `IncludeDeferred`** in favor of explicit deferral predicates — so the default "available now" view is itself just a query, not magic behavior.

The bubbles `list` component has a built-in `/` filter, but it only does client-side substring matching on rendered titles. We want server-side, structured filtering, so we override `/` with our own query bar.

## Goals / Non-Goals

**Goals:**
- A compact, typed query language: `key:value` tokens + free-text terms.
- Keys: `status`, `kind`, `assignee`, `due`, `defer`, `ready`.
- Free-text terms match title/description/assignee (case-insensitive substring), ANDed.
- Replace `IncludeDeferred` with explicit deferral predicates so the default view is a representable query, not implicit behavior.
- A pure, unit-testable parser (`internal/taskquery`) decoupled from the TUI.
- Query bar on the task list: focus with `/`, apply on Enter, default `status:pending ready:now`, inline parse errors.

**Non-Goals:**
- Boolean operators (OR/NOT), grouping/parentheses, quoted multi-word values — deferred.
- Date comparison operators (`due:<2026-06-01`) and windows (lower + upper bound) — dates resolve to a single threshold in v1.
- Exact-single-day date match — `due:`/`defer:` are threshold comparisons, not point matches.
- Multi-value keys (`status:done status:dropped` as OR) — single value per key, last wins.
- Saved/named queries, query history.
- Filtering across other entities (projects, inbox) — task list only.

## Decisions

### Decision 1: Grammar — whitespace-tokenized, key:value vs free-text

A query is split on whitespace into tokens. A token containing `:` where the part before `:` is a **recognized key** is a structured filter; everything else (including `foo:bar` with an unrecognized key) is a free-text term. Multiple free-text terms are ANDed; each term matches if it appears (case-insensitive substring) in title, description, OR assignee.

**Rationale:** No quoting/escaping keeps the parser and the mental model trivial. Treating unknown `key:value` as free text avoids surprising "unknown key" errors for words that happen to contain a colon.

**Alternatives:** Full lexer with quotes/operators (rejected — disproportionate for v1); unknown keys as hard errors (rejected — hostile to free typing).

### Decision 2: Single value per key, last-wins

Each recognized key sets one structured field. A repeated key overwrites (`status:done status:dropped` → dropped). Multi-status OR would require `TaskFilter.Status` to become a slice, which is a larger change to the just-archived `task-service` capability.

**Rationale:** Keeps `TaskFilter` pointer-field shape intact; covers the common cases. OR-semantics is an additive future change.

### Decision 3: Date predicates as resolved thresholds, three date keys

The three date keys (`due`, `ready`, `defer`) take a value that resolves to a reference time, or the null/not-null variants. A value may be:

- `now` — the current **instant** (not rounded to a day boundary)
- a relative duration: `-5d`, `7d`, `2w` (units `d` = days, `w` = weeks; sign = past/future relative to today)
- a keyword alias: `overdue` ≡ `-1d`, `today` ≡ `0d`, `week` ≡ `7d`
- an explicit ISO date: `YYYY-MM-DD`
- `none` / `any` — the null / not-null variants (no threshold)

Except for `now`, a value resolves to **end-of-day** of the named date in the **local** timezone, then converts to UTC for the query (`now` resolves to the current instant). Each key applies a fixed comparison:

```
due:X    → due_date ≤ X                              "must-deliver as of X"  (cumulative; excludes NULL)
ready:X  → defer_until IS NULL OR defer_until ≤ X     "available as of X"     (includes never-deferred)
defer:X  → defer_until > X                            "still parked as of X"  (excludes NULL)
due:none / defer:none → column IS NULL
due:any  / defer:any  → column IS NOT NULL
```

`ready` and `defer` are complementary directions on the same `defer_until` column: `ready:X` is "the gate has opened (or there is none) by X"; `defer:X` is "the gate is still in the future at X". The default view's "available now" is exactly `ready:now`.

`DatePredicate` carries a resolved time plus a kind discriminator: `OnOrBefore` (due), `AvailableAsOf` (ready), `After` (defer), `IsNull` (none), `IsNotNull` (any). One predicate type serves all three keys; each key accepts the subset of kinds shown above. Boundary/`now` computation lives in one helper in `taskquery`.

**Rationale:** A single resolved-threshold rule covers the GTD questions: `due ≤ X` = "what must I deliver by X"; `ready:now` = "what can I act on right now" (the default); `defer:X` = "what's still parked as of X" (the review). Keywords are friendly aliases over the same machinery. `now`-vs-`today` matters because a task deferred to 3pm should still be hidden at 9am — the default uses the instant, not end-of-day.

**Alternatives:**
- Per-shape vocabulary (exact-day `today`, window `week`, range `overdue`): rejected — one threshold suffices; `today`/`week` are cumulative aliases.
- A single bidirectional `defer:` key (no `ready:`): rejected — the default "available now" view is the complement of "parked" and must be explicitly representable; one key with one direction can't express both.
- Comparison operators (`<`, `>`, `>=`) and windows: deferred to a later change.
- Exact-single-day match: dropped — threshold semantics is more useful for GTD.

### Decision 4: No implicit deferral hiding — the default view is an explicit query

`IncludeDeferred` is removed. `TaskFilter` performs no automatic deferral filtering: a nil date predicate means "do not filter on that column." Hiding future-deferred tasks is expressed explicitly as `ready:now`, which is baked into the default query string (`status:pending ready:now`), not into `ListTasks`.

**Rationale:** Every view's deferral stance is visible in its query — no hidden behavior, and "available now" / "still parked" / "everything" are all representable. Cost: the default query is slightly longer, and a caller building a raw `TaskFilter` must add a `Ready` predicate to hide deferred tasks rather than getting it for free. Worth it for the no-magic payoff the rest of the design depends on.

### Decision 5: Parser returns (TaskFilter, error); errors carry a range

- Unrecognized key → not an error; the token becomes free text.
- Recognized key with an invalid value (`status:bogus`, `kind:foo`, `due:notadate`) → a parse error naming the offending token.
- A parse error SHALL carry the `[start, end)` rune-offset range of the offending substring within the input, so the query bar can highlight exactly the bad section (not just print a message).
- On any parse error the TUI does NOT apply the query: it keeps the last successfully-applied results and shows the error inline with the offending range highlighted.

**Rationale:** Distinguishes "I typed a word" from "I asked for a filter that can't exist." A positional range lets the editor underline the problem rather than making the user hunt for it. Not clobbering prior results on a typo keeps editing low-friction.

### Decision 6: Parser lives in `internal/taskquery`, decoupled from TUI

`taskquery.Parse(string) (gtd.TaskFilter, error)` is a pure function with table-driven tests. The TUI only calls it and renders the result/error.

**Rationale:** The grammar is the riskiest logic; isolating it makes it testable without driving bubbletea, and reusable if other views want the same syntax later.

### Decision 7: TUI query bar — live parse, apply on Enter

A `textinput` sits above the list, focused with `/`. Two distinct actions, decoupled:

- **Parse** (validation + error feedback) runs *live*: on Enter, and on a 2-second debounce after the last keystroke. Its only effect is updating the error display — on failure, the offending range (Decision 5) is highlighted in the bar; on success, the error clears. Parsing never reloads the list.
- **Apply** (run `ListTasks`, reload the list) runs *only on Enter*, and only when the query parses successfully.

Esc reverts the bar to the last-applied query. The bar always shows the active query (seeded `status:pending ready:now`). The bubbles list built-in filter keybinding is disabled to avoid two competing filters.

**Rationale:** Middle ground between live-apply (reload churn, jarring) and pure apply-on-Enter (no feedback until you commit). You get red-squiggle-style validation as you type without the list thrashing; the commit to actually filter stays an explicit Enter. The 2-second debounce keeps parsing from firing on every keystroke.

**Alternatives:** Live-apply on every keystroke (rejected — reload churn); validate only on Enter (rejected — no early feedback on a malformed query).

## Risks / Trade-offs

- **[Risk] LIKE `%term%` scans** → For a single-user local SQLite task list the row counts are tiny; full scans are fine. Revisit (FTS5) only if it ever matters.
- **[Risk] Local-vs-UTC date boundary bugs** → Centralize boundary computation in one helper in `taskquery`, unit-tested across timezones, so the conversion lives in exactly one place.
- **[Trade-off] No OR/quoting in v1** → Some queries are inexpressible (e.g. done-or-dropped in one view). Accepted; both are additive later and the single-value/AND model covers the common path.
- **[Risk] Overriding `/` surprises muscle memory** → The query bar is always visible showing the active query, so the affordance is discoverable; `/` to focus mirrors the list's prior filter key.

## Future Considerations

- **Free-text over project name (once `implement-projects` lands).** When tasks gain a project relationship, free-text terms should also match the linked project's name (a JOIN against projects in the sqlite `ListTasks` query). Out of scope here because the project relationship doesn't exist yet; capture as a follow-on once `implement-projects` is built. The query grammar needs no change — only the sqlite free-text clause widens.
