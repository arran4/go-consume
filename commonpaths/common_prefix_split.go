package commonpaths

import "slices"

type MatchPair struct {
	Match string
	Path  []string
}

func createNewPath(path []string, newSegment string) []string {
	if newSegment == "" {
		return nil
	}
	newPath := make([]string, len(path)+1)
	copy(newPath, path)
	newPath[len(path)] = newSegment
	return newPath
}

func commonPrefixSplitMapRecursive(result []*MatchPair, all, path []string, sp int, matching []int) []*MatchPair {
	switch len(matching) {
	case 0:
		return result
	}
	ended := map[int]int{}
	var endedSeq []int
	var bucket map[rune][]int
	var runeSeq []rune
	i := sp
	for ; len(matching) > 0; i++ {
		bucket = map[rune][]int{}
		runeSeq = nil
		for _, eachI := range matching {
			each := all[eachI]
			if _, ok := ended[eachI]; ok {
				continue
			}
			if i >= len(each) {
				ended[eachI] = i
				endedSeq = append(endedSeq, eachI)
				continue
			}
			r := []rune(each)[i]
			if _, ok := bucket[r]; !ok {
				runeSeq = append(runeSeq, r)
			}
			bucket[r] = append(bucket[r], eachI)
		}
		if len(bucket) != 1 || len(ended) > 0 {
			break
		}
	}
	for _, each := range endedSeq {
		at := ended[each]
		newSegment := all[each][sp:at]
		newPath := createNewPath(path, newSegment)
		result = append(result, &MatchPair{
			Match: all[each],
			Path:  newPath,
		})
	}
	for _, r := range runeSeq {
		matchesWithEnded := bucket[r]
		newMatches := slices.Collect(func(yield func(int) bool) {
			for _, each := range matchesWithEnded {
				if _, hasEnded := ended[each]; hasEnded {
					continue
				}
				if !yield(each) {
					break
				}
			}
		})
		if len(newMatches) == 0 {
			continue
		}
		newSegment := all[newMatches[0]][sp:i]
		newPath := createNewPath(path, newSegment)
		result = commonPrefixSplitMapRecursive(result, all, newPath, i, newMatches)
	}
	return result
}

func commonPrefixSplitMapWrapper(all []string) []*MatchPair {
	result := make([]*MatchPair, 0, len(all))
	return commonPrefixSplitMapRecursive(result, all, nil, 0, slices.Collect(func(yield func(int) bool) {
		for i := range all {
			if !yield(i) {
				break
			}
		}
	}))
}

func CommonPrefixSplit(all []string) []*MatchPair {
	if len(all) == 0 {
		return []*MatchPair{}
	}
	sorted := make([]string, len(all))
	copy(sorted, all)
	slices.Sort(sorted)

	result := make([]*MatchPair, 0, len(all))
	var recurse func(start, end int, path []string, depth int)
	recurse = func(start, end int, path []string, depth int) {
		if start >= end {
			return
		}

		// Find common prefix length for the whole group
		lcp := 0
		s1 := sorted[start]
		s2 := sorted[end-1]
		minLen := len(s1)
		if len(s2) < minLen {
			minLen = len(s2)
		}

		// Byte-based comparison
		for i := 0; i < minLen-depth; i++ {
			if s1[depth+i] == s2[depth+i] {
				lcp++
			} else {
				break
			}
		}

		var currentPath []string
		if lcp > 0 {
			segment := s1[depth : depth+lcp]
			currentPath = createNewPath(path, segment)
			depth += lcp
		} else {
			currentPath = path
		}

		// Iterate through range and split
		idx := start

		// Handle strings ending at current depth
		for idx < end && len(sorted[idx]) == depth {
			result = append(result, &MatchPair{
				Match: sorted[idx],
				Path:  currentPath,
			})
			idx++
		}

		// Group remaining strings by next byte
		for idx < end {
			groupStart := idx
			char := sorted[idx][depth]
			// Find end of group with same char
			idx++
			for idx < end && len(sorted[idx]) > depth && sorted[idx][depth] == char {
				idx++
			}
			recurse(groupStart, idx, currentPath, depth)
		}
	}

	recurse(0, len(sorted), nil, 0)
	return result
}
