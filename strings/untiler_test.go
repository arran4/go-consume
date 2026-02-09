package bookmarks

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConsumeUntiler_Consume(t *testing.T) {
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
			ops:               []any{Inclusive(true)},
			expectedMatched:   "/",
			expectedSeparator: "/",
			expectedRemaining: "abc",
			expectedOk:        true,
		},
		{
			name:              "Slash match in middle",
			seps:              []string{"/"},
			input:             "foo/bar",
			ops:               []any{Inclusive(true)},
			expectedMatched:   "foo/",
			expectedSeparator: "/",
			expectedRemaining: "bar",
			expectedOk:        true,
		},
		{
			name:              "No match",
			seps:              []string{"/"},
			input:             "foobar",
			ops:               []any{Inclusive(true)},
			expectedMatched:   "",
			expectedSeparator: "",
			expectedRemaining: "foobar",
			expectedOk:        false,
		},
		{
			name:              "Longer separator",
			seps:              []string{"//"},
			input:             "foo//bar",
			ops:               []any{Inclusive(true)},
			expectedMatched:   "foo//",
			expectedSeparator: "//",
			expectedRemaining: "bar",
			expectedOk:        true,
		},
		{
			name:              "Multiple separators",
			seps:              []string{"/", "-"},
			input:             "foo-bar",
			ops:               []any{Inclusive(true)},
			expectedMatched:   "foo-",
			expectedSeparator: "-",
			expectedRemaining: "bar",
			expectedOk:        true,
		},
		{
			name:              "Empty string",
			seps:              []string{"/"},
			input:             "",
			ops:               []any{Inclusive(true)},
			expectedMatched:   "",
			expectedSeparator: "",
			expectedRemaining: "",
			expectedOk:        false,
		},
		{
			name:              "Separator at end",
			seps:              []string{"/"},
			input:             "foo/",
			ops:               []any{Inclusive(true)},
			expectedMatched:   "foo/",
			expectedSeparator: "/",
			expectedRemaining: "",
			expectedOk:        true,
		},
		{
			name:              "Just separator",
			seps:              []string{"/"},
			input:             "/",
			ops:               []any{Inclusive(true)},
			expectedMatched:   "/",
			expectedSeparator: "/",
			expectedRemaining: "",
			expectedOk:        true,
		},
		{
			name:              "No separators",
			seps:              []string{},
			input:             "/abc",
			ops:               []any{Inclusive(true)},
			expectedMatched:   "",
			expectedSeparator: "",
			expectedRemaining: "/abc",
			expectedOk:        false,
		},
		{
			name:              "Longest match preference",
			seps:              []string{"/", "//"},
			input:             "foo//bar",
			ops:               []any{Inclusive(true)},
			expectedMatched:   "foo//",
			expectedSeparator: "//",
			expectedRemaining: "bar",
			expectedOk:        true,
		},
		{
			name:              "Not inclusive",
			seps:              []string{"/"},
			input:             "foo/bar",
			ops:               []any{Inclusive(false)},
			expectedMatched:   "foo",
			expectedSeparator: "/",
			expectedRemaining: "/bar",
			expectedOk:        true,
		},
		{
			name:              "Start offset",
			seps:              []string{"/"},
			input:             "/foo/bar",
			ops:               []any{Inclusive(true), StartOffset(1)},
			expectedMatched:   "/foo/",
			expectedSeparator: "/",
			expectedRemaining: "bar",
			expectedOk:        true,
		},
		{
			name:              "Ignore 0 position match",
			seps:              []string{"/"},
			input:             "/foo/bar",
			ops:               []any{Inclusive(true), Ignore0PositionMatch(true)},
			expectedMatched:   "/foo/",
			expectedSeparator: "/",
			expectedRemaining: "bar",
			expectedOk:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cu := NewConsumeUntiler(tt.seps...)
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
