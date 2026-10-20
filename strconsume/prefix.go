package strconsume

import (
	"bufio"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/arran4/go-consume"
)

type trieNode struct {
	segment  string
	children []*trieNode
	isEnd    bool
	fullPath string
}

type PrefixConsumer struct {
	root *trieNode
}

func NewPrefixConsumer(paths ...string) *PrefixConsumer {
	if len(paths) == 0 {
		return &PrefixConsumer{root: &trieNode{}}
	}
	sorted := make([]string, len(paths))
	copy(sorted, paths)
	slices.Sort(sorted)

	root := &trieNode{}

	// Build trie from sorted list
	var insert func(parent *trieNode, start, end int, depth int)
	insert = func(parent *trieNode, start, end int, depth int) {
		if start >= end {
			return
		}

		i := start
		// Check if the items end here (are a prefix of others in this group)
		// Use loop to handle duplicates
		for i < end && len(sorted[i]) == depth {
			parent.fullPath = sorted[i]
			parent.isEnd = true
			i++
		}

		for i < end {
			// Identify group starting with same char
			char := sorted[i][depth]
			groupStart := i
			i++
			// Scan to find end of group
			for i < end && len(sorted[i]) > depth && sorted[i][depth] == char {
				i++
			}

			// Found a group [groupStart, i)
			// Find common prefix for this group to compress the trie (Patricia Trie optimization)
			lcp := 0
			s1 := sorted[groupStart]
			s2 := sorted[i-1]
			minLen := len(s1)
			if len(s2) < minLen {
				minLen = len(s2)
			}

			for k := 0; k < minLen-depth; k++ {
				if s1[depth+k] == s2[depth+k] {
					lcp++
				} else {
					break
				}
			}

			// Create node
			segment := s1[depth : depth+lcp]
			node := &trieNode{segment: segment}
			parent.children = append(parent.children, node)

			// Recurse
			insert(node, groupStart, i, depth+lcp)
		}
	}

	insert(root, 0, len(sorted), 0)

	return &PrefixConsumer{root: root}
}

// LongestPrefix finds the longest string in the set of paths that is a prefix of the input text.
// It returns the matching prefix and true if found, otherwise empty string and false.
func (ps *PrefixConsumer) LongestPrefix(text string) (string, bool) {
	match, found, _ := ps.longestPrefix(text, true)
	return match, found
}

// longestPrefix also reports whether more bytes could complete a longer prefix.
func (ps *PrefixConsumer) longestPrefix(text string, atEOF bool) (match string, found, needMore bool) {
	curr := ps.root
	if curr.isEnd {
		match, found = curr.fullPath, true
	}
	for idx := 0; ; {
		if idx == len(text) {
			return match, found, !atEOF && len(curr.children) > 0
		}
		var next *trieNode
		for _, child := range curr.children {
			if child.segment[0] != text[idx] {
				continue
			}
			rest := text[idx:]
			if len(rest) < len(child.segment) && strings.HasPrefix(child.segment, rest) {
				return match, found, !atEOF
			}
			if !strings.HasPrefix(rest, child.segment) {
				return match, found, false
			}
			next = child
			idx += len(child.segment)
			break
		}
		if next == nil {
			return match, found, false
		}
		curr = next
		if curr.isEnd {
			match, found = curr.fullPath, true
		}
	}
}

type prefixOptions struct {
	inclusive   bool
	startOffset int
	ignore0     bool
	followedBy  consume.MustBeFollowedBy
	atEnd       bool
}

func parsePrefixOptions(ops []any) (o prefixOptions) {
	for _, op := range ops {
		switch v := op.(type) {
		case consume.Inclusive:
			o.inclusive = bool(v)
		case consume.StartOffset:
			o.startOffset = int(v)
		case consume.Ignore0PositionMatch:
			o.ignore0 = bool(v)
		case consume.MustBeFollowedBy:
			o.followedBy = v
		case consume.MustBeAtEnd:
			o.atEnd = bool(v)
		}
	}
	if o.startOffset < 0 {
		o.startOffset = 0
	}
	return o
}

// EOF is an accepted follower boundary, as in Consume's original contract.
func (o prefixOptions) accepts(after string, atStart bool) bool {
	if o.ignore0 && atStart || o.atEnd && len(after) != 0 {
		return false
	}
	if o.followedBy != nil && len(after) != 0 {
		r, _ := utf8.DecodeRuneInString(after)
		return o.followedBy(r)
	}
	return true
}

// Consume scans the input string 'from' to find the longest matching prefix from the configured set.
// It searches starting from 'StartOffset' (default 0) and returns the first match found.
// It returns four values:
// 1. before: The string before the match (or the match itself if Inclusive is true).
// 2. match: The matched prefix.
// 3. remaining: The rest of the string after the match (or starting from the match if Inclusive is false).
// 4. found: True if a match was found, false otherwise.
// Options:
// - consume.Inclusive(true): If true, 'before' includes the matched prefix, and 'remaining' starts after it.
// - consume.StartOffset(n): Starts the search at byte offset n; negative offsets mean zero.
// - consume.Ignore0PositionMatch(true): Ignores matches at byte zero of the original input.
// - consume.MustBeFollowedBy(func(rune) bool): A following rune must satisfy the predicate; EOF is accepted.
// - consume.MustBeAtEnd(true): The match must be at the end of the string.
func (ps *PrefixConsumer) Consume(from string, ops ...any) (string, string, string, bool) {
	o := parsePrefixOptions(ops)
	for i := o.startOffset; i < len(from); i++ {
		match, found := ps.LongestPrefix(from[i:])
		if !found || !o.accepts(from[i+len(match):], i == 0) {
			continue
		}
		if o.inclusive {
			return from[:i+len(match)], match, from[i+len(match):], true
		}
		return from[:i], match, from[i:], true
	}
	return "", "", from, false
}

// Iterator yields (matched, remaining), with remaining always after the prefix.
// It matches sequentially at the current position; after an option rejects a
// candidate it searches byte by byte for a valid one, then resumes sequentially.
// StartOffset applies only initially. Ignore0PositionMatch refers to byte zero
// of the original input, independently of StartOffset. Negative offsets mean zero.
// Inclusive includes input skipped since the last yield in matched; otherwise
// matched is just the prefix. An accepted empty prefix is yielded once, then stops.
func (ps *PrefixConsumer) Iterator(from string, ops ...any) func(yield func(string, string) bool) {
	return func(yield func(string, string) bool) {
		o := parsePrefixOptions(ops)
		search := false
		lastEnd := 0
		for i := o.startOffset; i <= len(from); {
			match, found := ps.LongestPrefix(from[i:])
			if !found {
				if !search {
					return
				}
				i++
				continue
			}
			end := i + len(match)
			if !o.accepts(from[end:], i == 0) {
				search = true
				i++
				continue
			}
			matched := match
			if o.inclusive {
				matched = from[lastEnd:end]
			}
			if !yield(matched, from[end:]) || len(match) == 0 {
				return
			}
			i, lastEnd, search = end, end, false
		}
	}
}

// SplitFunc emits only matching prefixes: Inclusive has no effect on scanner
// tokens, which already include the entire prefix. Skipped input is discarded.
// Position options and rejection searches follow Iterator's contract. A definite
// mismatch during sequential matching ends scanning without emitting unmatched
// data. A prefix is held until a longer match and all constraints can be decided.
// An accepted empty prefix emits one final empty token and terminates scanning.
// Each returned split function holds state for one scanner.
func (ps *PrefixConsumer) SplitFunc(ops ...any) bufio.SplitFunc {
	o := parsePrefixOptions(ops)
	skip := o.startOffset
	atStart := true
	search := false
	done := false
	return func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if done {
			return 0, nil, bufio.ErrFinalToken
		}
		offset := 0
		if skip > 0 {
			n := min(skip, len(data))
			skip -= n
			if n > 0 {
				atStart = false
			}
			if skip > 0 {
				if atEOF {
					done = true
					return n, nil, bufio.ErrFinalToken
				}
				return n, nil, nil
			}
			// Evaluate the rest in this call: Scanner stops on a nil token at EOF.
			offset = n
		}
		text := string(data)
		for i := offset; i <= len(data); i++ {
			match, found, needMore := ps.longestPrefix(text[i:], atEOF)
			if needMore {
				// Discard only bytes already ruled out by a rejection search.
				if i > 0 {
					atStart = false
				}
				return i, nil, nil
			}
			if !found {
				if !search {
					done = true
					return i, nil, bufio.ErrFinalToken
				}
				continue
			}
			end := i + len(match)
			reject := o.ignore0 && atStart && i == 0 || o.atEnd && end != len(data)
			if !reject && o.atEnd && !atEOF {
				atStart = atStart && i == 0
				return i, nil, nil
			}
			if !reject && o.followedBy != nil {
				if !atEOF && (end == len(data) || !utf8.FullRune(data[end:])) {
					atStart = atStart && i == 0
					return i, nil, nil
				}
				if end < len(data) {
					r, _ := utf8.DecodeRune(data[end:])
					reject = !o.followedBy(r)
				}
			}
			if reject {
				search = true
				continue
			}
			atStart, search = false, false
			if len(match) == 0 {
				done = true
				return end, []byte{}, bufio.ErrFinalToken
			}
			return end, data[i:end], nil
		}
		atStart = atStart && len(data) == 0
		if atEOF {
			done = true
			return len(data), nil, bufio.ErrFinalToken
		}
		return len(data), nil, nil
	}
}
