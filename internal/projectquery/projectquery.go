// Package projectquery parses a compact project-list query string into a
// gtd.ProjectFilter. The grammar is whitespace-tokenized: a token of the form
// key:value whose key is recognized sets a structured filter field; any other
// token is a free-text search term.
package projectquery

import (
	"fmt"
	"strings"

	"github.com/qualidafial/gtd-tui"
	"github.com/qualidafial/gtd-tui/tui/components/querybar"
)

// token is a whitespace-delimited slice of the input with its rune offsets.
type token struct {
	text  string
	start int
	end   int
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
	"status": true,
}

// Parse converts a query string into a gtd.ProjectFilter. An empty query
// yields a zero-value filter and no error. A recognized key with an invalid
// value returns a *querybar.ParseError; an unrecognized key:value token is
// treated as free text. Repeated keys are last-wins.
func Parse(query string) (gtd.ProjectFilter, error) {
	var filter gtd.ProjectFilter

	for _, tok := range tokenize(query) {
		key, value, isKV := strings.Cut(tok.text, ":")
		if !isKV || !recognizedKeys[key] {
			filter.Search = append(filter.Search, tok.text)
			continue
		}

		switch key {
		case "status":
			s, err := parseStatus(value)
			if err != nil {
				return gtd.ProjectFilter{}, &querybar.ParseError{
					Message: fmt.Sprintf("invalid query token %q", tok.text),
					Start:   tok.start,
					End:     tok.end,
				}
			}
			filter.Status = &s
		}
	}

	return filter, nil
}

func parseStatus(v string) (gtd.ProjectStatus, error) {
	switch v {
	case "open":
		return gtd.ProjectStatusOpen, nil
	case "someday":
		return gtd.ProjectStatusSomeday, nil
	case "done":
		return gtd.ProjectStatusDone, nil
	case "dropped":
		return gtd.ProjectStatusDropped, nil
	}
	return "", fmt.Errorf("invalid status %q", v)
}
