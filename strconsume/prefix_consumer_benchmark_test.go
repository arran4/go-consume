package strconsume

import (
	"fmt"
	"testing"
)

// Reuse generatePaths from common_paths_benchmark_test.go (same package)

func BenchmarkPrefixConsumer_LongestPrefix(b *testing.B) {
	runPrefixConsumerBenchmark(b, func(ps *PrefixConsumer, text string) {
		ps.LongestPrefix(text)
	})
}

func runPrefixConsumerBenchmark(b *testing.B, fn func(*PrefixConsumer, string)) {
	inputSizes := []int{10, 100, 1000, 10000}

	for _, size := range inputSizes {
		paths := generatePaths(size, 5, 5)
		ps := NewPrefixConsumer(paths)

		// Generate some test inputs
		inputs := make([]string, 1000)
		for i := 0; i < 1000; i++ {
			// Pick a path and append something
			base := paths[i % len(paths)]
			inputs[i] = base + "/suffix"
		}

		b.Run(fmt.Sprintf("Size_%d_Hit", size), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				fn(ps, inputs[i % len(inputs)])
			}
		})

		// Miss cases
		missInputs := make([]string, 1000)
		for i := 0; i < 1000; i++ {
			missInputs[i] = fmt.Sprintf("/notfound/%d", i)
		}
		b.Run(fmt.Sprintf("Size_%d_Miss", size), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				fn(ps, missInputs[i % len(missInputs)])
			}
		})
	}
}

// Comparison with Map-based approach (naive iteration over map keys or sorted lengths)
// Simulate the map-based approach used in ConsumeUntiler (roughly)
type MapSearcher struct {
	matchers map[int]map[string]struct{}
	sizes    []int
}

func NewMapSearcher(paths []string) *MapSearcher {
	matchers := map[int]map[string]struct{}{}
	var sizes []int
	seenSizes := map[int]bool{}

	for _, p := range paths {
		l := len(p)
		if matchers[l] == nil {
			matchers[l] = map[string]struct{}{}
		}
		matchers[l][p] = struct{}{}
		if !seenSizes[l] {
			sizes = append(sizes, l)
			seenSizes[l] = true
		}
	}
	// Sort sizes descending
	// (omitted for brevity, assume simple iteration)
	return &MapSearcher{matchers: matchers, sizes: sizes}
}

func (ms *MapSearcher) Find(text string) (string, bool) {
	// Iterate all sizes (not sorted here properly but simulates checking map)
	// In reality we should sort descending for longest match
	// Just iterate map keys for simplicity of benchmark setup cost
	for size, m := range ms.matchers {
		if len(text) >= size {
			sub := text[:size]
			if _, ok := m[sub]; ok {
				return sub, true
			}
		}
	}
	return "", false
}

func BenchmarkMapSearch_Find(b *testing.B) {
	// This benchmark is approximate as it doesn't sort sizes exactly like ConsumeUntiler
	// But gives an idea of map lookup cost vs binary search
	inputSizes := []int{10, 100, 1000, 10000}

	for _, size := range inputSizes {
		paths := generatePaths(size, 5, 5)
		ms := NewMapSearcher(paths)

		inputs := make([]string, 1000)
		for i := 0; i < 1000; i++ {
			base := paths[i % len(paths)]
			inputs[i] = base + "/suffix"
		}

		b.Run(fmt.Sprintf("Size_%d_Hit", size), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				ms.Find(inputs[i % len(inputs)])
			}
		})
	}
}
