package viewport

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/google/go-cmp/cmp"
	"strings"
	"testing"
)

func wrap(line string, width int) []string {
	if width <= 0 {
		return []string{}
	}

	line = strings.TrimRight(line, " ")

	var res []string
	lineWidth := lipgloss.Width(line)
	for xOffset := 0; xOffset < lineWidth; xOffset += width {
		truncatedLine := truncateLine(line, xOffset, width, "")
		res = append(res, truncatedLine)
	}
	return res
}

func percent(a, b int) int {
	return int(float32(a) / float32(b) * 100)
}

// truncateLine returns the visible part of a line given an xOffset and width
func truncateLine(s string, xOffset, width int, lineContinuationIndicator string) string {
	// "the full line is like this"
	//      |xOffset     |xOffset+width
	//      |start       |end
	//      |..l line i..|  <- returned if lineContinuationIndicator = ".."

	if width <= 0 {
		return ""
	}

	ansiCodeIndexes := ansiPattern.FindAllStringIndex(s, -1)
	plainText := ansiPattern.ReplaceAllString(s, "")
	lenPlainText := len(plainText)

	indicatorLen := len(lineContinuationIndicator)
	if width <= indicatorLen && lenPlainText > width {
		return lineContinuationIndicator[:width]
	}

	start := xOffset
	end := xOffset + width
	if start < 0 {
		start = 0
	}
	if start >= lenPlainText {
		return reapplyANSI(s, "", ansiCodeIndexes, 0)
	}

	if end > lenPlainText {
		end = lenPlainText
	}

	if end < start {
		return reapplyANSI(s, "", ansiCodeIndexes, 0)
	}

	if end-start == lenPlainText && len(ansiCodeIndexes) == 0 {
		return s
	}

	if end-start <= indicatorLen && lenPlainText > indicatorLen {
		return lineContinuationIndicator[:min(indicatorLen, end-start)]
	}

	visible := plainText[start:end]
	if xOffset > 0 {
		visible = lineContinuationIndicator + visible[indicatorLen:]
	}
	if end < len(plainText) {
		visible = visible[:width-indicatorLen] + lineContinuationIndicator
	}

	return reapplyANSI(s, visible, ansiCodeIndexes, start)
}

func reapplyANSI(original, truncated string, ansiCodeIndexes [][]int, start int) string {
	var result []byte
	var lenAnsiAdded int

	for i := 0; i < len(truncated); i++ {
		for len(ansiCodeIndexes) > 0 {
			candidateAnsi := ansiCodeIndexes[0]
			codeStart, codeEnd := candidateAnsi[0], candidateAnsi[1]
			originalIdx := start + i + lenAnsiAdded
			if codeStart <= originalIdx || codeEnd <= originalIdx {
				result = append(result, original[codeStart:codeEnd]...)
				lenAnsiAdded += codeEnd - codeStart
				ansiCodeIndexes = ansiCodeIndexes[1:]
			} else {
				break
			}
		}
		result = append(result, truncated[i])
	}

	// add remaining ansi codes in order to end
	for _, codeIndexes := range ansiCodeIndexes {
		codeStart, codeEnd := codeIndexes[0], codeIndexes[1]
		result = append(result, original[codeStart:codeEnd]...)
	}

	return string(result)
}

// pad is a test helper function that pads the given lines to the given width and height.
// for example, pad(5, 4, []string{"a", "b", "c"}) will be padded to:
// "a    "
// "b    "
// "c    "
// "     "
// as a single string
func pad(width, height int, lines []string) string {
	var res []string
	for _, line := range lines {
		resLine := line
		numSpaces := width - lipgloss.Width(line)
		if numSpaces > 0 {
			resLine += strings.Repeat(" ", numSpaces)
		}
		res = append(res, resLine)
	}
	numEmptyLines := height - len(lines)
	for i := 0; i < numEmptyLines; i++ {
		res = append(res, strings.Repeat(" ", width))
	}
	return strings.Join(res, "\n")
}

// compare is a test helper function that compares two strings and fails the test if they are different
func compare(t *testing.T, expected, actual string) {
	if diff := cmp.Diff(expected, actual); diff != "" {
		t.Errorf("Mismatch (-want +got):\n%s", diff)
	}
}

func safeSliceUpToIdx[T any](s []T, i int) []T {
	if i > len(s) {
		return s
	}
	if i < 0 {
		return []T{}
	}
	return s[:i]
}

func safeSliceFromIdx(s []string, i int) []string {
	if i < 0 {
		return s
	}
	if i > len(s) {
		return []string{}
	}
	return s[i:]
}

func clampValMinMax(v, minimum, maximum int) int {
	return max(minimum, min(maximum, v))
}
