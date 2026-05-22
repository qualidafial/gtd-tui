// Package orderkey produces sortable fractional-indexing keys so list
// items can be reordered with single-row updates regardless of list
// length. A new key between two existing ones is always representable
// without renumbering: the algorithm picks the shortest string K such
// that A < K < B under lexicographic ordering.
package orderkey

import (
	"fmt"
	"strings"
)

// alphabet is the digit set used to encode keys. Characters are in
// strict ASCII order so lexicographic comparison matches digit order.
const alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

const base = len(alphabet)

// Between returns a key K such that a < K < b under lexicographic
// ordering. An empty a is treated as -∞; an empty b is treated as +∞.
// Both empty returns a canonical middle key suitable for the first
// item in a list.
//
// Between returns false when no key fits in the requested gap. This
// case only arises when a is empty and b sits at the alphabet floor
// (e.g. "0"); the caller should renumber the list rather than retry.
//
// Between panics when a >= b (with both non-empty).
func Between(a, b string) (string, bool) {
	if a != "" && b != "" && a >= b {
		panic(fmt.Sprintf("orderkey.Between: requires a < b, got a=%q b=%q", a, b))
	}

	var sb strings.Builder

	for i := 0; ; i++ {
		ao := -1
		bo := base
		if i < len(a) {
			ao = ordinal(a[i])
		}
		if i < len(b) {
			bo = ordinal(b[i])
		}

		gap := bo - ao
		switch {
		case gap > 1:
			mid := (ao + bo) / 2
			sb.WriteByte(alphabet[mid])
			return sb.String(), true
		case gap == 0:
			// Identical character in both inputs — commit it to the
			// shared prefix and keep walking.
			sb.WriteByte(alphabet[ao])
		case gap == 1 && ao >= 0:
			// a and b diverge here by a single digit. Commit a's digit
			// to the prefix and treat b as +∞ for the remaining
			// iterations: any extension we append is automatically less
			// than the rest of b.
			sb.WriteByte(alphabet[ao])
			b = ""
		case gap == 1 && ao < 0:
			// a is exhausted and b sits at the alphabet floor at this
			// position. Committing b's floor digit keeps us tied with
			// b's prefix; if b still has more digits after this one,
			// the committed prefix is strictly shorter than b and is
			// our answer. Otherwise no key fits between a and b.
			sb.WriteByte(alphabet[bo])
			if i+1 < len(b) {
				return sb.String(), true
			}
			return "", false // no key fits
		}
	}
}

// Renumber returns evenly-spaced canonical keys for n items. Use it to
// recover when keys have grown unwieldy or when backfilling a new
// ordering column. The returned slice has length n and the keys are
// strictly increasing.
func Renumber(n int) []string {
	if n <= 0 {
		return nil
	}
	keys := make([]string, n)
	// Reserve room at the beginning by distributing across the inner range
	// of the alphabet. The "+1" denominator places every key strictly
	// inside (0, base], never on the floor.
	for i := range keys {
		ord := (i + 1) * base / (n + 1)
		if ord < 1 {
			ord = 1
		}
		if ord >= base {
			ord = base - 1
		}
		keys[i] = string(alphabet[ord])
	}
	// Spacing collapses for large n; fall back to chained Between so
	// keys remain strictly increasing.
	for i := 1; i < n; i++ {
		if keys[i-1] >= keys[i] {
			keys[i], _ = Between(keys[i-1], "") // always succeeds
		}
	}
	return keys
}

func ordinal(c byte) int {
	switch {
	case c >= '0' && c <= '9':
		return int(c - '0')
	case c >= 'A' && c <= 'Z':
		return int(c-'A') + 10
	case c >= 'a' && c <= 'z':
		return int(c-'a') + 36
	}
	panic(fmt.Sprintf("orderkey: invalid character %q in key", c))
}
