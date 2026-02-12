package strconsume

import (
	"testing"

	"github.com/arran4/go-consume"
	"github.com/stretchr/testify/assert"
)

func TestUntilConsumer_MixedNesting(t *testing.T) {
	cu := NewUntilConsumer(":")

	t.Run("Quotes inside brackets", func(t *testing.T) {
		matched, sep, remaining, found := cu.Consume(
			`(")")`,
			consume.Encasing{Start: "(", End: ")"},
			consume.Encasing{Start: "\"", End: "\""},
			consume.Inclusive(true),
		)
		assert.False(t, found)
		assert.Equal(t, "", matched)
		assert.Equal(t, "", sep)
		assert.Equal(t, `(")")`, remaining)
	})

	t.Run("Brackets inside quotes", func(t *testing.T) {
		matched, sep, remaining, found := cu.Consume(
			`"( )"`,
			consume.Encasing{Start: "(", End: ")"},
			consume.Encasing{Start: "\"", End: "\""},
			consume.Inclusive(true),
		)
		assert.False(t, found)
		assert.Equal(t, "", matched)
		assert.Equal(t, "", sep)
		assert.Equal(t, `"( )"`, remaining)
	})

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

	t.Run("SplitFunc Mixed Nesting", func(t *testing.T) {
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
