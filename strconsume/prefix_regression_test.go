package strconsume

import (
	"bufio"
	"strings"
	"testing"
	"github.com/arran4/go-consume"
)

func TestPrefixConsumer_Iterator_IgnoresOptions(t *testing.T) {
	pc := NewPrefixConsumer("foo")

	// MustBeAtEnd(true) should mean it only matches "foo" at the end.
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

	split := pc.SplitFunc(consume.MustBeAtEnd(true))
	scanner := bufio.NewScanner(strings.NewReader("foobar"))
	scanner.Split(split)

	if scanner.Scan() {
		if scanner.Text() == "foo" {
			t.Errorf("PrefixConsumer.SplitFunc ignored options (e.g., MustBeAtEnd)")
		}
	}
}
