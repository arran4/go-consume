package strconsume

import (
	"github.com/arran4/go-consume"
	"testing"
)

func TestPrefixConsumer_Iterator_Inclusive_Panic(t *testing.T) {
	pc := NewPrefixConsumer("foo")
	iter := pc.Iterator("foo", consume.Inclusive(true))
	iter(func(matched, remaining string) bool { return true })
}
