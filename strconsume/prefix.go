package strconsume

import (
	"bufio"
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
// - consume.MustMatchWholeString(true): The prefix must match the entire remaining string.
func (pc PrefixConsumer) Consume(from string, ops ...any) (string, string, bool) {
	caseInsensitive := false
	mustMatchWholeString := false
	for _, op := range ops {
		switch v := op.(type) {
		case consume.CaseInsensitive:
			caseInsensitive = bool(v)
		case consume.MustMatchWholeString:
			mustMatchWholeString = bool(v)
		}
	}

	for _, size := range pc.sizes {
		if size > len(from) {
			continue
		}
		extract := from[:size]

		match := false
		// Optimization: if not case insensitive, direct lookup
		if !caseInsensitive {
			if _, ok := pc.matchers[size][extract]; ok {
				match = true
			}
		} else {
			// slower path for case sensitivity
			for p := range pc.matchers[size] {
				if strings.EqualFold(extract, p) {
					match = true
					break
				}
			}
		}
		if match {
			remaining := from[size:]
			if mustMatchWholeString && len(remaining) > 0 {
				continue
			}
			return extract, remaining, true
		}
	}
	return "", from, false
}

func (pc PrefixConsumer) SplitFunc(ops ...any) bufio.SplitFunc {
	caseInsensitive := false
	for _, op := range ops {
		switch v := op.(type) {
		case consume.CaseInsensitive:
			caseInsensitive = bool(v)
		}
	}

	return func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}

		for _, size := range pc.sizes {
			if size > len(data) {
				continue
			}
			extract := data[:size]
			extractStr := string(extract)

			match := false
			if !caseInsensitive {
				if _, ok := pc.matchers[size][extractStr]; ok {
					match = true
				}
			} else {
				for p := range pc.matchers[size] {
					if strings.EqualFold(extractStr, p) {
						match = true
						break
					}
				}
			}

			if match {
				return size, extract, nil
			}
		}

		if atEOF {
			return 0, nil, nil
		}
		return 0, nil, nil
	}
}

// Iterator provides a func(yield func(string, string) bool) iterator pattern.
// It iterates over the input string, repeatedly consuming prefixes.
// The yielded values are (matched, remaining).
// The iteration stops when no prefix matches. The unmatched remainder is NOT yielded.
// Options:
// - consume.CaseInsensitive(true): Match prefixes case-insensitively.
func (pc PrefixConsumer) Iterator(from string, ops ...any) func(yield func(string, string) bool) {
	return func(yield func(string, string) bool) {
		for {
			matched, remaining, found := pc.Consume(from, ops...)
			if !found {
				return
			}
			if !yield(matched, remaining) {
				return
			}
			// Avoid infinite loop if prefix is empty
			if len(matched) == 0 {
				return
			}
			from = remaining
		}
	}
}
