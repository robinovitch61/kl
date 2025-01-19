package linebuffer

import (
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/mattn/go-runewidth"
	"github.com/robinovitch61/kl/internal/constants"
	"strings"
	"unicode/utf8"
)

// LineBuffer provides functionality to get sequential strings of a specified terminal width, accounting
// for the ansi escape codes styling the line.
type LineBuffer struct {
	line                string  // underlying string with ansi codes. utf-8 bytes
	leftRuneIdx         int     // left plaintext rune idx to start next PopLeft result from
	lineNoAnsi          string  // line without ansi codes. utf-8 bytes
	lineNoAnsiRunes     []rune  // runes of lineNoAnsi. len(lineNoAnsiRunes) == len(lineNoAnsiWidths)
	runeIdxToByteOffset []int   // idx of lineNoAnsiRunes to byte offset. len(runeIdxToByteOffset) == len(lineNoAnsiRunes)
	lineNoAnsiWidths    []int   // terminal cell widths of lineNoAnsi. len(lineNoAnsiWidths) == len(lineNoAnsiRunes)
	lineNoAnsiCumWidths []int   // cumulative lineNoAnsiWidths
	ansiCodeIndexes     [][]int // slice of startByte, endByte indexes of ansi codes in the line
}

// Fast ANSI scanner that avoids regex allocations
func findAnsiRanges(s string) [][]int {
	// Pre-count to allocate exact size
	count := strings.Count(s, "\x1b[")
	if count == 0 {
		return nil
	}

	ranges := make([][]int, 0, count)
	for i := 0; i < len(s); {
		if i+1 < len(s) && s[i] == '\x1b' && s[i+1] == '[' {
			// Found start of ANSI sequence
			start := i
			i += 2 // skip \x1b[

			// Find the 'm' that ends this sequence
			for i < len(s) && s[i] != 'm' {
				i++
			}

			if i < len(s) && s[i] == 'm' {
				ranges = append(ranges, []int{start, i + 1})
				i++
				continue
			}
		}
		i++
	}
	return ranges
}

func New(line string) LineBuffer {
	lb := LineBuffer{
		line:        line,
		leftRuneIdx: 0,
	}

	// Use custom scanner instead of regex
	lb.ansiCodeIndexes = findAnsiRanges(line)

	// Use same string strip logic but with our ranges
	if len(lb.ansiCodeIndexes) > 0 {
		var b strings.Builder
		b.Grow(len(line)) // Pre-allocate to avoid resizing
		lastPos := 0
		for _, r := range lb.ansiCodeIndexes {
			b.WriteString(line[lastPos:r[0]])
			lastPos = r[1]
		}
		b.WriteString(line[lastPos:])
		lb.lineNoAnsi = b.String()
	} else {
		lb.lineNoAnsi = line
	}

	// Get exact rune count
	n := utf8.RuneCountInString(lb.lineNoAnsi)

	// Allocate one large slice to hold all our integer data
	// Format: [runeIdxToByteOffset|widths|cumWidths]
	combined := make([]int, n+1+n+n)
	lb.runeIdxToByteOffset = combined[:n+1]
	lb.lineNoAnsiWidths = combined[n+1 : n+1+n]
	lb.lineNoAnsiCumWidths = combined[n+1+n:]

	// Still need runes slice separately since it's []rune not []int
	lb.lineNoAnsiRunes = make([]rune, n)

	// Single pass to fill all slices
	currentOffset := 0
	cumWidth := 0
	i := 0
	for _, r := range lb.lineNoAnsi {
		lb.runeIdxToByteOffset[i] = currentOffset
		currentOffset += utf8.RuneLen(r)

		lb.lineNoAnsiRunes[i] = r

		width := runewidth.RuneWidth(r)
		lb.lineNoAnsiWidths[i] = width
		cumWidth += width
		lb.lineNoAnsiCumWidths[i] = cumWidth
		i++
	}
	lb.runeIdxToByteOffset[n] = currentOffset

	return lb
}

func (l LineBuffer) TotalLines(width int) int {
	if width == 0 {
		return 0
	}
	return (l.fullWidth() + width - 1) / width
}

func (l *LineBuffer) SeekToLine(n, width int) {
	if n <= 0 {
		l.leftRuneIdx = 0
		return
	}
	l.leftRuneIdx = getLeftRuneIdx(n*width, l.lineNoAnsiCumWidths)
}

func (l *LineBuffer) SeekToWidth(width int) {
	// width can go past end, in which case PopLeft() returns "". Required when e.g. panning past line's end.
	if width <= 0 {
		l.leftRuneIdx = 0
		return
	}
	l.leftRuneIdx = getLeftRuneIdx(width, l.lineNoAnsiCumWidths)
}

func (l LineBuffer) fullWidth() int {
	if len(l.lineNoAnsiCumWidths) == 0 {
		return 0
	}
	return l.lineNoAnsiCumWidths[len(l.lineNoAnsiCumWidths)-1]
}

// getLeftRuneIdx does a binary search to find the first index at which vals[index-1] >= w
func getLeftRuneIdx(w int, vals []int) int {
	if w == 0 {
		return 0
	}
	if len(vals) == 0 {
		return 0
	}

	left, right := 0, len(vals)-1

	if vals[right] < w {
		return len(vals)
	}

	for left < right {
		mid := left + (right-left)/2

		if vals[mid] >= w {
			right = mid
		} else {
			left = mid + 1
		}
	}

	return left + 1
}

// PopLeft returns a string of the buffer's width from its current left offset, scrolling the left offset to the right
func (l *LineBuffer) PopLeft(width int, continuation, toHighlight string, highlightStyle lipgloss.Style) string {
	if l.leftRuneIdx >= len(l.lineNoAnsiRunes) || width == 0 {
		return ""
	}

	var result strings.Builder
	remainingWidth := width
	startRuneIdx := l.leftRuneIdx
	startByteOffset := l.runeIdxToByteOffset[startRuneIdx]

	runesWritten := 0
	for ; remainingWidth > 0 && l.leftRuneIdx < len(l.lineNoAnsiRunes); l.leftRuneIdx++ {
		r := l.lineNoAnsiRunes[l.leftRuneIdx]
		runeWidth := l.lineNoAnsiWidths[l.leftRuneIdx]
		if runeWidth > remainingWidth {
			break
		}

		result.WriteRune(r)
		runesWritten++
		remainingWidth -= runeWidth
	}

	res := result.String()

	// apply left/right line continuation indicators
	if len(continuation) > 0 && (startRuneIdx > 0 || l.leftRuneIdx < len(l.lineNoAnsiRunes)) {
		continuationRunes := []rune(continuation)

		// if more runes to the left of the result, replace start runes with continuation indicator, respecting width
		if startRuneIdx > 0 {
			res = replaceStartWithContinuation(res, continuationRunes)
		}

		// if more runes to the right, replace final runes in result with continuation indicator, respecting width
		if l.leftRuneIdx < len(l.lineNoAnsiRunes) {
			res = replaceEndWithContinuation(res, continuationRunes)
		}
	}

	// reapply original styling
	if len(l.ansiCodeIndexes) > 0 {
		res = reapplyAnsi(l.line, res, startByteOffset, l.ansiCodeIndexes)
	}

	// highlight the desired string
	res = l.highlightString(res, startByteOffset, toHighlight, highlightStyle)

	// remove empty sequences
	res = constants.EmptySequenceRegex.ReplaceAllString(res, "")

	return res
}

func (l *LineBuffer) WrappedLines(
	width int,
	maxLinesEachEnd int,
	toHighlight string,
	toHighlightStyle lipgloss.Style,
) []string {
	if width <= 0 {
		return []string{}
	}

	if maxLinesEachEnd <= 0 {
		maxLinesEachEnd = -1
	}

	// preserve empty lines
	if l.line == "" {
		return []string{l.line}
	}

	var res []string
	totalLines := l.TotalLines(width)

	l.SeekToWidth(0)
	if maxLinesEachEnd > 0 && totalLines > maxLinesEachEnd*2 {
		for nLines := 0; nLines < maxLinesEachEnd; nLines++ {
			res = append(res, l.PopLeft(width, "", toHighlight, toHighlightStyle))
		}

		l.SeekToLine(totalLines-maxLinesEachEnd, width)
		for nLines := 0; nLines < maxLinesEachEnd; nLines++ {
			res = append(res, l.PopLeft(width, "", toHighlight, toHighlightStyle))
		}
	} else {
		for nLines := 0; nLines < totalLines; nLines++ {
			res = append(res, l.PopLeft(width, "", toHighlight, toHighlightStyle))
		}
	}

	l.SeekToWidth(0)
	return res
}

func (l LineBuffer) highlightString(
	s string,
	startByteOffset int,
	toHighlight string,
	highlightStyle lipgloss.Style,
) string {
	if toHighlight != "" && len(highlightStyle.String()) > 0 {
		// highlight
		s = highlightLine(s, toHighlight, highlightStyle, 0, len(s))

		if left, endIdx := overflowsLeft(l.lineNoAnsi, startByteOffset, toHighlight); left {
			highlightLeft := l.lineNoAnsi[startByteOffset:endIdx]
			s = highlightLine(s, highlightLeft, highlightStyle, 0, len(highlightLeft))
		}
		endByteOffset := l.runeIdxToByteOffset[l.leftRuneIdx]
		if right, startIdx := overflowsRight(l.lineNoAnsi, endByteOffset, toHighlight); right {
			highlightRight := l.lineNoAnsi[startIdx:endByteOffset]
			lenPlainTextRes := len(stripAnsi(s))
			s = highlightLine(s, highlightRight, highlightStyle, lenPlainTextRes-len(highlightRight), lenPlainTextRes)
		}
	}

	return s
}

func reapplyAnsi(original, truncated string, truncByteOffset int, ansiCodeIndexes [][]int) string {
	var result []byte
	var lenAnsiAdded int
	isReset := true
	truncatedBytes := []byte(truncated)

	for i := 0; i < len(truncatedBytes); {
		// collect all ansi codes that should be applied immediately before the current runes
		var ansisToAdd []string
		for len(ansiCodeIndexes) > 0 {
			candidateAnsi := ansiCodeIndexes[0]
			codeStart, codeEnd := candidateAnsi[0], candidateAnsi[1]
			originalIdx := truncByteOffset + i + lenAnsiAdded
			if codeStart <= originalIdx {
				code := original[codeStart:codeEnd]
				isReset = code == "\x1b[m"
				ansisToAdd = append(ansisToAdd, code)
				lenAnsiAdded += codeEnd - codeStart
				ansiCodeIndexes = ansiCodeIndexes[1:]
			} else {
				break
			}
		}

		for _, ansi := range simplifyAnsiCodes(ansisToAdd) {
			result = append(result, ansi...)
		}

		// add the bytes of the current rune
		_, size := utf8.DecodeRune(truncatedBytes[i:])
		result = append(result, truncatedBytes[i:i+size]...)
		i += size
	}

	if !isReset {
		result = append(result, "\x1b[m"...)
	}

	return string(result)
}

// highlightLine highlights a string in a line that potentially has ansi codes in it without disrupting them
// start and end are the byte offsets for which highlighting is considered in the line, not counting ansi codes
func highlightLine(line, highlight string, highlightStyle lipgloss.Style, start, end int) string {
	if line == "" || highlight == "" {
		return line
	}

	renderedHighlight := highlightStyle.Render(highlight)
	lenHighlight := len(highlight)
	var result strings.Builder
	var activeStyles []string
	inAnsi := false
	nonAnsiBytes := 0

	i := 0
	for i < len(line) {
		if strings.HasPrefix(line[i:], "\x1b[") {
			// found start of ansi
			inAnsi = true
			ansiLen := strings.Index(line[i:], "m")
			if ansiLen != -1 {
				escEnd := i + ansiLen + 1
				ansi := line[i:escEnd]
				if ansi == "\x1b[m" {
					activeStyles = []string{} // reset
				} else {
					activeStyles = append(activeStyles, ansi) // add new active style
				}
				result.WriteString(ansi)
				i = escEnd
				inAnsi = false
				continue
			}
		}

		// check if current position starts a highlight match
		if len(highlight) > 0 && !inAnsi && nonAnsiBytes >= start && nonAnsiBytes < end && strings.HasPrefix(line[i:], highlight) {
			// reset current styles, if any
			if len(activeStyles) > 0 {
				result.WriteString("\x1b[m")
			}

			// apply highlight
			result.WriteString(renderedHighlight)
			nonAnsiBytes += lenHighlight

			// restore previous styles, if any
			if len(activeStyles) > 0 {
				for j := range activeStyles {
					result.WriteString(activeStyles[j])
				}
			}
			i += len(highlight)
			continue
		}

		result.WriteByte(line[i])
		nonAnsiBytes++
		i++
	}

	return result.String()
}

func stripAnsi(input string) string {
	return constants.AnsiRegex.ReplaceAllString(input, "")
}

func simplifyAnsiCodes(ansis []string) []string {
	//println()
	//for _, a := range ansis {
	//	println(fmt.Sprintf("%q", a))
	//}
	if len(ansis) == 0 {
		return []string{}
	}

	// if there's just a bunch of reset sequences, compress it to one
	allReset := true
	for _, ansi := range ansis {
		if ansi != "\x1b[m" {
			allReset = false
			break
		}
	}
	if allReset {
		return []string{"\x1b[m"}
	}

	// return all ansis to the right of the rightmost reset seq
	for i := len(ansis) - 1; i >= 0; i-- {
		if ansis[i] == "\x1b[m" {
			result := ansis[i+1:]
			// keep reset at the start if present
			if ansis[0] == "\x1b[m" {
				return append([]string{"\x1b[m"}, result...)
			}
			return result
		}
	}
	return ansis
}

// overflowsLeft checks if a substring overflows a string on the left if the string were to start at startByteIdx inclusive.
// assumes s has no ansi codes.
// It performs a case-sensitive comparison and returns two values:
//   - A boolean indicating whether there is overflow
//   - An integer indicating the ending string index (exclusive) of the overflow (0 if none)
//
// Examples:
//
//	                   01234567890
//		overflowsLeft("my str here", 3, "my str") returns (true, 6)
//		overflowsLeft("my str here", 3, "your str") returns (false, 0)
//		overflowsLeft("my str here", 6, "my str") returns (false, 0)
func overflowsLeft(s string, startByteIdx int, substr string) (bool, int) {
	if len(s) == 0 || len(substr) == 0 || len(substr) > len(s) {
		return false, 0
	}
	end := len(substr) + startByteIdx
	for offset := 1; offset < len(substr); offset++ {
		if startByteIdx-offset < 0 || end-offset > len(s) {
			continue
		}
		if s[startByteIdx-offset:end-offset] == substr {
			return true, end - offset
		}
	}
	return false, 0
}

// overflowsRight checks if a substring overflows a string on the right if the string were to end at endByteIdx exclusive.
// assumes s has no ansi codes.
// It performs a case-sensitive comparison and returns two values:
//   - A boolean indicating whether there is overflow
//   - An integer indicating the starting string startByteIdx of the overflow (0 if none)
//
// Examples:
//
//	                    01234567890
//		overflowsRight("my str here", 3, "y str") returns (true, 1)
//		overflowsRight("my str here", 3, "y strong") returns (false, 0)
//		overflowsRight("my str here", 6, "tr here") returns (true, 4)
func overflowsRight(s string, endByteIdx int, substr string) (bool, int) {
	if len(s) == 0 || len(substr) == 0 || len(substr) > len(s) {
		return false, 0
	}

	leftmostIdx := endByteIdx - len(substr) + 1
	for offset := 0; offset < len(substr); offset++ {
		startIdx := leftmostIdx + offset
		if startIdx < 0 || startIdx+len(substr) > len(s) {
			continue
		}
		sl := s[startIdx : startIdx+len(substr)]
		if sl == substr {
			return true, leftmostIdx + offset
		}
	}
	return false, 0
}

func replaceStartWithContinuation(result string, continuationRunes []rune) string {
	if len(result) == 0 {
		return result
	}

	resultRunes := []rune(result)
	totalContinuationRunes := len(continuationRunes)
	continuationRunesPlaced := 0
	resultRunesReplaced := 0

	for {
		if continuationRunesPlaced >= totalContinuationRunes {
			return string(resultRunes)
		}

		var widthToReplace int
		resultRuneToReplaceIdx := continuationRunesPlaced
		for {
			if resultRuneToReplaceIdx >= len(resultRunes) {
				return string(resultRunes)
			}

			widthToReplace = runewidth.RuneWidth(resultRunes[resultRuneToReplaceIdx])
			if widthToReplace > totalContinuationRunes-continuationRunesPlaced {
				return string(resultRunes)
			}

			if widthToReplace > 0 {
				// remove any following zero-width runes
				for resultRuneToReplaceIdx+1 < len(resultRunes) && runewidth.RuneWidth(resultRunes[resultRuneToReplaceIdx+1]) == 0 {
					resultRunes = replaceRuneWithRunes(resultRunes, resultRuneToReplaceIdx+1, []rune{})
				}
				break
			} else {
				// this can occur when two runes combine into the width of 1
				// e.g. e\u0301 (i.e. é) is 2 runes, the second of which has zero width so should not be replaced
				resultRunes = replaceRuneWithRunes(resultRunes, resultRuneToReplaceIdx, []rune{})
			}
		}

		// get a slice of continuation runes that will replace the result rune, e.g. ".." for double-width unicode char
		var replaceWith []rune
		for {
			if widthToReplace <= 0 {
				break
			}

			nextContinuationRuneIdx := continuationRunesPlaced
			if nextContinuationRuneIdx >= len(continuationRunes) {
				break
			}

			nextContinuationRune := continuationRunes[nextContinuationRuneIdx]
			replaceWith = append(replaceWith, nextContinuationRune)
			widthToReplace -= 1 // assumes continuation runes are of width 1
			continuationRunesPlaced += 1
		}

		resultRunes = replaceRuneWithRunes(resultRunes, resultRuneToReplaceIdx, replaceWith)
		resultRunesReplaced += 1
	}
}

func replaceEndWithContinuation(s string, continuationRunes []rune) string {
	if len(s) == 0 {
		return s
	}

	resultRunes := []rune(s)
	originalResultRunesLen := len(resultRunes)
	totalContinuationRunes := len(continuationRunes)
	continuationRunesPlaced := 0
	resultRunesReplaced := 0
	for {
		if continuationRunesPlaced >= totalContinuationRunes {
			return string(resultRunes)
		}

		var widthToReplace int
		var resultRuneToReplaceIdx int
		for {
			resultRuneToReplaceIdx = originalResultRunesLen - 1 - resultRunesReplaced
			if resultRuneToReplaceIdx < 0 {
				return string(resultRunes)
			}

			widthToReplace = runewidth.RuneWidth(resultRunes[resultRuneToReplaceIdx])
			if widthToReplace > totalContinuationRunes-continuationRunesPlaced {
				return string(resultRunes)
			}

			if widthToReplace > 0 {
				break
			} else {
				// this can occur when two runes combine into the width of 1
				// e.g. e\u0301 (i.e. é) is 2 runes, the second of which has zero width so should not be replaced
				resultRunes = replaceRuneWithRunes(resultRunes, resultRuneToReplaceIdx, []rune{})
				resultRunesReplaced++
			}
		}

		// get a slice of continuation runes that will replace the result rune, e.g. ".." for double-width unicode char
		var replaceWith []rune
		for {
			if widthToReplace <= 0 {
				break
			}
			nextContinuationRuneIdx := len(continuationRunes) - 1 - continuationRunesPlaced
			if nextContinuationRuneIdx < 0 {
				break
			}
			nextContinuationRune := continuationRunes[nextContinuationRuneIdx]
			replaceWith = append([]rune{nextContinuationRune}, replaceWith...)
			widthToReplace -= 1 // assumes continuation runes are of width 1
			continuationRunesPlaced += 1
		}

		resultRunes = replaceRuneWithRunes(resultRunes, resultRuneToReplaceIdx, replaceWith)
		resultRunesReplaced += 1
	}
}

func replaceRuneWithRunes(rs []rune, idxToReplace int, replaceWith []rune) []rune {
	result := make([]rune, len(rs)+len(replaceWith)-1)
	copy(result, rs[:idxToReplace])
	copy(result[idxToReplace:], replaceWith)
	copy(result[idxToReplace+len(replaceWith):], rs[idxToReplace+1:])
	return result
}
