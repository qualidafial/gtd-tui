package orderkey

import (
	"math/rand/v2"
	"slices"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBetween_BasicCases(t *testing.T) {
	tests := []struct {
		name   string
		a, b   string
		want   string
		wantOk bool
	}{
		{
			name:   "both empty",
			a:      "",
			b:      "",
			want:   "U",
			wantOk: true,
		},
		{
			name:   "prepend single char",
			a:      "",
			b:      "U",
			want:   "E",
			wantOk: true,
		},
		{
			name:   "append single char",
			a:      "V",
			b:      "",
			want:   "k",
			wantOk: true,
		},
		{
			name:   "large gap",
			a:      "A",
			b:      "z",
			want:   "Z",
			wantOk: true,
		},
		{
			name:   "single digit gap",
			a:      "A",
			b:      "B",
			want:   "AU",
			wantOk: true,
		},
		{
			name:   "common prefix, then digit gap",
			a:      "AB",
			b:      "AC",
			want:   "ABU",
			wantOk: true,
		},
		{
			name:   "left is prefix of right",
			a:      "A",
			b:      "AB",
			want:   "A5",
			wantOk: true,
		},
		{
			name:   "right diverges by one digit from left",
			a:      "AB",
			b:      "B",
			want:   "Aa",
			wantOk: true,
		},
		{
			name:   "both same length, single digit diff",
			a:      "Aa",
			b:      "Ab",
			want:   "AaU",
			wantOk: true,
		},
		{
			name:   "deep common prefix",
			a:      "abcdef",
			b:      "abcdfg",
			want:   "abcdep",
			wantOk: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := Between(tt.a, tt.b)
			if ok != tt.wantOk || got != tt.want {
				t.Fatalf("Between(%v, %v) = (%q, %v), want (%q, %v)", tt.a, tt.b, got, ok, tt.want, tt.wantOk)
			}
			if tt.wantOk {
				if tt.a != "" && !(tt.a < got) {
					t.Errorf("expected a (%q) < result (%q)", tt.a, got)
				}
				if tt.b != "" && !(got < tt.b) {
					t.Errorf("expected result (%q) < b (%q)", got, tt.b)
				}
			}
		})
	}
}

func TestBetween_MonotonicUnderRandomInserts(t *testing.T) {
	// Build up a list by inserting at random positions and verify the
	// keys stay in strictly increasing order throughout. If a prepend
	// drives keys all the way to the alphabet floor (or an append to
	// the ceiling) Between will panic — recover by renumbering, which
	// is the documented bailout for callers.
	rng := rand.New(rand.NewPCG(1, 2))
	first, _ := Between("", "")
	keys := []string{first}
	inserts := 1000
	var renumbers int
	for i := 0; i < inserts; i++ {
		idx := rng.IntN(len(keys) + 1)
		var left, right string
		if idx > 0 {
			left = keys[idx-1]
		}
		if idx < len(keys) {
			right = keys[idx]
		}

		var ok bool
		newKey, ok := Between(left, right)
		if !ok {
			renumbers++
			keys = Renumber(len(keys))
			continue
		}
		keys = slices.Insert(keys, idx, newKey)

		if !sort.StringsAreSorted(keys) {
			t.Fatalf("keys lost ordering after insert %d at index %d:\n%v", i, idx, keys)
		}
	}
	t.Logf("renumbered %d times across %d inserts", renumbers, inserts)
}

func TestBetween_PanicsOnInvalidBounds(t *testing.T) {
	tests := []struct {
		name      string
		a, b      string
		want      string
		wantOk    bool
		wantPanic bool
	}{
		{
			name:      "a equals b",
			a:         "K",
			b:         "K",
			wantPanic: true,
		},
		{
			name:      "a greater than b",
			a:         "Z",
			b:         "A",
			wantPanic: true,
		},
		{
			name:   "no room below alphabet floor",
			a:      "",
			b:      "0",
			want:   "",
			wantOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				r := recover()
				if tt.wantPanic && r == nil {
					t.Fatalf("expected panic for Between(%q, %q)", tt.a, tt.b)
				}
				if !tt.wantPanic && r != nil {
					t.Fatalf("unexpected panic for Between(%q, %q): %v", tt.a, tt.b, r)
				}
			}()
			got, ok := Between(tt.a, tt.b)
			if ok != tt.wantOk {
				t.Fatalf("Between(%q, %q) => ok %v; want %v", tt.a, tt.b, ok, tt.wantOk)
			}
			if tt.wantOk {
				if tt.a != "" && !(tt.a < got) {
					t.Errorf("expected a (%q) < result (%q)", tt.a, got)
				}
				if tt.b != "" && !(got < tt.b) {
					t.Errorf("expected result (%q) < b (%q)", got, tt.b)
				}
			}
		})
	}
}

func TestRenumber_ProducesStrictlyIncreasingKeys(t *testing.T) {
	tests := []int{0, 1, 2, 10, 100, 200}
	for _, n := range tests {
		keys := Renumber(n)
		if len(keys) != n {
			t.Errorf("Renumber(%d) returned %d keys", n, len(keys))
		}
		for i := 1; i < len(keys); i++ {
			if !(keys[i-1] < keys[i]) {
				t.Errorf("Renumber(%d): keys[%d]=%q not less than keys[%d]=%q",
					n, i-1, keys[i-1], i, keys[i])
			}
		}
	}
}

func TestRenumber_LeavesHeadroomAtBeginning(t *testing.T) {
	// Renumber output should never sit at the alphabet floor,
	// so callers can still prepend after backfilling.
	keys := Renumber(200)
	for _, k := range keys {
		if k == "0" {
			t.Errorf("Renumber produced floor key %q; cannot prepend below it", k)
		}
	}
	// Confirm we can insert before the first
	_, ok := Between("", keys[0])
	assert.True(t, ok)
}
