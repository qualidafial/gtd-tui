## Context

The task list (`tui/pages/tasks/tasklist`) uses `list.NewDefaultDelegate()` with `ShowDescription=false`, so `Item.Title()` (just `task.Title`) is the only thing rendered. The default delegate owns foreground color, width truncation, and the selected-row highlight, so embedding lipgloss markup in `Title()` would fight it. The TUI is Charm v2 (bubbletea/bubbles/lipgloss) with huh for forms. `tui/components/date/date.go` already demonstrates the patterns we need: `naturaltime` parsing, a midnight check to distinguish date-only from timed values (`formatDate`), and learning the terminal background via `tea.BackgroundColorMsg` to pick adaptive colors.

The query engine (`internal/taskquery`) is the input side of dates; this change is the output/display side. They share a relative-time vocabulary (`Nd`, `Nw`, `today`, weekday) but not code. The one place they must agree is the boundary instant where a task flips overdue/ready — hence the folded-in query fix.

## Goals / Non-Goals

**Goals:**
- A custom list delegate that renders status marker + title + inline chips with per-status and per-urgency colors.
- A single relative-time WHEN formatter with future and past ladders, shared by all four chip words.
- Display/filter agreement on the boundary day: due thresholds at end-of-day, defer/ready at start-of-day.

**Non-Goals:**
- No light-theme tuning beyond what huh's adaptive colors give for free (personal dark-theme tool).
- No changes to the domain, service, or sqlite layers; `Task` is unchanged.
- No changes to the query bar input/parsing UX beyond the defer/ready resolution boundary.
- No right-aligned columns; chips are left-aligned inline (decided to avoid visual row-matching fatigue).

## Decisions

**Custom delegate over DefaultDelegate.** Implement `list.ItemDelegate` directly so the delegate controls truncation (title truncates first, chips preserved), per-status title styling, and a selection highlight scoped to the title. The DefaultDelegate's whole-row recolor and width handling can't express status colors + multi-color chips, so extending it would be more friction than replacing it.

**One WHEN formatter, two ladders.** A pure function `formatWhen(ref, now time.Time) string` is the core. The chip layer decides the *word* (`due`/`overdue`, `defer`/`ready`) by comparing the reference instant to `now`, then calls the formatter for the WHEN. Future and past ladders differ (past has no tomorrow/weekday), so the formatter branches on sign. Day counts use calendar-day differences in local tz (truncate both to midnight, subtract), not 24h spans, so "tomorrow" stays "tomorrow" late at night. Beyond a 30-day window the formatter stops counting days and emits an absolute `YYYY-MM-DD` date (no week rollup), keeping distant chips unambiguous.

**Reference-instant rule lives in the chip layer, not storage.** A date-only value entered in the date field is already stored at local midnight (`parseDate`). For due, the chip treats the reference as end-of-that-day; for defer, start-of-that-day (= the stored midnight). Timed values use their exact instant for both. This keeps the flip semantics in one place and matches the query resolution.

**Query fix mirrors the rule.** `taskquery.resolveTime` currently sends every day-granularity value through `endOfLocalDay`. Split it: `due` stays end-of-day; `defer`/`ready` resolve via a new `startOfLocalDay`. `now` is unchanged (exact instant). This makes `defer:`/`ready:` filters agree with the displayed chip on the boundary day. The resolution branch needs the key (or predicate target) threaded into `resolveTime`/`parseDatePredicate`.

**Colors via huh theme + static fallback.** Pull adaptive colors from `huh.ThemeCharm` where they map cleanly (error/red, etc.); use static lipgloss colors for the urgency bands huh doesn't name (orange, teal, magenta). Learn the background through `tea.BackgroundColorMsg` plumbed into the delegate, same as the date field. Don't over-engineer — dark theme is the only target.

**Selection highlight scoped to title.** The delegate emphasizes the selected title (cursor indicator and/or bold) and leaves chips at their urgency colors, so urgency reads consistently regardless of cursor position.

## Risks / Trade-offs

- [Query resolution change alters existing `defer:`/`ready:` filter boundaries] → Covered by the task-query delta spec and its scenarios; the change is intentional and makes filter match display. Verify the existing taskquery tests are updated, not just passing.
- [Threading the key into `resolveTime` touches a shared parse path] → Keep the signature change minimal (pass the target column/kind already known at the call site in `parseDatePredicate`); avoid broad refactors of the parser.
- [Chip truncation under narrow widths] → Title truncates first with ellipsis; chips are short and kept. Extremely narrow terminals may still clip — acceptable for a personal tool; no special handling.
- [Color legibility on the selected row] → Mitigated by keeping chips at full urgency color and only restyling the title; eyeball during verification (dropped `[-]`/strikethrough especially).
- [Assignee magenta may clash with `list.Model`'s default selected-row highlight] → Flag for manual validation; if `@assignee` and the selection highlight aren't visually distinct, pick a different assignee color.

## Open Questions

- Exact glyph/emphasis for selection (cursor `> ` vs. bold vs. both) — settle visually during implementation.
- Whether `dropped` `[-]` + strikethrough reads clearly in the target terminal — confirm by eye; fall back to a different marker if not.
