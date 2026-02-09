# go-consume

`go-consume` is a Go library for consuming strings with various matchers. It provides a `ConsumeUntiler` to scan an input string for configured separators and return the matched substring, the separator, and the remaining string.

## Installation

```bash
go get github.com/arran4/go-consume
```

## Usage

### Basic Usage

```go
package main

import (
	"fmt"
	"github.com/arran4/go-consume"
	"github.com/arran4/go-consume/strconsume"
)

func main() {
	// Create a new ConsumeUntiler with separators "/"
	cu := strconsume.NewConsumeUntiler("/")

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

### Options

The `Consume` method accepts optional arguments to control behavior:

- `consume.Inclusive(true)`: Include the separator in the returned `matched` string. The `remaining` string will start after the separator.
- `consume.StartOffset(n)`: Start scanning from index `n`.
- `consume.Ignore0PositionMatch(true)`: Ignore matches at the very beginning of the string (index 0).

```go
// Example with Inclusive(true)
matched, separator, remaining, found := cu.Consume("path/to/resource", consume.Inclusive(true))
// matched: "path/"
// remaining: "to/resource"
```

## License

BSD 3-Clause License. See [LICENSE](LICENSE) for details.
