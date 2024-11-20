package viewport

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/google/go-cmp/cmp"
	"github.com/robinovitch61/kl/internal/linebuffer"
	"regexp"
	"strings"
	"testing"
	"unicode"
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
	//println(fmt.Sprintf("after trim time: %v", time.Since(start)))

	// preserve empty lines
	if line == "" {
		return []string{line}
	}

	var res []string
	lineWidth := lipgloss.Width(line)
	totalLines := (lineWidth + width - 1) / width

	lineBuffer := linebuffer.New(line, "")

	if maxLinesEachEnd > 0 && totalLines > maxLinesEachEnd*2 {
		i := 0
		for xOffset := 0; xOffset < width*maxLinesEachEnd; xOffset += width {
			i += 1
			res = append(res, lineBuffer.Truncate(xOffset, width))
		}

		startOffset := lineWidth - (maxLinesEachEnd * width)
		i = 0
		for xOffset := startOffset; xOffset < lineWidth; xOffset += width {
			i += 1
			res = append(res, lineBuffer.Truncate(xOffset, width))
		}
	} else {
		for xOffset := 0; xOffset < lineWidth; xOffset += width {
			res = append(res, lineBuffer.Truncate(xOffset, width))
		}
	}

	return res
}

func percent(a, b int) int {
	return int(float32(a) / float32(b) * 100)
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
