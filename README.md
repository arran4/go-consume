# go-consume

`go-consume` is a Go library for consuming strings with various matchers. It provides flexible consumers to scan input strings, making it easy to parse paths, commands, or structured text.

## Installation

```bash
go get github.com/arran4/go-consume
```

## Usage

### UntilConsumer

`UntilConsumer` scans an input string for configured separators.

```go
package main

import (
	"fmt"
	"github.com/arran4/go-consume"
	"github.com/arran4/go-consume/strconsume"
)

func main() {
	// Create a new UntilConsumer with separators "/"
	cu := strconsume.NewUntilConsumer("/")

	input := "path/to/resource"
	
	// Consume until the first separator
	matched, separator, remaining, found := cu.Consume(input)
	
	if found {
		fmt.Printf("Matched: %s\n", matched)       // Output: path
		fmt.Printf("Separator: %s\n", separator)   // Output: /
		fmt.Printf("Remaining: %s\n", remaining)   // Output: to/resource
	}
}
```

### PrefixConsumer

`PrefixConsumer` checks if the input string starts with any of the configured prefixes.

```go
package main

import (
	"fmt"
	"github.com/arran4/go-consume"
	"github.com/arran4/go-consume/strconsume"
)

func main() {
	// Create a new PrefixConsumer
	pc := strconsume.NewPrefixConsumer("foo", "bar")

	input := "foobar"
	
	// Consume prefix
	matched, remaining, found := pc.Consume(input)
	
	if found {
		fmt.Printf("Matched: %s\n", matched)     // Output: foo
		fmt.Printf("Remaining: %s\n", remaining) // Output: bar
	}
}
```

### Options

The `Consume` methods accept optional arguments to control behavior.

#### Options for `UntilConsumer`

- `consume.Inclusive(true)`: Include the separator in the returned `matched` string. The `remaining` string will start after the separator.
- `consume.StartOffset(n)`: Start scanning from index `n`.
- `consume.Ignore0PositionMatch(true)`: Ignore matches at the very beginning of the string (index 0).
- `consume.CaseInsensitive(true)`: Match separators case-insensitively.

```go
// Example with Inclusive(true)
matched, separator, remaining, found := cu.Consume("path/to/resource", consume.Inclusive(true))
// matched: "path/"
// remaining: "to/resource"
```

#### Options for `PrefixConsumer`

- `consume.CaseInsensitive(true)`: Match prefixes case-insensitively.

```go
// Example with CaseInsensitive(true)
pc := strconsume.NewPrefixConsumer("Foo")
matched, remaining, found := pc.Consume("foobar", consume.CaseInsensitive(true))
// matched: "foo" (returns the matched part from input)
// remaining: "bar"
```

## License

BSD 3-Clause License. See [LICENSE](LICENSE) for details.
