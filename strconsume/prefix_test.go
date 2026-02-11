package strconsume

import (
	"testing"

	"github.com/arran4/go-consume"
	"github.com/stretchr/testify/assert"
)

func TestPrefixConsumer_Consume(t *testing.T) {
	tests := []struct {
		name              string
		prefixes          []string
		input             string
		ops               []any
		expectedMatched   string
		expectedRemaining string
		expectedFound     bool
	}{
		{
			name:              "Single prefix match",
			prefixes:          []string{"foo"},
			input:             "foobar",
			expectedMatched:   "foo",
			expectedRemaining: "bar",
			expectedFound:     true,
		},
		{
			name:              "No match",
			prefixes:          []string{"foo"},
			input:             "barfoo",
			expectedMatched:   "",
			expectedRemaining: "barfoo",
			expectedFound:     false,
		},
		{
			name:              "Longest match preference",
			prefixes:          []string{"foo", "foobar"},
			input:             "foobarbaz",
			expectedMatched:   "foobar",
			expectedRemaining: "baz",
			expectedFound:     true,
		},
		{
			name:              "Empty string",
			prefixes:          []string{"foo"},
			input:             "",
			expectedMatched:   "",
			expectedRemaining: "",
			expectedFound:     false,
		},
		{
			name:              "Prefix longer than input",
			prefixes:          []string{"foo"},
			input:             "fo",
			expectedMatched:   "",
			expectedRemaining: "fo",
			expectedFound:     false,
		},
		{
			name:              "Case insensitive match",
			prefixes:          []string{"Foo"},
			input:             "foobar",
			ops:               []any{consume.CaseInsensitive(true)},
			expectedMatched:   "foo", // returns extracted part
			expectedRemaining: "bar",
			expectedFound:     true,
		},
		{
			name:              "Case sensitive mismatch",
			prefixes:          []string{"Foo"},
			input:             "foobar",
			ops:               []any{consume.CaseInsensitive(false)},
			expectedMatched:   "",
			expectedRemaining: "foobar",
			expectedFound:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pc := NewPrefixConsumer(tt.prefixes...)
			matched, remaining, found := pc.Consume(tt.input, tt.ops...)
			assert.Equal(t, tt.expectedFound, found)
			if found {
				assert.Equal(t, tt.expectedMatched, matched)
				assert.Equal(t, tt.expectedRemaining, remaining)
			} else {
				assert.Equal(t, tt.expectedRemaining, remaining)
				assert.Equal(t, "", matched)
			}
		})
	}
}

func TestPrefixConsumer_MustMatchWholeString(t *testing.T) {
	pc := NewPrefixConsumer("foo")

	// Normal behavior
	matched, remaining, found := pc.Consume("foobar")
	assert.True(t, found)
	assert.Equal(t, "foo", matched)
	assert.Equal(t, "bar", remaining)

	// With MustMatchWholeString
	matched, remaining, found = pc.Consume("foobar", consume.MustMatchWholeString(true))
	assert.False(t, found)
	assert.Equal(t, "", matched)
	assert.Equal(t, "foobar", remaining)

	matched, remaining, found = pc.Consume("foo", consume.MustMatchWholeString(true))
	assert.True(t, found)
	assert.Equal(t, "foo", matched)
	assert.Equal(t, "", remaining)
}

func TestPrefixConsumer_MustMatchWholeString_CaseInsensitive(t *testing.T) {
	pc := NewPrefixConsumer("foo")

	// With MustMatchWholeString and CaseInsensitive
	matched, remaining, found := pc.Consume("Foo", consume.MustMatchWholeString(true), consume.CaseInsensitive(true))
	assert.True(t, found)
	assert.Equal(t, "Foo", matched) // Preserves case of input
	assert.Equal(t, "", remaining)

	matched, remaining, found = pc.Consume("Foobar", consume.MustMatchWholeString(true), consume.CaseInsensitive(true))
	assert.False(t, found)
	assert.Equal(t, "", matched)
	assert.Equal(t, "Foobar", remaining)
}

func TestPrefixConsumer_Iterator_MustMatchWholeString(t *testing.T) {
	pc := NewPrefixConsumer("foo")

	// Iterator with MustMatchWholeString
	// Should yield only if the whole string matches the prefix.
	// Since iterator repeatedly consumes, if we match whole string, remaining is empty, so loop terminates.

	iter := pc.Iterator("foo", consume.MustMatchWholeString(true))
	count := 0
	iter(func(matched, remaining string) bool {
		count++
		assert.Equal(t, "foo", matched)
		assert.Equal(t, "", remaining)
		return true
	})
	assert.Equal(t, 1, count)

	// If not matching whole string
	iter = pc.Iterator("foobar", consume.MustMatchWholeString(true))
	count = 0
	iter(func(matched, remaining string) bool {
		count++
		return true
	})
	assert.Equal(t, 0, count)
}
