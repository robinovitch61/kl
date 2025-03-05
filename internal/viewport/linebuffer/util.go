package linebuffer

import (
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/mattn/go-runewidth"
	"regexp"
	"strings"
	"unicode/utf8"
)

var emptySequenceRegex = regexp.MustCompile("\x1b\\[[0-9;]+m\x1b\\[m")

var (
	red     = lipgloss.Color("#FF0000")
	blue    = lipgloss.Color("#0000FF")
	green   = lipgloss.Color("#00FF00")
	redFg   = lipgloss.NewStyle().Foreground(red)
	redBg   = lipgloss.NewStyle().Background(red)
	blueBg  = lipgloss.NewStyle().Background(blue)
	greenBg = lipgloss.NewStyle().Background(green)
)

// reapplyAnsi reconstructs ANSI escape sequences in a truncated string based on their positions in the original.
// It ensures that any active text formatting (colors, styles) from the original string is correctly maintained
// in the truncated output, and adds proper reset codes where needed.
//
// Parameters:
//   - original: the source string containing ANSI escape sequences
//   - truncated: the truncated version of the string, without ANSI sequences
//   - truncByteOffset: byte offset in the original string where truncation started
//   - ansiCodeIndexes: pairs of start/end byte positions of ANSI codes in the original string
//
// Returns a string with ANSI escape sequences reapplied at appropriate positions,
// maintaining the original text formatting while preserving proper UTF-8 encoding.
func reapplyAnsi(original, truncated string, truncByteOffset int, ansiCodeIndexes [][]uint32) string {
	var result strings.Builder
	result.Grow(len(truncated))
	var lenAnsiAdded int
	isReset := true

	for i := 0; i < len(truncated); {
		// collect all ansi codes that should be applied immediately before the current runes
		var ansisToAdd []string
		for len(ansiCodeIndexes) > 0 {
			candidateAnsi := ansiCodeIndexes[0]
			codeStart, codeEnd := int(candidateAnsi[0]), int(candidateAnsi[1])
			originalByteIdx := truncByteOffset + i + lenAnsiAdded
			if codeStart <= originalByteIdx {
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
			result.WriteString(ansi)
		}

		// add the bytes of the current rune
		_, size := utf8.DecodeRuneInString(truncated[i:])
		result.WriteString(truncated[i : i+size])
		i += size
	}

	if !isReset {
		result.WriteString("\x1b[m")
	}
	return result.String()
}

// getNonAnsiBytes extracts a substring of specified length from the input string, excluding ANSI escape sequences.
// It reads from the given start position until it has collected the requested number of non-ANSI bytes.
//
// Parameters:
//   - s: The input string that may contain ANSI escape sequences
//   - startIdx: The byte position in the input to start reading from
//   - numBytes: The number of non-ANSI bytes to collect
//
// Returns a string containing bytesToExtract bytes of the input with ANSI sequences removed. If the input text ends
// before collecting bytesToExtract bytes, returns all available non-ANSI bytes.
func getNonAnsiBytes(s string, startIdx, numBytes int) string {
	var result strings.Builder
	currentPos := startIdx
	bytesCollected := 0
	for currentPos < len(s) && bytesCollected < numBytes {
		if strings.HasPrefix(s[currentPos:], "\x1b[") {
			escEnd := currentPos + strings.Index(s[currentPos:], "m") + 1
			currentPos = escEnd
			continue
		}
		result.WriteByte(s[currentPos])
		bytesCollected++
		currentPos++
	}
	return result.String()
}

// highlightLine highlights a string in a line that potentially has ansi codes in it without disrupting them
// start and end are the byte offsets for which highlighting is considered in the line, not counting ansi codes
func highlightLine(line, highlight string, highlightStyle lipgloss.Style, start, end int) string {
	if line == "" || highlight == "" {
		return line
	}

	renderedHighlight := highlightStyle.Render(highlight)
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
		if !inAnsi && nonAnsiBytes >= start && nonAnsiBytes < end {
			textToCheck := getNonAnsiBytes(line, i, len(highlight))
			if textToCheck == highlight {
				// reset current styles, if any
				if len(activeStyles) > 0 {
					result.WriteString("\x1b[m")
				}
				// apply highlight
				result.WriteString(renderedHighlight)
				// restore previous styles, if any
				if len(activeStyles) > 0 {
					for j := range activeStyles {
						result.WriteString(activeStyles[j])
					}
				}

				// skip to end of matched text
				count := 0
				for count < len(highlight) {
					if strings.HasPrefix(line[i:], "\x1b[") {
						escEnd := i + strings.Index(line[i:], "m") + 1
						result.WriteString(line[i:escEnd])
						i = escEnd
						continue
					}
					i++
					count++
					nonAnsiBytes++
				}
				continue
			}
		}
		result.WriteByte(line[i])
		nonAnsiBytes++
		i++
	}
	return removeEmptyAnsiSequences(result.String())
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
	ranges := findAnsiByteRanges(input)
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
	if len(s) == 0 || len(continuationRunes) == 0 {
		return s
	}

	var sb strings.Builder
	ansiCodeIndexes := findAnsiRuneRanges(s)
	runes := []rune(s)

	for runeIdx := 0; runeIdx < len(runes); {
		if len(ansiCodeIndexes) > 0 {
			codeStart, codeEnd := int(ansiCodeIndexes[0][0]), int(ansiCodeIndexes[0][1])
			if runeIdx == codeStart {
				for j := codeStart; j < codeEnd; j++ {
					sb.WriteRune(runes[j])
				}
				// skip ansi
				runeIdx = codeEnd
				ansiCodeIndexes = ansiCodeIndexes[1:]
				continue
			}
		}
		if len(continuationRunes) > 0 {
			rWidth := runewidth.RuneWidth(runes[runeIdx])

			// if rune is wider than remaining continuation width, cut off the continuation
			remainingContinuationWidth := 0
			for _, cr := range continuationRunes {
				remainingContinuationWidth += runewidth.RuneWidth(cr)
			}
			if rWidth > remainingContinuationWidth {
				sb.WriteRune(runes[runeIdx])
				continuationRunes = nil
			}

			// replace current rune with continuation runes
			for rWidth > 0 && len(continuationRunes) > 0 {
				currContinuationRune := continuationRunes[0]
				sb.WriteRune(currContinuationRune)
				continuationRunes = continuationRunes[1:]
				rWidth -= runewidth.RuneWidth(currContinuationRune)
			}

			// skip subsequent zero-width runes that are not ansi sequences
			nextIdx := runeIdx + 1
			for nextIdx < len(runes) {
				nextRWidth := runewidth.RuneWidth(runes[nextIdx])
				if nextRWidth == 0 && nextIdx < len(runes) && !runesHaveAnsiPrefix(runes[nextIdx:]) {
					runeIdx += 1
					nextIdx = runeIdx + 1
				} else {
					break
				}
			}
		} else {
			sb.WriteRune(runes[runeIdx])
		}
		runeIdx += 1
	}

	return sb.String()
}

func replaceEndWithContinuation(s string, continuationRunes []rune) string {
	if len(s) == 0 || len(continuationRunes) == 0 {
		return s
	}

	var result string
	ansiCodeIndexes := findAnsiRuneRanges(s)
	runes := []rune(s)

	for runeIdx := len(runes) - 1; runeIdx >= 0; {
		if len(ansiCodeIndexes) > 0 {
			lastAnsiCodeIndexes := ansiCodeIndexes[len(ansiCodeIndexes)-1]
			codeStart, codeEnd := int(lastAnsiCodeIndexes[0]), int(lastAnsiCodeIndexes[1])
			if runeIdx == codeEnd-1 {
				for j := codeEnd - 1; j >= codeStart; j-- {
					result = string(runes[j]) + result
				}
				// skip ansi
				runeIdx = codeStart - 1
				ansiCodeIndexes = ansiCodeIndexes[:len(ansiCodeIndexes)-1]
				continue
			}
		}
		if len(continuationRunes) > 0 {
			rWidth := runewidth.RuneWidth(runes[runeIdx])

			// if rune is wider than remaining continuation width, cut off the continuation
			remainingContinuationWidth := 0
			for _, cr := range continuationRunes {
				remainingContinuationWidth += runewidth.RuneWidth(cr)
			}
			if rWidth > remainingContinuationWidth {
				result = string(runes[runeIdx]) + result
				continuationRunes = nil
			}

			// replace current rune with continuation runes
			for rWidth > 0 && len(continuationRunes) > 0 {
				currContinuationRune := continuationRunes[len(continuationRunes)-1]
				result = string(currContinuationRune) + result
				continuationRunes = continuationRunes[:len(continuationRunes)-1]
				rWidth -= runewidth.RuneWidth(currContinuationRune)
			}
		} else {
			result = string(runes[runeIdx]) + result
		}
		runeIdx -= 1
	}

	return result
}

func runesHaveAnsiPrefix(runes []rune) bool {
	return len(runes) >= 2 && runes[0] == '\x1b' && runes[1] == '['
}

func findAnsiByteRanges(s string) [][]uint32 {
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

func findAnsiRuneRanges(s string) [][]uint32 {
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
	runes := []rune(s)
	for i := 0; i < len(runes); {
		if i+1 < len(runes) && runes[i] == '\x1b' && runes[i+1] == '[' {
			start := i
			i += 2 // skip \x1b[

			// find the 'm' that ends this sequence
			for i < len(runes) && runes[i] != 'm' {
				i++
			}

			if i < len(runes) && runes[i] == 'm' {
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

// getBytesLeftOfWidth returns nBytes of content to the left of startBufferIdx while excluding ANSI codes
func getBytesLeftOfWidth(nBytes int, buffers []LineBuffer, startBufferIdx int, widthToLeft int) string {
	if nBytes < 0 {
		panic("nBytes must be greater than 0")
	}
	if nBytes == 0 || len(buffers) == 0 || startBufferIdx >= len(buffers) {
		return ""
	}

	// first try to get bytes from the current buffer
	var result string
	currentBuffer := buffers[startBufferIdx]
	runeIdx := currentBuffer.findRuneIndexWithWidthToLeft(widthToLeft)
	if runeIdx > 0 {
		var startByteOffset uint32
		if runeIdx >= currentBuffer.numNoAnsiRunes {
			startByteOffset = uint32(len(currentBuffer.lineNoAnsi))
		} else {
			startByteOffset = currentBuffer.getByteOffsetAtRuneIdx(runeIdx)
		}
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

// getBytesRightOfWidth returns nBytes of content to the right of endBufferIdx while excluding ANSI codes
func getBytesRightOfWidth(nBytes int, buffers []LineBuffer, endBufferIdx int, widthToRight int) string {
	if nBytes < 0 {
		panic("nBytes must be greater than 0")
	}
	if nBytes == 0 || len(buffers) == 0 || endBufferIdx >= len(buffers) {
		return ""
	}

	// first try to get bytes from the current buffer
	var result string
	currentBuffer := buffers[endBufferIdx]
	if widthToRight > 0 {
		currentBufferWidth := currentBuffer.Width()
		widthToLeft := currentBufferWidth - widthToRight
		startRuneIdx := currentBuffer.findRuneIndexWithWidthToLeft(widthToLeft)
		if startRuneIdx < currentBuffer.numNoAnsiRunes {
			startByteOffset := currentBuffer.getByteOffsetAtRuneIdx(startRuneIdx)
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

// getWrappedLines is logic shared by WrappedLines in single and multi LineBuffers
// it is well-tested as part of the tests of those methods
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

func removeEmptyAnsiSequences(s string) string {
	return emptySequenceRegex.ReplaceAllString(s, "")
}
