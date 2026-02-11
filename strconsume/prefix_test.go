package strconsume

import (
	"github.com/arran4/go-consume"
	"testing"
)

func TestPrefixConsumer(t *testing.T) {
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
			expected: "", // "" is prefix of everything
			found:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ps := NewPrefixConsumer(tt.paths...)
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

func TestPrefixConsumer_Consume(t *testing.T) {
	pc := NewPrefixConsumer("/sep", "/foo")

	// Test basic consume
	matched, separator, remaining, found := pc.Consume("prefix/sep/suffix")
	if !found {
		t.Errorf("Consume failed to find separator")
	}
	if matched != "prefix" {
		t.Errorf("Consume matched = %q, expected %q", matched, "prefix")
	}
	if separator != "/sep" {
		t.Errorf("Consume separator = %q, expected %q", separator, "/sep")
	}
	if remaining != "/sep/suffix" {
		t.Errorf("Consume remaining = %q, expected %q", remaining, "/sep/suffix")
	}

	// Test Inclusive
	matched, _, remaining, found = pc.Consume("prefix/sep/suffix", consume.Inclusive(true))
	if !found {
		t.Errorf("Consume (inclusive) failed")
	}
	if matched != "prefix/sep" {
		t.Errorf("Consume (inclusive) matched = %q, expected %q", matched, "prefix/sep")
	}
	if remaining != "/suffix" {
		t.Errorf("Consume (inclusive) remaining = %q, expected %q", remaining, "/suffix")
	}

	// Test StartOffset
	_, separator, _, found = pc.Consume("prefix/sep/suffix", consume.StartOffset(7))
	// Offset 7 is after /sep start.
	if found {
		t.Errorf("Consume (offset 7) found unexpected match: %s", separator)
	}

	// Test StartOffset matching
	_, separator, _, found = pc.Consume("prefix/sep/suffix", consume.StartOffset(6))
	if !found {
		t.Errorf("Consume (offset 6) failed")
	}
	if separator != "/sep" {
		t.Errorf("Consume (offset 6) separator = %q, expected %q", separator, "/sep")
	}

	// Test Ignore0PositionMatch
	_, separator, _, found = pc.Consume("/sep/suffix", consume.Ignore0PositionMatch(true))
	if found {
		// Should skip 0 position. No match later (unlike /s matching /suffix).
		t.Errorf("Consume (Ignore0) found unexpected match: %s", separator)
	}

	// Test Ignore0PositionMatch with later match
	matched, _, _, found = pc.Consume("/sep/sep", consume.Ignore0PositionMatch(true))
	if !found {
		t.Errorf("Consume (Ignore0 with later) failed")
	}
	if matched != "/sep" {
		t.Errorf("Consume (Ignore0 with later) matched = %q, expected %q", matched, "/sep")
	}
}

func TestPrefixConsumer_Consume_MustBeFollowedBy(t *testing.T) {
	pc := NewPrefixConsumer("/sep")
	delimiter := func(r rune) bool { return r == '/' }

	// Input: prefix/sep/suffix
	// Match /sep. Next char /. Matches delimiter.
	_, separator, _, found := pc.Consume("prefix/sep/suffix", consume.MustBeFollowedBy(delimiter))
	if !found {
		t.Errorf("Consume failed")
	}
	if separator != "/sep" {
		t.Errorf("Got %q, expected %q", separator, "/sep")
	}

	// Input: prefix/sepSuffix
	// Match /sep. Next char S. Not delimiter. Should FAIL to match at this position?
	// But Consume scans.
	// Is there another match? No.
	_, separator, _, found = pc.Consume("prefix/sepSuffix", consume.MustBeFollowedBy(delimiter))
	if found {
		t.Errorf("Consume found match when not followed by delimiter: %s", separator)
	}

	// Input: prefix/sep
	// Match /sep. Next char EOF. Should pass.
	_, separator, _, found = pc.Consume("prefix/sep", consume.MustBeFollowedBy(delimiter))
	if !found {
		t.Errorf("Consume failed at EOF")
	}
	if separator != "/sep" {
		t.Errorf("Got %q, expected %q", separator, "/sep")
	}
}

func TestPrefixConsumer_Consume_MustBeAtEnd(t *testing.T) {
	pc := NewPrefixConsumer("/sep")

	// Match at end
	_, _, _, found := pc.Consume("prefix/sep", consume.MustBeAtEnd(true))
	if !found {
		t.Errorf("Consume (MustBeAtEnd) failed at end")
	}

	// Match not at end
	_, _, _, found = pc.Consume("prefix/sep/suffix", consume.MustBeAtEnd(true))
	if found {
		t.Errorf("Consume (MustBeAtEnd) found match not at end")
	}
}
