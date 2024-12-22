package viewport

import (
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/robinovitch61/kl/internal/constants"
	"github.com/robinovitch61/kl/internal/linebuffer"
	"strings"
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

	// preserve empty lines
	if line == "" {
		return []string{line}
	}

	var res []string
	lineWidth := lipgloss.Width(line)
	totalLines := (lineWidth + width - 1) / width

	lineBuffer := linebuffer.New(line, "")

	if maxLinesEachEnd > 0 && totalLines > maxLinesEachEnd*2 {
		for xOffset := 0; xOffset < width*maxLinesEachEnd; xOffset += width {
			res = append(res, lineBuffer.Truncate(xOffset, width))
		}

		startOffset := lineWidth - (maxLinesEachEnd * width)
		for xOffset := startOffset; xOffset < lineWidth; xOffset += width {
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

	// Helper function to check if we're inside an ANSI escape sequence
	isInAnsiCode := func(s string, pos int) bool {
		// Look back for ESC character
		for i := pos; i >= 0; i-- {
			if s[i] == '\x1b' {
				return true
			} else if s[i] == 'm' {
				return false
			}
		}
		return false
	}

	result := &strings.Builder{}
	i := 0
	activeStyle := ""

	for i < len(line) {
		if strings.HasPrefix(line[i:], "\x1b[") {
			// Found start of ANSI sequence
			escEnd := strings.Index(line[i:], "m")
			if escEnd != -1 {
				escEnd += i + 1
				currentSequence := line[i:escEnd]
				if currentSequence == "\x1b[m" {
					activeStyle = "" // Reset style
				} else {
					activeStyle = currentSequence // Set new active style
				}
				result.WriteString(currentSequence)
				i = escEnd
				continue
			}
		}

		// Check if current position starts a highlight match
		if len(highlight) > 0 && strings.HasPrefix(line[i:], highlight) && !isInAnsiCode(line, i) {
			// Reset current style if any
			if activeStyle != "" {
				result.WriteString("\x1b[m")
			}

			// Apply highlight
			result.WriteString(highlightStyle.Render(highlight))

			// Restore previous style if there was one
			if activeStyle != "" {
				result.WriteString(activeStyle)
			}

			i += len(highlight)
			continue
		}

		// Regular character
		result.WriteByte(line[i])
		i++
	}

	// removing empty sequences may hurt performance, but helps legibility
	return constants.EmptySequenceRegex.ReplaceAllString(result.String(), "")
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
