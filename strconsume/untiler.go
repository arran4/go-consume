package strconsume

import (
	"sort"

	"github.com/arran4/go-consume"
)

func NewConsumeUntiler(s ...string) ConsumeUntiler {
	matchers := map[int]map[string]struct{}{}
	var sizes []int
	for _, se := range s {
		me, ok := matchers[len(se)]
		if !ok || me == nil {
			me = map[string]struct{}{}
			sizes = append(sizes, len(se))
		}
		me[se] = struct{}{}
		matchers[len(se)] = me
	}
	sort.Sort(sort.Reverse(sort.IntSlice(sizes)))
	return ConsumeUntiler{
		matchers: matchers,
		sizes:    sizes,
	}
}

type ConsumeUntiler struct {
	matchers map[int]map[string]struct{}
	sizes    []int
}

// Consume scans the input string 'from' for any of the configured separators.
// It returns four values:
// 1. matched: The substring before the found separator.
// 2. separator: The separator that was found.
// 3. remaining: The rest of the string. If inclusive is true, this starts after the separator. If false, it starts at the separator.
// 4. found: True if a separator was found, false otherwise.
// If no separator is found, it returns ("", "", from, false).
func (cu ConsumeUntiler) Consume(from string, ops ...any) (string, string, string, bool) {
	inclusive := false
	startOffset := 0
	ignore0PositionMatch := false
	for _, op := range ops {
		switch v := op.(type) {
		case consume.Inclusive:
			inclusive = bool(v)
		case consume.StartOffset:
			startOffset = int(v)
		case consume.Ignore0PositionMatch:
			ignore0PositionMatch = bool(v)
		}
	}
	for i := startOffset; i < len(from); i++ {
		for _, size := range cu.sizes {
			if i+size > len(from) {
				continue
			}
			extract := from[i : i+size]
			if _, ok := cu.matchers[size][extract]; ok {
				if i == 0 && ignore0PositionMatch {
					continue
				}
				matched := from[:i]
				if inclusive {
					return matched + extract, extract, from[i+size:], true
				}
				return matched, extract, from[i:], true
			}
		}
	}
	return "", "", from, false
}
