package viewport

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/google/go-cmp/cmp"
	"strings"
	"testing"
)

// TODO LEO: test
func splitLineIntoSizedChunks(line string, chunkSize int) []string {
	var wrappedLines []string
	for {
		lineWidth := stringWidth(line)
		if lineWidth == 0 {
			break
		}

		width := chunkSize
		if lineWidth < chunkSize {
			width = lineWidth
		}

		wrappedLines = append(wrappedLines, line[0:width])
		line = line[width:]
	}
	return wrappedLines
}

func percent(a, b int) int {
	return int(float32(a) / float32(b) * 100)
}

// stringWidth is a function in case in the future something like utf8.RuneCountInString or lipgloss.Width is better
func stringWidth(s string) int {
	// NOTE: lipgloss.Width is significantly less performant than len
	return lipgloss.Width(s)
}

func getVisiblePartOfLine(s string, xOffset, width int, lineContinuationIndicator string) string {
	if width <= 0 {
		return ""
	}

	ansiCodeIndexes := ansiPattern.FindAllStringIndex(s, -1)
	plainText := ansiPattern.ReplaceAllString(s, "")
	lenPlainText := len(plainText)

	indicatorLen := len(lineContinuationIndicator)
	if width <= indicatorLen {
		return lineContinuationIndicator[:width]
	}

	start := xOffset
	end := xOffset + width
	if start < 0 {
		start = 0
	}
	if start >= lenPlainText {
		return ""
	}

	if end > lenPlainText {
		end = lenPlainText
	}

	if end < start {
		return ""
	}

	// TODO LEO: fix this
	println(s, xOffset, width, start, end)
	if end-start < width {
		return s[start:end]
	}

	if end-start <= indicatorLen {
		return lineContinuationIndicator[:min(indicatorLen, end-start)]
	}

	if width == 2*indicatorLen {
		return lineContinuationIndicator + lineContinuationIndicator
	}

	visible := plainText[start:end]
	if xOffset > 0 {
		visible = lineContinuationIndicator + visible[indicatorLen:]
	}
	if end < len(plainText) {
		visible = visible[:width-indicatorLen] + lineContinuationIndicator
	}

	return reapplyANSI(s, visible, ansiCodeIndexes, start, end)
}

func reapplyANSI(original, truncated string, ansiCodeIndexes [][]int, start, end int) string {
	var result []byte
	var lenAnsiAdded int

	// step through the truncated string and add back ansi codes at the relevant locations
	for i := 0; i < len(truncated); i++ {
		originalIdx := start + i + lenAnsiAdded

		for j, codeIndexes := range ansiCodeIndexes {
			codeStart, codeEnd := codeIndexes[0], codeIndexes[1]
			if codeStart <= originalIdx && originalIdx < codeEnd {
				result = append(result, original[codeStart:codeEnd]...)
				lenAnsiAdded += codeEnd - codeStart

				// remove the added code
				ansiCodeIndexes = append(ansiCodeIndexes[:j], ansiCodeIndexes[j+1:]...)
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
