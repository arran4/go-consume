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

func NewPrefixConsumer(paths []string) *PrefixConsumer {
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
				// If at end of string, we assume match is valid (boundary reached) unless controlled by another option?
				// mustBeFollowedBy checks "if followed by X". EOF is valid boundary.
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
		for {
			matched, found := ps.LongestPrefix(from)
			if !found {
				return
			}
			remaining := from[len(matched):]
			if !yield(matched, remaining) {
				return
			}
			if len(matched) == 0 {
				return
			}
			from = remaining
		}
	}
}

func (ps *PrefixConsumer) SplitFunc(ops ...any) bufio.SplitFunc {
	return func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		if atEOF && len(data) == 0 {
			return 0, nil, nil
		}
		// This is inefficient but functional for now: convert to string
		text := string(data)
		matched, found := ps.LongestPrefix(text)
		if found {
			return len(matched), []byte(matched), nil
		}
		return 0, nil, nil
	}
}
