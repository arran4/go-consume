package strconsume

import (
	"bufio"
	"bytes"
	"sort"
	"strings"
	"unicode/utf8"

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
// Options:
// - consume.Inclusive(true): If true, matched includes the separator, and remaining starts after it.
// - consume.StartOffset(n): Starts the search at offset n.
// - consume.Ignore0PositionMatch(true): Ignores matches at the start of the string.
// - consume.CaseInsensitive(true): Matches separators case-insensitively.
// - consume.ConsumeRemainingIfNotFound(true): If no separator is found, return the whole string as matched, empty separator, and true.
// - consume.Escape("string"): Specifies an escape string (e.g. "\\"). Can be specified multiple times.
// - consume.Encasing{Start: "(", End: ")"}: Specifies an encasing pair. Can be specified multiple times.
// - consume.EscapeBreaksEncasing(true): If true, escape strings work inside encasings.
func (cu UntilConsumer) Consume(from string, ops ...any) (string, string, string, bool) {
	inclusive := false
	startOffset := 0
	ignore0PositionMatch := false
	caseInsensitive := false
	consumeRemainingIfNotFound := false
	var escapes []string
	var encasings []consume.Encasing
	escapeBreaksEncasing := false
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
		case consume.ConsumeRemainingIfNotFound:
			consumeRemainingIfNotFound = bool(v)
		case consume.Escape:
			if len(string(v)) == 0 {
				panic("consume: escape string cannot be empty")
			}
			escapes = append(escapes, string(v))
		case consume.Encasing:
			if len(v.Start) == 0 {
				panic("consume: encasing start cannot be empty")
			}
			encasings = append(encasings, v)
		case consume.EscapeBreaksEncasing:
			escapeBreaksEncasing = bool(v)
		}
	}

	var encasingStack []consume.Encasing

	for i := startOffset; i < len(from); {
		if len(encasingStack) > 0 {
			current := encasingStack[len(encasingStack)-1]

			if escapeBreaksEncasing && len(escapes) > 0 {
				foundEscape := false
				for _, esc := range escapes {
					if strings.HasPrefix(from[i:], esc) {
						i += len(esc)
						if i < len(from) {
							_, w := utf8.DecodeRuneInString(from[i:])
							i += w
						}
						foundEscape = true
						break
					}
				}
				if foundEscape {
					continue
				}
			}

			if strings.HasPrefix(from[i:], current.End) {
				encasingStack = encasingStack[:len(encasingStack)-1]
				i += len(current.End)
				continue
			}

			if current.Start != current.End {
				foundStart := false
				for _, enc := range encasings {
					if strings.HasPrefix(from[i:], enc.Start) {
						encasingStack = append(encasingStack, enc)
						i += len(enc.Start)
						foundStart = true
						break
					}
				}
				if foundStart {
					continue
				}
			}

			_, w := utf8.DecodeRuneInString(from[i:])
			i += w
			continue
		}

		if len(escapes) > 0 {
			foundEscape := false
			for _, esc := range escapes {
				if strings.HasPrefix(from[i:], esc) {
					i += len(esc)
					if i < len(from) {
						_, w := utf8.DecodeRuneInString(from[i:])
						i += w
					}
					foundEscape = true
					break
				}
			}
			if foundEscape {
				continue
			}
		}

		if len(encasings) > 0 {
			foundStart := false
			for _, enc := range encasings {
				if strings.HasPrefix(from[i:], enc.Start) {
					encasingStack = append(encasingStack, enc)
					i += len(enc.Start)
					foundStart = true
					break
				}
			}
			if foundStart {
				continue
			}
		}

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
					break
				}
				matched := from[:i]
				if inclusive {
					return matched + separator, separator, from[i+size:], true
				}
				return matched, separator, from[i:], true
			}
		}
		_, w := utf8.DecodeRuneInString(from[i:])
		i += w
	}
	if consumeRemainingIfNotFound {
		return from, "", "", true
	}
	return "", "", from, false
}

func (cu UntilConsumer) SplitFunc(ops ...any) bufio.SplitFunc {
	inclusive := false
	startOffset := 0
	ignore0PositionMatch := false
	caseInsensitive := false
	var escapes []string
	var encasings []consume.Encasing
	escapeBreaksEncasing := false
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
		case consume.Escape:
			if len(string(v)) == 0 {
				panic("consume: escape string cannot be empty")
			}
			escapes = append(escapes, string(v))
		case consume.Encasing:
			if len(v.Start) == 0 {
				panic("consume: encasing start cannot be empty")
			}
			encasings = append(encasings, v)
		case consume.EscapeBreaksEncasing:
			escapeBreaksEncasing = bool(v)
		}
	}
	var escapesBytes [][]byte
	for _, e := range escapes {
		escapesBytes = append(escapesBytes, []byte(e))
	}
	type encasingBytes struct {
		Start, End []byte
	}
	var encasingsBytes []encasingBytes
	for _, e := range encasings {
		encasingsBytes = append(encasingsBytes, encasingBytes{[]byte(e.Start), []byte(e.End)})
	}

	return func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}

		var encasingStack []encasingBytes

		for i := startOffset; i < len(data); {
			if len(encasingStack) > 0 {
				current := encasingStack[len(encasingStack)-1]

				if escapeBreaksEncasing && len(escapesBytes) > 0 {
					foundEscape := false
					for _, esc := range escapesBytes {
						if bytes.HasPrefix(data[i:], esc) {
							i += len(esc)
							if i < len(data) {
								_, w := utf8.DecodeRune(data[i:])
								i += w
							}
							foundEscape = true
							break
						}
					}
					if foundEscape {
						continue
					}
				}

				if bytes.HasPrefix(data[i:], current.End) {
					encasingStack = encasingStack[:len(encasingStack)-1]
					i += len(current.End)
					continue
				}

				if !bytes.Equal(current.Start, current.End) {
					foundStart := false
					for _, enc := range encasingsBytes {
						if bytes.HasPrefix(data[i:], enc.Start) {
							encasingStack = append(encasingStack, enc)
							i += len(enc.Start)
							foundStart = true
							break
						}
					}
					if foundStart {
						continue
					}
				}

				_, w := utf8.DecodeRune(data[i:])
				i += w
				continue
			}

			if len(escapesBytes) > 0 {
				foundEscape := false
				for _, esc := range escapesBytes {
					if bytes.HasPrefix(data[i:], esc) {
						i += len(esc)
						if i < len(data) {
							_, w := utf8.DecodeRune(data[i:])
							i += w
						}
						foundEscape = true
						break
					}
				}
				if foundEscape {
					continue
				}
			}

			if len(encasingsBytes) > 0 {
				foundStart := false
				for _, enc := range encasingsBytes {
					if bytes.HasPrefix(data[i:], enc.Start) {
						encasingStack = append(encasingStack, enc)
						i += len(enc.Start)
						foundStart = true
						break
					}
				}
				if foundStart {
					continue
				}
			}

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
						break
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
			_, w := utf8.DecodeRune(data[i:])
			i += w
		}

		if atEOF {
			return len(data), data, nil
		}
		return 0, nil, nil
	}
}

// Iterator provides a func(yield func(string, string) bool) iterator pattern.
// It iterates over the input string, splitting it by the configured separators.
// The yielded values are (matched, separator).
// The last yielded value will be the remaining string with an empty separator, unless the input was fully consumed by separators (e.g. empty string or ends with separator, but Wait: strings.Split yields empty string at end if string ends with separator).
// Options:
// - consume.Inclusive(true): If true, matched includes the separator, and remaining starts after it.
// - consume.StartOffset(n): Starts the first search at offset n. Subsequent searches start from the beginning of the remaining string.
// - consume.Ignore0PositionMatch(true): Ignores matches at the start of the string (for each iteration step).
// - consume.CaseInsensitive(true): Matches separators case-insensitively.
func (cu UntilConsumer) Iterator(from string, ops ...any) func(yield func(string, string) bool) {
	return func(yield func(string, string) bool) {
		inclusive := false

		var loopOps []any
		firstRun := true

		for _, op := range ops {
			switch v := op.(type) {
			case consume.Inclusive:
				inclusive = bool(v)
				loopOps = append(loopOps, op)
			case consume.StartOffset:
				// StartOffset only applies to the first run, so we don't add it to loopOps
			default:
				loopOps = append(loopOps, op)
			}
		}

		for {
			var currentOps []any
			if firstRun {
				currentOps = ops
			} else {
				currentOps = loopOps
			}

			matched, separator, remaining, found := cu.Consume(from, currentOps...)
			firstRun = false

			if !found {
				yield(from, "")
				return
			}

			if !yield(matched, separator) {
				return
			}

			if inclusive {
				from = remaining
			} else {
				if len(separator) > 0 {
					from = remaining[len(separator):]
				} else {
					if len(matched) > 0 {
						from = remaining
					} else {
						if len(from) > 0 {
							from = from[1:]
						} else {
							return
						}
					}
				}
			}
		}
	}
}
