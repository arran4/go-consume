package strconsume

import (
	"testing"

	"github.com/arran4/go-consume"
	"github.com/stretchr/testify/assert"
)

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
