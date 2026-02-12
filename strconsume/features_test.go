package strconsume

import (
	"testing"

	"github.com/arran4/go-consume"
	"github.com/stretchr/testify/assert"
)

func TestUntilConsumer_Features(t *testing.T) {
	// Test Escape
	t.Run("Escape character", func(t *testing.T) {
		// Separator ":", Escape "\"
		// Input "foo\:bar:baz" -> Match "foo\:bar", Sep ":", Remaining "baz"
		cu := NewUntilConsumer(":")
		matched, sep, remaining, found := cu.Consume(
			"foo\\:bar:baz",
			consume.Escape("\\"),
			consume.Inclusive(true),
		)
		assert.True(t, found)
		assert.Equal(t, "foo\\:bar:", matched) // Inclusive includes separator
		assert.Equal(t, ":", sep)
		assert.Equal(t, "baz", remaining)
	})

	t.Run("Escape string", func(t *testing.T) {
		// Separator ":", Escape "ESC"
		// Input "fooESC:bar:baz" -> Match "fooESC:bar", Sep ":", Remaining "baz"
		cu := NewUntilConsumer(":")
		matched, sep, remaining, found := cu.Consume(
			"fooESC:bar:baz",
			consume.Escape("ESC"),
			consume.Inclusive(true),
		)
		assert.True(t, found)
		assert.Equal(t, "fooESC:bar:", matched)
		assert.Equal(t, ":", sep)
		assert.Equal(t, "baz", remaining)
	})

	// Test Encasing
	t.Run("Encasing quotes", func(t *testing.T) {
		// Separator ":", Encasing ""
		// Input `foo"bar:baz":qux` -> Match `foo"bar:baz"`, Sep `:`, Remaining `qux`
		cu := NewUntilConsumer(":")
		matched, sep, remaining, found := cu.Consume(
			`foo"bar:baz":qux`,
			consume.Encasing{Start: "\"", End: "\""},
			consume.Inclusive(true),
		)
		assert.True(t, found)
		assert.Equal(t, `foo"bar:baz":`, matched)
		assert.Equal(t, ":", sep)
		assert.Equal(t, "qux", remaining)
	})

	t.Run("Encasing brackets", func(t *testing.T) {
		// Separator ":", Encasing ()
		// Input `foo(bar:baz):qux` -> Match `foo(bar:baz)`, Sep `:`, Remaining `qux`
		cu := NewUntilConsumer(":")
		matched, sep, remaining, found := cu.Consume(
			`foo(bar:baz):qux`,
			consume.Encasing{Start: "(", End: ")"},
			consume.Inclusive(true),
		)
		assert.True(t, found)
		assert.Equal(t, `foo(bar:baz):`, matched)
		assert.Equal(t, ":", sep)
		assert.Equal(t, "qux", remaining)
	})

	t.Run("Encasing nested brackets", func(t *testing.T) {
		// Separator ":", Encasing ()
		// Input `foo((bar:baz)):qux` -> Match `foo((bar:baz))`, Sep `:`, Remaining `qux`
		cu := NewUntilConsumer(":")
		matched, sep, remaining, found := cu.Consume(
			`foo((bar:baz)):qux`,
			consume.Encasing{Start: "(", End: ")"},
			consume.Inclusive(true),
		)
		assert.True(t, found)
		assert.Equal(t, `foo((bar:baz)):`, matched)
		assert.Equal(t, ":", sep)
		assert.Equal(t, "qux", remaining)
	})

	// Test EscapeBreaksEncasing
	t.Run("Escape breaks encasing", func(t *testing.T) {
		// Separator ":", Encasing "", Escape "\"
		// Input `foo"bar\"baz":qux`
		cu := NewUntilConsumer(":")
		matched, sep, remaining, found := cu.Consume(
			`foo"bar\"baz":qux`,
			consume.Encasing{Start: "\"", End: "\""},
			consume.Escape("\\"),
			consume.EscapeBreaksEncasing(true),
			consume.Inclusive(true),
		)
		assert.True(t, found)
		assert.Equal(t, `foo"bar\"baz":`, matched)
		assert.Equal(t, ":", sep)
		assert.Equal(t, "qux", remaining)
	})
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
}

func TestUntilConsumer_Features_EdgeCases(t *testing.T) {
	t.Run("Escape at end", func(t *testing.T) {
		cu := NewUntilConsumer(":")
		matched, sep, remaining, found := cu.Consume(
			"foo\\",
			consume.Escape("\\"),
			consume.Inclusive(true),
		)
		assert.False(t, found)
		assert.Equal(t, "", matched)
		assert.Equal(t, "", sep)
		assert.Equal(t, "foo\\", remaining)
	})

	t.Run("Encasing Start same as Separator", func(t *testing.T) {
		cu := NewUntilConsumer("(")
		matched, sep, remaining, found := cu.Consume(
			"((foo))",
			consume.Encasing{Start: "(", End: ")"},
			consume.Inclusive(true),
		)
		assert.False(t, found)
		assert.Equal(t, "", matched)
		assert.Equal(t, "", sep)
		assert.Equal(t, "((foo))", remaining)
	})
}
