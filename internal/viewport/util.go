package viewport

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/google/go-cmp/cmp"
	"github.com/robinovitch61/kl/internal/linebuffer"
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
func highlightLine(line, highlight string, highlightStyle lipgloss.Style) string {
	if line == "" || highlight == "" {
		return line
	}

	// Pre-compute highlight style string, removing final reset
	highlightStr := strings.TrimSuffix(highlightStyle.String(), "\x1b[0m")

	var result strings.Builder
	result.Grow(len(line) * 2) // Pre-allocate buffer

	searchStart := 0
	writeStart := 0
	currentStyleStart := -1
	inAnsi := false

	lowLine := strings.ToLower(line)
	lowHighlight := strings.ToLower(highlight)

	for {
		idx := strings.Index(lowLine[searchStart:], lowHighlight)
		if idx == -1 {
			result.WriteString(line[writeStart:])
			break
		}

		idx += searchStart

		// Check if we're in an ANSI sequence
		for i := writeStart; i < idx; i++ {
			if line[i] == '\x1b' {
				inAnsi = true
				if line[i+1] == '[' {
					if currentStyleStart == -1 {
						currentStyleStart = i
					}
				}
			} else if inAnsi && line[i] == 'm' {
				inAnsi = false
				if i >= 2 && line[i-1] == '0' && line[i-2] == '[' && line[i-3] == '\x1b' {
					currentStyleStart = -1
				}
			}
		}

		// Skip if we're in an ANSI sequence
		if inAnsi {
			searchStart = idx + 1
			continue
		}

		// Write up to match
		result.WriteString(line[writeStart:idx])

		// Close current style if needed
		if currentStyleStart != -1 {
			result.WriteString("\x1b[0m")
		}

		// Write highlighted section
		result.WriteString(highlightStr)
		result.WriteString(line[idx : idx+len(highlight)])
		result.WriteString("\x1b[0m")

		// Restore previous style if needed
		if currentStyleStart != -1 {
			// Find end of style sequence
			styleEnd := currentStyleStart
			for styleEnd < len(line) {
				if line[styleEnd] == 'm' {
					result.WriteString(line[currentStyleStart : styleEnd+1])
					break
				}
				styleEnd++
			}
		}

		writeStart = idx + len(highlight)
		searchStart = writeStart

		// Reset style tracking for next section if we're at a reset sequence
		if writeStart >= 4 &&
			line[writeStart-4:writeStart] == "\x1b[0m" {
			currentStyleStart = -1
		}
	}

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
