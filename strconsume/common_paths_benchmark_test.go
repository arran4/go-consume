package strconsume

import (
	"fmt"
	"testing"
)

// Helper to generate paths
func generatePaths(count int, depth int, width int) []string {
	paths := make([]string, 0, count)

	// Characters to use
	chars := "abcdefghijklmnopqrstuvwxyz"

	// Recursive helper
	var generate func(currentPath string, currentDepth int)
	generate = func(currentPath string, currentDepth int) {
		if len(paths) >= count {
			return
		}
		if currentDepth >= depth {
			paths = append(paths, currentPath)
			return
		}

		for i := 0; i < width; i++ {
			// Deterministic pseudo-random char selection
			charIdx := (len(paths) + i + currentDepth) % len(chars)
			segment := string(chars[charIdx])
			generate(currentPath+"/"+segment, currentDepth+1)
			if len(paths) >= count {
				return
			}
		}
	}

	generate("", 0)

	// If we didn't fill up, just duplicate
	for len(paths) < count {
		paths = append(paths, paths[0])
	}

	return paths
}

func BenchmarkCommonPaths_Map(b *testing.B) {
	runBenchmark(b, CommonPathsMap)
}

func BenchmarkCommonPaths_Sort(b *testing.B) {
	runBenchmark(b, CommonPaths)
}

func runBenchmark(b *testing.B, fn func([]string) []*MatchPair) {
	inputSizes := []int{10, 100, 1000, 10000}

	for _, size := range inputSizes {
		// Use a fixed seed or deterministic generation so both benchmarks run on same input
		input := generatePaths(size, 5, 5)
		b.Run(fmt.Sprintf("Size_%d", size), func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				fn(input)
			}
		})
	}

	// Long paths scenario
	longPaths := generatePaths(100, 50, 2)
	b.Run("LongPaths_100", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			fn(longPaths)
		}
	})

	// No shared prefix
	noShared := make([]string, 100)
	for i := 0; i < 100; i++ {
		noShared[i] = fmt.Sprintf("/%d", i)
	}
	b.Run("NoShared_100", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			fn(noShared)
		}
	})

	// Dense prefixes scenario (highly overlapping)
	dense := generatePaths(1000, 10, 2) // Depth 10, width 2 -> 1024 leaves
	b.Run("DensePrefixes_1000", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			fn(dense)
		}
	})
}
