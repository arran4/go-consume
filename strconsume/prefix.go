package strconsume

import (
	"sort"
	"strings"

	"github.com/arran4/go-consume"
)

func NewPrefixConsumer(prefixes ...string) PrefixConsumer {
	matchers := map[int]map[string]struct{}{}
	var sizes []int
	for _, prefix := range prefixes {
		me, ok := matchers[len(prefix)]
		if !ok || me == nil {
			me = map[string]struct{}{}
			sizes = append(sizes, len(prefix))
		}
		me[prefix] = struct{}{}
		matchers[len(prefix)] = me
	}
	sort.Sort(sort.Reverse(sort.IntSlice(sizes)))
	return PrefixConsumer{
		matchers: matchers,
		sizes:    sizes,
	}
}

type PrefixConsumer struct {
	matchers map[int]map[string]struct{}
	sizes    []int
}

// Consume checks if the input string 'from' starts with any of the configured prefixes.
// It returns three values:
// 1. matched: The matched prefix (from the input string itself, preserving case).
// 2. remaining: The rest of the string after the prefix.
// 3. found: True if a prefix was found, false otherwise.
// Options:
// - consume.CaseInsensitive(true): Match prefixes case-insensitively.
func (pc PrefixConsumer) Consume(from string, ops ...any) (string, string, bool) {
	caseInsensitive := false
	for _, op := range ops {
		switch v := op.(type) {
		case consume.CaseInsensitive:
			caseInsensitive = bool(v)
		}
	}

	for _, size := range pc.sizes {
		if size > len(from) {
			continue
		}
		extract := from[:size]

		// Optimization: if not case insensitive, direct lookup
		if !caseInsensitive {
			if _, ok := pc.matchers[size][extract]; ok {
				return extract, from[size:], true
			}
			continue
		}

		// slower path for case sensitivity
		for p := range pc.matchers[size] {
			if strings.EqualFold(extract, p) {
				return extract, from[size:], true
			}
		}
	}
	return "", from, false
}
