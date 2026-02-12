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
		{
			name:              "Escape character",
			seps:              []string{":"},
			input:             "foo\\:bar:baz",
			ops:               []any{consume.Escape("\\"), consume.Inclusive(true)},
			expectedMatched:   "foo\\:bar:",
			expectedSeparator: ":",
			expectedRemaining: "baz",
			expectedOk:        true,
		},
		{
			name:              "Escape string",
			seps:              []string{":"},
			input:             "fooESC:bar:baz",
			ops:               []any{consume.Escape("ESC"), consume.Inclusive(true)},
			expectedMatched:   "fooESC:bar:",
			expectedSeparator: ":",
			expectedRemaining: "baz",
			expectedOk:        true,
		},
		{
			name:              "Encasing quotes",
			seps:              []string{":"},
			input:             `foo"bar:baz":qux`,
			ops:               []any{consume.Encasing{Start: "\"", End: "\""}, consume.Inclusive(true)},
			expectedMatched:   `foo"bar:baz":`,
			expectedSeparator: ":",
			expectedRemaining: "qux",
			expectedOk:        true,
		},
		{
			name:              "Encasing brackets",
			seps:              []string{":"},
			input:             `foo(bar:baz):qux`,
			ops:               []any{consume.Encasing{Start: "(", End: ")"}, consume.Inclusive(true)},
			expectedMatched:   `foo(bar:baz):`,
			expectedSeparator: ":",
			expectedRemaining: "qux",
			expectedOk:        true,
		},
		{
			name:              "Encasing nested brackets",
			seps:              []string{":"},
			input:             `foo((bar:baz)):qux`,
			ops:               []any{consume.Encasing{Start: "(", End: ")"}, consume.Inclusive(true)},
			expectedMatched:   `foo((bar:baz)):`,
			expectedSeparator: ":",
			expectedRemaining: "qux",
			expectedOk:        true,
		},
		{
			name:              "Escape breaks encasing",
			seps:              []string{":"},
			input:             `foo"bar\"baz":qux`,
			ops:               []any{consume.Encasing{Start: "\"", End: "\""}, consume.Escape("\\"), consume.EscapeBreaksEncasing(true), consume.Inclusive(true)},
			expectedMatched:   `foo"bar\"baz":`,
			expectedSeparator: ":",
			expectedRemaining: "qux",
			expectedOk:        true,
		},
		{
			name:              "Escape at end",
			seps:              []string{":"},
			input:             "foo\\",
			ops:               []any{consume.Escape("\\"), consume.Inclusive(true)},
			expectedMatched:   "",
			expectedSeparator: "",
			expectedRemaining: "foo\\",
			expectedOk:        false,
		},
		{
			name:              "Encasing Start same as Separator",
			seps:              []string{"("},
			input:             "((foo))",
			ops:               []any{consume.Encasing{Start: "(", End: ")"}, consume.Inclusive(true)},
			expectedMatched:   "",
			expectedSeparator: "",
			expectedRemaining: "((foo))",
			expectedOk:        false,
		},
		{
			name:              "Mixed Nesting: Quotes inside brackets",
			seps:              []string{":"},
			input:             `(")")`,
			ops:               []any{consume.Encasing{Start: "(", End: ")"}, consume.Encasing{Start: "\"", End: "\""}, consume.Inclusive(true)},
			expectedMatched:   "",
			expectedSeparator: "",
			expectedRemaining: `(")")`,
			expectedOk:        false,
		},
		{
			name:              "Mixed Nesting: Brackets inside quotes",
			seps:              []string{":"},
			input:             `"( )"`,
			ops:               []any{consume.Encasing{Start: "(", End: ")"}, consume.Encasing{Start: "\"", End: "\""}, consume.Inclusive(true)},
			expectedMatched:   "",
			expectedSeparator: "",
			expectedRemaining: `"( )"`,
			expectedOk:        false,
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

func TestUntilConsumer_SplitFunc_Features(t *testing.T) {
	cu := NewUntilConsumer(":")

	t.Run("Escape character", func(t *testing.T) {
		split := cu.SplitFunc(
			consume.Escape("\\"),
			consume.Inclusive(true),
		)
		data := []byte("foo\\:bar:baz")
		advance, token, err := split(data, false)
		assert.NoError(t, err)
		assert.Equal(t, 9, advance) // "foo\:bar:" len is 9
		assert.Equal(t, "foo\\:bar:", string(token))
	})

	t.Run("Encasing", func(t *testing.T) {
		split := cu.SplitFunc(
			consume.Encasing{Start: "(", End: ")"},
			consume.Inclusive(true),
		)
		data := []byte("foo(bar:baz):qux")
		advance, token, err := split(data, false)
		assert.NoError(t, err)
		assert.Equal(t, 13, advance) // "foo(bar:baz):" len is 13
		assert.Equal(t, "foo(bar:baz):", string(token))
	})

	t.Run("Mixed Nesting", func(t *testing.T) {
		split := cu.SplitFunc(
			consume.Encasing{Start: "(", End: ")"},
			consume.Encasing{Start: "\"", End: "\""},
			consume.Inclusive(true),
		)
		data := []byte(`(")")`)
		advance, token, err := split(data, true)
		assert.NoError(t, err)
		assert.Equal(t, 5, advance)
		assert.Equal(t, `(")")`, string(token))
	})
}

func TestUntilConsumer_Validation(t *testing.T) {
	cu := NewUntilConsumer(":")

	t.Run("Invalid Empty Escape", func(t *testing.T) {
		assert.Panics(t, func() {
			cu.Consume("foo", consume.Escape(""))
		})
	})

	t.Run("Invalid Empty Encasing Start", func(t *testing.T) {
		assert.Panics(t, func() {
			cu.Consume("foo", consume.Encasing{Start: "", End: "x"})
		})
	})
}
