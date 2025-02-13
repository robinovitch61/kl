package linebuffer

import (
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/robinovitch61/kl/internal/constants"
	"strings"
)

// MultiLineBuffer implements LineBufferer by wrapping multiple LineBuffers without extra memory allocation
type MultiLineBuffer struct {
	buffers    []LineBuffer
	totalWidth int // cached total width across all buffers
}

// type assertion that MultiLineBuffer implements LineBufferer
var _ LineBufferer = MultiLineBuffer{}

// type assertion that *MultiLineBuffer implements LineBufferer
var _ LineBufferer = (*MultiLineBuffer)(nil)

func NewMulti(buffers ...LineBuffer) MultiLineBuffer {
	if len(buffers) == 0 {
		return MultiLineBuffer{}
	}

	totalWidth := 0
	for _, buf := range buffers {
		totalWidth += buf.Width()
	}

	return MultiLineBuffer{
		buffers:    buffers,
		totalWidth: totalWidth,
	}
}

func (m MultiLineBuffer) Width() int {
	return m.totalWidth
}

// TODO LEO: don't use this for e.g. search, instead inject filter into a Matches() bool method here
// TODO LEO: or store the concatenated string on init?
func (m MultiLineBuffer) Content() string {
	if len(m.buffers) == 0 {
		return ""
	}

	if len(m.buffers) == 1 {
		return m.buffers[0].Content()
	}

	totalLen := 0
	for _, buf := range m.buffers {
		totalLen += len(buf.Content())
	}

	var builder strings.Builder
	builder.Grow(totalLen)

	for _, buf := range m.buffers {
		builder.WriteString(buf.Content())
	}

	return builder.String()
}

func (m MultiLineBuffer) Take(
	startWidth, takeWidth int,
	continuation, toHighlight string,
	highlightStyle lipgloss.Style,
) (string, int) {
	if len(m.buffers) == 0 {
		return "", 0
	}

	if len(m.buffers) == 1 {
		return m.buffers[0].Take(startWidth, takeWidth, continuation, toHighlight, highlightStyle)
	}

	if startWidth >= m.totalWidth {
		return "", 0
	}

	// find which buffer contains our start position
	skippedWidth := 0
	firstBufferIdx := 0
	startWidthFirstBuffer := startWidth

	for i, buf := range m.buffers {
		bufWidth := buf.Width()
		if skippedWidth+bufWidth > startWidth {
			firstBufferIdx = i
			startWidthFirstBuffer = startWidth - skippedWidth
			break
		}
		skippedWidth += bufWidth
		startWidthFirstBuffer -= bufWidth
	}

	// TODO: adjust this based on content taken from the left of res
	// we should take enough bytes from the left of res to ensure that an overlapping highlight can be found, 2x the length of toHighlight
	// ansi sequences should be stripped from the content taken from the left and right of res when formulating lineNoAnsi
	startByteOffset := 0

	// take from first buffer
	res, takenWidth := m.buffers[firstBufferIdx].Take(startWidthFirstBuffer, takeWidth, "", "", lipgloss.NewStyle())
	remainingWidth := takeWidth - takenWidth

	// if we have more width to take and more buffers available, continue
	currentBufferIdx := firstBufferIdx + 1
	for remainingWidth > 0 && currentBufferIdx < len(m.buffers) {
		nextPart, partWidth := m.buffers[currentBufferIdx].Take(0, remainingWidth, "", "", lipgloss.NewStyle())
		if partWidth == 0 {
			break
		}
		res += nextPart
		remainingWidth -= partWidth
		currentBufferIdx++
	}

	// TODO: adjust this based on content taken from the right of res
	// we should take enough bytes from the right of res to ensure that an overlapping highlight can be found, 2x the length of toHighlight
	// ansi sequences should be stripped from the content taken from the left and right of res when formulating lineNoAnsi
	endByteOffset := startByteOffset + len(res)

	// apply continuation indicators if needed
	if len(continuation) > 0 {
		contentToLeft := startWidth > 0
		contentToRight := m.totalWidth-startWidth > takeWidth-remainingWidth

		if contentToLeft || contentToRight {
			continuationRunes := []rune(continuation)
			if contentToLeft {
				res = replaceStartWithContinuation(res, continuationRunes)
			}
			if contentToRight {
				res = replaceEndWithContinuation(res, continuationRunes)
			}
		}
	}

	// highlight the desired string
	res = highlightString(
		res,
		toHighlight,
		highlightStyle,
		lineNoAnsi, // TODO: this should include content taken from the left and right of res, all stripped of ansi codes
		startByteOffset,
		endByteOffset,
	)

	// remove empty sequences
	res = constants.EmptySequenceRegex.ReplaceAllString(res, "")

	return res, takeWidth - remainingWidth
}

func (m MultiLineBuffer) WrappedLines(
	width int,
	maxLinesEachEnd int,
	toHighlight string,
	toHighlightStyle lipgloss.Style,
) []string {
	// handle edge cases
	if width <= 0 {
		return []string{}
	}
	if maxLinesEachEnd <= 0 {
		maxLinesEachEnd = -1
	}
	if len(m.buffers) == 0 {
		return []string{}
	}
	if len(m.buffers) == 1 {
		return m.buffers[0].WrappedLines(width, maxLinesEachEnd, toHighlight, toHighlightStyle)
	}

	// calculate total number of lines
	totalLines := (m.totalWidth + width - 1) / width
	if totalLines == 0 {
		return []string{""}
	}

	var result []string
	startWidth := 0

	if maxLinesEachEnd > 0 && totalLines > maxLinesEachEnd*2 {
		// take maxLinesEachEnd from start
		for nLines := 0; nLines < maxLinesEachEnd; nLines++ {
			line, lineWidth := m.Take(startWidth, width, "", toHighlight, toHighlightStyle)
			result = append(result, line)
			startWidth += lineWidth
		}

		// take maxLinesEachEnd from end
		startWidth = (totalLines - maxLinesEachEnd) * width
		for nLines := 0; nLines < maxLinesEachEnd; nLines++ {
			line, lineWidth := m.Take(startWidth, width, "", toHighlight, toHighlightStyle)
			result = append(result, line)
			startWidth += lineWidth
		}
	} else {
		// take all lines
		for nLines := 0; nLines < totalLines; nLines++ {
			line, lineWidth := m.Take(startWidth, width, "", toHighlight, toHighlightStyle)
			result = append(result, line)
			startWidth += lineWidth
		}
	}

	return result
}

func (m MultiLineBuffer) Repr() string {
	v := "Multi("
	for i := range m.buffers {
		if i > 0 {
			v += ", "
		}
		v += m.buffers[i].Repr()
	}
	v += ")"
	return v
}
