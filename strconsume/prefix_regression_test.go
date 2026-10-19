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

func TestPrefixConsumer_Iterator_ContinuesSearching(t *testing.T) {
	pc := NewPrefixConsumer("foo")

	// Ignore0PositionMatch(true) should reject the first "foo", but it should find the second "foo"
	iter := pc.Iterator("foofoo", consume.Ignore0PositionMatch(true))

	count := 0
	iter(func(matched, remaining string) bool {
		count++
		return true
	})

	if count != 1 {
		t.Errorf("Expected exactly 1 match (the second 'foo'), got %d matches", count)
	}
}

func TestPrefixConsumer_SplitFunc_ChunkedAtEnd(t *testing.T) {
	pc := NewPrefixConsumer("foo")
	split := pc.SplitFunc(consume.MustBeAtEnd(true))

	// Chunk 1: "foo", not EOF. Must return 0, nil, nil.
	adv, tok, err := split([]byte("foo"), false)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if adv != 0 || tok != nil {
		t.Errorf("Expected 0 advance and nil token for 'foo' chunk without EOF, got adv=%d, tok=%q", adv, tok)
	}

	// Chunk 2: "foo", with EOF. Must return 3, "foo", nil.
	adv, tok, err = split([]byte("foo"), true)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if adv != 3 || string(tok) != "foo" {
		t.Errorf("Expected 3 advance and 'foo' token for 'foo' chunk at EOF, got adv=%d, tok=%q", adv, tok)
	}
}

func TestPrefixConsumer_SplitFunc_ChunkedFollowedBy(t *testing.T) {
	pc := NewPrefixConsumer("foo")
	follow := consume.MustBeFollowedBy(func(r rune) bool { return r == 'b' })
	split := pc.SplitFunc(follow)

	// Chunk 1: "foo", not EOF. Must return 0, nil, nil.
	adv, tok, err := split([]byte("foo"), false)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if adv != 0 || tok != nil {
		t.Errorf("Expected 0 advance and nil token for 'foo' chunk without EOF, got adv=%d, tok=%q", adv, tok)
	}

	// Chunk 2: "foob", not EOF. Must return 3, "foo", nil.
	adv, tok, err = split([]byte("foob"), false)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if adv != 3 || string(tok) != "foo" {
		t.Errorf("Expected 3 advance and 'foo' token for 'foob' chunk, got adv=%d, tok=%q", adv, tok)
	}
}

func TestPrefixConsumer_SplitFunc_MultipleOptions(t *testing.T) {
	pc := NewPrefixConsumer("foo")
	follow := consume.MustBeFollowedBy(func(r rune) bool { return r == 'b' })
	split := pc.SplitFunc(follow, consume.MustBeAtEnd(true)) // these are mutually exclusive basically for non-empty string, but let's test a valid failure scenario.

	adv, tok, _ := split([]byte("foob"), true)
	if adv != 0 || tok != nil {
		t.Errorf("Expected failure because 'foob' is not at end for 'foo', got adv=%d", adv)
	}
}
