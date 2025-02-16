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
	line                string     // underlying string with ansi codes. utf-8 bytes
	lineNoAnsi          string     // line without ansi codes. utf-8 bytes
	runeIdxToByteOffset []uint32   // idx of lineNoAnsiRunes to byte offset
	lineNoAnsiWidths    []uint8    // terminal cell widths
	lineNoAnsiCumWidths []uint32   // cumulative lineNoAnsiWidths
	ansiCodeIndexes     [][]uint32 // slice of startByte, endByte indexes of ansi codes
}

// type assertion that LineBuffer implements LineBufferer
var _ LineBufferer = LineBuffer{}

// type assertion that *LineBuffer implements LineBufferer
var _ LineBufferer = (*LineBuffer)(nil)

func New(line string) LineBuffer {
	lb := LineBuffer{
		line: line,
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

	n := utf8.RuneCountInString(lb.lineNoAnsi)

	// single allocation for all integer slices
	combined := make([]uint32, n+1+n)
	lb.runeIdxToByteOffset = combined[:n+1]
	lb.lineNoAnsiCumWidths = combined[n+1:]

	lb.lineNoAnsiWidths = make([]uint8, n)

	var currentOffset uint32
	var cumWidth uint32
	i := 0
	for pos := 0; pos < len(lb.lineNoAnsi); {
		lb.runeIdxToByteOffset[i] = currentOffset
		r, size := utf8.DecodeRuneInString(lb.lineNoAnsi[pos:])
		currentOffset += uint32(size)
		width := uint8(runewidth.RuneWidth(r))
		lb.lineNoAnsiWidths[i] = width
		cumWidth += uint32(width)
		lb.lineNoAnsiCumWidths[i] = cumWidth
		i++
		pos += size
	}
	lb.runeIdxToByteOffset[n] = currentOffset
	return lb
}

func (l LineBuffer) Width() int {
	if len(l.lineNoAnsiCumWidths) > 0 {
		return int(l.lineNoAnsiCumWidths[len(l.lineNoAnsiCumWidths)-1])
	}
	return 0
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

	startRuneIdx := getLeftRuneIdx(startWidth, l.lineNoAnsiCumWidths)

	if startRuneIdx >= len(l.lineNoAnsiWidths) || takeWidth == 0 {
		return "", 0
	}

	var result strings.Builder
	remainingWidth := takeWidth
	leftRuneIdx := startRuneIdx
	startByteOffset := l.runeIdxToByteOffset[startRuneIdx]

	runesWritten := 0
	for ; remainingWidth > 0 && leftRuneIdx < len(l.lineNoAnsiWidths); leftRuneIdx++ {
		r := l.runeAt(leftRuneIdx)
		runeWidth := l.lineNoAnsiWidths[leftRuneIdx]
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
		for ; leftRuneIdx < len(l.lineNoAnsiWidths); leftRuneIdx++ {
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
	if len(continuation) > 0 && (startRuneIdx > 0 || leftRuneIdx < len(l.lineNoAnsiWidths)) {
		continuationRunes := []rune(continuation)

		// if more runes to the left of the result, replace start runes with continuation indicator
		if startRuneIdx > 0 {
			res = replaceStartWithContinuation(res, continuationRunes)
		}

		// if more runes to the right, replace final runes in result with continuation indicator
		if leftRuneIdx < len(l.lineNoAnsiWidths) {
			res = replaceEndWithContinuation(res, continuationRunes)
		}
	}

	// highlight the desired string
	res = highlightString(
		res,
		toHighlight,
		highlightStyle,
		l.lineNoAnsi,
		int(startByteOffset),
		int(l.runeIdxToByteOffset[leftRuneIdx]),
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
	// preserve empty lines
	if l.line == "" {
		return []string{l.line}
	}

	return getWrappedLines(
		l,
		getTotalLines(l.lineNoAnsiCumWidths, uint32(width)),
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

func (l LineBuffer) runeAt(idx int) rune {
	if idx < 0 || idx >= len(l.lineNoAnsiWidths) {
		return -1
	}
	start := l.runeIdxToByteOffset[idx]
	end := l.runeIdxToByteOffset[idx+1]
	r, _ := utf8.DecodeRuneInString(l.lineNoAnsi[start:end])
	return r
}
