package linebuffer

import (
	"fmt"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/mattn/go-runewidth"
	"github.com/robinovitch61/kl/internal/constants"
	"github.com/robinovitch61/kl/internal/dev"
	"strings"
	"unicode/utf8"
)

// LineBuffer provides functionality to get sequential strings of a specified terminal width, accounting
// for the ansi escape codes styling the line.
type LineBuffer struct {
	line                string         // line with ansi codes. utf-8 bytes
	width               int            // width in terminal columns (not bytes or runes)
	continuation        string         // indicator for line continuation, e.g. "..."
	toHighlight         string         // string to highlight using highlightStyle
	highlightStyle      lipgloss.Style // style for toHighlight
	leftRuneIdx         int            // left plaintext rune idx to start next PopLeft result from
	lineRunes           []rune         // runes of line
	runeIdxToByteOffset []int          // idx of lineRunes to byte offset. len(runeIdxToByteOffset) == len(lineRunes)
	plainText           string         // line without ansi codes. utf-8 bytes
	plainTextRunes      []rune         // runes of plainText. len(plainTextRunes) == len(plainTextWidths)
	plainTextWidths     []int          // terminal column widths of plainText. len(plainTextWidths) == len(plainTextRunes)
	plainTextCumWidth   []int
	continuationRunes   []rune  // runes of continuation
	continuationWidths  []int   // terminal column widths of continuation
	ansiCodeIndexes     [][]int // slice of startByte, endByte indexes of ansi codes in the line
}

func New(line string, width int, continuation string, toHighlight string, highlightStyle lipgloss.Style) LineBuffer {
	lb := LineBuffer{
		line:           line,
		width:          width,
		continuation:   continuation,
		toHighlight:    stripAnsi(toHighlight),
		highlightStyle: highlightStyle,
		leftRuneIdx:    0,
	}

	if len(constants.AnsiRegex.FindAllStringIndex(continuation, -1)) > 0 {
		panic("continuation string cannot contain ansi codes")
	}

	lb.ansiCodeIndexes = constants.AnsiRegex.FindAllStringIndex(line, -1)
	lb.plainText = stripAnsi(line)

	lb.lineRunes = []rune(lb.line)
	lb.runeIdxToByteOffset = initByteOffsets(lb.lineRunes)

	lb.plainTextRunes = []rune(lb.plainText)

	lb.plainTextWidths = make([]int, len(lb.plainTextRunes))
	lb.plainTextCumWidth = make([]int, len(lb.plainTextRunes))
	for i := range lb.plainTextRunes {
		runeWidth := runewidth.RuneWidth(lb.plainTextRunes[i])
		lb.plainTextWidths[i] = runeWidth
		if i == 0 {
			lb.plainTextCumWidth[i] = runeWidth
		} else {
			lb.plainTextCumWidth[i] = lb.plainTextCumWidth[i-1] + runeWidth
		}
	}

	lb.continuationRunes = []rune(lb.continuation)
	lb.continuationWidths = make([]int, len(lb.continuationRunes))
	for i := range lb.continuationRunes {
		runeWidth := runewidth.RuneWidth(lb.continuationRunes[i])
		if runeWidth != 1 {
			panic(fmt.Sprintf("width != 1 rune '%v' not valid in continuation", lb.continuationRunes[i]))
		}
		lb.continuationWidths[i] = runeWidth
	}
	return lb
}

// PopLeft returns a string of the buffer's width from its current left offset, scrolling the left offset to the right
func (l *LineBuffer) PopLeft() string {
	if l.leftRuneIdx >= len(l.plainTextRunes) || l.width == 0 {
		return ""
	}

	var result strings.Builder
	remainingWidth := l.width
	startRuneIdx := l.leftRuneIdx

	runesWritten := 0
	for ; remainingWidth > 0 && l.leftRuneIdx < len(l.plainTextRunes); l.leftRuneIdx++ {
		// get either a rune from the continuation or the line
		r := l.plainTextRunes[l.leftRuneIdx]
		runeWidth := l.plainTextWidths[l.leftRuneIdx]
		if runeWidth > remainingWidth {
			break
		}

		result.WriteRune(r)
		runesWritten++
		remainingWidth -= runeWidth
	}

	// if more runes to the left of the result, replace start runes with continuation indicator, respecting width
	if startRuneIdx > 0 {
		result = l.replaceStartRunesWithContinuation(result)
	}
	// if more runes to the right, replace final runes in result with continuation indicator, respecting width
	if l.leftRuneIdx < len(l.plainTextRunes) {
		result = l.replaceEndRunesWithContinuation(result)
	}

	res := result.String()
	if len(l.ansiCodeIndexes) > 0 {
		res = reapplyANSI(l.line, res, l.runeIdxToByteOffset[startRuneIdx], l.ansiCodeIndexes)
	}

	res = applyToHighlight(l.line, res, l.runeIdxToByteOffset[startRuneIdx], l.ansiCodeIndexes, l.toHighlight, l.highlightStyle)

	return res
}

func (l LineBuffer) replaceStartRunesWithContinuation(result strings.Builder) strings.Builder {
	if result.Len() == 0 {
		return result
	}

	var res strings.Builder
	resultRunes := []rune(result.String())
	totalContinuationRunes := len(l.continuationRunes)
	continuationRunesPlaced := 0
	resultRunesReplaced := 0

	for {
		if continuationRunesPlaced >= totalContinuationRunes {
			res.WriteString(string(resultRunes))
			return res
		}

		resultRuneToReplaceIdx := resultRunesReplaced
		if resultRuneToReplaceIdx >= len(resultRunes) {
			res.WriteString(string(resultRunes))
			return res
		}

		widthToReplace := runewidth.RuneWidth(resultRunes[resultRuneToReplaceIdx])

		// get a slice of continuation runes that will replace the result rune, e.g. ".." for double-width unicode char
		var continuationRunes []rune
		for {
			if widthToReplace <= 0 {
				break
			}

			nextContinuationRuneIdx := continuationRunesPlaced
			if nextContinuationRuneIdx >= len(l.continuationRunes) {
				break
			}

			nextContinuationRune := l.continuationRunes[nextContinuationRuneIdx]
			continuationRunes = append(continuationRunes, nextContinuationRune)
			widthToReplace -= 1 // assumes continuation runes are of width 1
			continuationRunesPlaced += 1
		}

		var leftResult []rune
		if resultRuneToReplaceIdx > 0 {
			leftResult = resultRunes[:resultRuneToReplaceIdx]
		}
		rightResult := resultRunes[resultRuneToReplaceIdx+1:]
		resultRunes = append(append(leftResult, continuationRunes...), rightResult...)
		resultRunesReplaced += 1
	}
}

func (l LineBuffer) replaceEndRunesWithContinuation(result strings.Builder) strings.Builder {
	if result.Len() == 0 {
		return result
	}

	var res strings.Builder

	resultRunes := []rune(result.String())
	totalContinuationRunes := len(l.continuationRunes)
	continuationRunesPlaced := 0
	resultRunesReplaced := 0
	for {
		if continuationRunesPlaced >= totalContinuationRunes {
			res.WriteString(string(resultRunes))
			return res
		}

		resultRuneToReplaceIdx := len(resultRunes) - 1 - resultRunesReplaced
		if resultRuneToReplaceIdx < 0 {
			res.WriteString(string(resultRunes))
			return res
		}
		widthToReplace := runewidth.RuneWidth(resultRunes[resultRuneToReplaceIdx])

		// get a slice of continuation runes that will replace the result rune, e.g. ".." for double-width unicode char
		var continuationRunes []rune
		for {
			if widthToReplace <= 0 {
				break
			}
			nextContinuationRuneIdx := len(l.continuationRunes) - 1 - continuationRunesPlaced
			if nextContinuationRuneIdx < 0 {
				break
			}
			nextContinuationRune := l.continuationRunes[nextContinuationRuneIdx]
			continuationRunes = append([]rune{nextContinuationRune}, continuationRunes...)
			widthToReplace -= 1 // assumes continuation runes are of width 1
			continuationRunesPlaced += 1
		}

		leftResult := append(resultRunes[:resultRuneToReplaceIdx], continuationRunes...)
		var rightResult []rune
		if resultRuneToReplaceIdx+1 < len(resultRunes) {
			rightResult = resultRunes[resultRuneToReplaceIdx+1:]
		}
		resultRunes = append(leftResult, rightResult...)
		resultRunesReplaced += 1
	}
}

func reapplyANSI(original, truncated string, truncByteOffset int, ansiCodeIndexes [][]int) string {
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

// applyToHighlight styles the toHighlight string in the truncated input in the highlightStyle without disrupting
// other styled portions of the truncated string and including toHighlight matches that overflow the truncated
// string from the left or the right
func applyToHighlight(
	original string, // original string from which truncated was created
	truncated string, // truncated string with ansi sequences from the original already reapplied
	truncByteOffset int, // the byte offset in the original string from which truncated begins
	ansiCodeIndexes [][]int, // slice of startByte, endByte indexes of ansi codes in the line
	toHighlight string, // string toHighlight in truncated - does not contain any ansi sequences
	highlightStyle lipgloss.Style, // style with which to highlight toHighlight
) string {
	// If nothing to highlight or truncated string is empty, return as is
	if toHighlight == "" || truncated == "" {
		return truncated
	}

	// First, find all matches of toHighlight in the original string
	// This helps us handle cases where matches might be partially visible in truncated
	type match struct {
		start, end int
	}
	var matches []match

	// Get matches from original string to handle partial matches at truncation boundaries
	lastIndex := 0
	for {
		index := strings.Index(original[lastIndex:], toHighlight)
		if index == -1 {
			break
		}
		start := lastIndex + index
		end := start + len(toHighlight)
		matches = append(matches, match{start, end})
		lastIndex = start + 1
	}

	// If no matches found, return original truncated string
	if len(matches) == 0 {
		return truncated
	}

	// Calculate the end offset of truncated in original
	truncEndOffset := truncByteOffset + len([]byte(stripAnsi(truncated)))

	// Filter matches to only those that overlap with truncated section
	var relevantMatches []match
	for _, m := range matches {
		// Check if match overlaps with truncated section
		if m.start < truncEndOffset && m.end > truncByteOffset {
			relevantMatches = append(relevantMatches, m)
		}
	}

	// If no relevant matches, return original truncated string
	if len(relevantMatches) == 0 {
		return truncated
	}

	// Build result by applying highlight style while preserving existing ANSI codes
	var result strings.Builder
	currentPos := 0
	truncatedStripped := stripAnsi(truncated)

	for _, m := range relevantMatches {
		// Convert match positions from original to truncated string coordinates
		relativeStart := m.start - truncByteOffset
		relativeEnd := m.end - truncByteOffset

		// Adjust positions to fit within truncated bounds
		if relativeStart < 0 {
			relativeStart = 0
		}
		if relativeEnd > len(truncatedStripped) {
			relativeEnd = len(truncatedStripped)
		}

		// Add text before match with original styling
		beforeMatch := getTextWithAnsi(truncated, currentPos, relativeStart, ansiCodeIndexes)
		result.WriteString(beforeMatch)

		// Add highlighted match text while preserving internal ANSI codes
		matchText := getTextWithAnsi(truncated, relativeStart, relativeEnd, ansiCodeIndexes)
		result.WriteString(highlightStyle.Render(matchText))

		currentPos = relativeEnd
	}

	// Add remaining text after last match
	if currentPos < len(truncatedStripped) {
		remaining := getTextWithAnsi(truncated, currentPos, len(truncatedStripped), ansiCodeIndexes)
		result.WriteString(remaining)
	}

	return result.String()
}

// stripAnsi removes all ANSI escape sequences from the input string
func stripAnsi(input string) string {
	return constants.AnsiRegex.ReplaceAllString(input, "")
}

// getTextWithAnsi extracts text between start and end positions while preserving ANSI codes
func getTextWithAnsi(input string, start, end int, ansiCodeIndexes [][]int) string {
	var result strings.Builder

	// Add any ANSI codes that are active at the start position
	for _, codeRange := range ansiCodeIndexes {
		if codeRange[0] <= start && codeRange[1] > start {
			result.WriteString(input[codeRange[0]:codeRange[1]])
		}
	}

	// Add the actual text content
	stripped := stripAnsi(input)
	if start < len(stripped) {
		endPos := end
		if endPos > len(stripped) {
			endPos = len(stripped)
		}
		result.WriteString(stripped[start:endPos])
	}

	return result.String()
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

func initByteOffsets(runes []rune) []int {
	offsets := make([]int, len(runes)+1)
	currentOffset := 0
	for i, r := range runes {
		offsets[i] = currentOffset
		runeLen := utf8.RuneLen(r)
		if runeLen == -1 {
			// invalid utf-8 value, assume 1 byte
			dev.Debug(fmt.Sprintf("invalid utf-8 value: %v", r))
			runeLen = 1
		}
		currentOffset += runeLen
	}
	offsets[len(runes)] = currentOffset
	return offsets
}
