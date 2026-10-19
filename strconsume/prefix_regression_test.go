package strconsume

import (
	"bufio"
	"github.com/arran4/go-consume"
	"strings"
	"testing"
)

func TestPrefixConsumer_Iterator_IgnoresOptions(t *testing.T) {
	pc := NewPrefixConsumer("foo")

	// MustBeAtEnd(true) should mean it only matches "foo" at the end.
	// In "foobar", "foo" is at the start, so it shouldn't match.
	iter := pc.Iterator("foobar", consume.MustBeAtEnd(true))
	matchedFirst := false
	iter(func(matched, remaining string) bool {
		if matched == "foo" {
			matchedFirst = true
		}
		return false
	})

	if matchedFirst {
		t.Errorf("PrefixConsumer.Iterator ignored options (e.g., MustBeAtEnd)")
	}
}

func TestPrefixConsumer_SplitFunc_IgnoresOptions(t *testing.T) {
	pc := NewPrefixConsumer("foo")

	// Inclusive(true) is an option, or MustBeAtEnd(true)
	split := pc.SplitFunc(consume.MustBeAtEnd(true))
	scanner := bufio.NewScanner(strings.NewReader("foobar"))
	scanner.Split(split)

	if scanner.Scan() {
		if scanner.Text() == "foo" {
			t.Errorf("PrefixConsumer.SplitFunc ignored options (e.g., MustBeAtEnd)")
		}
	}
}
