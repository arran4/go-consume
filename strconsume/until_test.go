package strconsume

import (
	"testing"

	"github.com/arran4/go-consume"
	"github.com/stretchr/testify/assert"
)

func TestUntilConsumer_Consume(t *testing.T) {
	tests := []struct {
		name              string
		seps              []string
		input             string
		ops               []any
		expectedMatched   string
		expectedSeparator string
		expectedRemaining string
		expectedOk        bool
	}{
		{
			name:              "Slash match at start",
			seps:              []string{"/"},
			input:             "/abc",
			ops:               []any{consume.Inclusive(true)},
			expectedMatched:   "/",
			expectedSeparator: "/",
			expectedRemaining: "abc",
			expectedOk:        true,
		},
		{
			name:              "Slash match in middle",
			seps:              []string{"/"},
			input:             "foo/bar",
			ops:               []any{consume.Inclusive(true)},
			expectedMatched:   "foo/",
			expectedSeparator: "/",
			expectedRemaining: "bar",
			expectedOk:        true,
		},
		{
			name:              "No match",
			seps:              []string{"/"},
			input:             "foobar",
			ops:               []any{consume.Inclusive(true)},
			expectedMatched:   "",
			expectedSeparator: "",
			expectedRemaining: "foobar",
			expectedOk:        false,
		},
		{
			name:              "Longer separator",
			seps:              []string{"//"},
			input:             "foo//bar",
			ops:               []any{consume.Inclusive(true)},
			expectedMatched:   "foo//",
			expectedSeparator: "//",
			expectedRemaining: "bar",
			expectedOk:        true,
		},
		{
			name:              "Multiple separators",
			seps:              []string{"/", "-"},
			input:             "foo-bar",
			ops:               []any{consume.Inclusive(true)},
			expectedMatched:   "foo-",
			expectedSeparator: "-",
			expectedRemaining: "bar",
			expectedOk:        true,
		},
		{
			name:              "Empty string",
			seps:              []string{"/"},
			input:             "",
			ops:               []any{consume.Inclusive(true)},
			expectedMatched:   "",
			expectedSeparator: "",
			expectedRemaining: "",
			expectedOk:        false,
		},
		{
			name:              "Separator at end",
			seps:              []string{"/"},
			input:             "foo/",
			ops:               []any{consume.Inclusive(true)},
			expectedMatched:   "foo/",
			expectedSeparator: "/",
			expectedRemaining: "",
			expectedOk:        true,
		},
		{
			name:              "Just separator",
			seps:              []string{"/"},
			input:             "/",
			ops:               []any{consume.Inclusive(true)},
			expectedMatched:   "/",
			expectedSeparator: "/",
			expectedRemaining: "",
			expectedOk:        true,
		},
		{
			name:              "No separators",
			seps:              []string{},
			input:             "/abc",
			ops:               []any{consume.Inclusive(true)},
			expectedMatched:   "",
			expectedSeparator: "",
			expectedRemaining: "/abc",
			expectedOk:        false,
		},
		{
			name:              "Longest match preference",
			seps:              []string{"/", "//"},
			input:             "foo//bar",
			ops:               []any{consume.Inclusive(true)},
			expectedMatched:   "foo//",
			expectedSeparator: "//",
			expectedRemaining: "bar",
			expectedOk:        true,
		},
		{
			name:              "Not inclusive",
			seps:              []string{"/"},
			input:             "foo/bar",
			ops:               []any{consume.Inclusive(false)},
			expectedMatched:   "foo",
			expectedSeparator: "/",
			expectedRemaining: "/bar",
			expectedOk:        true,
		},
		{
			name:              "Start offset",
			seps:              []string{"/"},
			input:             "/foo/bar",
			ops:               []any{consume.Inclusive(true), consume.StartOffset(1)},
			expectedMatched:   "/foo/",
			expectedSeparator: "/",
			expectedRemaining: "bar",
			expectedOk:        true,
		},
		{
			name:              "Ignore 0 position match",
			seps:              []string{"/"},
			input:             "/foo/bar",
			ops:               []any{consume.Inclusive(true), consume.Ignore0PositionMatch(true)},
			expectedMatched:   "/foo/",
			expectedSeparator: "/",
			expectedRemaining: "bar",
			expectedOk:        true,
		},
		{
			name:              "Case insensitive match",
			seps:              []string{"/Foo/"},
			input:             "/foo/bar",
			ops:               []any{consume.Inclusive(true), consume.CaseInsensitive(true)},
			expectedMatched:   "/foo/",
			expectedSeparator: "/foo/",
			expectedRemaining: "bar",
			expectedOk:        true,
		},
		{
			name:              "ConsumeRemainingIfNotFound - Normal behavior (not found)",
			seps:              []string{";"},
			input:             "foobar",
			ops:               []any{},
			expectedMatched:   "",
			expectedSeparator: "",
			expectedRemaining: "foobar",
			expectedOk:        false,
		},
		{
			name:              "ConsumeRemainingIfNotFound - Match remaining",
			seps:              []string{";"},
			input:             "foobar",
			ops:               []any{consume.ConsumeRemainingIfNotFound(true)},
			expectedMatched:   "foobar",
			expectedSeparator: "",
			expectedRemaining: "",
			expectedOk:        true,
		},
		{
			name:              "ConsumeRemainingIfNotFound - Normal behavior (found)",
			seps:              []string{";"},
			input:             "foo;bar",
			ops:               []any{consume.ConsumeRemainingIfNotFound(true)},
			expectedMatched:   "foo",
			expectedSeparator: ";",
			expectedRemaining: ";bar",
			expectedOk:        true,
		},
		{
			name:              "ConsumeRemainingIfNotFound - Inclusive",
			seps:              []string{";"},
			input:             "foobar",
			ops:               []any{consume.ConsumeRemainingIfNotFound(true), consume.Inclusive(true)},
			expectedMatched:   "foobar",
			expectedSeparator: "",
			expectedRemaining: "",
			expectedOk:        true,
		},
		{
			name:              "ConsumeRemainingIfNotFound - Empty input",
			seps:              []string{";"},
			input:             "",
			ops:               []any{consume.ConsumeRemainingIfNotFound(true)},
			expectedMatched:   "",
			expectedSeparator: "",
			expectedRemaining: "",
			expectedOk:        true,
		},
		{
			name:              "ConsumeRemainingIfNotFound - Start offset",
			seps:              []string{";"},
			input:             "abc",
			ops:               []any{consume.StartOffset(1), consume.ConsumeRemainingIfNotFound(true)},
			expectedMatched:   "abc",
			expectedSeparator: "",
			expectedRemaining: "",
			expectedOk:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cu := NewUntilConsumer(tt.seps...)
			matched, sep, remaining, ok := cu.Consume(tt.input, tt.ops...)
			assert.Equal(t, tt.expectedOk, ok)
			if ok {
				assert.Equal(t, tt.expectedMatched, matched)
				assert.Equal(t, tt.expectedSeparator, sep)
				assert.Equal(t, tt.expectedRemaining, remaining)
			} else {
				assert.Equal(t, tt.expectedRemaining, remaining)
				assert.Equal(t, "", matched)
				assert.Equal(t, "", sep)
			}
		})
	}
}

func TestUntilConsumer_Iterator_ConsumeRemainingIfNotFound(t *testing.T) {
	uc := NewUntilConsumer(";")

	// Iterator with ConsumeRemainingIfNotFound
	// Should yield parts split by separator, and if the last part has no separator, it should yield it too.
	// Normal iterator yields (matched, separator).
	// If no separator found, it yields (from, "").

	iter := uc.Iterator("foo;bar", consume.ConsumeRemainingIfNotFound(true))
	var results []string
	iter(func(matched, separator string) bool {
		results = append(results, matched)
		return true
	})
	// "foo;bar" -> "foo" (sep ";"), remaining "bar"
	// "bar" -> not found -> "bar" (sep ""), remaining ""
	// Then loop continues with empty string. Not found (or found as empty).
	// Iterator yields remainder, which is empty.
	assert.Equal(t, []string{"foo", "bar", ""}, results)

	// With no separator at all
	iter = uc.Iterator("foobar", consume.ConsumeRemainingIfNotFound(true))
	results = nil
	iter(func(matched, separator string) bool {
		results = append(results, matched)
		return true
	})
	assert.Equal(t, []string{"foobar", ""}, results)
}
