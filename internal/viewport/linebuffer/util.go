package linebuffer

import (
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/mattn/go-runewidth"
	"strings"
	"unicode/utf8"
)

var (
	red     = lipgloss.Color("#FF0000")
	blue    = lipgloss.Color("#0000FF")
	green   = lipgloss.Color("#00FF00")
	redBg   = lipgloss.NewStyle().Background(red)
	blueBg  = lipgloss.NewStyle().Background(blue)
	greenBg = lipgloss.NewStyle().Background(green)
)

func reapplyAnsi(original, truncated string, truncByteOffset int, ansiCodeIndexes [][]uint32) string {
	var result []byte
	var lenAnsiAdded int
	isReset := true
	truncatedBytes := []byte(truncated)

	for i := 0; i < len(truncatedBytes); {
		// collect all ansi codes that should be applied immediately before the current runes
		var ansisToAdd []string
		for len(ansiCodeIndexes) > 0 {
			candidateAnsi := ansiCodeIndexes[0]
			codeStart, codeEnd := int(candidateAnsi[0]), int(candidateAnsi[1])
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

// highlightString applies highlighting to a segment of text while handling cases where the highlight
// might overflow the segment boundaries. It preserves any existing ANSI styling in the segment.
//
// Parameters:
//   - styledSegment: the text segment to highlight, which may contain ANSI codes
//   - toHighlight: the substring to search for and highlight
//   - highlightStyle: the style to apply to matched substrings
//   - plainLine: the complete line without any ANSI codes, used for overflow detection
//   - segmentStart: byte offset where this segment starts in plainLine
//   - segmentEnd: byte offset where this segment ends in plainLine
//
// Returns the segment with highlighting applied, preserving original ANSI codes.
func highlightString(
	styledSegment string,
	toHighlight string,
	highlightStyle lipgloss.Style,
	plainLine string,
	segmentStart int,
	segmentEnd int,
) string {
	if toHighlight != "" && len(highlightStyle.String()) > 0 {
		styledSegment = highlightLine(styledSegment, toHighlight, highlightStyle, 0, len(styledSegment))

		if left, endIdx := overflowsLeft(plainLine, segmentStart, toHighlight); left {
			highlightLeft := plainLine[segmentStart:endIdx]
			styledSegment = highlightLine(styledSegment, highlightLeft, highlightStyle, 0, len(highlightLeft))
		}
		if right, startIdx := overflowsRight(plainLine, segmentEnd, toHighlight); right {
			highlightRight := plainLine[startIdx:segmentEnd]
			lenPlainTextRes := len(stripAnsi(styledSegment))
			styledSegment = highlightLine(styledSegment, highlightRight, highlightStyle, lenPlainTextRes-len(highlightRight), lenPlainTextRes)
		}
	}

	return styledSegment
}

func stripAnsi(input string) string {
	ranges := findAnsiRanges(input)
	if len(ranges) == 0 {
		return input
	}

	totalAnsiLen := 0
	for _, r := range ranges {
		totalAnsiLen += int(r[1] - r[0])
	}

	finalLen := len(input) - totalAnsiLen
	var builder strings.Builder
	builder.Grow(finalLen)

	lastPos := 0
	for _, r := range ranges {
		builder.WriteString(input[lastPos:int(r[0])])
		lastPos = int(r[1])
	}

	builder.WriteString(input[lastPos:])
	return builder.String()
}

func simplifyAnsiCodes(ansis []string) []string {
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

func replaceStartWithContinuation(s string, continuationRunes []rune) string {
	if len(s) == 0 {
		return s
	}

	ansiRanges := findAnsiRanges(s)
	if len(ansiRanges) != 0 {
		plainText := stripAnsi(s)
		replacedPlainText := replaceStartWithContinuation(plainText, continuationRunes)
		return reapplyAnsi(s, replacedPlainText, 0, ansiRanges)
	}

	resultRunes := []rune(s)
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

	ansiRanges := findAnsiRanges(s)
	if len(ansiRanges) != 0 {
		plainText := stripAnsi(s)
		replacedPlainText := replaceEndWithContinuation(plainText, continuationRunes)
		return reapplyAnsi(s, replacedPlainText, 0, ansiRanges)
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

func findAnsiRanges(s string) [][]uint32 {
	// pre-count to allocate exact size
	count := strings.Count(s, "\x1b[")
	if count == 0 {
		return nil
	}

	allRanges := make([]uint32, count*2)
	ranges := make([][]uint32, count)

	for i := 0; i < count; i++ {
		ranges[i] = allRanges[i*2 : i*2+2]
	}

	rangeIdx := 0
	for i := 0; i < len(s); {
		if i+1 < len(s) && s[i] == '\x1b' && s[i+1] == '[' {
			start := i
			i += 2 // skip \x1b[

			// find the 'm' that ends this sequence
			for i < len(s) && s[i] != 'm' {
				i++
			}

			if i < len(s) && s[i] == 'm' {
				allRanges[rangeIdx*2] = uint32(start)
				allRanges[rangeIdx*2+1] = uint32(i + 1)
				rangeIdx++
				i++
				continue
			}
		}
		i++
	}
	return ranges[:rangeIdx]
}

// getLeftRuneIdx does a binary search to find the first index at which vals[index-1] >= w
func getLeftRuneIdx(w int, vals []uint32) int {
	if w == 0 {
		return 0
	}
	if len(vals) == 0 {
		return 0
	}

	left, right := 0, len(vals)-1

	if vals[right] < uint32(w) {
		return len(vals)
	}

	for left < right {
		mid := left + (right-left)/2

		if vals[mid] >= uint32(w) {
			right = mid
		} else {
			left = mid + 1
		}
	}

	return left + 1
}

func getTotalLines(cumWidths []uint32, lineWidth uint32) int {
	if len(cumWidths) == 0 {
		return 0
	}
	if lineWidth == 0 {
		return 0
	}

	fullWidth := cumWidths[len(cumWidths)-1]
	return int((fullWidth + lineWidth - 1) / lineWidth)
}

// TODO LEO: test
// getBytesLeftOfWidth returns nBytes of content to the left of startBufferIdx while excluding ANSI codes
func getBytesLeftOfWidth(nBytes int, buffers []LineBuffer, startBufferIdx int, startWidth int) string {
	if nBytes <= 0 || len(buffers) == 0 || startBufferIdx >= len(buffers) {
		return ""
	}

	// first try to get bytes from the current buffer
	var result string
	currentBuffer := buffers[startBufferIdx]
	leftRuneIdx := getLeftRuneIdx(startWidth, currentBuffer.lineNoAnsiCumWidths)
	if leftRuneIdx > 0 {
		startByteOffset := currentBuffer.runeIdxToByteOffset[leftRuneIdx]
		noAnsiContent := currentBuffer.lineNoAnsi[:startByteOffset]
		if len(noAnsiContent) >= nBytes {
			return noAnsiContent[len(noAnsiContent)-nBytes:]
		}
		result = noAnsiContent
		nBytes -= len(noAnsiContent)
	}

	// if we need more bytes, look in previous buffers
	for i := startBufferIdx - 1; i >= 0 && nBytes > 0; i-- {
		prevBuffer := buffers[i]
		noAnsiContent := prevBuffer.lineNoAnsi
		if len(noAnsiContent) >= nBytes {
			result = noAnsiContent[len(noAnsiContent)-nBytes:] + result
			break
		}
		result = noAnsiContent + result
		nBytes -= len(noAnsiContent)
	}

	return result
}

// TODO LEO: test
// getBytesRightOfWidth returns nBytes of content to the right of endBufferIdx while excluding ANSI codes
func getBytesRightOfWidth(nBytes int, buffers []LineBuffer, endBufferIdx int, remainingWidth int) string {
	if nBytes <= 0 || len(buffers) == 0 || endBufferIdx >= len(buffers) {
		return ""
	}

	// first try to get bytes from the current buffer
	var result string
	currentBuffer := buffers[endBufferIdx]
	if remainingWidth > 0 {
		currentWidth := currentBuffer.Width()
		startWidth := currentWidth - remainingWidth
		if startWidth < 0 {
			startWidth = 0
		}
		leftRuneIdx := getLeftRuneIdx(startWidth, currentBuffer.lineNoAnsiCumWidths)
		if leftRuneIdx < len(currentBuffer.lineNoAnsiRunes) {
			startByteOffset := currentBuffer.runeIdxToByteOffset[leftRuneIdx]
			noAnsiContent := currentBuffer.lineNoAnsi[startByteOffset:]
			if len(noAnsiContent) >= nBytes {
				return noAnsiContent[:nBytes]
			}
			result = noAnsiContent
			nBytes -= len(noAnsiContent)
		}
	}

	// if we need more bytes, look in subsequent buffers
	for i := endBufferIdx + 1; i < len(buffers) && nBytes > 0; i++ {
		nextBuffer := buffers[i]
		noAnsiContent := nextBuffer.lineNoAnsi
		if len(noAnsiContent) >= nBytes {
			result += noAnsiContent[:nBytes]
			break
		}
		result += noAnsiContent
		nBytes -= len(noAnsiContent)
	}

	return result
}

// TODO LEO: test
func getWrappedLines(
	l LineBufferer,
	totalLines int,
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

	var res []string
	startWidth := 0
	if maxLinesEachEnd > 0 && totalLines > maxLinesEachEnd*2 {
		for nLines := 0; nLines < maxLinesEachEnd; nLines++ {
			line, lineWidth := l.Take(startWidth, width, "", toHighlight, toHighlightStyle)
			res = append(res, line)
			startWidth += lineWidth
		}

		startWidth = (totalLines - maxLinesEachEnd) * width
		for nLines := 0; nLines < maxLinesEachEnd; nLines++ {
			line, lineWidth := l.Take(startWidth, width, "", toHighlight, toHighlightStyle)
			res = append(res, line)
			startWidth += lineWidth
		}
	} else {
		for nLines := 0; nLines < totalLines; nLines++ {
			line, lineWidth := l.Take(startWidth, width, "", toHighlight, toHighlightStyle)
			res = append(res, line)
			startWidth += lineWidth
		}
	}

	return res
}
