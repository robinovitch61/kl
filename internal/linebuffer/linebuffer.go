package linebuffer

import (
	"fmt"
	"github.com/mattn/go-runewidth"
	"github.com/robinovitch61/kl/internal/constants"
	"strings"
)

// LineBuffer provides functionality to get sequential strings of a specified terminal width, accounting
// for the ansi escape codes styling the line.
type LineBuffer struct {
	line           string // line with ansi codes. utf-8 bytes
	width          int    // width in terminal columns (not bytes or runes)
	continuation   string // indicator for line continuation, e.g. "..."
	leftRuneIdx    int    // left plaintext rune idx
	rightRuneIdx   int    // right plaintext rune idx
	lineRunes      []rune // runes of line
	plainText      string // line without ansi codes. utf-8 bytes
	plainTextRunes []rune // runes of plainText. len(plainTextRunes) == len(plainTextWidths)
	// TODO LEO: differentiate between CumWidths and Widths
	plainTextWidths    []int   // terminal column widths of plainText. len(plainTextWidths) == len(plainTextRunes)
	continuationRunes  []rune  // runes of continuation
	continuationWidths []int   // terminal column widths of continuation
	ansiCodeIndexes    [][]int // slice of startByte, endByte indexes of ansi codes in the line
}

func New(line string, width int, continuation string) LineBuffer {
	lb := LineBuffer{
		line:         line,
		width:        width,
		continuation: continuation,
		leftRuneIdx:  0,
	}

	if len(constants.AnsiRegex.FindAllStringIndex(continuation, -1)) > 0 {
		panic("continuation string cannot contain ansi codes")
	}

	lb.ansiCodeIndexes = constants.AnsiRegex.FindAllStringIndex(line, -1)
	lb.plainText = constants.AnsiRegex.ReplaceAllString(line, "")

	lb.lineRunes = []rune(lb.line)
	lb.plainTextRunes = []rune(lb.plainText)

	lb.rightRuneIdx = len(lb.plainTextRunes)

	lb.plainTextWidths = make([]int, len(lb.plainTextRunes))
	for i := range lb.plainTextRunes {
		lb.plainTextWidths[i] = runewidth.RuneWidth(lb.plainTextRunes[i])
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
	runesToLeft := l.leftRuneIdx > 0

	runesWritten := 0
	for ; remainingWidth > 0 && l.leftRuneIdx < len(l.plainTextRunes); l.leftRuneIdx++ {
		// get either a rune from the continuation or the line
		r := l.plainTextRunes[l.leftRuneIdx]
		runeWidth := l.plainTextWidths[l.leftRuneIdx]
		if runeWidth > remainingWidth {
			break
		}

		// TODO LEO : this should take runes from continuation up to the rune width its replacing
		// consider putting this at the end when resultRunes are known
		if runesToLeft && len(l.continuationRunes) > 0 && runesWritten < len(l.continuationRunes) {
			r = l.continuationRunes[runesWritten]
			runeWidth = l.continuationWidths[runesWritten]
		}

		result.WriteRune(r)
		runesWritten++
		remainingWidth -= runeWidth
	}

	// if more runes to the right, replace final runes in result with continuation indicator, respecting width
	// assumes all continuation runes are of width 1
	if result.Len() > 0 && l.leftRuneIdx < len(l.plainTextRunes) {
		result = l.replaceRightRunesWithContinuation(result)
	}

	return result.String()
}

// PopRight returns a string of the buffer's width from its current right offset, scrolling the right offset to the left
func (l *LineBuffer) PopRight() string {
	return "TODO"
}

func (l LineBuffer) replaceRightRunesWithContinuation(result strings.Builder) strings.Builder {
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
		widthToReplace := l.plainTextWidths[resultRuneToReplaceIdx]

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

//func (l LineBuffer) Truncate(xOffset, width int) string {
//	if width <= 0 {
//		return ""
//	}
//	if xOffset == 0 && len(l.lineContinuationIndicator) == 0 && !strings.Contains(l.line, "\x1b") {
//		if len(l.lineRunes) <= width {
//			return l.line
//		}
//		return string(l.lineRunes[:width])
//	}
//
//	lenPlainText := len(l.plainTextRunes)
//
//	if lenPlainText == 0 || xOffset >= lenPlainText {
//		return ""
//	}
//
//	indicatorLen := utf8.RuneCountInString(l.lineContinuationIndicator)
//	if width <= indicatorLen && lenPlainText > width {
//		return l.lineContinuationIndicator[:width]
//	}
//
//	var b strings.Builder
//
//	start := xOffset
//	if start < 0 {
//		start = 0
//	}
//	end := xOffset + width
//	if end > lenPlainText {
//		end = lenPlainText
//	}
//	if end <= start {
//		if len(l.ansiCodeIndexes) > 0 {
//			return ""
//		}
//		return ""
//	}
//
//	if start == 0 && end == lenPlainText && len(l.ansiCodeIndexes) == 0 {
//		return l.line
//	}
//
//	visible := string(l.plainTextRunes[start:end])
//
//	if indicatorLen > 0 {
//		if end-start <= indicatorLen && lenPlainText > indicatorLen {
//			return l.lineContinuationIndicator[:min(indicatorLen, end-start)]
//		}
//		visLen := utf8.RuneCountInString(visible)
//		if xOffset > 0 && visLen > indicatorLen {
//			b.WriteString(l.lineContinuationIndicator)
//			b.WriteString(string([]rune(visible)[indicatorLen:]))
//			visible = b.String()
//		} else if xOffset > 0 {
//			visible = l.lineContinuationIndicator
//		}
//		if end < lenPlainText && visLen > indicatorLen {
//			b.Reset()
//			b.WriteString(string([]rune(visible)[:visLen-indicatorLen]))
//			b.WriteString(l.lineContinuationIndicator)
//			visible = b.String()
//		} else if end < lenPlainText {
//			visible = l.lineContinuationIndicator
//		}
//	}
//
//	if len(l.ansiCodeIndexes) > 0 {
//		//println(fmt.Sprintf("reapplyAnsi(%q, %q, %d, %v) = %q", l.line, visible, l.byteOffsets[start], l.ansiCodeIndexes, reapplied))
//		return reapplyANSI(l.line, visible, l.byteOffsets[start], l.ansiCodeIndexes)
//	}
//	return visible
//}
//
//func reapplyANSI(original, truncated string, truncByteOffset int, ansiCodeIndexes [][]int) string {
//	var result []byte
//	var lenAnsiAdded int
//	isReset := true
//	truncatedBytes := []byte(truncated)
//
//	for i := 0; i < len(truncatedBytes); {
//		// collect all ansi codes that should be applied immediately before the current runes
//		var ansisToAdd []string
//		for len(ansiCodeIndexes) > 0 {
//			candidateAnsi := ansiCodeIndexes[0]
//			codeStart, codeEnd := candidateAnsi[0], candidateAnsi[1]
//			originalIdx := truncByteOffset + i + lenAnsiAdded
//			if codeStart <= originalIdx {
//				code := original[codeStart:codeEnd]
//				isReset = code == "\x1b[m"
//				ansisToAdd = append(ansisToAdd, code)
//				lenAnsiAdded += codeEnd - codeStart
//				ansiCodeIndexes = ansiCodeIndexes[1:]
//			} else {
//				break
//			}
//		}
//
//		for _, ansi := range simplifyAnsiCodes(ansisToAdd) {
//			result = append(result, ansi...)
//		}
//
//		// add the bytes of the current rune
//		_, size := utf8.DecodeRune(truncatedBytes[i:])
//		result = append(result, truncatedBytes[i:i+size]...)
//		i += size
//	}
//
//	if !isReset {
//		result = append(result, "\x1b[m"...)
//	}
//
//	return string(result)
//}

//func initByteOffsets(runes []rune) []int {
//	offsets := make([]int, len(runes)+1)
//	currentOffset := 0
//	for i, r := range runes {
//		offsets[i] = currentOffset
//		runeLen := utf8.RuneLen(r)
//		if runeLen == -1 {
//			// invalid utf-8 value, assume 1 byte
//			dev.Debug(fmt.Sprintf("invalid utf-8 value: %v", r))
//			runeLen = 1
//		}
//		currentOffset += runeLen
//	}
//	offsets[len(runes)] = currentOffset
//	return offsets
//}
