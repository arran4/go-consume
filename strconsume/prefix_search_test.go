package strconsume

import (
	"testing"

)

func TestPrefixSearcher(t *testing.T) {
	tests := []struct {
		name     string
		paths    []string
		input    string
		expected string
		found    bool
	}{
		{
			name:     "Simple prefix match",
			paths:    []string{"AAB", "ABB"},
			input:    "AAB",
			expected: "AAB",
			found:    true,
		},
		{
			name:     "Longest prefix match",
			paths:    []string{"A", "AA", "AAA"},
			input:    "AAAA",
			expected: "AAA",
			found:    true,
		},
		{
			name:     "No match",
			paths:    []string{"B", "C"},
			input:    "A",
			expected: "",
			found:    false,
		},
		{
			name:     "Empty input",
			paths:    []string{"A"},
			input:    "",
			expected: "",
			found:    false,
		},
		{
			name:     "Empty paths",
			paths:    []string{},
			input:    "A",
			expected: "",
			found:    false,
		},
		{
			name:     "Match shorter prefix",
			paths:    []string{"A", "B"},
			input:    "AA",
			expected: "A",
			found:    true,
		},
		{
			name:     "Match longer than text",
			paths:    []string{"AAA"},
			input:    "AA",
			expected: "",
			found:    false,
		},
		{
			name:     "Multiple candidates with same prefix",
			paths:    []string{"foo", "foobar"},
			input:    "foobarbaz",
			expected: "foobar",
			found:    true,
		},
		{
			name:     "Dense paths",
			paths:    []string{"/a", "/ab", "/abc", "/abd"},
			input:    "/abd/foo",
			expected: "/abd",
			found:    true,
		},
		{
			name:     "Paths with no common prefix with text",
			paths:    []string{"foo", "bar"},
			input:    "baz",
			expected: "",
			found:    false,
		},
		{
			name:     "Empty path in list",
			paths:    []string{"", "a"},
			input:    "abc",
			expected: "a", // Should prefer longest non-empty
			found:    true,
		},
		{
			name:     "Empty path match",
			paths:    []string{""},
			input:    "abc",
			expected: "", // Depending on implementation, empty string is prefix of everything?
			// strings.HasPrefix("abc", "") is true.
			// Ideally we want non-empty match?
			// But "" is a valid prefix.
			found:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := NewPrefixSearcher(tt.paths)
			got, found := ps.LongestPrefix(tt.input)
			if found != tt.found {
				t.Errorf("LongestPrefix() found = %v, expected %v", found, tt.found)
			}
			if got != tt.expected {
				t.Errorf("LongestPrefix() got = %q, expected %q", got, tt.expected)
			}
		})
	}
}
