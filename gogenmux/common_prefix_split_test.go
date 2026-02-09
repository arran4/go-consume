package gogenmux

import (
	"sort"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func sortMatchPairs(mps []*MatchPair) {
	sort.Slice(mps, func(i, j int) bool {
		if mps[i].Match != mps[j].Match {
			return mps[i].Match < mps[j].Match
		}
		// If matches are equal, compare paths
		// Join path for comparison
		p1 := strings.Join(mps[i].Path, "")
		p2 := strings.Join(mps[j].Path, "")
		return p1 < p2
	})
}

func TestCommonPrefixSplit(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected []*MatchPair
	}{
		// Simple examples resembling regular language prefixes
		{
			name:  "Simple prefix match",
			input: []string{"AAB", "ABB"},
			expected: []*MatchPair{
				{Match: "AAB", Path: []string{"A", "AB"}},
				{Match: "ABB", Path: []string{"A", "BB"}},
			},
		},
		{
			name:  "Different different lengths",
			input: []string{"AA", "BBBB"},
			expected: []*MatchPair{
				{Match: "AA", Path: []string{"AA"}},
				{Match: "BBBB", Path: []string{"BBBB"}},
			},
		},
		{
			name:  "Same different lengths",
			input: []string{"AA", "AAAA"},
			expected: []*MatchPair{
				{Match: "AA", Path: []string{"AA"}},
				{Match: "AAAA", Path: []string{"AA", "AA"}},
			},
		},
		{
			name:  "Same different lengths - reversed - maintains order",
			input: []string{"AAAA", "AA"},
			expected: []*MatchPair{
				{Match: "AA", Path: []string{"AA"}},
				{Match: "AAAA", Path: []string{"AA", "AA"}},
			},
		},
		{
			name:  "Simple prefix match fully different",
			input: []string{"AAB", "ABB", "ABC"},
			expected: []*MatchPair{
				{Match: "AAB", Path: []string{"A", "AB"}},
				{Match: "ABB", Path: []string{"A", "B", "B"}},
				{Match: "ABC", Path: []string{"A", "B", "C"}},
			},
		},
		{
			name:  "Identical strings",
			input: []string{"AAA", "AAA"},
			expected: []*MatchPair{
				{Match: "AAA", Path: []string{"AAA"}},
				{Match: "AAA", Path: []string{"AAA"}},
			},
		},

		// REST-like examples
		{
			name:  "Basic REST-like paths",
			input: []string{"/user/123", "/user/456", "/admin/789", "/test/abc"},
			expected: []*MatchPair{
				{Match: "/user/123", Path: []string{"/", "user/", "123"}},
				{Match: "/user/456", Path: []string{"/", "user/", "456"}},
				{Match: "/admin/789", Path: []string{"/", "admin/789"}},
				{Match: "/test/abc", Path: []string{"/", "test/abc"}},
			},
		},
		{
			name:  "REST paths with nested segments",
			input: []string{"/product/1", "/product/1/details", "/product/2", "/order/1001"},
			expected: []*MatchPair{
				{Match: "/product/1", Path: []string{"/", "product/", "1"}},
				{Match: "/product/1/details", Path: []string{"/", "product/", "1", "/details"}},
				{Match: "/product/2", Path: []string{"/", "product/", "2"}},
				{Match: "/order/1001", Path: []string{"/", "order/1001"}},
			},
		},
		{
			name:  "REST paths with overlapping prefixes",
			input: []string{"/api/v1/user", "/api/v1/admin", "/api/v2/user", "/static/css"},
			expected: []*MatchPair{
				{Match: "/api/v1/user", Path: []string{"/", "api/v", "1/", "user"}},
				{Match: "/api/v1/admin", Path: []string{"/", "api/v", "1/", "admin"}},
				{Match: "/api/v2/user", Path: []string{"/", "api/v", "2/user"}},
				{Match: "/static/css", Path: []string{"/", "static/css"}},
			},
		},

		// Edge cases
		{
			name:     "Empty input",
			input:    []string{},
			expected: []*MatchPair{},
		},
		{
			name:  "Single input",
			input: []string{"/unique/path"},
			expected: []*MatchPair{
				{Match: "/unique/path", Path: []string{"/unique/path"}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CommonPrefixSplit(tt.input)

			// Sort both expected and result to ignore order
			sortMatchPairs(result)

			// Copy expected before sorting
			expectedCopy := make([]*MatchPair, len(tt.expected))
			copy(expectedCopy, tt.expected)
			sortMatchPairs(expectedCopy)

			if diff := cmp.Diff(expectedCopy, result); diff != "" {
				t.Errorf("CommonPrefixSplit() mismatch (-expected +got):\n%s", diff)
			}
		})
	}
}
