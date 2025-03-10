package linebuffer

import (
	"fmt"
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/mattn/go-runewidth"
	"github.com/robinovitch61/kl/internal/filter"
	"strings"
	"unicode/utf8"
)

// LineBuffer provides functionality to get sequential strings of a specified terminal cell width, accounting
// for the ansi escape codes styling the line.
type LineBuffer struct {
	line                 string     // underlying string with ansi codes. utf-8 encoded bytes
	lineNoAnsi           string     // line without ansi codes. utf-8 encoded bytes
	lineNoAnsiRuneWidths []uint8    // packed terminal cell widths, 4 widths per byte (2 bits each)
	ansiCodeIndexes      [][]uint32 // slice of startByte, endByte indexes of ansi codes
	numNoAnsiRunes       int        // number of runes in lineNoAnsi
	totalWidth           int        // total width in terminal cells

	sparsity                        int      // interval for which to store cumulative cell width
	sparseRuneIdxToNoAnsiByteOffset []uint32 // rune idx to byte offset of lineNoAnsi, stored every sparsity runes
	sparseLineNoAnsiCumRuneWidths   []uint32 // cumulative terminal cell width, stored every sparsity runes
}

// type assertion that LineBuffer implements LineBufferer
var _ LineBufferer = LineBuffer{}

// type assertion that *LineBuffer implements LineBufferer
var _ LineBufferer = (*LineBuffer)(nil)

func New(line string) LineBuffer {
	if len(line) <= 0 {
		return LineBuffer{line: line}
	}

	// keep sparsity 1 for short lines
	sparsity := 1
	if len(line) > 1000 {
		sparsity = 10 // tradeoff between memory usage and CPU. 10 seems to be a good balance
	}

	lb := LineBuffer{
		line:     line,
		sparsity: sparsity,
	}

	lb.ansiCodeIndexes = findAnsiByteRanges(line)

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

	numRunes := utf8.RuneCountInString(lb.lineNoAnsi)

	// calculate size needed for sparse cumulative widths
	sparseLen := (numRunes + lb.sparsity - 1) / lb.sparsity
	lb.sparseRuneIdxToNoAnsiByteOffset = make([]uint32, sparseLen)
	lb.sparseLineNoAnsiCumRuneWidths = make([]uint32, sparseLen)

	// calculate size needed for packed rune widths (4 widths per byte)
	packedLen := (numRunes + 3) / 4
	lb.lineNoAnsiRuneWidths = make([]uint8, packedLen)

	var currentOffset uint32
	var cumWidth uint32
	runeIdx := 0
	for byteOffset := 0; byteOffset < len(lb.lineNoAnsi); {
		r, runeNumBytes := utf8.DecodeRuneInString(lb.lineNoAnsi[byteOffset:])
		width := uint8(runewidth.RuneWidth(r))

		// pack 4 widths per byte (2 bits each)
		packedIdx := runeIdx / 4
		bitPos := (runeIdx % 4) * 2
		// clear the 2 bits at the position and set the new width
		lb.lineNoAnsiRuneWidths[packedIdx] &= ^(uint8(3) << bitPos)
		lb.lineNoAnsiRuneWidths[packedIdx] |= width << bitPos

		cumWidth += uint32(width)
		if runeIdx%lb.sparsity == 0 {
			lb.sparseRuneIdxToNoAnsiByteOffset[runeIdx/lb.sparsity] = currentOffset
			lb.sparseLineNoAnsiCumRuneWidths[runeIdx/lb.sparsity] = cumWidth
		}
		if runeIdx == numRunes-1 {
			lb.totalWidth = int(cumWidth)
		}
		currentOffset += uint32(runeNumBytes)
		runeIdx++
		byteOffset += runeNumBytes
	}
	lb.numNoAnsiRunes = runeIdx

	return lb
}

func (l LineBuffer) Width() int {
	if len(l.line) == 0 {
		return 0
	}
	return l.totalWidth
}

func (l LineBuffer) Content() string {
	return l.line
}

// Take returns a string of the buffer's width from its current left offset
func (l LineBuffer) Take(
	widthToLeft, takeWidth int,
	continuation, toHighlight string,
	highlightStyle lipgloss.Style,
) (string, int) {
	if widthToLeft < 0 {
		widthToLeft = 0
	}

	widthToLeft = min(widthToLeft, l.Width())
	startRuneIdx := l.findRuneIndexWithWidthToLeft(widthToLeft)

	if startRuneIdx >= l.numNoAnsiRunes || takeWidth == 0 {
		return "", 0
	}

	var result strings.Builder
	remainingWidth := takeWidth
	leftRuneIdx := startRuneIdx
	startByteOffset := l.getByteOffsetAtRuneIdx(startRuneIdx)

	runesWritten := 0
	for ; remainingWidth > 0 && leftRuneIdx < l.numNoAnsiRunes; leftRuneIdx++ {
		r := l.runeAt(leftRuneIdx)
		runeWidth := l.getRuneWidth(leftRuneIdx)
		if int(runeWidth) > remainingWidth {
			break
		}

		result.WriteRune(r)
		runesWritten++
		remainingWidth -= int(runeWidth)
	}

	// if only zero-width runes were written, return ""
	for i := 0; i < runesWritten; i++ {
		if runewidth.RuneWidth(l.runeAt(startRuneIdx+i)) > 0 {
			break
		}
		if i == runesWritten-1 {
			return "", 0
		}
	}

	// write the subsequent zero-width runes, e.g. the accent on an 'e'
	if result.Len() > 0 {
		for ; leftRuneIdx < l.numNoAnsiRunes; leftRuneIdx++ {
			r := l.runeAt(leftRuneIdx)
			if runewidth.RuneWidth(r) == 0 {
				result.WriteRune(r)
			} else {
				break
			}
		}
	}

	res := result.String()

	// reapply original styling
	if len(l.ansiCodeIndexes) > 0 {
		res = reapplyAnsi(l.line, res, int(startByteOffset), l.ansiCodeIndexes)
	}

	// apply left/right line continuation indicators
	if len(continuation) > 0 && (startRuneIdx > 0 || leftRuneIdx < l.numNoAnsiRunes) {
		continuationRunes := []rune(continuation)

		// if more runes to the left of the result, replace start runes with continuation indicator
		if startRuneIdx > 0 {
			res = replaceStartWithContinuation(res, continuationRunes)
		}

		// if more runes to the right, replace final runes in result with continuation indicator
		if leftRuneIdx < l.numNoAnsiRunes {
			res = replaceEndWithContinuation(res, continuationRunes)
		}
	}

	// highlight the desired string
	var endByteOffset int
	if leftRuneIdx < l.numNoAnsiRunes {
		endByteOffset = int(l.getByteOffsetAtRuneIdx(leftRuneIdx))
	} else {
		endByteOffset = len(l.lineNoAnsi)
	}
	res = highlightString(
		res,
		toHighlight,
		highlightStyle,
		l.lineNoAnsi,
		int(startByteOffset),
		endByteOffset,
	)

	res = removeEmptyAnsiSequences(res)
	return res, takeWidth - remainingWidth
}

func (l LineBuffer) WrappedLines(
	width int,
	maxLinesEachEnd int,
	toHighlight string,
	toHighlightStyle lipgloss.Style,
) []string {
	if width == 0 {
		return []string{}
	}
	// preserve empty lines
	if l.line == "" {
		return []string{l.line}
	}

	lastRuneIdx := l.numNoAnsiRunes - 1
	totalWidth := l.getCumulativeWidthAtRuneIdx(lastRuneIdx)
	totalLines := (int(totalWidth) + width - 1) / width
	return getWrappedLines(
		l,
		totalLines,
		width,
		maxLinesEachEnd,
		toHighlight,
		toHighlightStyle,
	)
}

func (l LineBuffer) Matches(f filter.Model) bool {
	return f.Matches(l.lineNoAnsi)
}

func (l LineBuffer) Repr() string {
	return fmt.Sprintf("LB(%q)", l.line)
}

// runeAt decodes the desired rune from the lineNoAnsi string
// it serves as a memory-saving technique compared to storing all the runes in a slice
func (l LineBuffer) runeAt(runeIdx int) rune {
	if runeIdx < 0 || runeIdx >= l.numNoAnsiRunes {
		return -1
	}
	start := l.getByteOffsetAtRuneIdx(runeIdx)
	var end uint32
	if runeIdx+1 >= l.numNoAnsiRunes {
		end = uint32(len(l.lineNoAnsi))
	} else {
		end = l.getByteOffsetAtRuneIdx(runeIdx + 1)
	}
	r, _ := utf8.DecodeRuneInString(l.lineNoAnsi[start:end])
	return r
}

func (l LineBuffer) getByteOffsetAtRuneIdx(runeIdx int) uint32 {
	if runeIdx < 0 {
		panic("runeIdx must be greater or equal to 0")
	}
	if runeIdx == 0 || len(l.line) == 0 || l.sparsity == 0 {
		return 0
	}
	if runeIdx >= l.numNoAnsiRunes {
		panic("rune index greater than num runes")
	}

	// get the last stored byte offset before this index
	sparseIdx := runeIdx / l.sparsity
	baseRuneIdx := sparseIdx * l.sparsity

	if baseRuneIdx == runeIdx {
		return l.sparseRuneIdxToNoAnsiByteOffset[sparseIdx]
	}

	currRuneIdx := baseRuneIdx
	byteOffset := l.sparseRuneIdxToNoAnsiByteOffset[sparseIdx]
	for ; currRuneIdx != runeIdx; currRuneIdx++ {
		_, nBytes := utf8.DecodeRuneInString(l.lineNoAnsi[byteOffset:])
		byteOffset += uint32(nBytes)
	}
	return byteOffset
}

// getRuneWidth extracts the width of a rune from the packed array
func (l LineBuffer) getRuneWidth(runeIdx int) uint8 {
	if runeIdx < 0 || runeIdx >= l.numNoAnsiRunes {
		return 0
	}

	packedIdx := runeIdx / 4
	bitPos := (runeIdx % 4) * 2
	return (l.lineNoAnsiRuneWidths[packedIdx] >> bitPos) & 3
}

func (l LineBuffer) getCumulativeWidthAtRuneIdx(runeIdx int) uint32 {
	if runeIdx < 0 {
		return 0
	}
	if runeIdx >= l.numNoAnsiRunes {
		panic("runeIdx greater than num runes")
	}

	// get the last stored cumulative width before this index
	sparseIdx := runeIdx / l.sparsity
	baseRuneIdx := sparseIdx * l.sparsity

	if baseRuneIdx == runeIdx {
		return l.sparseLineNoAnsiCumRuneWidths[sparseIdx]
	}

	// sum the widths from the last stored point to our target index
	var additionalWidth uint32
	for i := baseRuneIdx + 1; i <= runeIdx; i++ {
		additionalWidth += uint32(l.getRuneWidth(i))
	}

	return l.sparseLineNoAnsiCumRuneWidths[sparseIdx] + additionalWidth
}

// findRuneIndexWithWidthToLeft returns the index of the rune that has the input width to the left of it
func (l LineBuffer) findRuneIndexWithWidthToLeft(widthToLeft int) int {
	if widthToLeft < 0 {
		panic("widthToLeft less than 0")
	}
	if widthToLeft == 0 || l.numNoAnsiRunes == 0 {
		return 0
	}
	if widthToLeft > l.Width() {
		panic("widthToLeft greater than total width")
	}

	left, right := 0, l.numNoAnsiRunes-1
	if l.getCumulativeWidthAtRuneIdx(right) < uint32(widthToLeft) {
		return l.numNoAnsiRunes
	}

	for left < right {
		mid := left + (right-left)/2
		if l.getCumulativeWidthAtRuneIdx(mid) >= uint32(widthToLeft) {
			right = mid
		} else {
			left = mid + 1
		}
	}

	// skip over zero-width runes
	w := l.getCumulativeWidthAtRuneIdx(left)
	nextLeft := left + 1
	for nextLeft < l.numNoAnsiRunes && l.getCumulativeWidthAtRuneIdx(nextLeft) == w {
		left = nextLeft
		nextLeft += 1
	}

	return left + 1
}
