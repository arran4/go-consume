package strconsume_test

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/arran4/go-consume"
	"github.com/arran4/go-consume/strconsume"
)

func ExamplePrefixConsumer_SplitFunc() {
	pc := strconsume.NewPrefixConsumer("foo", "bar")
	input := "foobarfoo"
	scanner := bufio.NewScanner(strings.NewReader(input))
	scanner.Split(pc.SplitFunc())

	for scanner.Scan() {
		fmt.Printf("Token: %q\n", scanner.Text())
	}
	// Output:
	// Token: "foo"
	// Token: "bar"
	// Token: "foo"
}

func ExampleUntilConsumer_SplitFunc() {
	cu := strconsume.NewUntilConsumer("/")
	input := "path/to/resource"
	scanner := bufio.NewScanner(strings.NewReader(input))
	scanner.Split(cu.SplitFunc())

	for scanner.Scan() {
		fmt.Printf("Token: %q\n", scanner.Text())
	}
	// Output:
	// Token: "path"
	// Token: "to"
	// Token: "resource"
}

func ExampleUntilConsumer_SplitFunc_inclusive() {
	cu := strconsume.NewUntilConsumer("/")
	input := "path/to/resource"
	scanner := bufio.NewScanner(strings.NewReader(input))
	scanner.Split(cu.SplitFunc(consume.Inclusive(true)))

	for scanner.Scan() {
		fmt.Printf("Token: %q\n", scanner.Text())
	}
	// Output:
	// Token: "path/"
	// Token: "to/"
	// Token: "resource"
}
