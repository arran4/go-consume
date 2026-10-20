package strconsume

import (
	"bufio"
	"fmt"
	"io"
	"reflect"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/arran4/go-consume"
)

type prefixPair struct{ matched, remaining string }

func TestPrefixConsumer_OptionContracts(t *testing.T) {
	slash := consume.MustBeFollowedBy(func(r rune) bool { return r == '/' })
	never := consume.MustBeFollowedBy(func(r rune) bool { return false })
	unicodeFollower := consume.MustBeFollowedBy(func(r rune) bool { return r == '界' })
	tests := []struct {
		name   string
		paths  []string
		input  string
		ops    []any
		pairs  []prefixPair
		tokens []string
	}{
		{name: "sequential", input: "foofoo", pairs: []prefixPair{{"foo", "foo"}, {"foo", ""}}, tokens: []string{"foo", "foo"}},
		{name: "sequence stops at gap", paths: []string{"foo", "bar"}, input: "foobarbazfoo", pairs: []prefixPair{{"foo", "barbazfoo"}, {"bar", "bazfoo"}}, tokens: []string{"foo", "bar"}},
		{name: "non-prefix", input: "xfoo"},
		{name: "no match", input: "xxxx"},
		{name: "empty input"},
		{name: "inclusive empty suffix", input: "foo", ops: []any{consume.Inclusive(true)}, pairs: []prefixPair{{"foo", ""}}, tokens: []string{"foo"}},
		{name: "inclusive short suffix", input: "foox", ops: []any{consume.Inclusive(true)}, pairs: []prefixPair{{"foo", "x"}}, tokens: []string{"foo"}},
		{name: "inclusive false", input: "foofoo", ops: []any{consume.Inclusive(false)}, pairs: []prefixPair{{"foo", "foo"}, {"foo", ""}}, tokens: []string{"foo", "foo"}},
		{name: "inclusive offset", input: "xfoofoo", ops: []any{consume.Inclusive(true), consume.StartOffset(1)}, pairs: []prefixPair{{"xfoo", "foo"}, {"foo", ""}}, tokens: []string{"foo", "foo"}},
		{name: "exclusive offset", input: "xfoofoo", ops: []any{consume.StartOffset(1)}, pairs: []prefixPair{{"foo", "foo"}, {"foo", ""}}, tokens: []string{"foo", "foo"}},
		{name: "offset is anchored", input: "xxfoo", ops: []any{consume.StartOffset(1)}},
		{name: "negative offset", input: "foo", ops: []any{consume.StartOffset(-1)}, pairs: []prefixPair{{"foo", ""}}, tokens: []string{"foo"}},
		{name: "offset at end", input: "foo", ops: []any{consume.StartOffset(3)}},
		{name: "offset past end", input: "foo", ops: []any{consume.StartOffset(100)}},
		{name: "offset integer limit", input: "foo", ops: []any{consume.StartOffset(int(^uint(0) >> 1))}},
		{name: "ignore initial only", input: "foofoofoo", ops: []any{consume.Ignore0PositionMatch(true)}, pairs: []prefixPair{{"foo", "foo"}, {"foo", ""}}, tokens: []string{"foo", "foo"}},
		{name: "ignore searches gap", input: "fooxfoofoo", ops: []any{consume.Ignore0PositionMatch(true)}, pairs: []prefixPair{{"foo", "foo"}, {"foo", ""}}, tokens: []string{"foo", "foo"}},
		{name: "ignore includes skipped input", input: "fooxfoofoo", ops: []any{consume.Ignore0PositionMatch(true), consume.Inclusive(true)}, pairs: []prefixPair{{"fooxfoo", "foo"}, {"foo", ""}}, tokens: []string{"foo", "foo"}},
		{name: "ignore does not enable initial search", input: "xfoo", ops: []any{consume.Ignore0PositionMatch(true)}},
		{name: "ignore with offset", input: "xfoofoo", ops: []any{consume.StartOffset(1), consume.Ignore0PositionMatch(true)}, pairs: []prefixPair{{"foo", "foo"}, {"foo", ""}}, tokens: []string{"foo", "foo"}},
		{name: "at end accepts", input: "foo", ops: []any{consume.MustBeAtEnd(true)}, pairs: []prefixPair{{"foo", ""}}, tokens: []string{"foo"}},
		{name: "at end rejects", input: "foobar", ops: []any{consume.MustBeAtEnd(true)}},
		{name: "at end searches gap", input: "fooxfoo", ops: []any{consume.MustBeAtEnd(true)}, pairs: []prefixPair{{"foo", ""}}, tokens: []string{"foo"}},
		{name: "at end searches overlap", paths: []string{"aba"}, input: "ababa", ops: []any{consume.MustBeAtEnd(true)}, pairs: []prefixPair{{"aba", ""}}, tokens: []string{"aba"}},
		{name: "follower accepts", input: "foo/", ops: []any{slash}, pairs: []prefixPair{{"foo", "/"}}, tokens: []string{"foo"}},
		{name: "follower rejects", input: "foox", ops: []any{slash}},
		{name: "follower EOF accepts", input: "foo", ops: []any{never}, pairs: []prefixPair{{"foo", ""}}, tokens: []string{"foo"}},
		{name: "follower searches gap", input: "fooxfoo/", ops: []any{slash}, pairs: []prefixPair{{"foo", "/"}}, tokens: []string{"foo"}},
		{name: "multibyte follower", input: "foo界", ops: []any{unicodeFollower}, pairs: []prefixPair{{"foo", "界"}}, tokens: []string{"foo"}},
		{name: "invalid follower rune", input: "foo\xff", ops: []any{consume.MustBeFollowedBy(func(r rune) bool { return r == utf8.RuneError })}, pairs: []prefixPair{{"foo", "\xff"}}, tokens: []string{"foo"}},
		{name: "truncated follower at EOF", input: "foo\xe7", ops: []any{consume.MustBeFollowedBy(func(r rune) bool { return r == utf8.RuneError })}, pairs: []prefixPair{{"foo", "\xe7"}}, tokens: []string{"foo"}},
		{name: "at end and follower", input: "fooxfoo", ops: []any{consume.MustBeAtEnd(true), never}, pairs: []prefixPair{{"foo", ""}}, tokens: []string{"foo"}},
		{name: "at end and follower reject", input: "foo/", ops: []any{consume.MustBeAtEnd(true), slash}},
		{name: "ignore and constraints", input: "fooxfoo", ops: []any{consume.Ignore0PositionMatch(true), consume.MustBeAtEnd(true), never}, pairs: []prefixPair{{"foo", ""}}, tokens: []string{"foo"}},
		{name: "ignore at EOF", input: "foo", ops: []any{consume.Ignore0PositionMatch(true), consume.MustBeAtEnd(true), never}},
		{name: "repeated inclusive last false", input: "xfoo", ops: []any{consume.StartOffset(1), consume.Inclusive(true), consume.Inclusive(false)}, pairs: []prefixPair{{"foo", ""}}, tokens: []string{"foo"}},
		{name: "repeated inclusive last true", input: "xfoo", ops: []any{consume.StartOffset(1), consume.Inclusive(false), consume.Inclusive(true)}, pairs: []prefixPair{{"xfoo", ""}}, tokens: []string{"foo"}},
		{name: "repeated offset", input: "xfoofoo", ops: []any{consume.StartOffset(100), consume.StartOffset(1)}, pairs: []prefixPair{{"foo", "foo"}, {"foo", ""}}, tokens: []string{"foo", "foo"}},
		{name: "explicit false clears constraints", input: "foox", ops: []any{consume.Ignore0PositionMatch(true), consume.Ignore0PositionMatch(false), consume.MustBeAtEnd(true), consume.MustBeAtEnd(false)}, pairs: []prefixPair{{"foo", "x"}}, tokens: []string{"foo"}},
		{name: "repeated follower", input: "foo/", ops: []any{never, slash}, pairs: []prefixPair{{"foo", "/"}}, tokens: []string{"foo"}},
		{name: "nil follower clears", input: "foox", ops: []any{never, consume.MustBeFollowedBy(nil)}, pairs: []prefixPair{{"foo", "x"}}, tokens: []string{"foo"}},
		{name: "unsupported typed options", input: "foox", ops: []any{consume.CaseInsensitive(true), consume.MustMatchWholeString(true), consume.ConsumeRemainingIfNotFound(true), consume.Escape(""), consume.Encasing{}, consume.EscapeBreaksEncasing(true), true, 1, nil}, pairs: []prefixPair{{"foo", "x"}}, tokens: []string{"foo"}},
		{name: "unsupported case folding", input: "FOO", ops: []any{consume.CaseInsensitive(true)}},
		{name: "longest prefix", paths: []string{"f", "foo", "foobar"}, input: "foobarfoo", pairs: []prefixPair{{"foobar", "foo"}, {"foo", ""}}, tokens: []string{"foobar", "foo"}},
		{name: "partial longer prefix at EOF", paths: []string{"foo", "foobar"}, input: "fooba", pairs: []prefixPair{{"foo", "ba"}}, tokens: []string{"foo"}},
		{name: "longest rejected no shorter fallback", paths: []string{"foo", "foobar"}, input: "foobar/", ops: []any{consume.MustBeFollowedBy(func(r rune) bool { return r == 'b' })}},
		{name: "byte offset", paths: []string{"界"}, input: "é界界", ops: []any{consume.StartOffset(2), consume.Inclusive(true)}, pairs: []prefixPair{{"é界", "界"}, {"界", ""}}, tokens: []string{"界", "界"}},
		{name: "offset inside rune", paths: []string{"界"}, input: "é界", ops: []any{consume.StartOffset(1)}},
		{name: "empty prefix nonempty input", paths: []string{""}, input: "abc", pairs: []prefixPair{{"", "abc"}}, tokens: []string{""}},
		{name: "empty prefix empty input", paths: []string{""}, pairs: []prefixPair{{"", ""}}, tokens: []string{""}},
		{name: "empty prefix ignored", paths: []string{""}, input: "x", ops: []any{consume.Ignore0PositionMatch(true)}, pairs: []prefixPair{{"", ""}}, tokens: []string{""}},
		{name: "empty prefix ignored empty input", paths: []string{""}, ops: []any{consume.Ignore0PositionMatch(true)}},
		{name: "empty prefix at end", paths: []string{""}, input: "abc", ops: []any{consume.MustBeAtEnd(true), consume.Inclusive(true)}, pairs: []prefixPair{{"abc", ""}}, tokens: []string{""}},
		{name: "empty prefix offset end", paths: []string{""}, input: "abc", ops: []any{consume.StartOffset(3), consume.Inclusive(true)}, pairs: []prefixPair{{"abc", ""}}, tokens: []string{""}},
		{name: "empty prefix offset past end", paths: []string{""}, input: "abc", ops: []any{consume.StartOffset(4)}},
		{name: "empty and longer prefix", paths: []string{"", "foo", "foobar"}, input: "foobar", pairs: []prefixPair{{"foobar", ""}, {"", ""}}, tokens: []string{"foobar", ""}},
		{name: "no configured prefixes", paths: []string{}, input: "foo"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paths := tt.paths
			if paths == nil {
				paths = []string{"foo"}
			}
			pc := NewPrefixConsumer(paths...)
			iter := pc.Iterator(tt.input, tt.ops...)
			for run := 0; run < 2; run++ {
				var got []prefixPair
				iter(func(m, r string) bool {
					if len(got) > len(tt.input)+1 {
						t.Fatal("iterator did not terminate")
					}
					got = append(got, prefixPair{m, r})
					return true
				})
				if !reflect.DeepEqual(got, tt.pairs) {
					t.Fatalf("iterator = %#v; want %#v", got, tt.pairs)
				}
			}
			for _, chunk := range []int{1, 2, 3, 4, 7, 1024} {
				for _, eofWithData := range []bool{false, true} {
					t.Run(fmt.Sprintf("scanner/chunk=%d/EOF-with-data=%t", chunk, eofWithData), func(t *testing.T) {
						scanner := bufio.NewScanner(&prefixChunkReader{reader: strings.NewReader(tt.input), chunk: chunk, eofWithData: eofWithData})
						scanner.Split(pc.SplitFunc(tt.ops...))
						var got []string
						for scanner.Scan() {
							if len(got) > len(tt.input)+1 {
								t.Fatal("scanner did not terminate")
							}
							got = append(got, scanner.Text())
						}
						if err := scanner.Err(); err != nil {
							t.Fatalf("scanner.Err() = %v", err)
						}
						if !reflect.DeepEqual(got, tt.tokens) {
							t.Fatalf("tokens = %#v; want %#v", got, tt.tokens)
						}
					})
				}
			}
		})
	}
}

type prefixChunkReader struct {
	reader      *strings.Reader
	chunk       int
	eofWithData bool
}

func (r *prefixChunkReader) Read(p []byte) (int, error) {
	n, err := r.reader.Read(p[:min(len(p), r.chunk)])
	if r.eofWithData && r.reader.Len() == 0 {
		return n, io.EOF
	}
	return n, err
}

func TestPrefixConsumer_Iterator_EarlyStop(t *testing.T) {
	calls := 0
	NewPrefixConsumer("foo").Iterator("foofoo")(func(m, r string) bool {
		calls++
		if m != "foo" || r != "foo" {
			t.Fatalf("pair = (%q, %q)", m, r)
		}
		return false
	})
	if calls != 1 {
		t.Fatalf("callbacks = %d; want 1", calls)
	}
}

func TestPrefixConsumer_SplitFunc_Lookahead(t *testing.T) {
	tests := []struct {
		name    string
		paths   []string
		ops     []any
		chunks  []string
		final   string
		eof     bool
		advance int
		token   string
	}{
		{"EOF", []string{"foo"}, []any{consume.MustBeAtEnd(true)}, []string{"f", "foo"}, "foo", true, 3, "foo"},
		{"follower", []string{"foo"}, []any{consume.MustBeFollowedBy(func(r rune) bool { return r == '界' })}, []string{"foo", "foo\xe7", "foo\xe7\x95"}, "foo界", false, 3, "foo"},
		{"longer prefix", []string{"foo", "foobar"}, nil, []string{"f", "foo", "foob", "fooba"}, "foobar", false, 6, "foobar"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			split := NewPrefixConsumer(tt.paths...).SplitFunc(tt.ops...)
			for _, chunk := range tt.chunks {
				a, tok, err := split([]byte(chunk), false)
				if a != 0 || tok != nil || err != nil {
					t.Fatalf("split(%q, false) = (%d, %q, %v); want more data", chunk, a, tok, err)
				}
			}
			a, tok, err := split([]byte(tt.final), tt.eof)
			if a != tt.advance || string(tok) != tt.token || err != nil {
				t.Fatalf("final split = (%d, %q, %v)", a, tok, err)
			}
		})
	}
}

func TestPrefixConsumer_SplitFunc_BoundedSkipping(t *testing.T) {
	gap := strings.Repeat("x", 2*bufio.MaxScanTokenSize)
	for _, tt := range []struct {
		name, input string
		ops         []any
		want        []string
	}{
		{"definite mismatch stops", gap + "foo", nil, nil},
		{"large offset", gap + "foo", []any{consume.StartOffset(len(gap))}, []string{"foo"}},
		{"rejection searches long gap", "foo" + gap + "foo", []any{consume.MustBeAtEnd(true)}, []string{"foo"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			scanner := bufio.NewScanner(&prefixChunkReader{reader: strings.NewReader(tt.input), chunk: 7})
			scanner.Buffer(make([]byte, 32), 32)
			scanner.Split(NewPrefixConsumer("foo").SplitFunc(tt.ops...))
			var got []string
			for scanner.Scan() {
				got = append(got, scanner.Text())
			}
			if err := scanner.Err(); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("tokens = %#v; want %#v", got, tt.want)
			}
		})
	}
}

func TestPrefixConsumer_Consume_Contracts(t *testing.T) {
	slash := consume.MustBeFollowedBy(func(r rune) bool { return r == '/' })
	never := consume.MustBeFollowedBy(func(r rune) bool { return false })
	tests := []struct {
		name, input              string
		paths                    []string
		ops                      []any
		before, match, remaining string
		found                    bool
	}{
		{name: "search", input: "xfoo/", before: "x", match: "foo", remaining: "foo/", found: true},
		{name: "inclusive", input: "xfoo/", ops: []any{consume.Inclusive(true)}, before: "xfoo", match: "foo", remaining: "/", found: true},
		{name: "inclusive empty suffix", input: "xfoo", ops: []any{consume.Inclusive(true)}, before: "xfoo", match: "foo", found: true},
		{name: "offset still searches", input: "xxfoo", ops: []any{consume.StartOffset(1)}, before: "xx", match: "foo", remaining: "foo", found: true},
		{name: "negative offset", input: "foo", ops: []any{consume.StartOffset(-10)}, match: "foo", remaining: "foo", found: true},
		{name: "offset at end", input: "foo", ops: []any{consume.StartOffset(3)}, remaining: "foo"},
		{name: "offset past end", input: "foo", ops: []any{consume.StartOffset(100)}, remaining: "foo"},
		{name: "ignore original zero", input: "foofoo", ops: []any{consume.Ignore0PositionMatch(true)}, before: "foo", match: "foo", remaining: "foo", found: true},
		{name: "offset and ignore", input: "xfoo", ops: []any{consume.StartOffset(1), consume.Ignore0PositionMatch(true)}, before: "x", match: "foo", remaining: "foo", found: true},
		{name: "follower accepts", input: "foo/", ops: []any{slash}, match: "foo", remaining: "foo/", found: true},
		{name: "follower rejects", input: "foox", ops: []any{slash}, remaining: "foox"},
		{name: "end and follower accept EOF", input: "fooxfoo", ops: []any{consume.MustBeAtEnd(true), never}, before: "foox", match: "foo", remaining: "foo", found: true},
		{name: "end and follower reject", input: "foo/", ops: []any{consume.MustBeAtEnd(true), slash}, remaining: "foo/"},
		{name: "end and ignore reject", input: "foo", ops: []any{consume.MustBeAtEnd(true), consume.Ignore0PositionMatch(true)}, remaining: "foo"},
		{name: "repeated options", input: "xfoo/", ops: []any{consume.StartOffset(100), consume.StartOffset(1), consume.Inclusive(true), consume.Inclusive(false), consume.Ignore0PositionMatch(true), consume.Ignore0PositionMatch(false), consume.MustBeAtEnd(true), consume.MustBeAtEnd(false), never, slash}, before: "x", match: "foo", remaining: "foo/", found: true},
		{name: "nil follower", input: "foox", ops: []any{never, consume.MustBeFollowedBy(nil)}, match: "foo", remaining: "foox", found: true},
		{name: "unsupported options", input: "foo", ops: []any{consume.Escape(""), consume.Encasing{}, consume.CaseInsensitive(true), nil}, match: "foo", remaining: "foo", found: true},
		{name: "empty input", input: ""},
		{name: "empty prefix empty input", paths: []string{""}},
		{name: "empty prefix", paths: []string{""}, input: "abc", remaining: "abc", found: true},
		{name: "empty prefix ignored", paths: []string{""}, input: "abc", ops: []any{consume.Ignore0PositionMatch(true), consume.Inclusive(true)}, before: "a", remaining: "bc", found: true},
		{name: "empty prefix does not search final position", paths: []string{""}, input: "abc", ops: []any{consume.MustBeAtEnd(true)}, remaining: "abc"},
		{name: "longest no shorter fallback", paths: []string{"foo", "foobar"}, input: "foobar/", ops: []any{consume.MustBeFollowedBy(func(r rune) bool { return r == 'b' })}, remaining: "foobar/"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paths := tt.paths
			if paths == nil {
				paths = []string{"foo"}
			}
			b, m, r, ok := NewPrefixConsumer(paths...).Consume(tt.input, tt.ops...)
			if b != tt.before || m != tt.match || r != tt.remaining || ok != tt.found {
				t.Fatalf("Consume = (%q, %q, %q, %t); want (%q, %q, %q, %t)", b, m, r, ok, tt.before, tt.match, tt.remaining, tt.found)
			}
		})
	}
}

// Check trie selection against a simple reference and scanner chunk invariance
// against whole-string iteration, including arbitrary byte strings and overlaps.
func FuzzPrefixConsumer_Streaming(f *testing.F) {
	for _, seed := range []struct {
		input, a, b string
		flags       uint8
	}{
		{"foobarfoo", "foo", "foobar", 0},
		{"fooxfoo", "foo", "f", 2},
		{"foo界", "foo", "", 4},
		{"ababa", "aba", "", 3},
		{"", "", "", 0},
		{"xfoofoo", "foo", "bar", 8},
	} {
		f.Add(seed.input, seed.a, seed.b, seed.flags)
	}
	f.Fuzz(func(t *testing.T, input, a, b string, flags uint8) {
		// Bound each fuzz case while retaining arbitrary bytes and token boundaries.
		if len(input) > 512 || len(a) > 64 || len(b) > 64 {
			t.Skip()
		}
		pc := NewPrefixConsumer(a, b)
		wantMatch, wantFound := "", false
		for _, path := range []string{a, b} {
			if strings.HasPrefix(input, path) && (!wantFound || len(path) > len(wantMatch)) {
				wantMatch, wantFound = path, true
			}
		}
		match, found := pc.LongestPrefix(input)
		if match != wantMatch || found != wantFound {
			t.Fatalf("LongestPrefix = (%q, %t); want (%q, %t)", match, found, wantMatch, wantFound)
		}
		ops := []any{consume.Ignore0PositionMatch(flags&1 != 0), consume.MustBeAtEnd(flags&2 != 0), consume.StartOffset(int(flags>>3)%5 - 1)}
		if flags&4 != 0 {
			ops = append(ops, consume.MustBeFollowedBy(func(r rune) bool { return r == '界' || r == '/' }))
		}
		var want []string
		pc.Iterator(input, ops...)(func(m, r string) bool {
			if len(want) > len(input)+1 {
				t.Fatal("iterator did not terminate")
			}
			want = append(want, m)
			return true
		})
		for _, chunk := range []int{1, 3, 1024} {
			scanner := bufio.NewScanner(&prefixChunkReader{reader: strings.NewReader(input), chunk: chunk, eofWithData: flags&16 != 0})
			scanner.Split(pc.SplitFunc(ops...))
			var got []string
			for scanner.Scan() {
				if len(got) > len(input)+1 {
					t.Fatal("scanner did not terminate")
				}
				got = append(got, scanner.Text())
			}
			if err := scanner.Err(); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("chunk=%d tokens = %#v; iterator = %#v", chunk, got, want)
			}
		}
	})
}
