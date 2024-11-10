package viewport

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/google/go-cmp/cmp"
	"regexp"
	"strings"
	"testing"
	"unicode"
	"unicode/utf8"
)

func wrap(line string, width int, maxLinesEachEnd int) []string {
	if width <= 0 {
		return []string{}
	}

	if maxLinesEachEnd <= 0 {
		maxLinesEachEnd = -1
	}

	// if line has non-whitespace, trim trailing spaces
	if strings.TrimSpace(line) != "" {
		line = strings.TrimRightFunc(line, unicode.IsSpace)
	}

	// preserve empty lines
	if line == "" {
		return []string{line}
	}

	var res []string
	lineWidth := lipgloss.Width(line)
	totalLines := (lineWidth + width - 1) / width

	if maxLinesEachEnd > 0 && totalLines > maxLinesEachEnd*2 {
		for xOffset := 0; xOffset < width*maxLinesEachEnd; xOffset += width {
			truncatedLine := truncateLine(line, xOffset, width, "")
			res = append(res, truncatedLine)
		}

		remainingLines := totalLines - maxLinesEachEnd
		startOffset := remainingLines * width
		for xOffset := startOffset; xOffset < lineWidth; xOffset += width {
			truncatedLine := truncateLine(line, xOffset, width, "")
			res = append(res, truncatedLine)
		}
	} else {
		for xOffset := 0; xOffset < lineWidth; xOffset += width {
			truncatedLine := truncateLine(line, xOffset, width, "")
			res = append(res, truncatedLine)
		}
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

	if xOffset == 0 && len(lineContinuationIndicator) == 0 && !strings.Contains(s, "\x1b") {
		if len(s) <= width {
			return s
		}
		return string([]rune(s)[:width])
	}

	ansiCodeIndexes := ansiPattern.FindAllStringIndex(s, -1)

	var plainText string
	if len(ansiCodeIndexes) > 0 {
		plainText = ansiPattern.ReplaceAllString(s, "")
	} else {
		plainText = s
	}

	plainRunes := []rune(plainText)
	lenPlainText := len(plainRunes)

	if lenPlainText == 0 || xOffset >= lenPlainText {
		if len(ansiCodeIndexes) > 0 {
			return reapplyANSI(s, "", ansiCodeIndexes, 0)
		}
		return ""
	}

	indicatorLen := len([]rune(lineContinuationIndicator))
	if width <= indicatorLen && lenPlainText > width {
		return lineContinuationIndicator[:width]
	}

	start := xOffset
	if start < 0 {
		start = 0
	}

	end := xOffset + width
	if end > lenPlainText {
		end = lenPlainText
	}

	if end <= start {
		if len(ansiCodeIndexes) > 0 {
			return reapplyANSI(s, "", ansiCodeIndexes, 0)
		}
		return ""
	}

	if start == 0 && end == lenPlainText && len(ansiCodeIndexes) == 0 {
		return s
	}

	visible := string(plainRunes[start:end])

	if indicatorLen > 0 {
		if end-start <= indicatorLen && lenPlainText > indicatorLen {
			return lineContinuationIndicator[:min(indicatorLen, end-start)]
		}

		visLen := len([]rune(visible))
		if xOffset > 0 && visLen > indicatorLen {
			visible = lineContinuationIndicator + string([]rune(visible)[indicatorLen:])
		} else if xOffset > 0 {
			visible = lineContinuationIndicator
		}

		if end < lenPlainText && visLen > indicatorLen {
			visible = string([]rune(visible)[:visLen-indicatorLen]) + lineContinuationIndicator
		} else if end < lenPlainText {
			visible = lineContinuationIndicator
		}
	}

	if len(ansiCodeIndexes) > 0 {
		byteOffset := 0
		for i := 0; i < start; i++ {
			_, size := utf8.DecodeRuneInString(string(plainRunes[i]))
			byteOffset += size
		}
		return reapplyANSI(s, visible, ansiCodeIndexes, byteOffset)
	}

	return visible
}

func reapplyANSI(original, truncated string, ansiCodeIndexes [][]int, startBytes int) string {
	var result []byte
	var lenAnsiAdded int
	truncatedBytes := []byte(truncated)

	for i := 0; i < len(truncatedBytes); {
		for len(ansiCodeIndexes) > 0 {
			candidateAnsi := ansiCodeIndexes[0]
			codeStart, codeEnd := candidateAnsi[0], candidateAnsi[1]
			originalIdx := startBytes + i + lenAnsiAdded
			if codeStart <= originalIdx || codeEnd <= originalIdx {
				result = append(result, original[codeStart:codeEnd]...)
				lenAnsiAdded += codeEnd - codeStart
				ansiCodeIndexes = ansiCodeIndexes[1:]
			} else {
				break
			}
		}

		_, size := utf8.DecodeRune(truncatedBytes[i:])
		result = append(result, truncatedBytes[i:i+size]...)
		i += size
	}

	// add remaining ansi codes in order to end
	for _, codeIndexes := range ansiCodeIndexes {
		codeStart, codeEnd := codeIndexes[0], codeIndexes[1]
		result = append(result, original[codeStart:codeEnd]...)
	}

	return string(result)
}

// highlightLine highlights a line that potentially has ansi codes in it without disrupting them
func highlightLine(line string, highlight string, highlightStyle lipgloss.Style) string {
	ansiCodeIndexes := ansiPattern.FindAllStringIndex(line, -1)

	if len(ansiCodeIndexes) == 0 {
		// no pre-existing ansi codes, easy
		return strings.ReplaceAll(line, highlight, highlightStyle.Render(highlight))
	}

	if len(ansiCodeIndexes)%2 != 0 {
		// odd number of ansi codes, meaning something like one is not reset
		// doesn't currently handle this, naively pretend there are no ansi codes
		return strings.ReplaceAll(line, highlight, highlightStyle.Render(highlight))
	}

	highlightRegexp := regexp.MustCompile(regexp.QuoteMeta(highlight))
	highlightIndexes := highlightRegexp.FindAllStringIndex(line, -1)

	resetCode := "\x1b[0m"
	var result strings.Builder
	lastIndex := 0
	for _, highlightIndex := range highlightIndexes {
		start, end := highlightIndex[0], highlightIndex[1]

		// check if the highlight is within any ansi code
		inAnsiCode := false
		for _, ansiIndex := range ansiCodeIndexes {
			if start >= ansiIndex[0] && start < ansiIndex[1] {
				inAnsiCode = true
				break
			}
		}

		if inAnsiCode {
			// if the highlight is within an ansi code, don't apply highlighting
			result.WriteString(line[lastIndex:end])
		} else {
			// add the part before the highlight
			result.WriteString(line[lastIndex:start])

			// check if the highlight is within any ansi code range (between start and end codes)
			inAnsiRange := false
			var activeAnsiCode string
			for i := 0; i < len(ansiCodeIndexes); i += 2 {
				if ansiCodeIndexes[i][0] <= start && end <= ansiCodeIndexes[i+1][0] {
					inAnsiRange = true
					activeAnsiCode = line[ansiCodeIndexes[i][0]:ansiCodeIndexes[i][1]]
					break
				}
			}

			if inAnsiRange {
				// reset, apply highlight, then reapply active ansi code
				result.WriteString(resetCode)
				result.WriteString(highlightStyle.Render(line[start:end]))
				result.WriteString(activeAnsiCode)
			} else {
				// just apply highlight without resetting or reapplying ansi codes
				result.WriteString(highlightStyle.Render(line[start:end]))
			}
		}

		lastIndex = end
	}

	// add the remaining part of the line
	result.WriteString(line[lastIndex:])
	return result.String()
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