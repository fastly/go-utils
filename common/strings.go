package common

import (
	"sort"
	"strings"
)

// EmbeddedLines returns a sorted slice of lines that
// are common in every input string when the string is split
// by \n. That is, the input string has embedded newlines.
func EmbeddedLines(inputs []string) []string {
	commonalities := make(map[string]int)
	for _, input := range inputs {
		split := strings.Split(input, "\n")
		for _, line := range split {
			commonalities[line] = commonalities[line] + 1
		}
	}

	max := 0
	for _, count := range commonalities {
		if count > max {
			max = count
		}
	}

	commonLines := []string{} // != nil

	for line, count := range commonalities {
		if count == max {
			commonLines = append(commonLines, line)
		}
	}

	sort.Strings(commonLines)

	return commonLines
}
