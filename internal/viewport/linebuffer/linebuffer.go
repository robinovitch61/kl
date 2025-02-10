package linebuffer

import (
	"fmt"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/mattn/go-runewidth"
	"github.com/robinovitch61/kl/internal/constants"
	"strings"
	"unicode/utf8"
)

// LineBuffer provides functionality to get sequential strings of a specified terminal cell width, accounting
// for the ansi escape codes styling the line.
type LineBuffer struct {
	content             string     // underlying string with ansi codes. utf-8 bytes
	lineNoAnsi          string     // line without ansi codes. utf-8 bytes
	lineNoAnsiRunes     []rune     // runes of lineNoAnsi. len(lineNoAnsiRunes) == len(lineNoAnsiWidths)
	runeIdxToByteOffset []uint32   // idx of lineNoAnsiRunes to byte offset. len(runeIdxToByteOffset) == len(lineNoAnsiRunes)
	lineNoAnsiWidths    []uint8    // terminal cell widths of lineNoAnsi. len(lineNoAnsiWidths) == len(lineNoAnsiRunes)
	lineNoAnsiCumWidths []uint32   // cumulative lineNoAnsiWidths
	ansiCodeIndexes     [][]uint32 // slice of startByte, endByte indexes of ansi codes in the line
}

// type assertion that LineBuffer implements LineBufferer
var _ LineBufferer = LineBuffer{}

// type assertion that *LineBuffer implements LineBufferer
var _ LineBufferer = (*LineBuffer)(nil)

func New(line string) *LineBuffer {
	lb := LineBuffer{
		content: line,
	}

	lb.ansiCodeIndexes = findAnsiRanges(line)

	if len(lb.ansiCodeIndexes) > 0 {
		totalLen := len(line)
		for _, r := range lb.ansiCodeIndexes {
			totalLen -= int(r[1] - r[0])
		}

		buf := make([]byte, 0, totalLen)
		lastPos := 0
		for _, r := range lb.ansiCodeIndexes {
			buf = append(buf, line[lastPos:int(r[0])]...)
			lastPos = int(r[1])
		}
		buf = append(buf, line[lastPos:]...)
		lb.lineNoAnsi = string(buf)
	} else {
		lb.lineNoAnsi = line
	}

	n := utf8.RuneCountInString(lb.lineNoAnsi)

	// single allocation for all integer slices
	combined := make([]uint32, n+1+n)
	lb.runeIdxToByteOffset = combined[:n+1]
	lb.lineNoAnsiCumWidths = combined[n+1:]

	lb.lineNoAnsiWidths = make([]uint8, n)
	lb.lineNoAnsiRunes = make([]rune, n)

	var currentOffset uint32
	var cumWidth uint32
	i := 0
	for _, r := range lb.lineNoAnsi {
		lb.runeIdxToByteOffset[i] = currentOffset
		currentOffset += uint32(utf8.RuneLen(r))
		lb.lineNoAnsiRunes[i] = r
		width := uint8(runewidth.RuneWidth(r))
		lb.lineNoAnsiWidths[i] = width
		cumWidth += uint32(width)
		lb.lineNoAnsiCumWidths[i] = cumWidth
		i++
	}
	lb.runeIdxToByteOffset[n] = currentOffset
	return &lb
}

func (l LineBuffer) Width() int {
	if len(l.lineNoAnsiCumWidths) > 0 {
		return int(l.lineNoAnsiCumWidths[len(l.lineNoAnsiCumWidths)-1])
	}
	return 0
}

// TODO LEO: don't use this for e.g. search, instead inject filter into a Matches() bool method here
func (l LineBuffer) Content() string {
	return l.content
}

// Take returns a string of the buffer's width from its current left offset
func (l LineBuffer) Take(
	startWidth, takeWidth int,
	continuation, toHighlight string,
	highlightStyle lipgloss.Style,
) (string, int) {
	if startWidth < 0 {
		startWidth = 0
	}

	startRuneIdx := getLeftRuneIdx(startWidth, l.lineNoAnsiCumWidths)

	if startRuneIdx >= len(l.lineNoAnsiRunes) || takeWidth == 0 {
		return "", 0
	}

	var result strings.Builder
	remainingWidth := takeWidth
	leftRuneIdx := startRuneIdx
	startByteOffset := l.runeIdxToByteOffset[startRuneIdx]

	runesWritten := 0
	for ; remainingWidth > 0 && leftRuneIdx < len(l.lineNoAnsiRunes); leftRuneIdx++ {
		r := l.lineNoAnsiRunes[leftRuneIdx]
		runeWidth := l.lineNoAnsiWidths[leftRuneIdx]
		if int(runeWidth) > remainingWidth {
			break
		}

		result.WriteRune(r)
		runesWritten++
		remainingWidth -= int(runeWidth)
	}

	res := result.String()

	// apply left/right line continuation indicators
	if len(continuation) > 0 && (startRuneIdx > 0 || leftRuneIdx < len(l.lineNoAnsiRunes)) {
		continuationRunes := []rune(continuation)

		// if more runes to the left of the result, replace start runes with continuation indicator, respecting width
		if startRuneIdx > 0 {
			res = replaceStartWithContinuation(res, continuationRunes)
		}

		// if more runes to the right, replace final runes in result with continuation indicator, respecting width
		if leftRuneIdx < len(l.lineNoAnsiRunes) {
			res = replaceEndWithContinuation(res, continuationRunes)
		}
	}

	// reapply original styling
	if len(l.ansiCodeIndexes) > 0 {
		res = reapplyAnsi(l.content, res, int(startByteOffset), l.ansiCodeIndexes)
	}

	// highlight the desired string
	res = l.highlightString(res, int(startByteOffset), int(l.runeIdxToByteOffset[leftRuneIdx]), toHighlight, highlightStyle)

	// remove empty sequences
	res = constants.EmptySequenceRegex.ReplaceAllString(res, "")

	return res, takeWidth - remainingWidth
}

func (l LineBuffer) WrappedLines(
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
	if l.content == "" {
		return []string{l.content}
	}

	var res []string
	totalLines := l.totalLines(width)

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

func (l LineBuffer) Repr() string {
	return fmt.Sprintf("LB(%q)", l.content)
}

func (l LineBuffer) highlightString(
	s string,
	startByteOffset int,
	endByteOffset int,
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
		if right, startIdx := overflowsRight(l.lineNoAnsi, endByteOffset, toHighlight); right {
			highlightRight := l.lineNoAnsi[startIdx:endByteOffset]
			lenPlainTextRes := len(stripAnsi(s))
			s = highlightLine(s, highlightRight, highlightStyle, lenPlainTextRes-len(highlightRight), lenPlainTextRes)
		}
	}

	return s
}

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

func (l LineBuffer) totalLines(width int) int {
	if width == 0 {
		return 0
	}
	return (int(l.fullWidth()) + width - 1) / width
}

func (l LineBuffer) fullWidth() uint32 {
	if len(l.lineNoAnsiCumWidths) == 0 {
		return 0
	}
	return l.lineNoAnsiCumWidths[len(l.lineNoAnsiCumWidths)-1]
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
