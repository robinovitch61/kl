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
	line       string // underlying string with ansi codes
	lineNoAnsi string // line without ansi codes
	// TODO LEO: make runeIdxToNoAnsiByteOffset sparse
	runeIdxToNoAnsiByteOffset []uint32 // idx of lineNoAnsi runes to byte offset
	// TODO LEO: consider making lineNoAnsiRuneWidths sparse
	lineNoAnsiRuneWidths    []uint8    // terminal cell widths
	sparsity                int        // interval for which to store cumulative cell width
	lineNoAnsiCumRuneWidths []uint32   // cumulative terminal cell width, stored every sparsity runes
	ansiCodeIndexes         [][]uint32 // slice of startByte, endByte indexes of ansi codes
}

// type assertion that LineBuffer implements LineBufferer
var _ LineBufferer = LineBuffer{}

// type assertion that *LineBuffer implements LineBufferer
var _ LineBufferer = (*LineBuffer)(nil)

func New(line string) LineBuffer {
	if len(line) <= 0 {
		return LineBuffer{line: line}
	}

	lb := LineBuffer{
		line:     line,
		sparsity: 100, // tradeoff between memory usage and CPU. 100 seems to be a good balance
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

	// single memory allocation for common types
	combined := make([]uint32, numRunes+sparseLen)
	lb.runeIdxToNoAnsiByteOffset = combined[:numRunes]
	lb.lineNoAnsiCumRuneWidths = combined[numRunes:]

	lb.lineNoAnsiRuneWidths = make([]uint8, numRunes)

	var currentOffset uint32
	var cumWidth uint32
	idx := 0
	for byteOffset := 0; byteOffset < len(lb.lineNoAnsi); {
		lb.runeIdxToNoAnsiByteOffset[idx] = currentOffset
		r, runeNumBytes := utf8.DecodeRuneInString(lb.lineNoAnsi[byteOffset:])
		currentOffset += uint32(runeNumBytes)
		width := uint8(runewidth.RuneWidth(r))
		lb.lineNoAnsiRuneWidths[idx] = width
		cumWidth += uint32(width)
		if idx%lb.sparsity == 0 {
			lb.lineNoAnsiCumRuneWidths[idx/lb.sparsity] = cumWidth
		}
		idx++
		byteOffset += runeNumBytes
	}

	return lb
}

func (l LineBuffer) Width() int {
	if len(l.line) == 0 {
		return 0
	}
	lastRuneIdx := len(l.lineNoAnsiRuneWidths) - 1
	return int(l.getCumulativeWidthAtRuneIdx(lastRuneIdx))
}

func (l LineBuffer) Content() string {
	return l.line
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

	startRuneIdx := getLeftRuneIdx(startWidth, l)

	if startRuneIdx >= len(l.lineNoAnsiRuneWidths) || takeWidth == 0 {
		return "", 0
	}

	var result strings.Builder
	remainingWidth := takeWidth
	leftRuneIdx := startRuneIdx
	startByteOffset := l.runeIdxToNoAnsiByteOffset[startRuneIdx]

	runesWritten := 0
	for ; remainingWidth > 0 && leftRuneIdx < len(l.lineNoAnsiRuneWidths); leftRuneIdx++ {
		r := l.runeAt(leftRuneIdx)
		runeWidth := l.lineNoAnsiRuneWidths[leftRuneIdx]
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
		for ; leftRuneIdx < len(l.lineNoAnsiRuneWidths); leftRuneIdx++ {
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
	if len(continuation) > 0 && (startRuneIdx > 0 || leftRuneIdx < len(l.lineNoAnsiRuneWidths)) {
		continuationRunes := []rune(continuation)

		// if more runes to the left of the result, replace start runes with continuation indicator
		if startRuneIdx > 0 {
			res = replaceStartWithContinuation(res, continuationRunes)
		}

		// if more runes to the right, replace final runes in result with continuation indicator
		if leftRuneIdx < len(l.lineNoAnsiRuneWidths) {
			res = replaceEndWithContinuation(res, continuationRunes)
		}
	}

	// highlight the desired string
	var endByteOffset int
	if leftRuneIdx < len(l.runeIdxToNoAnsiByteOffset) {
		endByteOffset = int(l.runeIdxToNoAnsiByteOffset[leftRuneIdx])
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

	lastRuneIdx := len(l.lineNoAnsiRuneWidths) - 1
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
	if runeIdx < 0 || runeIdx >= len(l.runeIdxToNoAnsiByteOffset) {
		return -1
	}
	start := l.runeIdxToNoAnsiByteOffset[runeIdx]
	var end uint32
	if runeIdx+1 >= len(l.runeIdxToNoAnsiByteOffset) {
		end = uint32(len(l.lineNoAnsi))
	} else {
		end = l.runeIdxToNoAnsiByteOffset[runeIdx+1]
	}
	r, _ := utf8.DecodeRuneInString(l.lineNoAnsi[start:end])
	return r
}

func (l LineBuffer) getCumulativeWidthAtRuneIdx(runeIdx int) uint32 {
	if runeIdx < 0 {
		return 0
	}
	if runeIdx >= len(l.lineNoAnsiRuneWidths) {
		runeIdx = len(l.lineNoAnsiRuneWidths) - 1
	}

	// get the last stored cumulative width before this index
	sparseIdx := runeIdx / l.sparsity
	baseRuneIdx := sparseIdx * l.sparsity

	if baseRuneIdx == runeIdx {
		return l.lineNoAnsiCumRuneWidths[sparseIdx]
	}

	// sum the widths from the last stored point to our target index
	var additionalWidth uint32
	for i := baseRuneIdx + 1; i <= runeIdx; i++ {
		additionalWidth += uint32(l.lineNoAnsiRuneWidths[i])
	}

	return l.lineNoAnsiCumRuneWidths[sparseIdx] + additionalWidth
}
