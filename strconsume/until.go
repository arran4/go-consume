package strconsume

import (
	"bufio"
	"sort"
	"strings"

	"github.com/arran4/go-consume"
)

func NewUntilConsumer(s ...string) UntilConsumer {
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
	return UntilConsumer{
		matchers: matchers,
		sizes:    sizes,
	}
}

type UntilConsumer struct {
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
func (cu UntilConsumer) Consume(from string, ops ...any) (string, string, string, bool) {
	inclusive := false
	startOffset := 0
	ignore0PositionMatch := false
	caseInsensitive := false
	for _, op := range ops {
		switch v := op.(type) {
		case consume.Inclusive:
			inclusive = bool(v)
		case consume.StartOffset:
			startOffset = int(v)
		case consume.Ignore0PositionMatch:
			ignore0PositionMatch = bool(v)
		case consume.CaseInsensitive:
			caseInsensitive = bool(v)
		}
	}
	for i := startOffset; i < len(from); i++ {
		for _, size := range cu.sizes {
			if i+size > len(from) {
				continue
			}
			extract := from[i : i+size]

			match := false
			separator := ""

			if !caseInsensitive {
				if _, ok := cu.matchers[size][extract]; ok {
					match = true
					separator = extract
				}
			} else {
				for s := range cu.matchers[size] {
					if strings.EqualFold(extract, s) {
						match = true
						separator = extract // matched from input
						break
					}
				}
			}

			if match {
				if i == 0 && ignore0PositionMatch {
					continue
				}
				matched := from[:i]
				if inclusive {
					return matched + separator, separator, from[i+size:], true
				}
				return matched, separator, from[i:], true
			}
		}
	}
	return "", "", from, false
}

func (cu UntilConsumer) SplitFunc(ops ...any) bufio.SplitFunc {
	inclusive := false
	startOffset := 0
	ignore0PositionMatch := false
	caseInsensitive := false
	for _, op := range ops {
		switch v := op.(type) {
		case consume.Inclusive:
			inclusive = bool(v)
		case consume.StartOffset:
			startOffset = int(v)
		case consume.Ignore0PositionMatch:
			ignore0PositionMatch = bool(v)
		case consume.CaseInsensitive:
			caseInsensitive = bool(v)
		}
	}
	return func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}

		for i := startOffset; i < len(data); i++ {
			for _, size := range cu.sizes {
				if i+size > len(data) {
					continue
				}
				extract := data[i : i+size]
				extractStr := string(extract)

				match := false

				if !caseInsensitive {
					if _, ok := cu.matchers[size][extractStr]; ok {
						match = true
					}
				} else {
					for s := range cu.matchers[size] {
						if strings.EqualFold(extractStr, s) {
							match = true
							break
						}
					}
				}

				if match {
					if i == 0 && ignore0PositionMatch {
						continue
					}

					advance = i + size
					if inclusive {
						token = data[:i+size]
					} else {
						token = data[:i]
					}
					return advance, token, nil
				}
			}
		}

		if atEOF {
			return len(data), data, nil
		}
		return 0, nil, nil
	}
}
