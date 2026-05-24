## 1. Relative-time WHEN formatter

- [x] 1.1 Add a `formatWhen(ref, now time.Time) string` helper (in tasklist or a small internal pkg) implementing the future ladder: timed-today→clock, date-only-today→`today`, +1→`tomorrow`, 2–6d→weekday, 7–30d→`Nd`, >30d→absolute `YYYY-MM-DD`
- [x] 1.2 Implement the past ladder in the same helper: timed-earlier-today→clock, 1–30d ago→`Nd`, >30d ago→absolute `YYYY-MM-DD` (no tomorrow/weekday)
- [x] 1.3 Use calendar-day differences in local tz (truncate to midnight) for all day counts; reuse the midnight check pattern from `date.formatDate` to detect date-only vs timed
- [x] 1.4 Unit-test the ladder boundaries: tomorrow, weekday band edges (2d, 6d), 7d, 30d, 31d→absolute date, timed today, and past `3d`

## 2. Chip model and formatting

- [x] 2.1 Add chip builders that, given a `gtd.Task` and `now`, produce the due/overdue chip (reference = end-of-day for date-only due, exact instant for timed) and the defer/ready chip (reference = start-of-day for date-only defer, exact instant for timed)
- [x] 2.2 Add the `@assignee` chip builder
- [x] 2.3 Apply suppression rules: due/defer suppressed on done & dropped; assignee kept on done; dropped shows no chips
- [x] 2.4 Order chips left-to-right: due/overdue, defer/ready, assignee
- [x] 2.5 Unit-test chip selection: due-today vs overdue (date-only EoD), defer→ready flip at SoD, ready day-count, suppression by status, assignee on done

## 3. Colors

- [x] 3.1 Define urgency color styles (overdue red, due-today orange, due 2–6d yellow, due 7d+ dim, defer dim blue, ready teal, assignee magenta), sourcing from `huh.ThemeCharm` where available and static lipgloss colors otherwise
- [x] 3.2 Plumb terminal background detection into the delegate via `tea.BackgroundColorMsg` (mirror `tui/components/date/date.go`) so adaptive colors resolve
- [x] 3.3 Define per-status title styles: done dim green, dropped dim gray + strikethrough, pending default

## 4. Custom list delegate

- [x] 4.1 Implement a `list.ItemDelegate` for tasks that renders `<marker> <title>  <chips>` with status markers `[ ]`/`[x]`/`[-]`
- [x] 4.2 Truncate the title first under width pressure; keep chips intact; left-align chips inline after the title
- [x] 4.3 Scope the selection highlight to the title (cursor indicator and/or bold); keep chips at their urgency colors on the selected row
- [x] 4.4 Replace `list.NewDefaultDelegate()` wiring in `tasklist.New` (model.go ~69-72) with the custom delegate; preserve existing short/full help funcs and `ShowDescription=false` behavior
- [x] 4.5 Remove/repurpose `Item.Description()` now that rendering moved into the delegate; keep `FilterValue`/`Title` as needed by the list

## 5. Query-side defer/ready boundary fix

- [x] 5.1 Add a `startOfLocalDay` helper alongside `endOfLocalDay` in `internal/taskquery/taskquery.go`
- [x] 5.2 Thread the predicate target into `resolveTime`/`parseDatePredicate` so `due` resolves day-granularity to end-of-day and `defer`/`ready` resolve to start-of-day; `now` stays the exact instant
- [x] 5.3 Update existing taskquery tests to the new defer/ready start-of-day boundaries; add cases asserting `due:` end-of-day and `defer:`/`ready:` start-of-day for keyword, relative, and ISO forms

## 6. Verification

- [x] 6.1 Run `go test ./...`
- [x] 6.2 Run the TUI and eyeball: pending/done/dropped markers, strikethrough legibility, each chip word, urgency colors, multi-chip rows, and selection highlight keeping chip colors
- [x] 6.3 Verify the `@assignee` magenta chip is visually distinct from `list.Model`'s default selected-row highlight; if not, pick a different assignee color
- [x] 6.4 Confirm a date-only defer-today shows `ready:today` and the matching `defer:`/`ready:` filter selects consistently on the boundary day
