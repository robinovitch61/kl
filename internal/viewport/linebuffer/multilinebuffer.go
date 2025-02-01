package linebuffer

import "github.com/charmbracelet/lipgloss/v2"

// MultiLineBuffer implements LineBufferer by wrapping multiple LineBuffers without extra memory allocation
type MultiLineBuffer struct {
	buffers          []LineBuffer
	currentBufferIdx int // tracks which buffer we're currently reading from
	totalWidth       int // cached total width across all buffers
}

// type assertion that *MultiLineBuffer implements LineBufferer
var _ LineBufferer = (*MultiLineBuffer)(nil)

func NewMulti(buffers ...LineBuffer) *MultiLineBuffer {
	if len(buffers) == 0 {
		return &MultiLineBuffer{}
	}

	// calculate total width
	totalWidth := 0
	for _, buf := range buffers {
		totalWidth += buf.Width()
	}

	return &MultiLineBuffer{
		buffers:    buffers,
		totalWidth: totalWidth,
	}
}

func (m *MultiLineBuffer) Width() int {
	return m.totalWidth
}

// TODO LEO: don't use this for e.g. search, instead inject filter into a Matches() bool method here
func (m *MultiLineBuffer) Content() string {
	totalLen := 0
	for _, buf := range m.buffers {
		totalLen += len(buf.Content())
	}

	result := make([]byte, 0, totalLen)
	for _, buf := range m.buffers {
		result = append(result, buf.Content()...)
	}
	return string(result)
}

func (m *MultiLineBuffer) SeekToWidth(width int) {
	if width <= 0 {
		m.currentBufferIdx = 0
		if len(m.buffers) > 0 {
			m.buffers[0].SeekToWidth(0)
		}
		return
	}

	// find which buffer contains the target width
	remainingWidth := width
	for i, buf := range m.buffers {
		bufWidth := buf.Width()
		if remainingWidth <= bufWidth {
			// found the buffer containing our target
			m.currentBufferIdx = i
			m.buffers[i].SeekToWidth(remainingWidth)
			return
		}
		remainingWidth -= bufWidth
	}

	// if we get here, we're seeking past the end
	if len(m.buffers) > 0 {
		m.currentBufferIdx = len(m.buffers) - 1
		m.buffers[m.currentBufferIdx].SeekToWidth(m.buffers[m.currentBufferIdx].Width())
	}
}

func (m *MultiLineBuffer) PopLeft(width int, continuation, toHighlight string, highlightStyle lipgloss.Style) string {
	if len(m.buffers) == 0 || width == 0 {
		return ""
	}

	// get content from current buffer
	result := m.buffers[m.currentBufferIdx].PopLeft(width, continuation, toHighlight, highlightStyle)

	// if we got less than requested width and have more buffers, move to next buffer
	if resultWidth := lipgloss.Width(result); resultWidth < width && m.currentBufferIdx < len(m.buffers)-1 {
		m.currentBufferIdx++
		m.buffers[m.currentBufferIdx].SeekToWidth(0)
		// get remaining content from next buffer
		remainingWidth := width - resultWidth
		nextResult := m.buffers[m.currentBufferIdx].PopLeft(remainingWidth, continuation, toHighlight, highlightStyle)
		result += nextResult
	}

	return result
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

	if maxLinesEachEnd <= 0 {
		maxLinesEachEnd = -1
	}

	// preserve empty lines
	if m.Content() == "" {
		return []string{""}
	}

	var res []string
	totalLines := m.totalLines(width)

	m.SeekToWidth(0)
	if maxLinesEachEnd > 0 && totalLines > maxLinesEachEnd*2 {
		for nLines := 0; nLines < maxLinesEachEnd; nLines++ {
			res = append(res, m.PopLeft(width, "", toHighlight, toHighlightStyle))
		}

		m.seekToLine(totalLines-maxLinesEachEnd, width)
		for nLines := 0; nLines < maxLinesEachEnd; nLines++ {
			res = append(res, m.PopLeft(width, "", toHighlight, toHighlightStyle))
		}
	} else {
		for nLines := 0; nLines < totalLines; nLines++ {
			res = append(res, m.PopLeft(width, "", toHighlight, toHighlightStyle))
		}
	}

	m.SeekToWidth(0)
	return res
}

func (m MultiLineBuffer) totalLines(width int) int {
	if width == 0 {
		return 0
	}

	total := 0
	for _, buf := range m.buffers {
		bufferWidth := int(buf.fullWidth())
		if bufferWidth > 0 {
			total += (bufferWidth + width - 1) / width
		}
	}
	return total
}

func (m *MultiLineBuffer) seekToLine(line int, width int) {
	if line <= 0 {
		m.SeekToWidth(0)
		return
	}

	// find which buffer contains our target line
	remainingLines := line
	for i, buf := range m.buffers {
		bufferLines := (int(buf.fullWidth()) + width - 1) / width
		if remainingLines < bufferLines {
			m.currentBufferIdx = i
			m.buffers[i].seekToLine(remainingLines, width)
			return
		}
		remainingLines -= bufferLines
	}

	if len(m.buffers) > 0 {
		m.currentBufferIdx = len(m.buffers) - 1
		lastBuf := &m.buffers[m.currentBufferIdx]
		lastBuf.seekToLine((int(lastBuf.fullWidth())+width-1)/width, width)
	}
}
