package consume

type Inclusive bool
type StartOffset int
type Ignore0PositionMatch bool
type MustBeFollowedBy func(rune) bool
type MustBeAtEnd bool
type CaseInsensitive bool
type MustMatchWholeString bool
type ConsumeRemainingIfNotFound bool

type Escape string
type Encasing struct {
	Start, End string
}
type EscapeBreaksEncasing bool
