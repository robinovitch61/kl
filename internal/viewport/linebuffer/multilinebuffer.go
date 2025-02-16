package linebuffer

import (
	"github.com/charmbracelet/lipgloss/v2"
	"github.com/robinovitch61/kl/internal/filter"
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

	for i := range m.buffers {
		bufWidth := m.buffers[i].Width()
		if skippedWidth+bufWidth > startWidth {
			firstBufferIdx = i
			startWidthFirstBuffer = startWidth - skippedWidth
			break
		}
		skippedWidth += bufWidth
		startWidthFirstBuffer -= bufWidth
	}

	// get content before our start position for highlight context
	nBytesLeftContext := len(toHighlight) * 2
	leftContext := getBytesLeftOfWidth(nBytesLeftContext, m.buffers, firstBufferIdx, startWidthFirstBuffer)

	// take from first buffer
	res, takenWidth := m.buffers[firstBufferIdx].Take(startWidthFirstBuffer, takeWidth, "", "", lipgloss.NewStyle())
	remainingTotalWidth := takeWidth - takenWidth
	remainingBufferWidth := m.buffers[firstBufferIdx].Width() - takenWidth

	// if we have more width to take and more buffers available, continue
	currentBufferIdx := firstBufferIdx + 1
	for remainingTotalWidth > 0 && currentBufferIdx < len(m.buffers) {
		nextPart, partWidth := m.buffers[currentBufferIdx].Take(0, remainingTotalWidth, "", "", lipgloss.NewStyle())
		remainingBufferWidth = m.buffers[currentBufferIdx].Width() - partWidth
		if partWidth == 0 {
			break
		}
		res += nextPart
		remainingTotalWidth -= partWidth
		currentBufferIdx++
	}

	// get content after our result for highlight context
	currentBufferIdx -= 1
	nBytesRightContext := len(toHighlight) * 2
	rightContext := getBytesRightOfWidth(nBytesRightContext, m.buffers, currentBufferIdx, remainingBufferWidth)

	// apply continuation indicators if needed
	if len(continuation) > 0 {
		contentToLeft := startWidth > 0
		contentToRight := m.totalWidth-startWidth > takeWidth-remainingTotalWidth
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
	resNoAnsi := stripAnsi(res)
	lineNoAnsi := leftContext + resNoAnsi + rightContext
	res = highlightString(
		res,
		toHighlight,
		highlightStyle,
		lineNoAnsi,
		len(leftContext),
		len(leftContext)+len(resNoAnsi),
	)

	res = removeEmptyAnsiSequences(res)
	return res, takeWidth - remainingTotalWidth
}

func (m MultiLineBuffer) WrappedLines(
	width int,
	maxLinesEachEnd int,
	toHighlight string,
	toHighlightStyle lipgloss.Style,
) []string {
	if width <= 0 {
		return []string{}
	}
	if len(m.buffers) == 0 {
		return []string{}
	}
	if len(m.buffers) == 1 {
		return m.buffers[0].WrappedLines(width, maxLinesEachEnd, toHighlight, toHighlightStyle)
	}

	totalLines := (m.totalWidth + width - 1) / width
	if totalLines == 0 {
		return []string{""}
	}

	return getWrappedLines(
		m,
		totalLines,
		width,
		maxLinesEachEnd,
		toHighlight,
		toHighlightStyle,
	)
}

func (m MultiLineBuffer) Matches(f filter.Model) bool {
	// TODO LEO: inefficient
	var builder strings.Builder
	for i := range m.buffers {
		builder.WriteString(m.buffers[i].lineNoAnsi)
	}
	lineNoAnsi := builder.String()
	return f.Matches(lineNoAnsi)
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
