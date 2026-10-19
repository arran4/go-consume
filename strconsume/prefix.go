package strconsume

import (
	"bufio"
	"slices"
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
	curr := ps.root
	match := ""
	hasMatch := false

	// Check match at root (empty string)
	if curr.isEnd {
		match = curr.fullPath
		hasMatch = true
	}

	idx := 0
	for idx < len(text) {
		var next *trieNode
		// Linear scan of children. Since we used CommonPaths logic (sorted), children are sorted by first char.
		for _, child := range curr.children {
			if len(child.segment) == 0 {
				continue // Should not happen with valid LCP logic
			}
			if child.segment[0] == text[idx] {
				segLen := len(child.segment)
				// Check if full segment matches
				if idx+segLen <= len(text) && text[idx:idx+segLen] == child.segment {
					next = child
					idx += segLen
					break
				} else {
					// Mismatch within segment or text too short
					return match, hasMatch
				}
			}
		}

		if next == nil {
			return match, hasMatch
		}
		curr = next
		if curr.isEnd {
			match = curr.fullPath
			hasMatch = true
		}
	}

	return match, hasMatch
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
// - consume.StartOffset(n): Starts the search at offset n.
// - consume.Ignore0PositionMatch(true): Ignores matches at the very start of the search (offset).
// - consume.MustBeFollowedBy(func(rune) bool): The match must be followed by a rune satisfying the predicate.
// - consume.MustBeAtEnd(true): The match must be at the end of the string.
func (ps *PrefixConsumer) Consume(from string, ops ...any) (string, string, string, bool) {
	inclusive := false
	startOffset := 0
	ignore0PositionMatch := false
	var mustBeFollowedBy consume.MustBeFollowedBy
	mustBeAtEnd := false
	for _, op := range ops {
		switch v := op.(type) {
		case consume.Inclusive:
			inclusive = bool(v)
		case consume.StartOffset:
			startOffset = int(v)
		case consume.Ignore0PositionMatch:
			ignore0PositionMatch = bool(v)
		case consume.MustBeFollowedBy:
			mustBeFollowedBy = v
		case consume.MustBeAtEnd:
			mustBeAtEnd = bool(v)
		}
	}
	for i := startOffset; i < len(from); i++ {
		match, found := ps.LongestPrefix(from[i:])
		if found {
			nextIdx := i + len(match)
			if mustBeAtEnd {
				if nextIdx != len(from) {
					continue
				}
			}

			if mustBeFollowedBy != nil {
				if nextIdx < len(from) {
					r, _ := utf8.DecodeRuneInString(from[nextIdx:])
					if !mustBeFollowedBy(r) {
						continue
					}
				}
			}

			if i == 0 && ignore0PositionMatch {
				continue
			}
			matched := from[:i]
			if inclusive {
				return matched + match, match, from[i+len(match):], true
			}
			return matched, match, from[i:], true
		}
	}
	return "", "", from, false
}

func (ps *PrefixConsumer) Iterator(from string, ops ...any) func(yield func(string, string) bool) {
	return func(yield func(string, string) bool) {
		startOffset := 0
		ignore0PositionMatch := false
		var mustBeFollowedBy consume.MustBeFollowedBy
		mustBeAtEnd := false

		for _, op := range ops {
			switch v := op.(type) {
			case consume.StartOffset:
				startOffset = int(v)
			case consume.Ignore0PositionMatch:
				ignore0PositionMatch = bool(v)
			case consume.MustBeFollowedBy:
				mustBeFollowedBy = v
			case consume.MustBeAtEnd:
				mustBeAtEnd = bool(v)
			}
		}

		if startOffset > 0 {
			if startOffset >= len(from) {
				return
			}
			from = from[startOffset:]
		}

		for {
			if len(from) == 0 {
				m, ok := ps.LongestPrefix("")
				if ok && m == "" {
					yield("", "")
				}
				return
			}

			match, found := ps.LongestPrefix(from)
			if !found {
				return
			}

			nextIdx := len(match)

			// Option constraints evaluation
			rejectMatch := false

			if mustBeAtEnd && nextIdx != len(from) {
				rejectMatch = true
			}

			if !rejectMatch && mustBeFollowedBy != nil && nextIdx < len(from) {
				r, _ := utf8.DecodeRuneInString(from[nextIdx:])
				if !mustBeFollowedBy(r) {
					rejectMatch = true
				}
			}

			if !rejectMatch && ignore0PositionMatch {
				rejectMatch = true
			}

			// Clear single-use options
			ignore0PositionMatch = false

			if rejectMatch {
				// If a match is rejected by options, advance past it to search for the next prefix match
				if len(match) > 0 {
					from = from[len(match):]
					continue
				} else {
					return
				}
			}

			remaining := from[nextIdx:]

			if !yield(match, remaining) {
				return
			}

			if len(match) == 0 {
				return
			}

			from = remaining
		}
	}
}

func (ps *PrefixConsumer) SplitFunc(ops ...any) bufio.SplitFunc {
	startOffset := 0
	ignore0PositionMatch := false
	var mustBeFollowedBy consume.MustBeFollowedBy
	mustBeAtEnd := false

	for _, op := range ops {
		switch v := op.(type) {
		case consume.StartOffset:
			startOffset = int(v)
		case consume.Ignore0PositionMatch:
			ignore0PositionMatch = bool(v)
		case consume.MustBeFollowedBy:
			mustBeFollowedBy = v
		case consume.MustBeAtEnd:
			mustBeAtEnd = bool(v)
		}
	}

	firstRun := true

	return func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}

		offset := 0
		if firstRun {
			offset = startOffset
			if offset > 0 {
				// We need to advance past startOffset but return no token to skip bytes.
				// However, if we don't have enough data, we wait.
				if offset >= len(data) {
					if atEOF {
						advance = len(data)
						firstRun = false
						return advance, nil, nil
					}
					return 0, nil, nil
				}
				advance = offset
				firstRun = false
				return advance, nil, nil
			}
		}

		text := string(data)
		m, ok := ps.LongestPrefix(text)

		if !ok {
			// If it doesn't match and atEOF is true, we must return the rest to avoid infinite loop.
			if atEOF {
				return len(data), nil, nil // Drop trailing unmatched bytes to avoid stall
			}
			return 0, nil, nil
		}

		nextIdx := len(m)
		rejectMatch := false

		if mustBeAtEnd {
			if nextIdx != len(text) {
				rejectMatch = true
			} else if !atEOF {
				return 0, nil, nil // Needs more data to ensure it's at EOF
			}
		}

		if !rejectMatch && mustBeFollowedBy != nil {
			if nextIdx < len(text) {
				r, _ := utf8.DecodeRuneInString(text[nextIdx:])
				if !mustBeFollowedBy(r) {
					rejectMatch = true
				}
			} else if !atEOF {
				return 0, nil, nil // Need more data to verify FollowedBy
			}
		}

		if !rejectMatch && firstRun && ignore0PositionMatch {
			rejectMatch = true
		}

		if rejectMatch {
			// Wait, if it fails, shouldn't we skip this match or return an error?
			// The original implementation just ignores the candidate.
			ignore0PositionMatch = false
			firstRun = false
			// Skip this match by advancing past it and returning nil token
			if len(m) > 0 {
				return len(m), nil, nil
			}
			if atEOF {
				return len(data), nil, nil
			}
			return 0, nil, nil
		}

		ignore0PositionMatch = false
		firstRun = false

		advance = len(m)
		token = []byte(m)
		return advance, token, nil
	}
}
