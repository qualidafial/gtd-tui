package projectquery

import (
	"errors"
	"testing"

	"github.com/qualidafial/gtd-tui"
	"github.com/qualidafial/gtd-tui/tui/components/querybar"
)

func open() *gtd.ProjectStatus    { return new(gtd.ProjectStatusOpen) }
func someday() *gtd.ProjectStatus { return new(gtd.ProjectStatusSomeday) }
func done() *gtd.ProjectStatus    { return new(gtd.ProjectStatusDone) }
func dropped() *gtd.ProjectStatus { return new(gtd.ProjectStatusDropped) }

func TestParse_Empty(t *testing.T) {
	f, err := Parse("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if f.Status != nil || len(f.Search) != 0 {
		t.Fatalf("empty query should yield zero filter, got %+v", f)
	}
}

func TestParse_StatusFilter(t *testing.T) {
	tests := []struct {
		query string
		want  *gtd.ProjectStatus
	}{
		{"status:open", open()},
		{"status:someday", someday()},
		{"status:done", done()},
		{"status:dropped", dropped()},
	}
	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			f, err := Parse(tt.query)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if f.Status == nil || *f.Status != *tt.want {
				t.Fatalf("Status = %v, want %v", f.Status, tt.want)
			}
		})
	}
}

func TestParse_InvalidStatus_ParseError(t *testing.T) {
	_, err := Parse("status:bogus")
	if err == nil {
		t.Fatal("expected error for invalid status")
	}
	var pe *querybar.ParseError
	if !errors.As(err, &pe) {
		t.Fatalf("expected *querybar.ParseError, got %T", err)
	}
	if pe.Start != 0 || pe.End != len("status:bogus") {
		t.Fatalf("ParseError range = [%d, %d), want [0, %d)", pe.Start, pe.End, len("status:bogus"))
	}
}

func TestParse_FreeTextSearch(t *testing.T) {
	f, err := Parse("shed plans")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(f.Search) != 2 || f.Search[0] != "shed" || f.Search[1] != "plans" {
		t.Fatalf("Search = %v, want [shed plans]", f.Search)
	}
}

func TestParse_UnrecognizedKeyTreatedAsFreeText(t *testing.T) {
	f, err := Parse("foo:bar")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(f.Search) != 1 || f.Search[0] != "foo:bar" {
		t.Fatalf("Search = %v, want [foo:bar]", f.Search)
	}
}

func TestParse_ErrorRange_MiddleToken(t *testing.T) {
	// "shed status:bogus plans" - status:bogus starts at rune 5
	_, err := Parse("shed status:bogus plans")
	if err == nil {
		t.Fatal("expected error")
	}
	var pe *querybar.ParseError
	if !errors.As(err, &pe) {
		t.Fatalf("expected *querybar.ParseError, got %T", err)
	}
	if pe.Start != 5 || pe.End != 17 {
		t.Fatalf("ParseError range = [%d, %d), want [5, 17)", pe.Start, pe.End)
	}
}
