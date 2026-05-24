// Package taskquery parses a compact task-list query string into a
// gtd.TaskFilter. The grammar is whitespace-tokenized: a token of the form
// key:value whose key is recognized sets a structured filter field; any other
// token is a free-text search term.
package taskquery

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/qualidafial/gtd-tui"
)

// ParseError reports an invalid value for a recognized key. Start and End are
// rune offsets into the original query marking the [start, end) range of the
// offending token, so a caller can highlight exactly the bad section.
type ParseError struct {
	Message string
	Start   int
	End     int
}

func (e *ParseError) Error() string { return e.Message }

// token is a whitespace-delimited slice of the input with its rune offsets.
type token struct {
	text  string
	start int // rune offset of first rune
	end   int // rune offset just past last rune
}

// tokenize splits s on whitespace, recording rune offsets for each token.
func tokenize(s string) []token {
	var tokens []token
	var cur []rune
	start := 0
	flush := func(end int) {
		if len(cur) > 0 {
			tokens = append(tokens, token{text: string(cur), start: start, end: end})
			cur = nil
		}
	}
	i := 0
	for _, r := range s {
		if r == ' ' || r == '\t' || r == '\n' || r == '\r' {
			flush(i)
		} else {
			if len(cur) == 0 {
				start = i
			}
			cur = append(cur, r)
		}
		i++
	}
	flush(i)
	return tokens
}

var recognizedKeys = map[string]bool{
	"status":   true,
	"kind":     true,
	"assignee": true,
	"due":      true,
	"defer":    true,
	"ready":    true,
}

// Parse converts a query string into a gtd.TaskFilter. An empty query yields a
// zero-value filter and no error. A recognized key with an invalid value
// returns a *ParseError; an unrecognized key:value token is treated as free
// text. Repeated keys are last-wins.
func Parse(query string) (gtd.TaskFilter, error) {
	return parseAtLoc(query, time.Now(), time.Local)
}

// parseAt is a testing convenience that resolves dates in the local timezone.
func parseAt(query string, now time.Time) (gtd.TaskFilter, error) {
	return parseAtLoc(query, now, time.Local)
}

// parseAtLoc is the testable core; now is the reference instant and loc is the
// timezone used to resolve day-granularity values to end-of-day.
func parseAtLoc(query string, now time.Time, loc *time.Location) (gtd.TaskFilter, error) {
	var filter gtd.TaskFilter

	for _, tok := range tokenize(query) {
		key, value, isKV := splitKeyValue(tok.text)
		if !isKV || !recognizedKeys[key] {
			filter.Search = append(filter.Search, tok.text)
			continue
		}

		if err := applyKeyValue(&filter, key, value, tok, now, loc); err != nil {
			return gtd.TaskFilter{}, err
		}
	}

	return filter, nil
}

// splitKeyValue splits "key:value" on the first colon. Returns isKV=false when
// there is no colon.
func splitKeyValue(s string) (key, value string, isKV bool) {
	return strings.Cut(s, ":")
}

func applyKeyValue(filter *gtd.TaskFilter, key, value string, tok token, now time.Time, loc *time.Location) error {
	switch key {
	case "status":
		s, err := parseStatus(value)
		if err != nil {
			return tokenError(tok)
		}
		filter.Status = &s
	case "kind":
		k, err := parseKind(value)
		if err != nil {
			return tokenError(tok)
		}
		filter.Kind = &k
	case "assignee":
		v := value
		filter.Assignee = &v
	case "due":
		p, err := parseDatePredicate(value, gtd.OnOrBefore, true, now, loc)
		if err != nil {
			return tokenError(tok)
		}
		filter.Due = p
	case "defer":
		p, err := parseDatePredicate(value, gtd.After, true, now, loc)
		if err != nil {
			return tokenError(tok)
		}
		filter.Defer = p
	case "ready":
		p, err := parseDatePredicate(value, gtd.AvailableAsOf, false, now, loc)
		if err != nil {
			return tokenError(tok)
		}
		filter.Ready = p
	}
	return nil
}

func tokenError(tok token) error {
	return &ParseError{
		Message: fmt.Sprintf("invalid query token %q", tok.text),
		Start:   tok.start,
		End:     tok.end,
	}
}

func parseStatus(v string) (gtd.TaskStatus, error) {
	switch v {
	case "pending":
		return gtd.TaskStatusPending, nil
	case "done":
		return gtd.TaskStatusDone, nil
	case "dropped":
		return gtd.TaskStatusDropped, nil
	}
	return "", fmt.Errorf("invalid status %q", v)
}

func parseKind(v string) (gtd.TaskKind, error) {
	switch v {
	case "next_action":
		return gtd.TaskKindNextAction, nil
	case "delegated":
		return gtd.TaskKindDelegated, nil
	}
	return "", fmt.Errorf("invalid kind %q", v)
}

// parseDatePredicate resolves a date value into a *gtd.DatePredicate. thresholdKind
// is the time-based kind for this key (OnOrBefore/After/AvailableAsOf).
// allowNullVariants enables none/any (IsNull/IsNotNull); ready disallows them.
func parseDatePredicate(v string, thresholdKind gtd.DatePredicateKind, allowNullVariants bool, now time.Time, loc *time.Location) (*gtd.DatePredicate, error) {
	switch v {
	case "none":
		if !allowNullVariants {
			return nil, fmt.Errorf("value %q not allowed here", v)
		}
		return &gtd.DatePredicate{Kind: gtd.IsNull}, nil
	case "any":
		if !allowNullVariants {
			return nil, fmt.Errorf("value %q not allowed here", v)
		}
		return &gtd.DatePredicate{Kind: gtd.IsNotNull}, nil
	}

	t, err := resolveTime(v, now, loc)
	if err != nil {
		return nil, err
	}
	return &gtd.DatePredicate{Kind: thresholdKind, Time: t.UTC()}, nil
}

// resolveTime resolves a date value to a time. `now` is the current instant;
// every other form resolves to end-of-day in loc.
func resolveTime(v string, now time.Time, loc *time.Location) (time.Time, error) {
	if v == "now" {
		return now, nil
	}

	// Keyword aliases over relative durations.
	switch v {
	case "overdue":
		v = "-1d"
	case "today":
		v = "0d"
	case "week":
		v = "7d"
	}

	if days, ok, err := parseRelative(v); err != nil {
		return time.Time{}, err
	} else if ok {
		day := now.In(loc).AddDate(0, 0, days)
		return endOfLocalDay(day, loc), nil
	}

	// ISO date.
	if t, err := time.ParseInLocation("2006-01-02", v, loc); err == nil {
		return endOfLocalDay(t, loc), nil
	}

	return time.Time{}, fmt.Errorf("invalid date %q", v)
}

// parseRelative parses -Nd / Nd / Nw into a signed number of days. ok is false
// when v is not a relative-duration form (so callers can try other forms).
func parseRelative(v string) (days int, ok bool, err error) {
	if v == "" {
		return 0, false, nil
	}
	unit := v[len(v)-1]
	if unit != 'd' && unit != 'w' {
		return 0, false, nil
	}
	numStr := v[:len(v)-1]
	n, perr := strconv.Atoi(numStr)
	if perr != nil {
		return 0, false, fmt.Errorf("invalid duration %q", v)
	}
	if unit == 'w' {
		n *= 7
	}
	return n, true, nil
}

// endOfLocalDay returns 23:59:59.999999999 of t's calendar day in loc.
func endOfLocalDay(t time.Time, loc *time.Location) time.Time {
	local := t.In(loc)
	return time.Date(local.Year(), local.Month(), local.Day(), 23, 59, 59, int(time.Second-time.Nanosecond), loc)
}
