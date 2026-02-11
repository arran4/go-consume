package strconsume

import (
	"testing"

	"github.com/arran4/go-consume"
	"github.com/stretchr/testify/assert"
)

func TestUntilConsumer_Iterator(t *testing.T) {
	tests := []struct {
		name     string
		seps     []string
		input    string
		ops      []any
		expected []struct {
			matched string
			sep     string
		}
	}{
		{
			name:  "Standard split",
			seps:  []string{"/"},
			input: "path/to/resource",
			expected: []struct {
				matched string
				sep     string
			}{
				{"path", "/"},
				{"to", "/"},
				{"resource", ""},
			},
		},
		{
			name:  "Ends with separator",
			seps:  []string{"/"},
			input: "a/",
			expected: []struct {
				matched string
				sep     string
			}{
				{"a", "/"},
				{"", ""},
			},
		},
		{
			name:  "Empty string",
			seps:  []string{"/"},
			input: "",
			expected: []struct {
				matched string
				sep     string
			}{
				{"", ""},
			},
		},
		{
			name:  "Only separators",
			seps:  []string{"/"},
			input: "//",
			expected: []struct {
				matched string
				sep     string
			}{
				{"", "/"},
				{"", "/"},
				{"", ""},
			},
		},
		{
			name:  "Inclusive",
			seps:  []string{"/"},
			input: "a/b",
			ops:   []any{consume.Inclusive(true)},
			expected: []struct {
				matched string
				sep     string
			}{
				{"a/", "/"},
				{"b", ""},
			},
		},
		{
			name:  "StartOffset",
			seps:  []string{"/"},
			input: "a/b/c",
			ops:   []any{consume.StartOffset(2)}, // Skip "a/"
			// "a/b/c" at offset 2 is "/b/c". matches "/" at 0?
			// Wait, consume at offset 2 of "a/b/c" -> "b/c".
			// matches "/" at 1. matched "b", sep "/", rem "c".
			// Yield ("a/b", "/")
			// Wait, Consume returns matched relative to input.
			// StartOffset=2. Input "a/b/c". Separator "/" found at index 3 (second /).
			// Matched "a/b".
			// Separator "/".
			// Remaining "c".
			// Next iteration: "c". No separator. Yield ("c", "").
			expected: []struct {
				matched string
				sep     string
			}{
				{"a/b", "/"},
				{"c", ""},
			},
		},
		{
			name:  "Ignore0PositionMatch",
			seps:  []string{"/"},
			input: "/a",
			ops:   []any{consume.Ignore0PositionMatch(true)},
			// Matches "/" at 0. Ignored.
			// Next match? No other "/" in "/a".
			// So no match found?
			// Consume returns not found.
			// Yield ("/a", "").
			expected: []struct {
				matched string
				sep     string
			}{
				{"/a", ""},
			},
		},
		{
			name:  "Ignore0PositionMatch with multiple",
			seps:  []string{"/"},
			input: "//a",
			ops:   []any{consume.Ignore0PositionMatch(true)},
			// Match "/" at 0 ignored.
			// Match "/" at 1 found.
			// Matched "/", Separator "/". Remaining "a".
			// Next iter: "a". No match. Yield "a", "".
			expected: []struct {
				matched string
				sep     string
			}{
				{"/", "/"},
				{"a", ""},
			},
		},
		{
			name:  "Case Insensitive",
			seps:  []string{"Foo"},
			input: "foobar",
			ops:   []any{consume.CaseInsensitive(true)},
			expected: []struct {
				matched string
				sep     string
			}{
				{"", "foo"}, // matched is "", separator is "foo" (from input)
				{"bar", ""},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cu := NewUntilConsumer(tt.seps...)
			iter := cu.Iterator(tt.input, tt.ops...)
			var actual []struct {
				matched string
				sep     string
			}
			iter(func(k, v string) bool {
				actual = append(actual, struct {
					matched string
					sep     string
				}{k, v})
				return true
			})
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestPrefixConsumer_Iterator(t *testing.T) {
	tests := []struct {
		name     string
		prefixes []string
		input    string
		ops      []any
		expected []struct {
			matched string
			rem     string
		}
	}{
		{
			name:     "Sequence",
			prefixes: []string{"foo", "bar"},
			input:    "foobar",
			expected: []struct {
				matched string
				rem     string
			}{
				{"foo", "bar"},
				{"bar", ""},
			},
		},
		{
			name:     "Partial match sequence",
			prefixes: []string{"foo", "bar"},
			input:    "foobarbaz",
			expected: []struct {
				matched string
				rem     string
			}{
				{"foo", "barbaz"},
				{"bar", "baz"},
			},
		},
		{
			name:     "No match at start",
			prefixes: []string{"foo"},
			input:    "baz",
			expected: nil,
		},
		{
			name:     "Empty input",
			prefixes: []string{"foo"},
			input:    "",
			expected: nil,
		},
		{
			name:     "Case Insensitive",
			prefixes: []string{"Foo"},
			input:    "foobar",
			ops:      []any{consume.CaseInsensitive(true)},
			expected: []struct {
				matched string
				rem     string
			}{
				{"foo", "bar"}, // matched is lower case from input
			},
		},
		{
			name:     "Empty prefix",
			prefixes: []string{""},
			input:    "abc",
			expected: []struct { // Should stop immediately to avoid infinite loop
				matched string
				rem     string
			}{
				// Actually, my implementation stops AFTER yielding empty match.
				// But wait, if matched is empty, yield ("","abc"), then stops.
				{"", "abc"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pc := NewPrefixConsumer(tt.prefixes...)
			iter := pc.Iterator(tt.input, tt.ops...)
			var actual []struct {
				matched string
				rem     string
			}
			iter(func(k, v string) bool {
				actual = append(actual, struct {
					matched string
					rem     string
				}{k, v})
				return true
			})
			assert.Equal(t, tt.expected, actual)
		})
	}
}
