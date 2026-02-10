package strconsume

import (
	"bufio"
	"strings"
	"testing"

	"github.com/arran4/go-consume"
	"github.com/stretchr/testify/assert"
)

func TestPrefixConsumer_SplitFunc(t *testing.T) {
	t.Run("Basic Prefix", func(t *testing.T) {
		pc := NewPrefixConsumer("foo", "bar")
		input := "foobarfoo"
		scanner := bufio.NewScanner(strings.NewReader(input))
		scanner.Split(pc.SplitFunc())

		var tokens []string
		for scanner.Scan() {
			tokens = append(tokens, scanner.Text())
		}
		// Expect "foo", "bar", "foo"
		assert.Equal(t, []string{"foo", "bar", "foo"}, tokens)
		assert.NoError(t, scanner.Err())
	})

	t.Run("Case Insensitive", func(t *testing.T) {
		pc := NewPrefixConsumer("foo")
		input := "FooFOOfoo"
		scanner := bufio.NewScanner(strings.NewReader(input))
		scanner.Split(pc.SplitFunc(consume.CaseInsensitive(true)))

		var tokens []string
		for scanner.Scan() {
			tokens = append(tokens, scanner.Text())
		}
		// Expect "Foo", "FOO", "foo" (the actual text from input)
		assert.Equal(t, []string{"Foo", "FOO", "foo"}, tokens)
		assert.NoError(t, scanner.Err())
	})
}

func TestUntilConsumer_SplitFunc(t *testing.T) {
	t.Run("Basic Until", func(t *testing.T) {
		cu := NewUntilConsumer("/")
		input := "a/b/c"
		scanner := bufio.NewScanner(strings.NewReader(input))
		scanner.Split(cu.SplitFunc())

		var tokens []string
		for scanner.Scan() {
			tokens = append(tokens, scanner.Text())
		}
		// Expect "a", "b", "c"
		assert.Equal(t, []string{"a", "b", "c"}, tokens)
		assert.NoError(t, scanner.Err())
	})

	t.Run("Inclusive", func(t *testing.T) {
		cu := NewUntilConsumer("/")
		input := "a/b/"
		scanner := bufio.NewScanner(strings.NewReader(input))
		scanner.Split(cu.SplitFunc(consume.Inclusive(true)))

		var tokens []string
		for scanner.Scan() {
			tokens = append(tokens, scanner.Text())
		}
		// Expect "a/", "b/"
		assert.Equal(t, []string{"a/", "b/"}, tokens)
		assert.NoError(t, scanner.Err())
	})
}
